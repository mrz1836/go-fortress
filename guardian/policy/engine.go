// Package policy provides workflow policy evaluation and validation.
package policy

import (
	"context"
	"sync"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Engine evaluates policies against workflows.
type Engine struct {
	mu                sync.RWMutex
	policies          []*Policy
	exceptions        *ExceptionConfig
	severityOverrides map[string]validator.Severity
	escalatedPolicies map[string]bool
}

// Info provides policy metadata.
type Info struct {
	ID          string
	Severity    string
	Description string
	HelpURL     string
	Tags        []string
}

// NewEngine creates a new policy engine with default policies.
func NewEngine() (*Engine, error) {
	e := &Engine{
		policies:          defaultPolicies(),
		exceptions:        &ExceptionConfig{},
		severityOverrides: make(map[string]validator.Severity),
		escalatedPolicies: make(map[string]bool),
	}

	return e, nil
}

// EscalateToError upgrades a policy's severity to error for enforcement.
// This is used to convert warnings to errors for specific rules when strict mode is enabled.
func (e *Engine) EscalateToError(policyID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.severityOverrides[policyID] = validator.SeverityError
	e.escalatedPolicies[policyID] = true
}

// SetSeverity overrides the severity for a specific policy.
func (e *Engine) SetSeverity(policyID string, severity validator.Severity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.severityOverrides[policyID] = severity
}

// IsEscalated returns true if the policy has been escalated to error.
func (e *Engine) IsEscalated(policyID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.escalatedPolicies[policyID]
}

// Evaluate runs all policies against a parsed workflow.
// Returns findings for any violations.
func (e *Engine) Evaluate(ctx context.Context, workflow *Workflow) ([]validator.Finding, error) {
	_ = ctx // reserved for future async evaluation

	var findings []validator.Finding

	for _, policy := range e.policies {
		policyFindings := policy.Check(workflow)

		for i := range policyFindings {
			// Check if finding is excepted
			if e.IsExcepted(&policyFindings[i]) {
				continue
			}

			// Apply severity escalation if configured
			policyFindings[i].Severity = e.getEffectiveSeverity(&policyFindings[i], policy.ID)

			findings = append(findings, policyFindings[i])
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
func (e *Engine) Policies() []Info {
	result := make([]Info, 0, len(e.policies))

	for _, p := range e.policies {
		result = append(result, Info{
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

// getEffectiveSeverity returns the effective severity for a finding,
// applying any overrides configured via EscalateToError or SetSeverity.
func (e *Engine) getEffectiveSeverity(finding *validator.Finding, policyID string) validator.Severity {
	e.mu.RLock()
	override, ok := e.severityOverrides[policyID]
	e.mu.RUnlock()

	if ok {
		return override
	}

	return finding.Severity
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
