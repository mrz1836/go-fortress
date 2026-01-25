package reporter_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestSummaryReporter_Name tests the reporter name.
func TestSummaryReporter_Name(t *testing.T) {
	t.Parallel()

	r := reporter.NewSummaryReporter()
	assert.Equal(t, "summary", r.Name())
}

// TestSummaryReporter_Write tests writing summary output.
func TestSummaryReporter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		report         *reporter.Report
		expectContains []string
	}{
		{
			name: "passed report",
			report: &reporter.Report{
				Version:  "1.0.0",
				Duration: 5 * time.Second,
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{},
				},
				ScenarioResults: []reporter.ScenarioResult{
					{ScenarioID: "TEST-001", Status: reporter.ResultPass, Duration: 1 * time.Second},
				},
				Summary: reporter.ReportSummary{
					PassedScenarios: 1,
					TotalScenarios:  1,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Guardian CI Validation Report",
				"PASSED",
				"No issues found",
				"Scenario Results",
				"Passed | 1",
			},
		},
		{
			name: "failed report with findings",
			report: &reporter.Report{
				Version:  "1.0.0",
				Duration: 10 * time.Second,
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							File:     "test.yml",
							Line:     10,
							Severity: validator.SeverityError,
							Message:  "test error message",
						},
						{
							File:     "test.yml",
							Line:     20,
							Severity: validator.SeverityWarning,
							Message:  "test warning message",
						},
					},
				},
				ScenarioResults: []reporter.ScenarioResult{
					{ScenarioID: "TEST-001", Status: reporter.ResultFail, Duration: 1 * time.Second, Error: "expected pass"},
				},
				Summary: reporter.ReportSummary{
					FailedScenarios: 1,
					TotalScenarios:  1,
					FindingsByLevel: map[validator.Severity]int{
						validator.SeverityError:   1,
						validator.SeverityWarning: 1,
					},
				},
			},
			expectContains: []string{
				"FAILED",
				"Severity | Count",
				"Findings",
				"`test.yml`",
				"Failed Scenarios",
				"expected pass",
			},
		},
		{
			name: "report with many findings truncates",
			report: &reporter.Report{
				Version:  "1.0.0",
				Duration: 5 * time.Second,
				StaticResults: &reporter.StaticResults{
					Findings: func() []validator.Finding {
						findings := make([]validator.Finding, 15)
						for i := range findings {
							findings[i] = validator.Finding{
								File:     "test.yml",
								Line:     i + 1,
								Severity: validator.SeverityWarning,
								Message:  "warning",
							}
						}
						return findings
					}(),
				},
				Summary: reporter.ReportSummary{
					FindingsByLevel: map[validator.Severity]int{
						validator.SeverityWarning: 15,
					},
				},
			},
			expectContains: []string{
				"...and 5 more findings",
			},
		},
		{
			name: "report with missing patterns shows them",
			report: &reporter.Report{
				Version:  "1.0.0",
				Duration: 5 * time.Second,
				ScenarioResults: []reporter.ScenarioResult{
					{
						ScenarioID:      "TEST-001",
						Status:          reporter.ResultFail,
						Duration:        1 * time.Second,
						MissingPatterns: []string{"expected_pattern"},
					},
				},
				Summary: reporter.ReportSummary{
					FailedScenarios: 1,
					TotalScenarios:  1,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"Missing patterns",
			},
		},
		{
			name: "report with error scenarios",
			report: &reporter.Report{
				Version:  "1.0.0",
				Duration: 5 * time.Second,
				ScenarioResults: []reporter.ScenarioResult{
					{
						ScenarioID: "TEST-001",
						Status:     reporter.ResultError,
						Duration:   1 * time.Second,
						Error:      "timeout",
					},
				},
				Summary: reporter.ReportSummary{
					ErrorScenarios:  1,
					TotalScenarios:  1,
					FindingsByLevel: map[validator.Severity]int{},
				},
			},
			expectContains: []string{
				"FAILED",
				"Errors | 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := reporter.NewSummaryReporter()
			var buf bytes.Buffer
			err := r.Write(context.Background(), tt.report, &buf)
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected,
					"expected output to contain %q", expected)
			}
		})
	}
}

// TestSummaryReporter_WriteFile tests writing to a file.
func TestSummaryReporter_WriteFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "summary.md")

	report := &reporter.Report{
		Version:  "1.0.0",
		Duration: 1 * time.Second,
		Summary:  reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
	}

	r := reporter.NewSummaryReporter()
	err := r.WriteFile(context.Background(), report, outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	assert.Contains(t, string(data), "Guardian CI Validation Report")
}

// TestSummaryReporter_WriteFile_Error tests write error handling.
func TestSummaryReporter_WriteFile_Error(t *testing.T) {
	t.Parallel()

	report := &reporter.Report{
		Version: "1.0.0",
	}

	r := reporter.NewSummaryReporter()
	err := r.WriteFile(context.Background(), report, "/nonexistent/path/summary.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}

// TestSummaryReporter_WriteToStepSummary_NotInGitHubActions tests behavior outside GitHub Actions.
func TestSummaryReporter_WriteToStepSummary_NotInGitHubActions(t *testing.T) {
	// Clear the env var if set
	originalVal := os.Getenv("GITHUB_STEP_SUMMARY")
	_ = os.Unsetenv("GITHUB_STEP_SUMMARY")
	defer func() {
		if originalVal != "" {
			_ = os.Setenv("GITHUB_STEP_SUMMARY", originalVal)
		}
	}()

	report := &reporter.Report{
		Version: "1.0.0",
		Summary: reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
	}

	r := reporter.NewSummaryReporter()
	err := r.WriteToStepSummary(context.Background(), report)
	require.NoError(t, err) // Should return nil when not in GitHub Actions
}

// TestSummaryReporter_WriteToStepSummary_InGitHubActions tests behavior in GitHub Actions.
func TestSummaryReporter_WriteToStepSummary_InGitHubActions(t *testing.T) {
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")

	// Create the file first
	f, err := os.Create(summaryFile) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Set the env var
	originalVal := os.Getenv("GITHUB_STEP_SUMMARY")
	require.NoError(t, os.Setenv("GITHUB_STEP_SUMMARY", summaryFile))
	defer func() {
		if originalVal == "" {
			_ = os.Unsetenv("GITHUB_STEP_SUMMARY")
		} else {
			_ = os.Setenv("GITHUB_STEP_SUMMARY", originalVal)
		}
	}()

	report := &reporter.Report{
		Version:  "1.0.0",
		Duration: 1 * time.Second,
		Summary:  reporter.ReportSummary{FindingsByLevel: map[validator.Severity]int{}},
	}

	r := reporter.NewSummaryReporter()
	err = r.WriteToStepSummary(context.Background(), report)
	require.NoError(t, err)

	// Verify content was written
	data, err := os.ReadFile(summaryFile) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	assert.Contains(t, string(data), "Guardian CI Validation Report")
}
