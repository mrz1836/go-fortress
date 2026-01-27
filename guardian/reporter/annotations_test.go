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

// TestAnnotationsReporter_Name tests the reporter name.
func TestAnnotationsReporter_Name(t *testing.T) {
	t.Parallel()

	r := reporter.NewAnnotationsReporter()
	assert.Equal(t, "annotations", r.Name())
}

// TestAnnotationsReporter_Write tests writing annotation output.
func TestAnnotationsReporter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		report         *reporter.Report
		expectContains []string
		expectAbsent   []string
	}{
		{
			name: "error severity",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test/error",
							Severity: validator.SeverityError,
							Message:  "error message",
							File:     "test.yml",
							Line:     10,
							Column:   5,
						},
					},
				},
			},
			expectContains: []string{
				"::error file=test.yml",
				"line=10",
				"col=5",
				"title=test/error",
				"::error message",
			},
		},
		{
			name: "warning severity",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test/warning",
							Severity: validator.SeverityWarning,
							Message:  "warning message",
							File:     "test.yml",
							Line:     20,
						},
					},
				},
			},
			expectContains: []string{
				"::warning file=test.yml",
				"line=20",
				"title=test/warning",
				"::warning message",
			},
			expectAbsent: []string{
				"col=",
			},
		},
		{
			name: "note severity",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test/note",
							Severity: validator.SeverityNote,
							Message:  "note message",
							File:     "test.yml",
							Line:     30,
						},
					},
				},
			},
			expectContains: []string{
				"::notice file=test.yml",
			},
		},
		{
			name: "info severity",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test/info",
							Severity: validator.SeverityInfo,
							Message:  "info message",
							File:     "test.yml",
							Line:     40,
						},
					},
				},
			},
			expectContains: []string{
				"::notice file=test.yml",
			},
		},
		{
			name: "with suggestion",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:     "test/rule",
							Severity:   validator.SeverityWarning,
							Message:    "message",
							File:       "test.yml",
							Line:       1,
							Suggestion: "fix it this way",
						},
					},
				},
			},
			expectContains: []string{
				"(Suggestion: fix it this way)",
			},
		},
		{
			name: "with end column",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:    "test/rule",
							Severity:  validator.SeverityError,
							Message:   "message",
							File:      "test.yml",
							Line:      1,
							Column:    5,
							EndColumn: 20,
						},
					},
				},
			},
			expectContains: []string{
				"endColumn=20",
			},
		},
		{
			name: "message with special characters",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test/rule",
							Severity: validator.SeverityWarning,
							Message:  "line1\nline2\rline3 with 100%",
							File:     "test.yml",
							Line:     1,
						},
					},
				},
			},
			expectContains: []string{
				"%0A", // newline escaped
				"%0D", // carriage return escaped
				"%25", // percent escaped
			},
		},
		{
			name: "no static results",
			report: &reporter.Report{
				Version:       "1.0.0",
				StaticResults: nil,
			},
			expectContains: []string{},
		},
		{
			name: "empty findings",
			report: &reporter.Report{
				Version: "1.0.0",
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{},
				},
			},
			expectContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := reporter.NewAnnotationsReporter()
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

// TestAnnotationsReporter_WriteAnnotations tests the WriteAnnotations method directly.
func TestAnnotationsReporter_WriteAnnotations(t *testing.T) {
	t.Parallel()

	findings := []validator.Finding{
		{
			RuleID:   "rule1",
			Severity: validator.SeverityError,
			Message:  "error 1",
			File:     "file1.yml",
			Line:     10,
		},
		{
			RuleID:   "rule2",
			Severity: validator.SeverityWarning,
			Message:  "warning 1",
			File:     "file2.yml",
			Line:     20,
		},
	}

	r := reporter.NewAnnotationsReporter()
	var buf bytes.Buffer
	err := r.WriteAnnotations(context.Background(), findings, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "::error file=file1.yml")
	assert.Contains(t, output, "::warning file=file2.yml")
}

// TestAnnotationsReporter_WriteFile tests writing to a file.
func TestAnnotationsReporter_WriteFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "annotations.txt")

	report := &reporter.Report{
		Version:  "1.0.0",
		Duration: 1 * time.Second,
		StaticResults: &reporter.StaticResults{
			Findings: []validator.Finding{
				{
					RuleID:   "test",
					Severity: validator.SeverityError,
					Message:  "error",
					File:     "test.yml",
					Line:     1,
				},
			},
		},
	}

	r := reporter.NewAnnotationsReporter()
	err := r.WriteFile(context.Background(), report, outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	assert.Contains(t, string(data), "::error")
}

// TestAnnotationsReporter_WriteFile_Error tests write error handling.
func TestAnnotationsReporter_WriteFile_Error(t *testing.T) {
	t.Parallel()

	report := &reporter.Report{
		Version: "1.0.0",
	}

	r := reporter.NewAnnotationsReporter()
	err := r.WriteFile(context.Background(), report, "/nonexistent/path/annotations.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}
