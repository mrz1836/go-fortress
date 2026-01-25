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

// TestEnvValidator_Name tests the validator name.
func TestEnvValidator_Name(t *testing.T) {
	t.Parallel()

	v := validator.NewEnvValidator()
	assert.Equal(t, "env", v.Name())
}

// TestEnvValidator_Validate tests .env.base validation.
func TestEnvValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filename       string
		content        string
		expectFindings int
		expectRuleIDs  []string
	}{
		{
			name:           "valid env file",
			filename:       ".env.base",
			content:        "API_KEY=abc123\nDB_HOST=localhost\n",
			expectFindings: 0,
		},
		{
			name:           "non-env-base file is skipped",
			filename:       ".env",
			content:        "invalid line without equals",
			expectFindings: 0, // Skipped entirely
		},
		{
			name:           "lowercase variable name",
			filename:       ".env.base",
			content:        "api_key=abc123\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/naming-convention"},
		},
		{
			name:           "mixed case variable name",
			filename:       ".env.base",
			content:        "ApiKey=abc123\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/naming-convention"},
		},
		{
			name:           "invalid syntax",
			filename:       ".env.base",
			content:        "this is not valid\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/invalid-syntax"},
		},
		{
			name:           "empty required value",
			filename:       ".env.base",
			content:        "API_KEY=\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/empty-required"},
		},
		{
			name:           "unquoted value with spaces",
			filename:       ".env.base",
			content:        "MESSAGE=hello world\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/unquoted-spaces"},
		},
		{
			name:           "quoted value with spaces is ok",
			filename:       ".env.base",
			content:        "MESSAGE=\"hello world\"\n",
			expectFindings: 0,
		},
		{
			name:           "single quoted value with spaces is ok",
			filename:       ".env.base",
			content:        "MESSAGE='hello world'\n",
			expectFindings: 0,
		},
		{
			name:           "boolean variable with non-boolean value",
			filename:       ".env.base",
			content:        "ENABLE_FEATURE=maybe\n",
			expectFindings: 1,
			expectRuleIDs:  []string{"env/boolean-value"},
		},
		{
			name:           "boolean variable with true is ok",
			filename:       ".env.base",
			content:        "ENABLE_FEATURE=true\n",
			expectFindings: 0,
		},
		{
			name:           "boolean variable with false is ok",
			filename:       ".env.base",
			content:        "DISABLE_LOGGING=false\n",
			expectFindings: 0,
		},
		{
			name:           "boolean variable with 1 is ok",
			filename:       ".env.base",
			content:        "FEATURE_ENABLED=1\n",
			expectFindings: 0,
		},
		{
			name:           "boolean variable with 0 is ok",
			filename:       ".env.base",
			content:        "FEATURE_DISABLED=0\n",
			expectFindings: 0,
		},
		{
			name:           "version variable suggestion",
			filename:       ".env.base",
			content:        "GO_VER=1.24\n",
			expectFindings: 1, // version naming only (GO_VER is valid UPPER_SNAKE_CASE)
			expectRuleIDs:  []string{"env/version-naming"},
		},
		{
			name:           "proper version variable is ok",
			filename:       ".env.base",
			content:        "GO_VERSION=1.24\n",
			expectFindings: 0,
		},
		{
			name:           "section headers are parsed",
			filename:       ".env.base",
			content:        "# === Database Settings ===\nDB_HOST=localhost\n",
			expectFindings: 0,
		},
		{
			name:           "section headers with dashes",
			filename:       ".env.base",
			content:        "# --- API Configuration ---\nAPI_URL=http://localhost\n",
			expectFindings: 0,
		},
		{
			name:           "empty lines are skipped",
			filename:       ".env.base",
			content:        "FOO=bar\n\n\nBAZ=qux\n",
			expectFindings: 0,
		},
		{
			name:           "comment lines are skipped",
			filename:       ".env.base",
			content:        "# This is a comment\nFOO=bar\n",
			expectFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, tt.filename)
			require.NoError(t, os.WriteFile(filePath, []byte(tt.content), 0o600))

			v := validator.NewEnvValidator()
			ctx := context.Background()
			findings, err := v.Validate(ctx, filePath)
			require.NoError(t, err)

			assert.Len(t, findings, tt.expectFindings)

			if len(tt.expectRuleIDs) > 0 && len(findings) > 0 {
				foundRuleIDs := make(map[string]bool)
				for _, f := range findings {
					foundRuleIDs[f.RuleID] = true
				}

				for _, expectedID := range tt.expectRuleIDs {
					assert.True(t, foundRuleIDs[expectedID],
						"expected rule %q to be in findings", expectedID)
				}
			}
		})
	}
}

// TestEnvValidator_Validate_FileNotFound tests error handling for missing files.
func TestEnvValidator_Validate_FileNotFound(t *testing.T) {
	t.Parallel()

	v := validator.NewEnvValidator()
	ctx := context.Background()
	_, err := v.Validate(ctx, "/nonexistent/.env.base")
	require.Error(t, err)
}

// TestEnvValidator_Validate_MultipleIssues tests multiple issues in one file.
func TestEnvValidator_Validate_MultipleIssues(t *testing.T) {
	t.Parallel()

	content := `# Config
api_key=
ENABLE_FEATURE=maybe
message=hello world
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, ".env.base")
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

	v := validator.NewEnvValidator()
	ctx := context.Background()
	findings, err := v.Validate(ctx, filePath)
	require.NoError(t, err)

	// Should have multiple findings
	assert.GreaterOrEqual(t, len(findings), 3)

	// Check all findings have source set
	for _, f := range findings {
		assert.Equal(t, validator.SourceEnv, f.Source)
		assert.NotEmpty(t, f.Suggestion)
	}
}
