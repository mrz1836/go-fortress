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

	// FlakeRetries is the max retries when a scenario exhibits the act/Docker
	// startup-flake signature (all expected log patterns missing AND the run
	// completed faster than FlakeFastFailThreshold). Set to 0 to disable.
	FlakeRetries int

	// FlakeFastFailThreshold is the duration below which a failed scenario
	// with missing log patterns is considered a flake. Real act runs spin up
	// a container and take seconds; flake fast-fails are typically sub-second.
	FlakeFastFailThreshold time.Duration

	// FlakeRetryBackoff is the wait between retry attempts. A brief pause
	// lets the Docker daemon settle between back-to-back launches.
	FlakeRetryBackoff time.Duration
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

		FlakeRetries:           1,
		FlakeFastFailThreshold: 2 * time.Second,
		FlakeRetryBackoff:      2 * time.Second,
	}
}

// LoadFromEnv loads configuration from environment variables.
// Environment variables override default values.
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	// Load string configurations
	loadStringEnv("GUARDIAN_ACT_PATH", &cfg.ActPath)
	loadStringEnv("GUARDIAN_ACTIONLINT_PATH", &cfg.ActionlintPath)
	loadStringEnv("GUARDIAN_WORKFLOWS_DIR", &cfg.WorkflowsDir)
	loadStringEnv("GUARDIAN_FIXTURES_DIR", &cfg.FixturesDir)
	loadStringEnv("GUARDIAN_OUTPUT_DIR", &cfg.OutputDir)
	loadStringEnv("GUARDIAN_EXCEPTIONS_FILE", &cfg.ExceptionsFile)
	loadStringEnv("GUARDIAN_SARIF_OUTPUT", &cfg.SARIFOutput)
	loadStringEnv("GUARDIAN_JSONL_OUTPUT", &cfg.JSONLOutput)

	// Load integer configurations
	loadIntEnv("GUARDIAN_PARALLEL_SCENARIOS", &cfg.ParallelScenarios)

	// FlakeRetries accepts 0 (kill switch), so it can't use loadIntEnv (>0).
	if v := os.Getenv("GUARDIAN_FLAKE_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.FlakeRetries = n
		}
	}

	// Load duration configurations
	loadDurationEnv("GUARDIAN_SCENARIO_TIMEOUT", &cfg.ScenarioTimeout)
	loadDurationEnv("GUARDIAN_STATIC_TIMEOUT", &cfg.StaticTimeout)
	loadDurationEnv("GUARDIAN_FLAKE_FAST_FAIL_THRESHOLD", &cfg.FlakeFastFailThreshold)
	loadDurationEnv("GUARDIAN_FLAKE_RETRY_BACKOFF", &cfg.FlakeRetryBackoff)

	// Load boolean configurations
	loadBoolEnv("GUARDIAN_VERBOSE", &cfg.Verbose)
	loadBoolEnv("GUARDIAN_DRY_RUN", &cfg.DryRun)
	loadBoolEnv("GUARDIAN_KEEP_CONTAINERS", &cfg.KeepContainers)
	loadBoolEnv("GUARDIAN_POLICY_STRICT", &cfg.PolicyStrict)

	return cfg
}

// loadStringEnv loads a string environment variable if set.
func loadStringEnv(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

// loadIntEnv loads an integer environment variable if set and valid.
func loadIntEnv(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			*target = n
		}
	}
}

// loadDurationEnv loads a duration environment variable if set and valid.
func loadDurationEnv(key string, target *time.Duration) {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			*target = d
		}
	}
}

// loadBoolEnv loads a boolean environment variable if set.
func loadBoolEnv(key string, target *bool) {
	if v := os.Getenv(key); v != "" {
		*target = parseBool(v)
	}
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
