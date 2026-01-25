package guardian

import (
	"os"
	"strconv"
	"time"
)

// Config holds all Guardian configuration.
type Config struct {
	// ActPath is the path to the act binary.
	ActPath string

	// ActionlintPath is the path to actionlint binary.
	ActionlintPath string

	// WorkflowsDir is the path to .github/workflows/.
	WorkflowsDir string

	// FixturesDir is the path to fixtures directory.
	FixturesDir string

	// OutputDir is where reports are written.
	OutputDir string

	// ParallelScenarios is max concurrent scenario execution.
	ParallelScenarios int

	// ScenarioTimeout is default timeout for scenarios.
	ScenarioTimeout time.Duration

	// StaticTimeout is timeout for static validation.
	StaticTimeout time.Duration

	// ExceptionsFile is path to guardian.yaml.
	ExceptionsFile string

	// SARIFOutput is the filename for SARIF output.
	SARIFOutput string

	// JSONLOutput is the filename for JSONL output.
	JSONLOutput string

	// Verbose enables detailed output.
	Verbose bool

	// DryRun skips actual execution.
	DryRun bool

	// KeepContainers preserves containers after execution (for debugging).
	KeepContainers bool

	// PolicyStrict enables strict policy enforcement.
	PolicyStrict bool
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ActPath:           "act",
		ActionlintPath:    "actionlint",
		WorkflowsDir:      ".github/workflows",
		FixturesDir:       ".github/ci-tester/fixtures",
		OutputDir:         ".mage-x",
		ParallelScenarios: 4,
		ScenarioTimeout:   30 * time.Second,
		StaticTimeout:     5 * time.Second,
		ExceptionsFile:    ".github/guardian.yaml",
		SARIFOutput:       "guardian.sarif",
		JSONLOutput:       "ci-results.jsonl",
		Verbose:           false,
		DryRun:            false,
		KeepContainers:    false,
		PolicyStrict:      true,
	}
}

// LoadFromEnv loads configuration from environment variables.
// Environment variables override default values.
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	// Tool paths
	if v := os.Getenv("GUARDIAN_ACT_PATH"); v != "" {
		cfg.ActPath = v
	}
	if v := os.Getenv("GUARDIAN_ACTIONLINT_PATH"); v != "" {
		cfg.ActionlintPath = v
	}

	// Directories
	if v := os.Getenv("GUARDIAN_WORKFLOWS_DIR"); v != "" {
		cfg.WorkflowsDir = v
	}
	if v := os.Getenv("GUARDIAN_FIXTURES_DIR"); v != "" {
		cfg.FixturesDir = v
	}
	if v := os.Getenv("GUARDIAN_OUTPUT_DIR"); v != "" {
		cfg.OutputDir = v
	}

	// Execution settings
	if v := os.Getenv("GUARDIAN_PARALLEL_SCENARIOS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ParallelScenarios = n
		}
	}
	if v := os.Getenv("GUARDIAN_SCENARIO_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ScenarioTimeout = d
		}
	}
	if v := os.Getenv("GUARDIAN_STATIC_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.StaticTimeout = d
		}
	}

	// Output configuration
	if v := os.Getenv("GUARDIAN_EXCEPTIONS_FILE"); v != "" {
		cfg.ExceptionsFile = v
	}
	if v := os.Getenv("GUARDIAN_SARIF_OUTPUT"); v != "" {
		cfg.SARIFOutput = v
	}
	if v := os.Getenv("GUARDIAN_JSONL_OUTPUT"); v != "" {
		cfg.JSONLOutput = v
	}

	// Debug settings
	if v := os.Getenv("GUARDIAN_VERBOSE"); v != "" {
		cfg.Verbose = parseBool(v)
	}
	if v := os.Getenv("GUARDIAN_DRY_RUN"); v != "" {
		cfg.DryRun = parseBool(v)
	}
	if v := os.Getenv("GUARDIAN_KEEP_CONTAINERS"); v != "" {
		cfg.KeepContainers = parseBool(v)
	}

	// Policy configuration
	if v := os.Getenv("GUARDIAN_POLICY_STRICT"); v != "" {
		cfg.PolicyStrict = parseBool(v)
	}

	return cfg
}

// parseBool parses a string as a boolean value.
// Accepts: "true", "1", "yes", "on" as true (case-insensitive).
// All other values return false.
func parseBool(s string) bool {
	switch s {
	case "true", "True", "TRUE", "1", "yes", "Yes", "YES", "on", "On", "ON":
		return true
	default:
		return false
	}
}
