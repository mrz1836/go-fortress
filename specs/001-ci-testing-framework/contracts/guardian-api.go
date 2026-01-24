// Package contracts provides CI testing framework design contracts.
//
// This file defines the public API interfaces for Fortress Guardian.
// These contracts specify what each component must implement.
// This is a design contract file, not implementation code.
// The actual implementations will be in the guardian/ package.
package contracts

import (
	"context"
	"io"
	"time"
)

// -----------------------------------------------------------------------------
// Core Guardian API
// -----------------------------------------------------------------------------

// Guardian is the main entry point for CI validation.
type Guardian interface {
	// RunStatic performs static validation only (no Docker required).
	// Returns findings from all validators and policy checks.
	// Target execution time: < 5 seconds.
	RunStatic(ctx context.Context) (*StaticResults, error)

	// RunTest executes quick validation scenarios.
	// Includes static validation plus fast failure scenarios.
	// Target execution time: < 60 seconds.
	RunTest(ctx context.Context) (*Report, error)

	// RunVerify executes comprehensive validation.
	// Includes all scenarios for pre-merge verification.
	// Target execution time: < 5 minutes.
	RunVerify(ctx context.Context) (*Report, error)

	// RunScenario executes a single scenario by ID.
	// Used for debugging specific CI behaviors.
	RunScenario(ctx context.Context, id string, opts ScenarioOptions) (*ScenarioResult, error)

	// ListScenarios returns all available scenarios.
	// Supports filtering by category or tags.
	ListScenarios(ctx context.Context, filter ScenarioFilter) ([]ScenarioInfo, error)
}

// NewGuardian creates a Guardian instance with the given configuration.
// func NewGuardian(cfg *Config) (Guardian, error)

// -----------------------------------------------------------------------------
// Validator Contracts
// -----------------------------------------------------------------------------

// Validator performs static analysis on workflow files.
type Validator interface {
	// Name returns the validator identifier (e.g., "actionlint", "schema").
	Name() string

	// Validate analyzes the given workflow file.
	// Returns findings for any issues detected.
	Validate(ctx context.Context, workflowPath string) ([]Finding, error)
}

// ValidatorRegistry manages available validators.
type ValidatorRegistry interface {
	// Register adds a validator to the registry.
	Register(v Validator)

	// Get returns a validator by name.
	Get(name string) (Validator, bool)

	// All returns all registered validators.
	All() []Validator

	// ValidateAll runs all validators against a workflow file.
	ValidateAll(ctx context.Context, workflowPath string) ([]Finding, error)
}

// -----------------------------------------------------------------------------
// Runner Contracts
// -----------------------------------------------------------------------------

// Runner executes GitHub Actions workflows locally.
type Runner interface {
	// Run executes a workflow using act.
	// Returns the execution result including output and exit code.
	Run(ctx context.Context, opts RunOptions) (*RunResult, error)

	// CheckAvailable verifies that act is installed and Docker is running.
	CheckAvailable(ctx context.Context) error
}

// RunOptions configures a workflow execution.
type RunOptions struct {
	// WorkingDir is the repository root for execution.
	WorkingDir string

	// WorkflowFile is the path to the workflow YAML file.
	WorkflowFile string

	// Job is the specific job to run (optional).
	Job string

	// EventFile is the path to the event payload JSON (optional).
	EventFile string

	// SecretsFile is the path to the secrets file (optional).
	SecretsFile string

	// Env is additional environment variables.
	Env map[string]string

	// Timeout is the maximum execution time.
	Timeout time.Duration

	// Verbose enables detailed output.
	Verbose bool

	// KeepContainer preserves the container after execution.
	KeepContainer bool
}

// RunResult contains the outcome of a workflow execution.
type RunResult struct {
	// ExitCode is the process exit code.
	ExitCode int

	// Output is the combined stdout/stderr.
	Output string

	// Duration is the execution time.
	Duration time.Duration

	// ContainerID is the container used (if KeepContainer was true).
	ContainerID string
}

// -----------------------------------------------------------------------------
// Scenario Contracts
// -----------------------------------------------------------------------------

// ScenarioRunner executes test scenarios.
type ScenarioRunner interface {
	// Execute runs a scenario and validates results.
	Execute(ctx context.Context, scenario *Scenario) (*ScenarioResult, error)

	// ExecuteAll runs multiple scenarios with parallelism control.
	ExecuteAll(ctx context.Context, scenarios []*Scenario) ([]ScenarioResult, error)
}

// ScenarioOptions configures single scenario execution.
type ScenarioOptions struct {
	// Verbose enables detailed output.
	Verbose bool

	// KeepContainer preserves container for debugging.
	KeepContainer bool

	// Timeout overrides scenario default timeout.
	Timeout time.Duration
}

// ScenarioFilter configures scenario listing.
type ScenarioFilter struct {
	// Category limits to specific category.
	Category string

	// Tags requires scenarios to have all specified tags.
	Tags []string

	// IncludeDisabled shows disabled scenarios.
	IncludeDisabled bool
}

// ScenarioInfo provides scenario metadata for listing.
type ScenarioInfo struct {
	// ID is the scenario identifier.
	ID string

	// Category is the scenario category.
	Category string

	// Description explains what the scenario tests.
	Description string

	// ExpectedStatus is "success" or "failure".
	ExpectedStatus string

	// Tags are scenario tags.
	Tags []string

	// Disabled indicates if scenario is currently disabled.
	Disabled bool
}

// -----------------------------------------------------------------------------
// Policy Engine Contracts
// -----------------------------------------------------------------------------

// PolicyEngine evaluates policies against workflows.
type PolicyEngine interface {
	// Evaluate runs all policies against a parsed workflow.
	// Returns findings for any violations.
	Evaluate(ctx context.Context, workflow *Workflow) ([]Finding, error)

	// LoadExceptions loads policy exceptions from configuration.
	LoadExceptions(ctx context.Context, configPath string) error

	// IsExcepted checks if a finding is covered by an exception.
	IsExcepted(finding *Finding) bool

	// Policies returns all registered policies.
	Policies() []PolicyInfo
}

// PolicyInfo provides policy metadata.
type PolicyInfo struct {
	// ID is the policy identifier.
	ID string

	// Severity is the default violation severity.
	Severity string

	// Description explains what the policy checks.
	Description string

	// HelpURL links to policy documentation.
	HelpURL string

	// Tags are policy tags.
	Tags []string
}

// -----------------------------------------------------------------------------
// Reporter Contracts
// -----------------------------------------------------------------------------

// Reporter generates output in a specific format.
type Reporter interface {
	// Name returns the reporter identifier (e.g., "jsonl", "sarif").
	Name() string

	// Write outputs the report to the given writer.
	Write(ctx context.Context, report *Report, w io.Writer) error

	// WriteFile outputs the report to a file.
	WriteFile(ctx context.Context, report *Report, path string) error
}

// ReporterRegistry manages available reporters.
type ReporterRegistry interface {
	// Register adds a reporter.
	Register(r Reporter)

	// Get returns a reporter by name.
	Get(name string) (Reporter, bool)

	// All returns all registered reporters.
	All() []Reporter
}

// TerminalReporter writes human-readable output to terminal.
type TerminalReporter interface {
	Reporter

	// SetColorEnabled enables/disables color output.
	SetColorEnabled(enabled bool)
}

// AnnotationReporter generates GitHub workflow commands for annotations.
type AnnotationReporter interface {
	Reporter

	// WriteAnnotations outputs findings as GitHub annotations.
	WriteAnnotations(ctx context.Context, findings []Finding, w io.Writer) error
}

// -----------------------------------------------------------------------------
// Data Types (from data-model.md)
// -----------------------------------------------------------------------------

// Scenario defines a CI test case.
type Scenario struct {
	ID          string
	Category    string
	Description string
	FixturePath string
	EventFile   string
	Workflow    string
	Job         string
	Expected    ExpectedResult
	Timeout     time.Duration
	Tags        []string
}

// ExpectedResult defines scenario success criteria.
type ExpectedResult struct {
	Status          string
	LogPatterns     []string
	ExcludePatterns []string
	ExitCode        *int
	Outputs         map[string]string
}

// Finding represents a detected issue.
type Finding struct {
	RuleID      string
	Severity    string
	Message     string
	File        string
	Line        int
	Column      int
	EndLine     int
	EndColumn   int
	Source      string
	Suggestion  string
	Fingerprint string
}

// Workflow is a parsed GitHub Actions workflow.
type Workflow struct {
	Name        string
	Path        string
	Permissions *Permissions
	Jobs        map[string]*Job
	Concurrency *Concurrency
	Raw         []byte
}

// Job represents a workflow job.
type Job struct {
	Name        string
	Permissions *Permissions
	Steps       []*Step
	RunsOn      []string
	If          string
	Needs       []string
	Concurrency *Concurrency
}

// Step represents a job step.
type Step struct {
	ID   string
	Name string
	Uses string
	Run  string
	With map[string]interface{}
	Env  map[string]string
	If   string
	Line int
}

// Permissions defines GitHub token permissions.
type Permissions struct {
	Actions        string
	Checks         string
	Contents       string
	Deployments    string
	Issues         string
	Packages       string
	PullRequests   string
	SecurityEvents string
	Statuses       string
}

// Concurrency defines concurrency settings.
type Concurrency struct {
	Group            string
	CancelInProgress bool
}

// Report contains Guardian execution results.
type Report struct {
	Version         string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Mode            string
	StaticResults   *StaticResults
	ScenarioResults []ScenarioResult
	Summary         ReportSummary
	Metadata        map[string]interface{}
}

// StaticResults contains static validation findings.
type StaticResults struct {
	Findings      []Finding
	ValidatorsRun []string
	Duration      time.Duration
}

// ScenarioResult contains scenario execution outcome.
type ScenarioResult struct {
	ScenarioID      string
	Status          string
	ActualStatus    string
	ExitCode        int
	Duration        time.Duration
	Output          string
	MatchedPatterns []string
	MissingPatterns []string
	Error           string
	LogPath         string
}

// ReportSummary provides aggregate statistics.
type ReportSummary struct {
	TotalScenarios    int
	PassedScenarios   int
	FailedScenarios   int
	SkippedScenarios  int
	ErrorScenarios    int
	TotalFindings     int
	FindingsByLevel   map[string]int
	ExceptionsApplied int
}

// Config holds Guardian configuration.
type Config struct {
	ActPath           string
	ActionlintPath    string
	WorkflowsDir      string
	FixturesDir       string
	OutputDir         string
	ParallelScenarios int
	ScenarioTimeout   time.Duration
	StaticTimeout     time.Duration
	ExceptionsFile    string
	Verbose           bool
	DryRun            bool
}
