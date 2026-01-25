package validator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Validator performs static analysis on workflow files.
type Validator interface {
	// Name returns the validator identifier (e.g., "actionlint", "schema").
	Name() string

	// Validate analyzes the given workflow file.
	// Returns findings for any issues detected.
	Validate(ctx context.Context, workflowPath string) ([]Finding, error)
}

// Registry manages available validators.
type Registry struct {
	validators map[string]Validator
	order      []string // maintains registration order
}

// NewRegistry creates a new validator registry.
func NewRegistry() *Registry {
	return &Registry{
		validators: make(map[string]Validator),
		order:      []string{},
	}
}

// Register adds a validator to the registry.
func (r *Registry) Register(v Validator) {
	name := v.Name()
	if _, exists := r.validators[name]; !exists {
		r.order = append(r.order, name)
	}
	r.validators[name] = v
}

// Get returns a validator by name.
func (r *Registry) Get(name string) (Validator, bool) {
	v, ok := r.validators[name]
	return v, ok
}

// All returns all registered validators in registration order.
func (r *Registry) All() []Validator {
	result := make([]Validator, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.validators[name])
	}
	return result
}

// ValidateAll runs all validators against a workflow file.
func (r *Registry) ValidateAll(ctx context.Context, workflowPath string) ([]Finding, error) {
	var allFindings []Finding

	for _, v := range r.All() {
		findings, err := v.Validate(ctx, workflowPath)
		if err != nil {
			// Log error but continue with other validators
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// FindWorkflowFiles returns all workflow YAML files in the given directory.
func FindWorkflowFiles(dir string) ([]string, error) {
	var workflows []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isWorkflowFile(path) {
			workflows = append(workflows, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return workflows, nil
}

// isWorkflowFile checks if the file is a GitHub Actions workflow file.
func isWorkflowFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yml" || ext == ".yaml"
}
