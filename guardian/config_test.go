package guardian_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian"
)

// TestDefaultConfig verifies that DefaultConfig returns sensible defaults.
func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := guardian.DefaultConfig()
	require.NotNil(t, cfg)

	// Check string defaults
	assert.Equal(t, "act", cfg.ActPath)
	assert.Equal(t, "actionlint", cfg.ActionlintPath)
	assert.Equal(t, ".github/workflows", cfg.WorkflowsDir)
	assert.Equal(t, ".github/ci-tester/fixtures", cfg.FixturesDir)
	assert.Equal(t, ".mage-x", cfg.OutputDir)
	assert.Equal(t, ".github/guardian.yaml", cfg.ExceptionsFile)
	assert.Equal(t, "guardian.sarif", cfg.SARIFOutput)
	assert.Equal(t, "ci-results.jsonl", cfg.JSONLOutput)

	// Check numeric defaults
	assert.Equal(t, 4, cfg.ParallelScenarios)
	assert.Equal(t, 30*time.Second, cfg.ScenarioTimeout)
	assert.Equal(t, 5*time.Second, cfg.StaticTimeout)

	// Check boolean defaults
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
	assert.False(t, cfg.KeepContainers)
	assert.True(t, cfg.PolicyStrict)
}

// TestLoadFromEnv tests that environment variables override defaults.
func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *guardian.Config)
	}{
		{
			name: "string env vars override defaults",
			envVars: map[string]string{
				"GUARDIAN_ACT_PATH":        "/custom/act",
				"GUARDIAN_ACTIONLINT_PATH": "/custom/actionlint",
				"GUARDIAN_WORKFLOWS_DIR":   "/custom/workflows",
				"GUARDIAN_FIXTURES_DIR":    "/custom/fixtures",
				"GUARDIAN_OUTPUT_DIR":      "/custom/output",
				"GUARDIAN_EXCEPTIONS_FILE": "/custom/exceptions.yaml",
				"GUARDIAN_SARIF_OUTPUT":    "custom.sarif",
				"GUARDIAN_JSONL_OUTPUT":    "custom.jsonl",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, "/custom/act", cfg.ActPath)
				assert.Equal(t, "/custom/actionlint", cfg.ActionlintPath)
				assert.Equal(t, "/custom/workflows", cfg.WorkflowsDir)
				assert.Equal(t, "/custom/fixtures", cfg.FixturesDir)
				assert.Equal(t, "/custom/output", cfg.OutputDir)
				assert.Equal(t, "/custom/exceptions.yaml", cfg.ExceptionsFile)
				assert.Equal(t, "custom.sarif", cfg.SARIFOutput)
				assert.Equal(t, "custom.jsonl", cfg.JSONLOutput)
			},
		},
		{
			name: "integer env var overrides default",
			envVars: map[string]string{
				"GUARDIAN_PARALLEL_SCENARIOS": "8",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, 8, cfg.ParallelScenarios)
			},
		},
		{
			name: "invalid integer env var keeps default",
			envVars: map[string]string{
				"GUARDIAN_PARALLEL_SCENARIOS": "invalid",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, 4, cfg.ParallelScenarios) // default
			},
		},
		{
			name: "zero or negative integer keeps default",
			envVars: map[string]string{
				"GUARDIAN_PARALLEL_SCENARIOS": "0",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, 4, cfg.ParallelScenarios) // default
			},
		},
		{
			name: "duration env vars override defaults",
			envVars: map[string]string{
				"GUARDIAN_SCENARIO_TIMEOUT": "2m",
				"GUARDIAN_STATIC_TIMEOUT":   "10s",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, 2*time.Minute, cfg.ScenarioTimeout)
				assert.Equal(t, 10*time.Second, cfg.StaticTimeout)
			},
		},
		{
			name: "invalid duration env var keeps default",
			envVars: map[string]string{
				"GUARDIAN_SCENARIO_TIMEOUT": "invalid",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, 30*time.Second, cfg.ScenarioTimeout) // default
			},
		},
		{
			name: "boolean env vars true values",
			envVars: map[string]string{
				"GUARDIAN_VERBOSE":         "true",
				"GUARDIAN_DRY_RUN":         "1",
				"GUARDIAN_KEEP_CONTAINERS": "yes",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.True(t, cfg.Verbose)
				assert.True(t, cfg.DryRun)
				assert.True(t, cfg.KeepContainers)
			},
		},
		{
			name: "boolean env vars false values",
			envVars: map[string]string{
				"GUARDIAN_VERBOSE": "false",
				"GUARDIAN_DRY_RUN": "0",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.False(t, cfg.Verbose)
				assert.False(t, cfg.DryRun)
			},
		},
		{
			name: "policy strict can be disabled",
			envVars: map[string]string{
				"GUARDIAN_POLICY_STRICT": "false",
			},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.False(t, cfg.PolicyStrict)
			},
		},
		{
			name:    "empty env vars keep defaults",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *guardian.Config) {
				assert.Equal(t, "act", cfg.ActPath)
				assert.Equal(t, 4, cfg.ParallelScenarios)
				assert.Equal(t, 30*time.Second, cfg.ScenarioTimeout)
				assert.False(t, cfg.Verbose)
				assert.True(t, cfg.PolicyStrict)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any existing env vars
			envVarsToClear := []string{
				"GUARDIAN_ACT_PATH", "GUARDIAN_ACTIONLINT_PATH", "GUARDIAN_WORKFLOWS_DIR",
				"GUARDIAN_FIXTURES_DIR", "GUARDIAN_OUTPUT_DIR", "GUARDIAN_EXCEPTIONS_FILE",
				"GUARDIAN_SARIF_OUTPUT", "GUARDIAN_JSONL_OUTPUT", "GUARDIAN_PARALLEL_SCENARIOS",
				"GUARDIAN_SCENARIO_TIMEOUT", "GUARDIAN_STATIC_TIMEOUT", "GUARDIAN_VERBOSE",
				"GUARDIAN_DRY_RUN", "GUARDIAN_KEEP_CONTAINERS", "GUARDIAN_POLICY_STRICT",
			}
			for _, key := range envVarsToClear {
				_ = os.Unsetenv(key)
			}

			// Set test env vars
			for key, val := range tt.envVars {
				require.NoError(t, os.Setenv(key, val))
			}

			// Cleanup after test
			t.Cleanup(func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			})

			cfg := guardian.LoadFromEnv()
			require.NotNil(t, cfg)
			tt.validate(t, cfg)
		})
	}
}

// TestParseBool tests the parseBool function through LoadFromEnv.
// parseBool is unexported, so we test it indirectly via boolean env vars.
func TestParseBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// True values
		{"lowercase true", "true", true},
		{"capitalized True", "True", true},
		{"uppercase TRUE", "TRUE", true},
		{"numeric 1", "1", true},
		{"lowercase yes", "yes", true},
		{"capitalized Yes", "Yes", true},
		{"uppercase YES", "YES", true},
		{"lowercase on", "on", true},
		{"capitalized On", "On", true},
		{"uppercase ON", "ON", true},

		// False values
		{"lowercase false", "false", false},
		{"uppercase FALSE", "FALSE", false},
		{"numeric 0", "0", false},
		{"lowercase no", "no", false},
		{"lowercase off", "off", false},
		{"random string", "random", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var
			_ = os.Unsetenv("GUARDIAN_VERBOSE")

			// Skip empty string test - empty means keep default, not set to false
			if tt.input == "" {
				return
			}

			require.NoError(t, os.Setenv("GUARDIAN_VERBOSE", tt.input))
			t.Cleanup(func() {
				_ = os.Unsetenv("GUARDIAN_VERBOSE")
			})

			cfg := guardian.LoadFromEnv()
			assert.Equal(t, tt.expected, cfg.Verbose,
				"parseBool(%q) should return %v", tt.input, tt.expected)
		})
	}
}
