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

// TestDeprecationValidator_Name tests the validator name.
func TestDeprecationValidator_Name(t *testing.T) {
	t.Parallel()

	v := validator.NewDeprecationValidator()
	assert.Equal(t, "deprecation", v.Name())
}

// TestDeprecationValidator_Validate tests deprecation detection.
func TestDeprecationValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yaml           string
		expectFindings int
		expectRuleID   string
	}{
		{
			name: "current action versions pass",
			yaml: `
name: Current
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`,
			expectFindings: 0,
		},
		{
			name: "deprecated checkout v3",
			yaml: `
name: Old Checkout
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated checkout v2",
			yaml: `
name: Old Checkout
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated checkout v1",
			yaml: `
name: Old Checkout
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated setup-go v4",
			yaml: `
name: Old Setup Go
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated setup-node v3",
			yaml: `
name: Old Setup Node
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v3
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated setup-python v4",
			yaml: `
name: Old Setup Python
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-python@v4
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated cache v3",
			yaml: `
name: Old Cache
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/cache@v3
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated upload-artifact v3",
			yaml: `
name: Old Upload
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/upload-artifact@v3
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated download-artifact v3",
			yaml: `
name: Old Download
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v3
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/action",
		},
		{
			name: "deprecated runner ubuntu-18.04",
			yaml: `
name: Old Ubuntu
jobs:
  build:
    runs-on: ubuntu-18.04
    steps:
      - run: echo hello
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/runner",
		},
		{
			name: "deprecated runner macos-11",
			yaml: `
name: Old macOS
jobs:
  build:
    runs-on: macos-11
    steps:
      - run: echo hello
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/runner",
		},
		{
			name: "deprecated runner windows-2019",
			yaml: `
name: Old Windows
jobs:
  build:
    runs-on: windows-2019
    steps:
      - run: echo hello
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/runner",
		},
		{
			name: "current runners pass",
			yaml: `
name: Current Runners
jobs:
  ubuntu:
    runs-on: ubuntu-latest
    steps:
      - run: echo ubuntu
  macos:
    runs-on: macos-14
    steps:
      - run: echo macos
  windows:
    runs-on: windows-2022
    steps:
      - run: echo windows
`,
			expectFindings: 0,
		},
		{
			name: "multiple deprecated items",
			yaml: `
name: Multiple Deprecated
jobs:
  build:
    runs-on: ubuntu-18.04
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v3
`,
			expectFindings: 3, // 1 runner + 2 actions
		},
		{
			name: "runs-on array with deprecated",
			yaml: `
name: Array Runner
jobs:
  build:
    runs-on: [self-hosted, ubuntu-18.04]
    steps:
      - run: echo hello
`,
			expectFindings: 1,
			expectRuleID:   "deprecation/runner",
		},
		{
			name: "SHA pinned action skipped",
			yaml: `
name: SHA Pinned
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd
`,
			expectFindings: 0, // Can't determine version from SHA
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")
			require.NoError(t, os.WriteFile(workflowPath, []byte(tt.yaml), 0o600))

			v := validator.NewDeprecationValidator()
			ctx := context.Background()
			findings, err := v.Validate(ctx, workflowPath)
			require.NoError(t, err)

			assert.Len(t, findings, tt.expectFindings)

			if tt.expectRuleID != "" && len(findings) > 0 {
				assert.Equal(t, tt.expectRuleID, findings[0].RuleID)
				assert.Equal(t, validator.SourceDeprecation, findings[0].Source)
				assert.NotEmpty(t, findings[0].Suggestion)
			}
		})
	}
}

// TestDeprecationValidator_Validate_InvalidYAML tests handling of invalid YAML.
func TestDeprecationValidator_Validate_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "invalid.yml")
	require.NoError(t, os.WriteFile(workflowPath, []byte("{{{invalid"), 0o600))

	v := validator.NewDeprecationValidator()
	ctx := context.Background()
	findings, err := v.Validate(ctx, workflowPath)

	// Should not error, just return nil findings
	require.NoError(t, err)
	assert.Nil(t, findings)
}

// TestDeprecationValidator_Validate_FileNotFound tests error handling for missing files.
func TestDeprecationValidator_Validate_FileNotFound(t *testing.T) {
	t.Parallel()

	v := validator.NewDeprecationValidator()
	ctx := context.Background()
	_, err := v.Validate(ctx, "/nonexistent/path/workflow.yml")
	require.Error(t, err)
}

// TestDeprecationValidator_VersionNormalization tests version normalization.
func TestDeprecationValidator_VersionNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		uses           string
		expectFindings int
	}{
		{
			name:           "v3.0.0 is deprecated",
			uses:           "actions/checkout@v3.0.0",
			expectFindings: 1,
		},
		{
			name:           "v3.1.2 is deprecated",
			uses:           "actions/checkout@v3.1.2",
			expectFindings: 1,
		},
		{
			name:           "v4 is current",
			uses:           "actions/checkout@v4",
			expectFindings: 0,
		},
		{
			name:           "v4.0.0 is current",
			uses:           "actions/checkout@v4.0.0",
			expectFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yaml := `
name: Test
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ` + tt.uses

			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")
			require.NoError(t, os.WriteFile(workflowPath, []byte(yaml), 0o600))

			v := validator.NewDeprecationValidator()
			ctx := context.Background()
			findings, err := v.Validate(ctx, workflowPath)
			require.NoError(t, err)

			assert.Len(t, findings, tt.expectFindings)
		})
	}
}
