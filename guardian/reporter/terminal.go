package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TerminalReporter writes human-readable output to terminal.
type TerminalReporter struct {
	colorEnabled bool
}

// ScenarioListItem represents a scenario for list output.
type ScenarioListItem struct {
	ID             string
	Category       string
	Description    string
	ExpectedStatus string
	Tags           []string
	Disabled       bool
}

// NewTerminalReporter creates a new terminal reporter.
func NewTerminalReporter() *TerminalReporter {
	return &TerminalReporter{
		colorEnabled: isTerminal(),
	}
}

// Name returns the reporter identifier.
func (r *TerminalReporter) Name() string {
	return "terminal"
}

// SetColorEnabled enables/disables color output.
func (r *TerminalReporter) SetColorEnabled(enabled bool) {
	r.colorEnabled = enabled
}

// Write outputs the report to the given writer.
func (r *TerminalReporter) Write(_ context.Context, report *Report, w io.Writer) error {
	// Header
	_, _ = fmt.Fprintf(w, "\n%s Guardian Report %s\n", r.icon("shield"), r.dim(report.Mode)) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	_, _ = fmt.Fprintf(w, "%s\n\n", strings.Repeat("-", 50))

	// Static results
	if report.StaticResults != nil {
		r.writeStaticResults(w, report.StaticResults)
	}

	// Scenario results or skipped notice
	if len(report.ScenarioResults) > 0 {
		r.writeScenarioResults(w, report.ScenarioResults)
	} else if report.RunnerStatus != nil && !report.RunnerStatus.Available {
		r.writeRunnerSkipped(w, report.RunnerStatus)
	}

	// Summary
	r.writeSummary(w, &report.Summary, report.Duration, report.RunnerStatus)

	return nil
}

// WriteFile outputs the report to a file.
func (r *TerminalReporter) WriteFile(ctx context.Context, report *Report, path string) error {
	f, err := os.Create(path) //nolint:gosec // path from trusted caller
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	defer func() { _ = f.Close() }()

	// Disable colors for file output
	origColor := r.colorEnabled
	r.colorEnabled = false

	defer func() { r.colorEnabled = origColor }()

	return r.Write(ctx, report, f)
}

// WriteScenarioList outputs a formatted list of scenarios.
func (r *TerminalReporter) WriteScenarioList(w io.Writer, scenarios []ScenarioListItem, filter string) {
	// Group by category
	byCategory := make(map[string][]ScenarioListItem)
	categoryOrder := []string{}

	for _, s := range scenarios {
		if _, exists := byCategory[s.Category]; !exists {
			categoryOrder = append(categoryOrder, s.Category)
		}

		byCategory[s.Category] = append(byCategory[s.Category], s)
	}

	// Count totals
	totalEnabled := 0

	for _, s := range scenarios {
		if !s.Disabled {
			totalEnabled++
		}
	}

	// Header
	if filter != "" {
		_, _ = fmt.Fprintf(w, "\n%s Available CI Scenarios (filter: %s)\n",
			r.icon("shield"), r.bold(filter))
	} else {
		_, _ = fmt.Fprintf(w, "\n%s Available CI Scenarios (%d total)\n",
			r.icon("shield"), totalEnabled)
	}

	_, _ = fmt.Fprintf(w, "%s\n\n", strings.Repeat("-", 50))

	// Output by category
	for _, category := range categoryOrder {
		items := byCategory[category]
		enabledCount := 0

		for _, item := range items {
			if !item.Disabled {
				enabledCount++
			}
		}

		_, _ = fmt.Fprintf(w, "%s (%d):\n", r.bold(category), enabledCount)

		for _, s := range items {
			if s.Disabled {
				continue
			}

			// Format: ID - Description [tags]
			var statusIcon string

			if s.ExpectedStatus == "failure" {
				statusIcon = r.red("✗")
			} else {
				statusIcon = r.green("✓")
			}

			tagsStr := ""
			if len(s.Tags) > 0 {
				tagsStr = " " + r.dim("["+strings.Join(s.Tags, ", ")+"]")
			}

			_, _ = fmt.Fprintf(w, "  %s %s - %s%s\n",
				statusIcon,
				r.cyan(s.ID),
				s.Description,
				tagsStr)
		}

		_, _ = fmt.Fprintln(w)
	}

	// Footer
	_, _ = fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(w, "Run a specific scenario: %s\n",
		r.dim("magex ci:scenario name=<ID>"))
	_, _ = fmt.Fprintf(w, "Filter by category:      %s\n",
		r.dim("magex ci:list filter=<category>"))
}

// writeStaticResults formats static validation results.
func (r *TerminalReporter) writeStaticResults(w io.Writer, results *StaticResults) {
	if len(results.Findings) == 0 {
		_, _ = fmt.Fprintf(w, "%s Static Validation: %s (%s)\n\n",
			r.icon("check"),
			r.green("PASSED"),
			formatDuration(results.Duration))

		return
	}

	_, _ = fmt.Fprintf(w, "%s Static Validation: %s (%d findings)\n\n",
		r.icon("warning"),
		r.yellow("WARNINGS"),
		len(results.Findings))

	// Group findings by file
	byFile := make(map[string][]validator.Finding)

	for _, f := range results.Findings {
		byFile[f.File] = append(byFile[f.File], f)
	}

	for file, findings := range byFile {
		_, _ = fmt.Fprintf(w, "  %s:\n", r.bold(file))

		for _, f := range findings {
			severity := r.formatSeverity(f.Severity)
			_, _ = fmt.Fprintf(w, "    %s:%d:%d %s %s\n",
				file, f.Line, f.Column,
				severity,
				f.Message)

			if f.Suggestion != "" {
				_, _ = fmt.Fprintf(w, "      %s %s\n", r.dim("Suggestion:"), f.Suggestion)
			}
		}

		_, _ = fmt.Fprintln(w)
	}
}

// writeScenarioResults formats scenario execution results.
func (r *TerminalReporter) writeScenarioResults(w io.Writer, results []ScenarioResult) {
	passed := 0
	failed := 0

	for _, res := range results {
		if res.Status == ResultPass {
			passed++
		} else {
			failed++
		}
	}

	status := r.green("PASSED")
	icon := r.icon("check")

	if failed > 0 {
		status = r.red("FAILED")
		icon = r.icon("cross")
	}

	_, _ = fmt.Fprintf(w, "%s Scenarios: %s (%d/%d passed)\n\n",
		icon, status, passed, len(results))

	for _, res := range results {
		statusIcon := r.icon("check")
		statusText := r.green("PASS")

		if res.Status != ResultPass {
			statusIcon = r.icon("cross")
			statusText = r.red(string(res.Status))
		}

		_, _ = fmt.Fprintf(w, "  %s %s %s (%s)\n",
			statusIcon,
			res.ScenarioID,
			statusText,
			formatDuration(res.Duration))

		if res.Error != "" {
			_, _ = fmt.Fprintf(w, "      %s %s\n", r.red("Error:"), res.Error)
		}

		if len(res.MissingPatterns) > 0 {
			_, _ = fmt.Fprintf(w, "      %s %v\n", r.yellow("Missing:"), res.MissingPatterns)
		}
	}

	_, _ = fmt.Fprintln(w)
}

// writeRunnerSkipped outputs information when scenarios cannot execute.
func (r *TerminalReporter) writeRunnerSkipped(w io.Writer, status *RunnerStatus) {
	// Determine a user-friendly message
	reason := status.UnavailableMsg
	if reason == "" {
		reason = "Runner not available"
	}

	// Check if it's a Docker-related issue
	isDockerIssue := strings.Contains(strings.ToLower(reason), "docker")

	_, _ = fmt.Fprintf(w, "%s Scenario Tests: %s - %s\n",
		r.icon("warning"),
		r.yellow("SKIPPED"),
		reason)

	_, _ = fmt.Fprintln(w, "  │")
	_, _ = fmt.Fprintf(w, "  │  %d scenarios cannot execute without Docker\n", status.RegisteredCount)
	_, _ = fmt.Fprintln(w, "  │")

	if isDockerIssue {
		_, _ = fmt.Fprintln(w, "  │  To enable scenario testing:")
		_, _ = fmt.Fprintln(w, "  │    1. Start Docker Desktop")
		_, _ = fmt.Fprintf(w, "  │    2. Run: %s (to verify)\n", r.cyan("docker info"))
		_, _ = fmt.Fprintf(w, "  │    3. Re-run: %s\n", r.cyan("magex ci:verify"))
	} else {
		_, _ = fmt.Fprintln(w, "  │  To enable scenario testing:")
		_, _ = fmt.Fprintf(w, "  │    Resolve: %s\n", reason)
	}

	_, _ = fmt.Fprintln(w, "  │")
	_, _ = fmt.Fprintf(w, "  │  To see available scenarios: %s\n", r.cyan("magex ci:list"))
	_, _ = fmt.Fprintln(w, "  │")
	_, _ = fmt.Fprintln(w)
}

// writeSummary formats the report summary.
func (r *TerminalReporter) writeSummary(w io.Writer, summary *ReportSummary, duration time.Duration, runnerStatus *RunnerStatus) {
	_, _ = fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(w, "Summary:\n")
	_, _ = fmt.Fprintf(w, "  Duration: %s\n", formatDuration(duration)) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response

	r.writeStaticSummary(w, summary)
	scenariosSkipped := r.writeScenarioSummary(w, summary, runnerStatus)

	if summary.ExceptionsApplied > 0 {
		_, _ = fmt.Fprintf(w, "  Exceptions: %d applied\n", summary.ExceptionsApplied) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	}

	r.writeFinalStatus(w, summary, scenariosSkipped)
}

// writeStaticSummary writes the static validation findings summary.
func (r *TerminalReporter) writeStaticSummary(w io.Writer, summary *ReportSummary) {
	if summary.TotalFindings == 0 {
		_, _ = fmt.Fprintln(w, "  Static: 0 findings")
		return
	}

	_, _ = fmt.Fprintf(w, "  Static: %s", pluralize(summary.TotalFindings, "finding"))

	parts := r.buildFindingsParts(summary.FindingsByLevel)
	if len(parts) > 0 {
		_, _ = fmt.Fprintf(w, " (%s)", strings.Join(parts, ", ")) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	}

	_, _ = fmt.Fprintln(w)
}

// buildFindingsParts builds the breakdown of findings by severity.
func (r *TerminalReporter) buildFindingsParts(findingsByLevel map[validator.Severity]int) []string {
	var parts []string

	if n, ok := findingsByLevel[validator.SeverityError]; ok && n > 0 {
		parts = append(parts, pluralize(n, "error"))
	}

	if n, ok := findingsByLevel[validator.SeverityWarning]; ok && n > 0 {
		parts = append(parts, pluralize(n, "warning"))
	}

	return parts
}

// writeScenarioSummary writes the scenario execution summary. Returns true if scenarios were skipped.
func (r *TerminalReporter) writeScenarioSummary(w io.Writer, summary *ReportSummary, runnerStatus *RunnerStatus) bool {
	scenariosSkipped := runnerStatus != nil && !runnerStatus.Available

	if scenariosSkipped {
		reason := formatSkipReason(runnerStatus.UnavailableMsg)
		_, _ = fmt.Fprintf(w, "  Scenarios: %d/%d executed (%s)\n",
			runnerStatus.ExecutedCount,
			runnerStatus.RegisteredCount,
			reason)
	} else if summary.TotalScenarios > 0 {
		_, _ = fmt.Fprintf(w, "  Scenarios: %d passed, %d failed, %d skipped, %d errors\n",
			summary.PassedScenarios,
			summary.FailedScenarios,
			summary.SkippedScenarios,
			summary.ErrorScenarios)
	}

	return scenariosSkipped
}

// writeFinalStatus writes the final status line.
func (r *TerminalReporter) writeFinalStatus(w io.Writer, summary *ReportSummary, scenariosSkipped bool) {
	_, _ = fmt.Fprintln(w)

	hasErrors := summary.FailedScenarios > 0 || summary.ErrorScenarios > 0 ||
		summary.FindingsByLevel[validator.SeverityError] > 0

	switch {
	case hasErrors:
		_, _ = fmt.Fprintf(w, "%s %s\n", r.icon("cross"), r.red("Some checks failed")) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	case scenariosSkipped:
		_, _ = fmt.Fprintf(w, "%s %s\n", r.icon("warning"), r.yellow("PARTIAL CHECK - scenario tests skipped")) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	default:
		_, _ = fmt.Fprintf(w, "%s %s\n", r.icon("check"), r.green("All checks passed!")) //nolint:gosec // G705: w is a terminal/file writer, not an HTTP response
	}
}

// formatSkipReason formats the runner unavailable message for display.
func formatSkipReason(msg string) string {
	if strings.Contains(strings.ToLower(msg), "docker") {
		return "Docker not running"
	}

	if msg == "" {
		return "runner unavailable"
	}

	return msg
}

// pluralize returns the word with proper singular/plural form.
func pluralize(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}

	return fmt.Sprintf("%d %ss", n, word)
}

// formatSeverity returns a colored severity indicator.
func (r *TerminalReporter) formatSeverity(s validator.Severity) string {
	switch s {
	case validator.SeverityError:
		return r.red("[ERROR]")
	case validator.SeverityWarning:
		return r.yellow("[WARN]")
	case validator.SeverityNote:
		return r.blue("[NOTE]")
	case validator.SeverityInfo:
		return r.dim("[INFO]")
	}

	return r.dim("[INFO]")
}

// green returns a green colored string.
func (r *TerminalReporter) green(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[32m" + s + "\033[0m"
}

// red returns a red colored string.
func (r *TerminalReporter) red(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[31m" + s + "\033[0m"
}

// yellow returns a yellow colored string.
func (r *TerminalReporter) yellow(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[33m" + s + "\033[0m"
}

// blue returns a blue colored string.
func (r *TerminalReporter) blue(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[34m" + s + "\033[0m"
}

// bold returns a bold string.
func (r *TerminalReporter) bold(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[1m" + s + "\033[0m"
}

// dim returns a dimmed string.
func (r *TerminalReporter) dim(s interface{}) string {
	str := fmt.Sprintf("%v", s)

	if !r.colorEnabled {
		return str
	}

	return "\033[2m" + str + "\033[0m"
}

// icon returns an icon for the given name.
func (r *TerminalReporter) icon(name string) string {
	if !r.colorEnabled {
		switch name {
		case "check":
			return "[OK]"
		case "cross":
			return "[FAIL]"
		case "warning":
			return "[WARN]"
		case "shield":
			return "[GUARDIAN]"
		default:
			return "[*]"
		}
	}

	switch name {
	case "check":
		return "\033[32m✓\033[0m"
	case "cross":
		return "\033[31m✗\033[0m"
	case "warning":
		return "\033[33m⚠\033[0m"
	case "shield":
		return "🛡️"
	default:
		return "•"
	}
}

// cyan returns a cyan colored string.
func (r *TerminalReporter) cyan(s string) string {
	if !r.colorEnabled {
		return s
	}

	return "\033[36m" + s + "\033[0m"
}

// isTerminal checks if stdout is a terminal.
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	return fmt.Sprintf("%.1fs", d.Seconds())
}
