package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// SummaryReporter generates GitHub Step Summary in markdown format.
type SummaryReporter struct{}

// NewSummaryReporter creates a new summary reporter.
func NewSummaryReporter() *SummaryReporter {
	return &SummaryReporter{}
}

// Name returns the reporter identifier.
func (r *SummaryReporter) Name() string {
	return "summary"
}

// Write outputs the report as a GitHub Step Summary.
func (r *SummaryReporter) Write(_ context.Context, report *Report, w io.Writer) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# Guardian CI Validation Report\n\n")

	// Overall status
	passed := report.Summary.FailedScenarios == 0 && report.Summary.ErrorScenarios == 0
	if passed {
		sb.WriteString("**Status**: :white_check_mark: PASSED\n\n")
	} else {
		sb.WriteString("**Status**: :x: FAILED\n\n")
	}

	// Duration
	sb.WriteString(fmt.Sprintf("**Duration**: %s\n\n", report.Duration))

	// Static Validation Summary
	if report.StaticResults != nil {
		sb.WriteString("## Static Validation\n\n")

		if len(report.StaticResults.Findings) == 0 {
			sb.WriteString(":white_check_mark: No issues found\n\n")
		} else {
			sb.WriteString("| Severity | Count |\n")
			sb.WriteString("|----------|-------|\n")

			for severity, count := range report.Summary.FindingsByLevel {
				icon := severityIcon(severity)
				sb.WriteString(fmt.Sprintf("| %s %s | %d |\n", icon, severity, count))
			}
			sb.WriteString("\n")

			// Findings table (limit to first 10)
			sb.WriteString("### Findings\n\n")
			sb.WriteString("| File | Line | Severity | Message |\n")
			sb.WriteString("|------|------|----------|----------|\n")

			limit := 10
			for i, f := range report.StaticResults.Findings {
				if i >= limit {
					sb.WriteString(fmt.Sprintf("\n*...and %d more findings*\n",
						len(report.StaticResults.Findings)-limit))
					break
				}
				sb.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s |\n",
					f.File, f.Line, f.Severity, truncateString(f.Message, 50)))
			}
			sb.WriteString("\n")
		}
	}

	// Scenario Results Summary
	if len(report.ScenarioResults) > 0 {
		sb.WriteString("## Scenario Results\n\n")

		sb.WriteString("| Metric | Count |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| :white_check_mark: Passed | %d |\n", report.Summary.PassedScenarios))
		sb.WriteString(fmt.Sprintf("| :x: Failed | %d |\n", report.Summary.FailedScenarios))
		sb.WriteString(fmt.Sprintf("| :warning: Errors | %d |\n", report.Summary.ErrorScenarios))
		sb.WriteString(fmt.Sprintf("| :fast_forward: Skipped | %d |\n", report.Summary.SkippedScenarios))
		sb.WriteString(fmt.Sprintf("| **Total** | **%d** |\n", report.Summary.TotalScenarios))
		sb.WriteString("\n")

		// Show failed/error scenarios details
		var failedScenarios []ScenarioResult
		for _, r := range report.ScenarioResults {
			if r.Status == ResultFail || r.Status == ResultError {
				failedScenarios = append(failedScenarios, r)
			}
		}

		if len(failedScenarios) > 0 {
			sb.WriteString("### Failed Scenarios\n\n")
			sb.WriteString("| Scenario | Status | Duration | Error |\n")
			sb.WriteString("|----------|--------|----------|-------|\n")

			for _, r := range failedScenarios {
				errMsg := r.Error
				if len(r.MissingPatterns) > 0 {
					errMsg = fmt.Sprintf("Missing patterns: %v", r.MissingPatterns)
				}
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
					r.ScenarioID, r.Status, r.Duration, truncateString(errMsg, 40)))
			}
			sb.WriteString("\n")
		}
	}

	// Write to output
	_, err := w.Write([]byte(sb.String()))
	return err
}

// WriteFile outputs the report to a file.
func (r *SummaryReporter) WriteFile(ctx context.Context, report *Report, path string) error {
	f, err := os.Create(path) //nolint:gosec // path from trusted caller
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	defer func() { _ = f.Close() }()

	return r.Write(ctx, report, f)
}

// WriteToStepSummary writes the summary to GitHub Step Summary file.
// Returns nil if not running in GitHub Actions.
func (r *SummaryReporter) WriteToStepSummary(ctx context.Context, report *Report) error {
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		// Not in GitHub Actions
		return nil
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // summary file for GitHub Actions
	if err != nil {
		return fmt.Errorf("opening step summary file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return r.Write(ctx, report, f)
}

// severityIcon returns an emoji icon for the severity level.
func severityIcon(s validator.Severity) string {
	switch s {
	case validator.SeverityError:
		return ":x:"
	case validator.SeverityWarning:
		return ":warning:"
	case validator.SeverityNote:
		return ":information_source:"
	case validator.SeverityInfo:
		return ":grey_question:"
	}
	return ":grey_question:"
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
