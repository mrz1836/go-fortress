# Data Model: Fortress Guardian CI Testing Framework

**Date**: 2026-01-24
**Branch**: `001-ci-testing-framework`

---

## Overview

This document defines the core entities, their relationships, and validation rules for the Fortress Guardian CI Testing Framework.

---

## Entity Diagram

```text
                                    ┌─────────────────┐
                                    │    Guardian     │
                                    │   (Orchestrator)│
                                    └────────┬────────┘
                                             │
                   ┌─────────────────────────┼─────────────────────────┐
                   │                         │                         │
                   ▼                         ▼                         ▼
          ┌───────────────┐         ┌───────────────┐         ┌───────────────┐
          │   Validator   │         │    Runner     │         │   Reporter    │
          │ (Static Checks)│         │ (Act Wrapper) │         │(Output Format)│
          └───────┬───────┘         └───────┬───────┘         └───────┬───────┘
                  │                         │                         │
                  ▼                         ▼                         ▼
          ┌───────────────┐         ┌───────────────┐         ┌───────────────┐
          │   Finding[]   │         │   Scenario[]  │         │    Report     │
          └───────────────┘         └───────┬───────┘         └───────────────┘
                                            │
                                            ▼
                                    ┌───────────────┐
                                    │   Fixture     │
                                    └───────────────┘

         ┌───────────────┐
         │    Policy     │◄────── PolicyEngine evaluates
         └───────┬───────┘        Workflow against policies
                 │
                 ▼
         ┌───────────────┐
         │   Exception   │
         └───────────────┘
```

---

## Core Entities

### 1. Scenario

A single test case that validates specific CI behavior.

```go
// scenarios/scenarios.go

// Scenario defines a CI test case with expected outcomes.
type Scenario struct {
    // ID uniquely identifies the scenario (e.g., "LINT-001").
    // Format: CATEGORY-NNN where NNN is zero-padded number.
    ID string `json:"id"`

    // Category groups related scenarios (Quality, Testing, Security, etc.).
    Category Category `json:"category"`

    // Description explains what the scenario tests.
    Description string `json:"description"`

    // FixturePath is the relative path to the fixture directory.
    // Path is relative to repository root (e.g., ".github/ci-tester/fixtures/lint-fail").
    FixturePath string `json:"fixture_path"`

    // EventFile is the path to the GitHub event JSON payload.
    // Optional; defaults to push event if not specified.
    EventFile string `json:"event_file,omitempty"`

    // Workflow is the workflow file to execute.
    // Path relative to fixture's .github/workflows/ directory.
    Workflow string `json:"workflow"`

    // Job is the specific job to run within the workflow.
    // Optional; runs all jobs if not specified.
    Job string `json:"job,omitempty"`

    // Expected defines the success criteria for this scenario.
    Expected ExpectedResult `json:"expected"`

    // Timeout is the maximum execution time for this scenario.
    // Defaults to 30s if not specified.
    Timeout time.Duration `json:"timeout,omitempty"`

    // Tags enable filtering scenarios (e.g., ["fast", "security", "p1"]).
    Tags []string `json:"tags,omitempty"`
}

// Category groups scenarios by type.
type Category string

const (
    CategoryQuality     Category = "Quality"
    CategoryTesting     Category = "Testing"
    CategorySecurity    Category = "Security"
    CategoryCoverage    Category = "Coverage"
    CategoryForkSafety  Category = "Fork Safety"
    CategoryConfig      Category = "Config"
    CategoryServices    Category = "Services"
    CategoryTooling     Category = "Tooling"
    CategoryArtifacts   Category = "Artifacts"
    CategorySupplyChain Category = "Supply Chain"
)

// ExpectedResult defines what the scenario should produce.
type ExpectedResult struct {
    // Status is the expected workflow conclusion: "success" or "failure".
    Status ExpectedStatus `json:"status"`

    // LogPatterns are regex patterns that MUST appear in the output.
    // All patterns must match for the scenario to pass.
    LogPatterns []string `json:"log_patterns,omitempty"`

    // ExcludePatterns are regex patterns that MUST NOT appear in output.
    ExcludePatterns []string `json:"exclude_patterns,omitempty"`

    // ExitCode is the expected process exit code.
    // Defaults to 0 for "success", 1 for "failure".
    ExitCode *int `json:"exit_code,omitempty"`

    // Outputs are key-value pairs expected in workflow outputs.
    Outputs map[string]string `json:"outputs,omitempty"`
}

// ExpectedStatus represents the expected workflow status.
type ExpectedStatus string

const (
    StatusSuccess ExpectedStatus = "success"
    StatusFailure ExpectedStatus = "failure"
)
```

**Validation Rules**:
- `ID` must match pattern `^[A-Z]+-[0-9]{3}$`
- `Category` must be one of the defined constants
- `FixturePath` must exist and contain a valid Go module
- `Workflow` must exist within the fixture
- `Timeout` must be between 1s and 10m
- At least one of `LogPatterns` or `Status` must be specified

**State Transitions**: N/A (Scenarios are immutable definitions)

---

### 2. Fixture

A pre-configured repository state designed to trigger specific failures.

```go
// runner/fixture.go

// Fixture represents a test repository with known characteristics.
type Fixture struct {
    // Path is the absolute path to the fixture directory.
    Path string `json:"path"`

    // Name is the fixture identifier (directory name).
    Name string `json:"name"`

    // Module is the Go module path (from go.mod).
    Module string `json:"module"`

    // FailureType describes what kind of failure this fixture produces.
    FailureType FailureType `json:"failure_type"`

    // Workflows lists available workflow files in the fixture.
    Workflows []string `json:"workflows"`
}

// FailureType categorizes how the fixture fails.
type FailureType string

const (
    FailureLint     FailureType = "lint"
    FailureTest     FailureType = "test"
    FailureRace     FailureType = "race"
    FailureSecurity FailureType = "security"
    FailureVuln     FailureType = "vulnerability"
    FailureCoverage FailureType = "coverage"
    FailureConfig   FailureType = "config"
    FailureNone     FailureType = "none" // For success scenarios
)
```

**Validation Rules**:
- `Path` must be an absolute path to an existing directory
- Fixture must contain `go.mod` file
- Fixture must contain `.github/workflows/` directory with at least one `.yml` file

---

### 3. Finding

A single validation result from static analysis or policy checks.

```go
// validator/finding.go

// Finding represents a single issue detected during validation.
type Finding struct {
    // RuleID identifies the rule that triggered this finding.
    // Format: "guardian/<rule-name>" or "actionlint/<kind>".
    RuleID string `json:"rule_id"`

    // Severity indicates the importance of this finding.
    Severity Severity `json:"severity"`

    // Message describes the issue found.
    Message string `json:"message"`

    // File is the path to the file containing the issue.
    // Path is relative to repository root.
    File string `json:"file"`

    // Line is the 1-based line number where the issue occurs.
    Line int `json:"line"`

    // Column is the 1-based column number where the issue starts.
    // May be 0 if not applicable.
    Column int `json:"column,omitempty"`

    // EndLine is the 1-based line where the issue ends.
    // Optional; same as Line if not specified.
    EndLine int `json:"end_line,omitempty"`

    // EndColumn is the 1-based column where the issue ends.
    EndColumn int `json:"end_column,omitempty"`

    // Source identifies what detected this finding.
    Source FindingSource `json:"source"`

    // Suggestion is an optional fix recommendation.
    Suggestion string `json:"suggestion,omitempty"`

    // Fingerprint is a hash for deduplication.
    Fingerprint string `json:"fingerprint,omitempty"`
}

// Severity levels for findings.
type Severity string

const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityNote    Severity = "note"
    SeverityInfo    Severity = "info"
)

// FindingSource identifies what tool generated the finding.
type FindingSource string

const (
    SourceActionlint   FindingSource = "actionlint"
    SourcePolicy       FindingSource = "policy"
    SourceSchema       FindingSource = "schema"
    SourceDeprecation  FindingSource = "deprecation"
)
```

**Validation Rules**:
- `RuleID` must not be empty
- `Severity` must be one of the defined constants
- `Message` must not be empty
- `File` must not be empty
- `Line` must be >= 1

---

### 4. Policy

A rule that validates workflow content against security/quality standards.

```go
// policy/rules.go

// Policy defines a validation rule for workflows.
type Policy struct {
    // ID uniquely identifies the policy (e.g., "sha-pinned-actions").
    ID string `json:"id"`

    // Severity is the default severity when this policy is violated.
    Severity Severity `json:"severity"`

    // Description explains what this policy checks.
    Description string `json:"description"`

    // HelpURL links to documentation about this policy.
    HelpURL string `json:"help_url,omitempty"`

    // Tags enable filtering policies (e.g., ["security", "required"]).
    Tags []string `json:"tags,omitempty"`

    // Check is the function that validates a workflow.
    // Returns findings for any violations detected.
    Check PolicyCheckFunc `json:"-"`
}

// PolicyCheckFunc is the signature for policy check functions.
// It receives a parsed workflow and returns any violations found.
type PolicyCheckFunc func(workflow *Workflow) []Finding

// Workflow represents a parsed GitHub Actions workflow.
// This is a simplified representation for policy checking.
type Workflow struct {
    // Name is the workflow name.
    Name string

    // Path is the file path relative to repository root.
    Path string

    // Permissions declared at workflow level.
    Permissions *Permissions

    // Jobs in this workflow.
    Jobs map[string]*Job

    // On defines the workflow triggers.
    On *WorkflowTriggers

    // Concurrency settings if defined.
    Concurrency *Concurrency

    // Raw is the original YAML content.
    Raw []byte
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
    ID       string
    Name     string
    Uses     string // Action reference (e.g., "actions/checkout@v4")
    Run      string
    With     map[string]interface{}
    Env      map[string]string
    If       string
    Line     int // Line number in workflow file
}

// Permissions defines GitHub token permissions.
type Permissions struct {
    Actions       string
    Checks        string
    Contents      string
    Deployments   string
    Issues        string
    Packages      string
    PullRequests  string
    SecurityEvents string
    Statuses      string
}
```

**Validation Rules**:
- `ID` must match pattern `^[a-z][a-z0-9-]*$`
- `Severity` must be one of the defined constants
- `Description` must not be empty
- `Check` must not be nil

---

### 5. Exception

A configured exemption from a policy rule.

```go
// policy/exceptions.go

// Exception allows bypassing a policy for specific files or patterns.
type Exception struct {
    // Policy is the ID of the policy to bypass.
    Policy string `yaml:"policy" json:"policy"`

    // Path is a glob pattern matching files to exempt.
    // Example: ".github/workflows/test.yml"
    Path string `yaml:"path" json:"path"`

    // Reason documents why this exception exists.
    Reason string `yaml:"reason" json:"reason"`

    // Expires is when this exception should be reviewed.
    // Optional; exceptions without expiration are permanent.
    Expires *time.Time `yaml:"expires,omitempty" json:"expires,omitempty"`

    // ApprovedBy records who approved this exception.
    ApprovedBy string `yaml:"approved_by,omitempty" json:"approved_by,omitempty"`

    // CreatedAt is when the exception was created.
    CreatedAt time.Time `yaml:"created_at" json:"created_at"`
}

// ExceptionConfig is the structure of .github/guardian.yaml.
type ExceptionConfig struct {
    // Exceptions lists all configured policy exemptions.
    Exceptions []Exception `yaml:"exceptions" json:"exceptions"`
}
```

**Validation Rules**:
- `Policy` must match an existing policy ID
- `Path` must be a valid glob pattern
- `Reason` must not be empty
- `Expires` if set, must be in the future
- `CreatedAt` must not be zero

---

### 6. Report

Aggregated results from a Guardian run.

```go
// reporter/report.go

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
    ModeStatic  RunMode = "static"
    ModeTest    RunMode = "test"
    ModeVerify  RunMode = "verify"
    ModeScenario RunMode = "scenario"
)

// StaticResults contains all static validation findings.
type StaticResults struct {
    // Findings from all validators and policies.
    Findings []Finding `json:"findings"`

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
    ActualStatus ExpectedStatus `json:"actual_status"`

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
    FindingsByLevel map[Severity]int `json:"findings_by_level"`

    // ExceptionsApplied counts policy exceptions used.
    ExceptionsApplied int `json:"exceptions_applied"`
}
```

**Validation Rules**:
- `Version` must not be empty
- `StartTime` must be before `EndTime`
- `Mode` must be one of the defined constants
- Summary counts must be consistent with results

---

## Entity Relationships

```text
Guardian (1) ─── contains ──→ (1) Config
    │
    ├── uses ──→ (n) Validator ──→ produces ──→ (n) Finding
    │
    ├── uses ──→ (1) PolicyEngine
    │               │
    │               ├── loads ──→ (n) Policy
    │               │
    │               └── loads ──→ (n) Exception
    │
    ├── uses ──→ (1) Runner
    │               │
    │               └── executes ──→ (n) Scenario
    │                                   │
    │                                   └── uses ──→ (1) Fixture
    │
    └── uses ──→ (n) Reporter ──→ generates ──→ (1) Report
```

---

## Configuration Entity

```go
// config.go

// Config holds all Guardian configuration.
type Config struct {
    // ActPath is the path to the act binary.
    ActPath string `json:"act_path"`

    // ActionlintPath is the path to actionlint binary.
    ActionlintPath string `json:"actionlint_path"`

    // WorkflowsDir is the path to .github/workflows/.
    WorkflowsDir string `json:"workflows_dir"`

    // FixturesDir is the path to fixtures directory.
    FixturesDir string `json:"fixtures_dir"`

    // OutputDir is where reports are written.
    OutputDir string `json:"output_dir"`

    // ParallelScenarios is max concurrent scenario execution.
    ParallelScenarios int `json:"parallel_scenarios"`

    // ScenarioTimeout is default timeout for scenarios.
    ScenarioTimeout time.Duration `json:"scenario_timeout"`

    // StaticTimeout is timeout for static validation.
    StaticTimeout time.Duration `json:"static_timeout"`

    // ExceptionsFile is path to guardian.yaml.
    ExceptionsFile string `json:"exceptions_file"`

    // Verbose enables detailed output.
    Verbose bool `json:"verbose"`

    // DryRun skips actual execution.
    DryRun bool `json:"dry_run"`
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
        Verbose:           false,
        DryRun:            false,
    }
}
```

---

## File Formats

### JSONL Output (ci-results.jsonl)

```jsonl
{"type":"run_start","timestamp":"2026-01-24T10:00:00Z","version":"1.0.0","mode":"verify"}
{"type":"static_complete","timestamp":"2026-01-24T10:00:02Z","findings":5,"duration_ms":1850}
{"type":"scenario","id":"LINT-001","status":"pass","duration_ms":4523,"exit_code":1}
{"type":"scenario","id":"TEST-001","status":"pass","duration_ms":6891,"exit_code":1}
{"type":"scenario","id":"SEC-001","status":"fail","duration_ms":3421,"error":"unexpected success"}
{"type":"run_end","timestamp":"2026-01-24T10:05:32Z","passed":34,"failed":1,"skipped":0}
```

### Policy Exceptions (.github/guardian.yaml)

```yaml
exceptions:
  - policy: sha-pinned-actions
    path: .github/workflows/test-experimental.yml
    reason: "Experimental workflow testing unpinned action behavior"
    expires: 2026-06-01
    approved_by: "@mrz1836"
    created_at: 2026-01-24

  - policy: explicit-permissions
    path: .github/workflows/dependabot-*.yml
    reason: "Dependabot workflows use inherited permissions"
    approved_by: "@mrz1836"
    created_at: 2026-01-24
```

---

## Index

| Entity | Package | Primary Key |
|--------|---------|-------------|
| Scenario | `guardian/scenarios` | `ID` |
| Fixture | `guardian/runner` | `Path` |
| Finding | `guardian/validator` | `Fingerprint` |
| Policy | `guardian/policy` | `ID` |
| Exception | `guardian/policy` | `Policy` + `Path` |
| Report | `guardian/reporter` | `StartTime` |
| Config | `guardian` | N/A (singleton) |
