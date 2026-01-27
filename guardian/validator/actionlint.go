// Package validator provides validation for GitHub Actions workflows.
package validator

import (
	"context"
	"io"
	"path/filepath"

	"github.com/rhysd/actionlint"
)

// ActionlintValidator wraps actionlint for workflow validation.
type ActionlintValidator struct {
	path string // path to actionlint binary (unused when using Go API)
}

// NewActionlintValidator creates a new actionlint validator.
func NewActionlintValidator(path string) *ActionlintValidator {
	return &ActionlintValidator{path: path}
}

// Name returns the validator identifier.
func (v *ActionlintValidator) Name() string {
	return "actionlint"
}

// Validate analyzes the given workflow file using actionlint's Go API.
func (v *ActionlintValidator) Validate(_ context.Context, workflowPath string) ([]Finding, error) {
	opts := &actionlint.LinterOptions{
		Verbose: false,
	}
	linter, err := actionlint.NewLinter(io.Discard, opts)
	if err != nil {
		return nil, err
	}

	// Detect project root for context.
	// Workflow files are in .github/workflows/, so project root is 2 dirs up.
	workflowDir := filepath.Dir(workflowPath) // .github/workflows
	githubDir := filepath.Dir(workflowDir)    // .github
	projectRoot := filepath.Dir(githubDir)    // project root

	project, err := actionlint.NewProject(projectRoot)
	if err != nil {
		// Non-fatal: continue without project context
		project = nil
	}

	errs, err := linter.LintFile(workflowPath, project)
	if err != nil {
		return nil, err
	}

	return convertActionlintErrors(errs, workflowPath), nil
}

// convertActionlintErrors converts actionlint errors to Guardian findings.
func convertActionlintErrors(errs []*actionlint.Error, workflowPath string) []Finding {
	findings := make([]Finding, 0, len(errs))

	for _, e := range errs {
		finding := Finding{
			RuleID:   "actionlint/" + e.Kind,
			Severity: mapActionlintSeverity(e.Kind),
			Message:  e.Message,
			File:     workflowPath,
			Line:     e.Line,
			Column:   e.Column,
			Source:   SourceActionlint,
		}

		// Add suggestion if available from error message
		if e.Kind == "deprecated-commands" {
			finding.Suggestion = "Use the environment file method instead"
		}

		findings = append(findings, finding)
	}

	return findings
}

// mapActionlintSeverity maps actionlint error kinds to Guardian severity.
func mapActionlintSeverity(kind string) Severity {
	switch kind {
	case "expression", "syntax", "step-id", "job-needs", "runner-label":
		return SeverityError
	case "permissions", "deprecated-commands", "shellcheck":
		return SeverityWarning
	default:
		return SeverityWarning
	}
}
