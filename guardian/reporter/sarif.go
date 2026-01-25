package reporter

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// SARIFReporter writes SARIF 2.1.0 format output for GitHub Security integration.
type SARIFReporter struct{}

// NewSARIFReporter creates a new SARIF reporter.
func NewSARIFReporter() *SARIFReporter {
	return &SARIFReporter{}
}

// Name returns the reporter identifier.
func (r *SARIFReporter) Name() string {
	return "sarif"
}

// Write outputs the report in SARIF format.
func (r *SARIFReporter) Write(_ context.Context, report *Report, w io.Writer) error {
	sarifReport := sarif.NewReport()

	// Create the tool component
	driver := sarif.NewToolComponent().
		WithName("Fortress Guardian").
		WithVersion(report.Version).
		WithInformationURI("https://github.com/mrz1836/go-fortress")

	// Create tool with driver
	tool := sarif.NewTool().WithDriver(driver)

	// Create run with tool
	run := sarif.NewRun().WithTool(tool)

	// Add rules and results from static findings
	if report.StaticResults != nil {
		rules := make(map[string]bool)
		for _, f := range report.StaticResults.Findings {
			// Add rule if not already added
			if !rules[f.RuleID] {
				shortDesc := sarif.NewMultiformatMessageString().WithText(f.Message)
				rule := sarif.NewRule(f.RuleID).
					WithShortDescription(shortDesc)
				if f.Suggestion != "" {
					help := sarif.NewMultiformatMessageString().WithText(f.Suggestion)
					rule.WithHelp(help)
				}
				run.Tool.Driver.AddRule(rule)
				rules[f.RuleID] = true
			}

			// Add result
			msg := sarif.NewMessage().WithText(f.Message)
			result := sarif.NewResult().
				WithRuleID(f.RuleID).
				WithLevel(mapSeverityToSARIF(f.Severity)).
				WithMessage(msg)

			// Add location
			if f.File != "" {
				physLoc := sarif.NewPhysicalLocation().
					WithArtifactLocation(sarif.NewSimpleArtifactLocation(f.File))

				if f.Line > 0 {
					region := sarif.NewRegion().
						WithStartLine(f.Line)
					if f.Column > 0 {
						region.WithStartColumn(f.Column)
					}
					if f.EndLine > 0 {
						region.WithEndLine(f.EndLine)
					}
					if f.EndColumn > 0 {
						region.WithEndColumn(f.EndColumn)
					}
					physLoc.WithRegion(region)
				}

				result.WithLocations([]*sarif.Location{
					sarif.NewLocation().WithPhysicalLocation(physLoc),
				})
			}

			// Add fingerprint for deduplication
			if f.Fingerprint != "" {
				result.WithPartialFingerprints(map[string]string{
					"guardian/fingerprint": f.Fingerprint,
				})
			}

			run.AddResult(result)
		}
	}

	sarifReport.AddRun(run)

	// Write to output
	return sarifReport.Write(w)
}

// WriteFile outputs the report to a file in SARIF format.
func (r *SARIFReporter) WriteFile(ctx context.Context, report *Report, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return r.Write(ctx, report, f)
}

// mapSeverityToSARIF converts Guardian severity to SARIF level.
func mapSeverityToSARIF(s validator.Severity) string {
	switch s {
	case validator.SeverityError:
		return "error"
	case validator.SeverityWarning:
		return "warning"
	case validator.SeverityNote:
		return "note"
	case validator.SeverityInfo:
		return "none"
	}
	return "none"
}
