package validator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

// SchemaValidator validates action.yml files against the expected schema.
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// Name returns the validator identifier.
func (v *SchemaValidator) Name() string {
	return "schema"
}

// Validate analyzes action.yml files for schema compliance.
// For workflow files, this validator passes through (actionlint handles those).
func (v *SchemaValidator) Validate(_ context.Context, path string) ([]Finding, error) {
	// Only validate action.yml files
	if !isActionFile(path) {
		return nil, nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // path from trusted validator input
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var action ActionYML
	if err := yaml.Unmarshal(data, &action); err != nil {
		return []Finding{{
			RuleID:   "schema/parse-error",
			Severity: SeverityError,
			Message:  fmt.Sprintf("failed to parse action.yml: %v", err),
			File:     path,
			Line:     1,
			Source:   SourceSchema,
		}}, nil
	}

	return v.validateAction(path, &action), nil
}

// ActionYML represents the structure of an action.yml file.
type ActionYML struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Author      string                  `yaml:"author,omitempty"`
	Inputs      map[string]ActionInput  `yaml:"inputs,omitempty"`
	Outputs     map[string]ActionOutput `yaml:"outputs,omitempty"`
	Runs        ActionRuns              `yaml:"runs"`
	Branding    *ActionBranding         `yaml:"branding,omitempty"`
}

// ActionInput represents an input parameter for an action.
type ActionInput struct {
	Description        string `yaml:"description"`
	Required           bool   `yaml:"required,omitempty"`
	Default            string `yaml:"default,omitempty"`
	DeprecationMessage string `yaml:"deprecationMessage,omitempty"`
}

// ActionOutput represents an output parameter for an action.
type ActionOutput struct {
	Description string `yaml:"description"`
	Value       string `yaml:"value,omitempty"`
}

// ActionRuns describes how the action is executed.
type ActionRuns struct {
	Using          string            `yaml:"using"`
	Main           string            `yaml:"main,omitempty"`
	Pre            string            `yaml:"pre,omitempty"`
	PreIf          string            `yaml:"pre-if,omitempty"`
	Post           string            `yaml:"post,omitempty"`
	PostIf         string            `yaml:"post-if,omitempty"`
	Image          string            `yaml:"image,omitempty"`
	PreEntrypoint  string            `yaml:"pre-entrypoint,omitempty"`
	Entrypoint     string            `yaml:"entrypoint,omitempty"`
	PostEntrypoint string            `yaml:"post-entrypoint,omitempty"`
	Args           []string          `yaml:"args,omitempty"`
	Env            map[string]string `yaml:"env,omitempty"`
	Steps          []interface{}     `yaml:"steps,omitempty"` // For composite actions
}

// ActionBranding describes the action's appearance in the marketplace.
type ActionBranding struct {
	Icon  string `yaml:"icon,omitempty"`
	Color string `yaml:"color,omitempty"`
}

// validateAction checks an action.yml for schema compliance.
func (v *SchemaValidator) validateAction(path string, action *ActionYML) []Finding {
	var findings []Finding

	// Required fields
	if action.Name == "" {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-name",
			Severity:   SeverityError,
			Message:    "action.yml must have a 'name' field",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Add a 'name' field to describe the action",
		})
	}

	if action.Description == "" {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-description",
			Severity:   SeverityError,
			Message:    "action.yml must have a 'description' field",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Add a 'description' field explaining what the action does",
		})
	}

	// Validate runs configuration
	findings = append(findings, v.validateRuns(path, &action.Runs)...)

	// Validate inputs
	for name, input := range action.Inputs {
		if input.Description == "" {
			findings = append(findings, Finding{
				RuleID:     "schema/input-missing-description",
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("input '%s' should have a description", name),
				File:       path,
				Line:       1,
				Source:     SourceSchema,
				Suggestion: fmt.Sprintf("Add a description for input '%s'", name),
			})
		}
	}

	// Validate outputs
	for name, output := range action.Outputs {
		if output.Description == "" {
			findings = append(findings, Finding{
				RuleID:     "schema/output-missing-description",
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("output '%s' should have a description", name),
				File:       path,
				Line:       1,
				Source:     SourceSchema,
				Suggestion: fmt.Sprintf("Add a description for output '%s'", name),
			})
		}
	}

	return findings
}

// validateRuns checks the runs configuration.
func (v *SchemaValidator) validateRuns(path string, runs *ActionRuns) []Finding {
	var findings []Finding

	validUsing := map[string]bool{
		"composite": true,
		"node12":    true,
		"node16":    true,
		"node20":    true,
		"docker":    true,
	}

	if runs.Using == "" {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-using",
			Severity:   SeverityError,
			Message:    "runs.using is required",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Specify 'runs.using' as 'composite', 'node20', or 'docker'",
		})
	} else if !validUsing[runs.Using] {
		findings = append(findings, Finding{
			RuleID:     "schema/invalid-using",
			Severity:   SeverityError,
			Message:    fmt.Sprintf("invalid runs.using value: %s", runs.Using),
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Use 'composite', 'node20', or 'docker'",
		})
	}

	// Check for deprecated node versions
	if runs.Using == "node12" {
		findings = append(findings, Finding{
			RuleID:     "schema/deprecated-node12",
			Severity:   SeverityWarning,
			Message:    "node12 is deprecated, upgrade to node20",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Change runs.using to 'node20'",
		})
	}
	if runs.Using == "node16" {
		findings = append(findings, Finding{
			RuleID:     "schema/deprecated-node16",
			Severity:   SeverityWarning,
			Message:    "node16 is deprecated, upgrade to node20",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Change runs.using to 'node20'",
		})
	}

	// Node/JavaScript actions require main
	if (runs.Using == "node12" || runs.Using == "node16" || runs.Using == "node20") && runs.Main == "" {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-main",
			Severity:   SeverityError,
			Message:    "runs.main is required for JavaScript actions",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Add 'runs.main' pointing to your entry point file",
		})
	}

	// Docker actions require image
	if runs.Using == "docker" && runs.Image == "" {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-image",
			Severity:   SeverityError,
			Message:    "runs.image is required for Docker actions",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Add 'runs.image' specifying the Docker image",
		})
	}

	// Composite actions require steps
	if runs.Using == "composite" && len(runs.Steps) == 0 {
		findings = append(findings, Finding{
			RuleID:     "schema/missing-steps",
			Severity:   SeverityError,
			Message:    "runs.steps is required for composite actions",
			File:       path,
			Line:       1,
			Source:     SourceSchema,
			Suggestion: "Add 'runs.steps' with at least one step",
		})
	}

	return findings
}

// isActionFile checks if the path is an action.yml file.
func isActionFile(path string) bool {
	base := filepath.Base(path)
	return base == "action.yml" || base == "action.yaml"
}
