package reporter

import (
	"time"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Report contains all results from a Guardian execution.
type Report struct {
	// Version is the Guardian version that generated this report.
	Version string `json:"version"`

	// StartTime is when the run began.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the run completed.
	EndTime time.Time `json:"end_time"`

	// Duration is the total execution time.
	Duration time.Duration `json:"duration"`

	// Mode indicates what type of run this was.
	Mode RunMode `json:"mode"`

	// StaticResults contains findings from static validation.
	StaticResults *StaticResults `json:"static_results,omitempty"`

	// ScenarioResults contains results from scenario execution.
	ScenarioResults []ScenarioResult `json:"scenario_results,omitempty"`

	// Summary provides aggregate statistics.
	Summary ReportSummary `json:"summary"`

	// Metadata contains additional run information.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RunMode indicates what type of validation was performed.
type RunMode string

const (
	ModeStatic   RunMode = "static"
	ModeTest     RunMode = "test"
	ModeVerify   RunMode = "verify"
	ModeScenario RunMode = "scenario"
)

// StaticResults contains all static validation findings.
type StaticResults struct {
	// Findings from all validators and policies.
	Findings []validator.Finding `json:"findings"`

	// ValidatorsRun lists which validators executed.
	ValidatorsRun []string `json:"validators_run"`

	// Duration is time spent on static validation.
	Duration time.Duration `json:"duration"`
}

// ScenarioResult captures the outcome of a single scenario.
type ScenarioResult struct {
	// ScenarioID is the ID of the executed scenario.
	ScenarioID string `json:"scenario_id"`

	// Status indicates if the scenario passed or failed.
	Status ResultStatus `json:"status"`

	// ActualStatus is what the workflow returned.
	ActualStatus string `json:"actual_status"`

	// ExitCode is the actual process exit code.
	ExitCode int `json:"exit_code"`

	// Duration is how long the scenario took.
	Duration time.Duration `json:"duration"`

	// Output is the captured stdout/stderr.
	// Truncated if exceeds limit.
	Output string `json:"output,omitempty"`

	// MatchedPatterns lists which expected patterns were found.
	MatchedPatterns []string `json:"matched_patterns,omitempty"`

	// MissingPatterns lists expected patterns not found.
	MissingPatterns []string `json:"missing_patterns,omitempty"`

	// Error contains error message if scenario failed unexpectedly.
	Error string `json:"error,omitempty"`

	// LogPath is the path to the full log file.
	LogPath string `json:"log_path,omitempty"`
}

// ResultStatus indicates scenario outcome.
type ResultStatus string

const (
	ResultPass  ResultStatus = "pass"
	ResultFail  ResultStatus = "fail"
	ResultSkip  ResultStatus = "skip"
	ResultError ResultStatus = "error"
)

// ReportSummary provides aggregate statistics.
type ReportSummary struct {
	// TotalScenarios is the number of scenarios executed.
	TotalScenarios int `json:"total_scenarios"`

	// PassedScenarios is scenarios that passed.
	PassedScenarios int `json:"passed_scenarios"`

	// FailedScenarios is scenarios that failed.
	FailedScenarios int `json:"failed_scenarios"`

	// SkippedScenarios is scenarios that were skipped.
	SkippedScenarios int `json:"skipped_scenarios"`

	// ErrorScenarios is scenarios that had execution errors.
	ErrorScenarios int `json:"error_scenarios"`

	// TotalFindings from static validation.
	TotalFindings int `json:"total_findings"`

	// FindingsByLevel breaks down findings by severity.
	FindingsByLevel map[validator.Severity]int `json:"findings_by_level"`

	// ExceptionsApplied counts policy exceptions used.
	ExceptionsApplied int `json:"exceptions_applied"`
}
