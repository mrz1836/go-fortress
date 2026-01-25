package scenarios

import (
	"regexp"
	"time"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/runner"
)

// Scenario defines a CI test case with expected outcomes.
type Scenario struct {
	// ID uniquely identifies the scenario (e.g., "LINT-001").
	// Format: CATEGORY-NNN where NNN is zero-padded number.
	ID string

	// Category groups related scenarios (Quality, Testing, Security, etc.).
	Category Category

	// Description explains what the scenario tests.
	Description string

	// FixturePath is the relative path to the fixture directory.
	// Path is relative to repository root (e.g., ".github/ci-tester/fixtures/lint-fail").
	FixturePath string

	// EventFile is the path to the GitHub event JSON payload.
	// Optional; defaults to push event if not specified.
	EventFile string

	// Workflow is the workflow file to execute.
	// Path relative to fixture's .github/workflows/ directory.
	Workflow string

	// Job is the specific job to run within the workflow.
	// Optional; runs all jobs if not specified.
	Job string

	// Expected defines the success criteria for this scenario.
	Expected ExpectedResult

	// Timeout is the maximum execution time for this scenario.
	// Defaults to 30s if not specified.
	Timeout time.Duration

	// Tags enable filtering scenarios (e.g., ["fast", "security", "p1"]).
	Tags []string

	// Disabled indicates if this scenario is currently disabled.
	Disabled bool
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
	Status ExpectedStatus

	// LogPatterns are regex patterns that MUST appear in the output.
	// All patterns must match for the scenario to pass.
	LogPatterns []string

	// ExcludePatterns are regex patterns that MUST NOT appear in output.
	ExcludePatterns []string

	// ExitCode is the expected process exit code.
	// Defaults to 0 for "success", 1 for "failure".
	ExitCode *int

	// Outputs are key-value pairs expected in workflow outputs.
	Outputs map[string]string
}

// ExpectedStatus represents the expected workflow status.
type ExpectedStatus string

const (
	StatusSuccess ExpectedStatus = "success"
	StatusFailure ExpectedStatus = "failure"
)

// Info provides scenario metadata for listing.
type Info struct {
	ID             string
	Category       string
	Description    string
	ExpectedStatus string
	Tags           []string
	Disabled       bool
}

// Validate checks if the run result matches this scenario's expectations.
func (s *Scenario) Validate(result *runner.RunResult) (reporter.ResultStatus, []string, []string) {
	var matchedPatterns []string
	var missingPatterns []string

	// Check status (exit code based)
	actualSuccess := result.ExitCode == 0
	expectedSuccess := s.Expected.Status == StatusSuccess

	if actualSuccess != expectedSuccess {
		if expectedSuccess {
			return reporter.ResultFail, nil, []string{"expected success but workflow failed"}
		}
		return reporter.ResultFail, nil, []string{"expected failure but workflow succeeded"}
	}

	// Check specific exit code if provided
	if s.Expected.ExitCode != nil && result.ExitCode != *s.Expected.ExitCode {
		return reporter.ResultFail, nil, []string{"unexpected exit code"}
	}

	// Check log patterns
	for _, pattern := range s.Expected.LogPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			missingPatterns = append(missingPatterns, "invalid pattern: "+pattern)
			continue
		}

		if re.MatchString(result.Output) {
			matchedPatterns = append(matchedPatterns, pattern)
		} else {
			missingPatterns = append(missingPatterns, pattern)
		}
	}

	// Check exclude patterns
	for _, pattern := range s.Expected.ExcludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(result.Output) {
			missingPatterns = append(missingPatterns, "excluded pattern found: "+pattern)
		}
	}

	// Determine final status
	if len(missingPatterns) > 0 {
		return reporter.ResultFail, matchedPatterns, missingPatterns
	}

	return reporter.ResultPass, matchedPatterns, nil
}
