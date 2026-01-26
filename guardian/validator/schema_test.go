package validator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestNewSchemaValidator tests validator creation.
func TestNewSchemaValidator(t *testing.T) {
	t.Parallel()

	v := validator.NewSchemaValidator()
	require.NotNil(t, v)
	assert.Equal(t, "schema", v.Name())
}

// TestSchemaValidator_Validate_NonActionFile tests skipping non-action files.
func TestSchemaValidator_Validate_NonActionFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{"workflow yml", "workflow.yml"},
		{"workflow yaml", "workflow.yaml"},
		{"ci yml", "ci.yml"},
		{"random yaml", "config.yaml"},
		{"not yaml", "readme.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte("name: Test"), 0o600)
			require.NoError(t, err)

			v := validator.NewSchemaValidator()
			findings, err := v.Validate(context.Background(), path)
			require.NoError(t, err)
			assert.Empty(t, findings, "non-action files should return no findings")
		})
	}
}

// TestSchemaValidator_Validate_ActionYML tests action.yml validation.
func TestSchemaValidator_Validate_ActionYML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		filename      string
		content       string
		expectedCount int
		expectedRules []string
	}{
		{
			name:     "valid composite action",
			filename: "action.yml",
			content: `name: My Action
description: A test action
runs:
  using: composite
  steps:
    - run: echo "Hello"
      shell: bash`,
			expectedCount: 0,
		},
		{
			name:     "valid node20 action",
			filename: "action.yml",
			content: `name: My Action
description: A test action
runs:
  using: node20
  main: dist/index.js`,
			expectedCount: 0,
		},
		{
			name:     "valid docker action",
			filename: "action.yml",
			content: `name: My Action
description: A test action
runs:
  using: docker
  image: Dockerfile`,
			expectedCount: 0,
		},
		{
			name:     "missing name",
			filename: "action.yml",
			content: `description: A test
runs:
  using: composite
  steps:
    - run: echo
      shell: bash`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-name"},
		},
		{
			name:     "missing description",
			filename: "action.yml",
			content: `name: Test
runs:
  using: composite
  steps:
    - run: echo
      shell: bash`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-description"},
		},
		{
			name:     "missing both name and description",
			filename: "action.yml",
			content: `runs:
  using: composite
  steps:
    - run: echo
      shell: bash`,
			expectedCount: 2,
			expectedRules: []string{"schema/missing-name", "schema/missing-description"},
		},
		{
			name:     "missing runs.using",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  main: index.js`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-using"},
		},
		{
			name:     "invalid runs.using value",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: node14`,
			expectedCount: 1,
			expectedRules: []string{"schema/invalid-using"},
		},
		{
			name:     "deprecated node12",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: node12
  main: index.js`,
			expectedCount: 1,
			expectedRules: []string{"schema/deprecated-node12"},
		},
		{
			name:     "deprecated node16",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: node16
  main: index.js`,
			expectedCount: 1,
			expectedRules: []string{"schema/deprecated-node16"},
		},
		{
			name:     "node action missing main",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: node20`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-main"},
		},
		{
			name:     "docker action missing image",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: docker`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-image"},
		},
		{
			name:     "composite action missing steps",
			filename: "action.yml",
			content: `name: Test
description: Test action
runs:
  using: composite`,
			expectedCount: 1,
			expectedRules: []string{"schema/missing-steps"},
		},
		{
			name:     "input without description",
			filename: "action.yml",
			content: `name: Test
description: Test action
inputs:
  my-input:
    required: true
runs:
  using: composite
  steps:
    - run: echo`,
			expectedCount: 1,
			expectedRules: []string{"schema/input-missing-description"},
		},
		{
			name:     "output without description",
			filename: "action.yml",
			content: `name: Test
description: Test action
outputs:
  my-output:
    value: ${{ steps.step1.outputs.result }}
runs:
  using: composite
  steps:
    - run: echo`,
			expectedCount: 1,
			expectedRules: []string{"schema/output-missing-description"},
		},
		{
			name:     "action.yaml extension also works",
			filename: "action.yaml",
			content: `name: Test
description: Test action
runs:
  using: composite
  steps:
    - run: echo`,
			expectedCount: 0,
		},
		{
			name:     "invalid YAML syntax",
			filename: "action.yml",
			content: `name: Test
  invalid: yaml: syntax`,
			expectedCount: 1,
			expectedRules: []string{"schema/parse-error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte(tt.content), 0o600)
			require.NoError(t, err)

			v := validator.NewSchemaValidator()
			findings, err := v.Validate(context.Background(), path)
			require.NoError(t, err)
			assert.Len(t, findings, tt.expectedCount)

			if len(tt.expectedRules) > 0 {
				foundRules := make([]string, len(findings))
				for i, f := range findings {
					foundRules[i] = f.RuleID
				}

				for _, expected := range tt.expectedRules {
					assert.Contains(t, foundRules, expected, "expected rule %s in findings", expected)
				}
			}
		})
	}
}

// TestSchemaValidator_Validate_FileReadError tests error handling.
func TestSchemaValidator_Validate_FileReadError(t *testing.T) {
	t.Parallel()

	v := validator.NewSchemaValidator()

	// Non-existent action file should return error
	_, err := v.Validate(context.Background(), "/nonexistent/action.yml")
	require.Error(t, err)
}

// TestSchemaValidator_Validate_FullAction tests a complete action.yml.
func TestSchemaValidator_Validate_FullAction(t *testing.T) {
	t.Parallel()

	content := `name: Complete Action
description: A fully featured action for testing
author: Test Author

inputs:
  name:
    description: The name to greet
    required: true
  greeting:
    description: The greeting message
    required: false
    default: Hello

outputs:
  result:
    description: The greeting result
    value: ${{ steps.greet.outputs.greeting }}

runs:
  using: composite
  steps:
    - id: greet
      run: echo "::set-output name=greeting::${{ inputs.greeting }}, ${{ inputs.name }}!"
      shell: bash

branding:
  icon: activity
  color: blue`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "action.yml")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSchemaValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)
	assert.Empty(t, findings, "fully valid action should have no findings")
}

// TestSchemaValidator_Validate_MultipleInputsOutputs tests multiple inputs/outputs validation.
func TestSchemaValidator_Validate_MultipleInputsOutputs(t *testing.T) {
	t.Parallel()

	content := `name: Test
description: Test action
inputs:
  input1:
    required: true
  input2:
    description: Has description
  input3:
    required: false
outputs:
  output1:
    value: test
  output2:
    description: Has description
  output3:
    value: test
runs:
  using: composite
  steps:
    - run: echo`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "action.yml")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSchemaValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)

	// Should have warnings for inputs/outputs without descriptions
	// input1, input3 missing description (2)
	// output1, output3 missing description (2)
	assert.Len(t, findings, 4)

	// Verify they're all warnings, not errors
	for _, f := range findings {
		assert.Equal(t, validator.SeverityWarning, f.Severity)
	}
}

// TestSchemaValidator_Validate_DockerWithEntrypoint tests docker action variations.
func TestSchemaValidator_Validate_DockerWithEntrypoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{
			name: "docker with image",
			content: `name: Test
description: Test
runs:
  using: docker
  image: Dockerfile`,
			expectedCount: 0,
		},
		{
			name: "docker with entrypoint",
			content: `name: Test
description: Test
runs:
  using: docker
  image: alpine
  entrypoint: /entrypoint.sh`,
			expectedCount: 0,
		},
		{
			name: "docker with pre and post",
			content: `name: Test
description: Test
runs:
  using: docker
  image: alpine
  pre-entrypoint: setup.sh
  entrypoint: main.sh
  post-entrypoint: cleanup.sh`,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "action.yml")
			err := os.WriteFile(path, []byte(tt.content), 0o600)
			require.NoError(t, err)

			v := validator.NewSchemaValidator()
			findings, err := v.Validate(context.Background(), path)
			require.NoError(t, err)
			assert.Len(t, findings, tt.expectedCount)
		})
	}
}

// TestSchemaValidator_Validate_NodeWithPrePost tests node action pre/post scripts.
func TestSchemaValidator_Validate_NodeWithPrePost(t *testing.T) {
	t.Parallel()

	content := `name: Test
description: Test
runs:
  using: node20
  main: dist/index.js
  pre: dist/setup.js
  pre-if: runner.os == 'Linux'
  post: dist/cleanup.js
  post-if: always()`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "action.yml")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSchemaValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

// TestSchemaValidator_FindingSource tests that findings have correct source.
func TestSchemaValidator_FindingSource(t *testing.T) {
	t.Parallel()

	content := `runs:
  using: composite
  steps:
    - run: echo`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "action.yml")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSchemaValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	for _, f := range findings {
		assert.Equal(t, validator.SourceSchema, f.Source)
	}
}

// TestSchemaValidator_FindingSuggestions tests that findings include suggestions.
func TestSchemaValidator_FindingSuggestions(t *testing.T) {
	t.Parallel()

	content := `runs:
  using: node20`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "action.yml")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSchemaValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	// Check that findings have suggestions
	hasSuggestion := false
	for _, f := range findings {
		if f.Suggestion != "" {
			hasSuggestion = true
			break
		}
	}
	assert.True(t, hasSuggestion, "expected at least one finding with a suggestion")
}
