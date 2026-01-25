package policy

import (
	"context"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Engine evaluates policies against workflows.
type Engine struct {
	policies   []*Policy
	exceptions *ExceptionConfig
}

// NewEngine creates a new policy engine with default policies.
func NewEngine() (*Engine, error) {
	e := &Engine{
		policies:   defaultPolicies(),
		exceptions: &ExceptionConfig{},
	}
	return e, nil
}

// Evaluate runs all policies against a parsed workflow.
// Returns findings for any violations.
func (e *Engine) Evaluate(ctx context.Context, workflow *Workflow) ([]validator.Finding, error) {
	var findings []validator.Finding

	for _, policy := range e.policies {
		policyFindings := policy.Check(workflow)
		for i := range policyFindings {
			// Check if finding is excepted
			if !e.IsExcepted(&policyFindings[i]) {
				findings = append(findings, policyFindings[i])
			}
		}
	}

	return findings, nil
}

// LoadExceptions loads policy exceptions from configuration file.
func (e *Engine) LoadExceptions(_ context.Context, configPath string) error {
	config, err := LoadExceptionConfig(configPath)
	if err != nil {
		return err
	}
	e.exceptions = config
	return nil
}

// IsExcepted checks if a finding is covered by an exception.
func (e *Engine) IsExcepted(finding *validator.Finding) bool {
	if e.exceptions == nil {
		return false
	}

	for _, exc := range e.exceptions.Exceptions {
		if exc.Matches(finding) {
			return true
		}
	}

	return false
}

// Policies returns all registered policies.
func (e *Engine) Policies() []PolicyInfo {
	result := make([]PolicyInfo, 0, len(e.policies))
	for _, p := range e.policies {
		result = append(result, PolicyInfo{
			ID:          p.ID,
			Severity:    string(p.Severity),
			Description: p.Description,
			HelpURL:     p.HelpURL,
			Tags:        p.Tags,
		})
	}
	return result
}

// RegisterPolicy adds a custom policy to the engine.
func (e *Engine) RegisterPolicy(p *Policy) {
	e.policies = append(e.policies, p)
}

// PolicyInfo provides policy metadata.
type PolicyInfo struct {
	ID          string
	Severity    string
	Description string
	HelpURL     string
	Tags        []string
}

// defaultPolicies returns the built-in policy set.
func defaultPolicies() []*Policy {
	return []*Policy{
		SHAPinnedActions,
		ExplicitPermissions,
		NoDangerousWorkflows,
		NoSecretLogging,
		ConcurrencyDefined,
		MinimalPermissions,
	}
}
