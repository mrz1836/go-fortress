package reporter_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestTerminalReporter_Write tests the Write method with various report configurations.
func TestTerminalReporter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		report         *reporter.Report
		expectContains []string
		expectAbsent   []string
	}{
		{
			name: "all checks passed with scenarios",
			report: &reporter.Report{
				Version:   "1.0.0",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(100 * time.Millisecond),
				Duration:  100 * time.Millisecond,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings:      []validator.Finding{},
					ValidatorsRun: []string{"actionlint"},
					Duration:      50 * time.Millisecond,
				},
				ScenarioResults: []reporter.ScenarioResult{
					{ScenarioID: "test-1", Status: reporter.ResultPass, Duration: 10 * time.Millisecond},
					{ScenarioID: "test-2", Status: reporter.ResultPass, Duration: 10 * time.Millisecond},
				},
				RunnerStatus: &reporter.RunnerStatus{
					Available:       true,
					RegisteredCount: 2,
					ExecutedCount:   2,
				},
				Summary: reporter.ReportSummary{
					TotalScenarios:  2,
					PassedScenarios: 2,
					TotalFindings:   0,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Guardian Report",
				"All checks passed!",
				"Static Validation",
				"Scenarios:",
				"2/2 passed",
			},
			expectAbsent: []string{
				"PARTIAL CHECK",
				"SKIPPED",
			},
		},
		{
			name: "runner unavailable - docker not running",
			report: &reporter.Report{
				Version:   "1.0.0",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(100 * time.Millisecond),
				Duration:  100 * time.Millisecond,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings:      []validator.Finding{},
					ValidatorsRun: []string{"actionlint"},
					Duration:      50 * time.Millisecond,
				},
				ScenarioResults: []reporter.ScenarioResult{},
				RunnerStatus: &reporter.RunnerStatus{
					Available:       false,
					UnavailableMsg:  "Docker is not running",
					RegisteredCount: 41,
					ExecutedCount:   0,
				},
				Summary: reporter.ReportSummary{
					TotalScenarios:  0,
					TotalFindings:   0,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Guardian Report",
				"SKIPPED",
				"Docker is not running",
				"41 scenarios cannot execute",
				"Start Docker Desktop",
				"docker info",
				"magex ci:verify",
				"magex ci:list",
				"0/41 executed",
				"PARTIAL CHECK - scenario tests skipped",
			},
			expectAbsent: []string{
				"All checks passed!",
			},
		},
		{
			name: "runner unavailable - generic error",
			report: &reporter.Report{
				Version:   "1.0.0",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(100 * time.Millisecond),
				Duration:  100 * time.Millisecond,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings:      []validator.Finding{},
					ValidatorsRun: []string{"actionlint"},
					Duration:      50 * time.Millisecond,
				},
				ScenarioResults: []reporter.ScenarioResult{},
				RunnerStatus: &reporter.RunnerStatus{
					Available:       false,
					UnavailableMsg:  "act binary not found",
					RegisteredCount: 10,
					ExecutedCount:   0,
				},
				Summary: reporter.ReportSummary{
					TotalScenarios:  0,
					TotalFindings:   0,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Guardian Report",
				"SKIPPED",
				"act binary not found",
				"10 scenarios cannot execute",
				"Resolve:",
				"0/10 executed",
				"PARTIAL CHECK",
			},
			expectAbsent: []string{
				"Start Docker Desktop",
				"All checks passed!",
			},
		},
		{
			name: "static warnings with runner unavailable",
			report: &reporter.Report{
				Version:   "1.0.0",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(100 * time.Millisecond),
				Duration:  100 * time.Millisecond,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							File:     "workflow.yml",
							Line:     10,
							Column:   5,
							Message:  "Warning message",
							Severity: validator.SeverityWarning,
						},
					},
					ValidatorsRun: []string{"actionlint"},
					Duration:      50 * time.Millisecond,
				},
				ScenarioResults: []reporter.ScenarioResult{},
				RunnerStatus: &reporter.RunnerStatus{
					Available:       false,
					UnavailableMsg:  "Docker daemon not responding",
					RegisteredCount: 5,
					ExecutedCount:   0,
				},
				Summary: reporter.ReportSummary{
					TotalScenarios:  0,
					TotalFindings:   1,
					FindingsByLevel: map[validator.Severity]int{validator.SeverityWarning: 1},
				},
			},
			expectContains: []string{
				"Guardian Report",
				"WARNINGS",
				"1 finding",
				"workflow.yml",
				"Warning message",
				"SKIPPED",
				"Docker",
				"0/5 executed",
				"PARTIAL CHECK",
				"1 warning",
			},
			expectAbsent: []string{
				"All checks passed!",
			},
		},
		{
			name: "failed scenarios - errors take precedence",
			report: &reporter.Report{
				Version:   "1.0.0",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(100 * time.Millisecond),
				Duration:  100 * time.Millisecond,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings:      []validator.Finding{},
					ValidatorsRun: []string{"actionlint"},
					Duration:      50 * time.Millisecond,
				},
				ScenarioResults: []reporter.ScenarioResult{
					{ScenarioID: "test-1", Status: reporter.ResultPass, Duration: 10 * time.Millisecond},
					{ScenarioID: "test-2", Status: reporter.ResultFail, Duration: 10 * time.Millisecond, Error: "expected pass"},
				},
				RunnerStatus: &reporter.RunnerStatus{
					Available:       true,
					RegisteredCount: 2,
					ExecutedCount:   2,
				},
				Summary: reporter.ReportSummary{
					TotalScenarios:  2,
					PassedScenarios: 1,
					FailedScenarios: 1,
					TotalFindings:   0,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Guardian Report",
				"Some checks failed",
				"FAILED",
			},
			expectAbsent: []string{
				"PARTIAL CHECK",
				"All checks passed!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := reporter.NewTerminalReporter()
			r.SetColorEnabled(false) // Disable colors for easier testing

			var buf bytes.Buffer
			err := r.Write(context.Background(), tt.report, &buf)
			require.NoError(t, err)

			output := buf.String()

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected,
					"expected output to contain %q", expected)
			}

			for _, absent := range tt.expectAbsent {
				assert.NotContains(t, output, absent,
					"expected output to NOT contain %q", absent)
			}
		})
	}
}

// TestTerminalReporter_WriteRunnerSkipped tests the runner skipped output specifically.
func TestTerminalReporter_WriteRunnerSkipped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		runnerStatus   *reporter.RunnerStatus
		expectContains []string
	}{
		{
			name: "docker not running shows specific instructions",
			runnerStatus: &reporter.RunnerStatus{
				Available:       false,
				UnavailableMsg:  "Cannot connect to the Docker daemon",
				RegisteredCount: 41,
				ExecutedCount:   0,
			},
			expectContains: []string{
				"SKIPPED",
				"Docker",
				"41 scenarios cannot execute",
				"Start Docker Desktop",
				"docker info",
				"magex ci:verify",
				"magex ci:list",
			},
		},
		{
			name: "generic error shows resolve instruction",
			runnerStatus: &reporter.RunnerStatus{
				Available:       false,
				UnavailableMsg:  "permission denied",
				RegisteredCount: 10,
				ExecutedCount:   0,
			},
			expectContains: []string{
				"SKIPPED",
				"permission denied",
				"10 scenarios cannot execute",
				"Resolve:",
				"magex ci:list",
			},
		},
		{
			name: "empty error message shows default",
			runnerStatus: &reporter.RunnerStatus{
				Available:       false,
				UnavailableMsg:  "",
				RegisteredCount: 5,
				ExecutedCount:   0,
			},
			expectContains: []string{
				"SKIPPED",
				"Runner not available",
				"5 scenarios cannot execute",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			report := &reporter.Report{
				Version:       "1.0.0",
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(100 * time.Millisecond),
				Duration:      100 * time.Millisecond,
				Mode:          reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{},
				RunnerStatus:  tt.runnerStatus,
				Summary:       reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
			}

			r := reporter.NewTerminalReporter()
			r.SetColorEnabled(false)

			var buf bytes.Buffer
			err := r.Write(context.Background(), report, &buf)
			require.NoError(t, err)

			output := buf.String()

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected,
					"expected output to contain %q", expected)
			}
		})
	}
}

// TestTerminalReporter_WriteSummary tests summary output formats.
func TestTerminalReporter_WriteSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		summary        reporter.ReportSummary
		runnerStatus   *reporter.RunnerStatus
		expectContains []string
	}{
		{
			name: "plural findings",
			summary: reporter.ReportSummary{
				TotalFindings:   3,
				FindingsByLevel: map[validator.Severity]int{validator.SeverityWarning: 3},
			},
			runnerStatus: &reporter.RunnerStatus{Available: true, RegisteredCount: 0, ExecutedCount: 0},
			expectContains: []string{
				"3 findings",
				"3 warnings",
			},
		},
		{
			name: "singular finding",
			summary: reporter.ReportSummary{
				TotalFindings:   1,
				FindingsByLevel: map[validator.Severity]int{validator.SeverityError: 1},
			},
			runnerStatus: &reporter.RunnerStatus{Available: true, RegisteredCount: 0, ExecutedCount: 0},
			expectContains: []string{
				"1 finding",
				"1 error",
			},
		},
		{
			name: "mixed errors and warnings",
			summary: reporter.ReportSummary{
				TotalFindings:   5,
				FindingsByLevel: map[validator.Severity]int{validator.SeverityError: 2, validator.SeverityWarning: 3},
			},
			runnerStatus: &reporter.RunnerStatus{Available: true, RegisteredCount: 0, ExecutedCount: 0},
			expectContains: []string{
				"5 findings",
				"2 errors",
				"3 warnings",
			},
		},
		{
			name: "scenarios skipped format",
			summary: reporter.ReportSummary{
				TotalFindings:   0,
				FindingsByLevel: map[validator.Severity]int{},
			},
			runnerStatus: &reporter.RunnerStatus{
				Available:       false,
				UnavailableMsg:  "Docker not found",
				RegisteredCount: 41,
				ExecutedCount:   0,
			},
			expectContains: []string{
				"0/41 executed",
				"Docker not running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			report := &reporter.Report{
				Version:       "1.0.0",
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(100 * time.Millisecond),
				Duration:      100 * time.Millisecond,
				Mode:          reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{},
				RunnerStatus:  tt.runnerStatus,
				Summary:       tt.summary,
			}

			r := reporter.NewTerminalReporter()
			r.SetColorEnabled(false)

			var buf bytes.Buffer
			err := r.Write(context.Background(), report, &buf)
			require.NoError(t, err)

			output := buf.String()

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected,
					"expected output to contain %q", expected)
			}
		})
	}
}

// TestTerminalReporter_Name tests the reporter name.
func TestTerminalReporter_Name(t *testing.T) {
	t.Parallel()

	r := reporter.NewTerminalReporter()
	assert.Equal(t, "terminal", r.Name())
}

// TestTerminalReporter_WriteFile tests writing to a file.
func TestTerminalReporter_WriteFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "output.txt")

	report := &reporter.Report{
		Version:       "1.0.0",
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(1 * time.Second),
		Duration:      1 * time.Second,
		Mode:          reporter.ModeVerify,
		StaticResults: &reporter.StaticResults{},
		RunnerStatus:  &reporter.RunnerStatus{Available: true},
		Summary:       reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
	}

	r := reporter.NewTerminalReporter()
	r.SetColorEnabled(false)
	err := r.WriteFile(context.Background(), report, outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	assert.Contains(t, string(data), "Guardian Report")
}

// TestTerminalReporter_WriteFile_Error tests write error handling.
func TestTerminalReporter_WriteFile_Error(t *testing.T) {
	t.Parallel()

	report := &reporter.Report{
		Version: "1.0.0",
		Summary: reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
	}

	r := reporter.NewTerminalReporter()
	err := r.WriteFile(context.Background(), report, "/nonexistent/path/output.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}

// TestRunnerStatus_Struct tests the RunnerStatus struct fields.
func TestRunnerStatus_Struct(t *testing.T) {
	t.Parallel()

	status := &reporter.RunnerStatus{
		Available:       false,
		UnavailableMsg:  "Docker is not running",
		RegisteredCount: 41,
		ExecutedCount:   0,
	}

	assert.False(t, status.Available)
	assert.Equal(t, "Docker is not running", status.UnavailableMsg)
	assert.Equal(t, 41, status.RegisteredCount)
	assert.Equal(t, 0, status.ExecutedCount)
}

// TestTerminalReporter_FinalStatus tests the final status line logic.
func TestTerminalReporter_FinalStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hasErrors    bool
		hasSkipped   bool
		expectStatus string
	}{
		{
			name:         "all passed",
			hasErrors:    false,
			hasSkipped:   false,
			expectStatus: "All checks passed!",
		},
		{
			name:         "partial check - scenarios skipped",
			hasErrors:    false,
			hasSkipped:   true,
			expectStatus: "PARTIAL CHECK - scenario tests skipped",
		},
		{
			name:         "errors take precedence over skipped",
			hasErrors:    true,
			hasSkipped:   true,
			expectStatus: "Some checks failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			summary := reporter.ReportSummary{
				FindingsByLevel: map[validator.Severity]int{},
			}
			if tt.hasErrors {
				summary.FailedScenarios = 1
			}

			var runnerStatus *reporter.RunnerStatus
			if tt.hasSkipped {
				runnerStatus = &reporter.RunnerStatus{
					Available:       false,
					UnavailableMsg:  "Docker not running",
					RegisteredCount: 10,
					ExecutedCount:   0,
				}
			} else {
				runnerStatus = &reporter.RunnerStatus{
					Available:       true,
					RegisteredCount: 5,
					ExecutedCount:   5,
				}
			}

			report := &reporter.Report{
				Version:       "1.0.0",
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(100 * time.Millisecond),
				Duration:      100 * time.Millisecond,
				Mode:          reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{},
				RunnerStatus:  runnerStatus,
				Summary:       summary,
			}

			r := reporter.NewTerminalReporter()
			r.SetColorEnabled(false)

			var buf bytes.Buffer
			err := r.Write(context.Background(), report, &buf)
			require.NoError(t, err)

			output := buf.String()

			// Check that the expected status appears in the output
			lines := strings.Split(output, "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, tt.expectStatus) {
					found = true
					break
				}
			}
			assert.True(t, found, "expected final status %q not found in output:\n%s", tt.expectStatus, output)
		})
	}
}
