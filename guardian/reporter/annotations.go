// Package reporter provides report generation in various formats.
package reporter

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// AnnotationsReporter generates GitHub workflow commands for annotations.
type AnnotationsReporter struct{}

// NewAnnotationsReporter creates a new annotations reporter.
func NewAnnotationsReporter() *AnnotationsReporter {
	return &AnnotationsReporter{}
}

// Name returns the reporter identifier.
func (r *AnnotationsReporter) Name() string {
	return "annotations"
}

// Write outputs findings as GitHub workflow commands.
func (r *AnnotationsReporter) Write(ctx context.Context, report *Report, w io.Writer) error {
	if report.StaticResults == nil {
		return nil
	}

	return r.WriteAnnotations(ctx, report.StaticResults.Findings, w)
}

// WriteFile outputs the report to a file.
func (r *AnnotationsReporter) WriteFile(ctx context.Context, report *Report, path string) error {
	f, err := os.Create(path) //nolint:gosec // path from trusted caller
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	defer func() { _ = f.Close() }()

	return r.Write(ctx, report, f)
}

// WriteAnnotations outputs findings as GitHub annotation commands.
func (r *AnnotationsReporter) WriteAnnotations(_ context.Context, findings []validator.Finding, w io.Writer) error {
	for _, f := range findings {
		cmd := mapSeverityToCommand(f.Severity)

		// Format: ::warning file={name},line={line},col={col},endColumn={endColumn}::{message}
		_, _ = fmt.Fprintf(w, "::%s file=%s", cmd, f.File)

		if f.Line > 0 {
			_, _ = fmt.Fprintf(w, ",line=%d", f.Line)
		}
		if f.Column > 0 {
			_, _ = fmt.Fprintf(w, ",col=%d", f.Column)
		}
		if f.EndColumn > 0 {
			_, _ = fmt.Fprintf(w, ",endColumn=%d", f.EndColumn)
		}

		// Title with rule ID
		_, _ = fmt.Fprintf(w, ",title=%s", f.RuleID)

		// Message
		_, _ = fmt.Fprintf(w, "::%s", escapeAnnotation(f.Message))

		// Add suggestion if present
		if f.Suggestion != "" {
			_, _ = fmt.Fprintf(w, " (Suggestion: %s)", escapeAnnotation(f.Suggestion))
		}

		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// mapSeverityToCommand converts Guardian severity to GitHub annotation command.
func mapSeverityToCommand(s validator.Severity) string {
	switch s {
	case validator.SeverityError:
		return "error"
	case validator.SeverityWarning:
		return "warning"
	case validator.SeverityNote, validator.SeverityInfo:
		return "notice"
	}
	return "notice"
}

// escapeAnnotation escapes special characters in annotation messages.
func escapeAnnotation(s string) string {
	// GitHub annotations use %0A for newline, %0D for carriage return, %25 for %
	result := ""
	for _, c := range s {
		switch c {
		case '%':
			result += "%25"
		case '\n':
			result += "%0A"
		case '\r':
			result += "%0D"
		default:
			result += string(c)
		}
	}
	return result
}
