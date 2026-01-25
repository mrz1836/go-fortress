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

// TestActionlintValidator_Name tests the validator name.
func TestActionlintValidator_Name(t *testing.T) {
	t.Parallel()

	v := validator.NewActionlintValidator("")
	assert.Equal(t, "actionlint", v.Name())
}

// TestActionlintValidator_Validate tests workflow validation.
func TestActionlintValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yaml           string
		expectFindings bool
		expectError    bool
	}{
		{
			name: "valid workflow",
			yaml: `
name: Valid
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo hello
`,
			expectFindings: false,
		},
		{
			name: "workflow with syntax error",
			yaml: `
name: Invalid
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ${{ invalid_expression }}
`,
			expectFindings: true,
		},
		{
			name: "workflow with undefined step id",
			yaml: `
name: Bad Step ID
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo ${{ steps.nonexistent.outputs.value }}
`,
			expectFindings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp workflow file
			tmpDir := t.TempDir()
			// Create proper directory structure
			workflowDir := filepath.Join(tmpDir, ".github", "workflows")
			require.NoError(t, os.MkdirAll(workflowDir, 0o750))
			workflowPath := filepath.Join(workflowDir, "test.yml")
			require.NoError(t, os.WriteFile(workflowPath, []byte(tt.yaml), 0o600))

			v := validator.NewActionlintValidator("")
			ctx := context.Background()
			findings, err := v.Validate(ctx, workflowPath)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectFindings {
				assert.NotEmpty(t, findings, "expected findings")
				for _, f := range findings {
					assert.NotEmpty(t, f.RuleID, "rule ID should be set")
					assert.Equal(t, validator.SourceActionlint, f.Source)
				}
			}
		})
	}
}

// TestActionlintValidator_Validate_FileNotFound tests error handling for missing files.
func TestActionlintValidator_Validate_FileNotFound(t *testing.T) {
	t.Parallel()

	v := validator.NewActionlintValidator("")
	ctx := context.Background()
	_, err := v.Validate(ctx, "/nonexistent/path/workflow.yml")
	require.Error(t, err)
}
