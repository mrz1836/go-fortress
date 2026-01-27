package scenarios_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/runner"
	"github.com/mrz1836/go-fortress/guardian/scenarios"
)

// TestScenario_Validate_Success tests scenarios that expect success.
func TestScenario_Validate_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scenario     *scenarios.Scenario
		result       *runner.RunResult
		expectStatus reporter.ResultStatus
	}{
		{
			name: "success scenario with exit code 0",
			scenario: &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status: scenarios.StatusSuccess,
				},
			},
			result: &runner.RunResult{
				ExitCode: 0,
				Output:   "Build succeeded",
			},
			expectStatus: reporter.ResultPass,
		},
		{
			name: "success scenario fails with non-zero exit",
			scenario: &scenarios.Scenario{
				ID: "TEST-002",
				Expected: scenarios.ExpectedResult{
					Status: scenarios.StatusSuccess,
				},
			},
			result: &runner.RunResult{
				ExitCode: 1,
				Output:   "Build failed",
			},
			expectStatus: reporter.ResultFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, _, _ := tt.scenario.Validate(tt.result)
			assert.Equal(t, tt.expectStatus, status)
		})
	}
}

// TestScenario_Validate_Failure tests scenarios that expect failure.
func TestScenario_Validate_Failure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scenario     *scenarios.Scenario
		result       *runner.RunResult
		expectStatus reporter.ResultStatus
	}{
		{
			name: "failure scenario with non-zero exit code",
			scenario: &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status: scenarios.StatusFailure,
				},
			},
			result: &runner.RunResult{
				ExitCode: 1,
				Output:   "Lint errors found",
			},
			expectStatus: reporter.ResultPass,
		},
		{
			name: "failure scenario fails with exit code 0",
			scenario: &scenarios.Scenario{
				ID: "TEST-002",
				Expected: scenarios.ExpectedResult{
					Status: scenarios.StatusFailure,
				},
			},
			result: &runner.RunResult{
				ExitCode: 0,
				Output:   "No errors",
			},
			expectStatus: reporter.ResultFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, _, _ := tt.scenario.Validate(tt.result)
			assert.Equal(t, tt.expectStatus, status)
		})
	}
}

// TestScenario_Validate_ExitCode tests specific exit code validation.
func TestScenario_Validate_ExitCode(t *testing.T) {
	t.Parallel()

	exitCode0 := 0
	exitCode1 := 1
	exitCode2 := 2

	tests := []struct {
		name         string
		expected     *int
		actual       int
		expectStatus reporter.ResultStatus
	}{
		{
			name:         "explicit exit code 0 matches",
			expected:     &exitCode0,
			actual:       0,
			expectStatus: reporter.ResultPass,
		},
		{
			name:         "explicit exit code 1 matches",
			expected:     &exitCode1,
			actual:       1,
			expectStatus: reporter.ResultPass,
		},
		{
			name:         "explicit exit code mismatch fails",
			expected:     &exitCode2,
			actual:       1,
			expectStatus: reporter.ResultFail,
		},
		{
			name:         "no explicit exit code uses status check",
			expected:     nil,
			actual:       0,
			expectStatus: reporter.ResultPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scenario := &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status:   scenarios.StatusSuccess,
					ExitCode: tt.expected,
				},
			}

			// Adjust expected status based on exit code
			if tt.expected == nil || *tt.expected == 0 {
				scenario.Expected.Status = scenarios.StatusSuccess
			} else {
				scenario.Expected.Status = scenarios.StatusFailure
			}

			result := &runner.RunResult{
				ExitCode: tt.actual,
				Output:   "test output",
			}

			status, _, _ := scenario.Validate(result)
			assert.Equal(t, tt.expectStatus, status)
		})
	}
}

// TestScenario_Validate_LogPatterns tests pattern matching in output.
func TestScenario_Validate_LogPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		patterns      []string
		output        string
		expectStatus  reporter.ResultStatus
		expectMatched []string
		expectMissing []string
	}{
		{
			name:          "all patterns match",
			patterns:      []string{"error", "line \\d+", "Build"},
			output:        "error on line 42: Build failed",
			expectStatus:  reporter.ResultPass,
			expectMatched: []string{"error", "line \\d+", "Build"},
		},
		{
			name:          "some patterns missing",
			patterns:      []string{"error", "warning", "fatal"},
			output:        "error: something went wrong",
			expectStatus:  reporter.ResultFail,
			expectMatched: []string{"error"},
			expectMissing: []string{"warning", "fatal"},
		},
		{
			name:          "regex pattern matches",
			patterns:      []string{"^Starting", "Completed in \\d+\\.\\d+s$"},
			output:        "Starting job\nProcessing...\nCompleted in 1.5s",
			expectStatus:  reporter.ResultPass,
			expectMatched: []string{"^Starting", "Completed in \\d+\\.\\d+s$"},
		},
		{
			name:          "no patterns always passes",
			patterns:      []string{},
			output:        "any output",
			expectStatus:  reporter.ResultPass,
			expectMatched: nil,
		},
		{
			name:          "invalid regex reports error",
			patterns:      []string{"[invalid"},
			output:        "any output",
			expectStatus:  reporter.ResultFail,
			expectMissing: []string{"invalid pattern: [invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scenario := &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status:      scenarios.StatusSuccess,
					LogPatterns: tt.patterns,
				},
			}

			result := &runner.RunResult{
				ExitCode: 0,
				Output:   tt.output,
			}

			status, matched, missing := scenario.Validate(result)
			assert.Equal(t, tt.expectStatus, status)

			if tt.expectMatched != nil {
				assert.Equal(t, tt.expectMatched, matched)
			}
			if tt.expectMissing != nil {
				assert.Equal(t, tt.expectMissing, missing)
			}
		})
	}
}

// TestScenario_Validate_ExcludePatterns tests exclude pattern detection.
func TestScenario_Validate_ExcludePatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		excludes      []string
		output        string
		expectStatus  reporter.ResultStatus
		expectMissing []string
	}{
		{
			name:         "no excluded patterns present passes",
			excludes:     []string{"FATAL", "panic"},
			output:       "normal log output",
			expectStatus: reporter.ResultPass,
		},
		{
			name:          "excluded pattern detected fails",
			excludes:      []string{"secret", "password"},
			output:        "exposing secret: abc123",
			expectStatus:  reporter.ResultFail,
			expectMissing: []string{"excluded pattern found: secret"},
		},
		{
			name:          "multiple excluded patterns detected",
			excludes:      []string{"error", "warning"},
			output:        "error occurred\nwarning: deprecated",
			expectStatus:  reporter.ResultFail,
			expectMissing: []string{"excluded pattern found: error", "excluded pattern found: warning"},
		},
		{
			name:         "regex exclude pattern",
			excludes:     []string{"token=[a-z0-9]+"},
			output:       "setting token=abc123xyz",
			expectStatus: reporter.ResultFail,
		},
		{
			name:         "empty excludes always passes",
			excludes:     []string{},
			output:       "any output with error and warning",
			expectStatus: reporter.ResultPass,
		},
		{
			name:         "invalid regex is skipped",
			excludes:     []string{"[invalid"},
			output:       "any output",
			expectStatus: reporter.ResultPass, // Invalid regex is skipped silently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scenario := &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status:          scenarios.StatusSuccess,
					ExcludePatterns: tt.excludes,
				},
			}

			result := &runner.RunResult{
				ExitCode: 0,
				Output:   tt.output,
			}

			status, _, missing := scenario.Validate(result)
			assert.Equal(t, tt.expectStatus, status)

			if tt.expectMissing != nil {
				assert.Equal(t, tt.expectMissing, missing)
			}
		})
	}
}

// TestScenario_Validate_Combined tests combined validation logic.
func TestScenario_Validate_Combined(t *testing.T) {
	t.Parallel()

	exitCode1 := 1

	tests := []struct {
		name         string
		scenario     *scenarios.Scenario
		result       *runner.RunResult
		expectStatus reporter.ResultStatus
	}{
		{
			name: "all conditions pass",
			scenario: &scenarios.Scenario{
				ID: "TEST-001",
				Expected: scenarios.ExpectedResult{
					Status:          scenarios.StatusFailure,
					ExitCode:        &exitCode1,
					LogPatterns:     []string{"error:", "line \\d+"},
					ExcludePatterns: []string{"panic", "FATAL"},
				},
			},
			result: &runner.RunResult{
				ExitCode: 1,
				Output:   "error: syntax error on line 42",
			},
			expectStatus: reporter.ResultPass,
		},
		{
			name: "status fails overrides everything",
			scenario: &scenarios.Scenario{
				ID: "TEST-002",
				Expected: scenarios.ExpectedResult{
					Status:      scenarios.StatusSuccess,
					LogPatterns: []string{"success"},
				},
			},
			result: &runner.RunResult{
				ExitCode: 1,
				Output:   "success message but exit code is wrong",
			},
			expectStatus: reporter.ResultFail,
		},
		{
			name: "exit code mismatch fails even with matching patterns",
			scenario: &scenarios.Scenario{
				ID: "TEST-003",
				Expected: scenarios.ExpectedResult{
					Status:      scenarios.StatusFailure,
					ExitCode:    &exitCode1,
					LogPatterns: []string{"error"},
				},
			},
			result: &runner.RunResult{
				ExitCode: 2, // Wrong exit code
				Output:   "error occurred",
			},
			expectStatus: reporter.ResultFail,
		},
		{
			name: "exclude pattern trumps log patterns",
			scenario: &scenarios.Scenario{
				ID: "TEST-004",
				Expected: scenarios.ExpectedResult{
					Status:          scenarios.StatusSuccess,
					LogPatterns:     []string{"completed"},
					ExcludePatterns: []string{"leaked"},
				},
			},
			result: &runner.RunResult{
				ExitCode: 0,
				Output:   "completed but leaked secret",
			},
			expectStatus: reporter.ResultFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, _, _ := tt.scenario.Validate(tt.result)
			assert.Equal(t, tt.expectStatus, status)
		})
	}
}
