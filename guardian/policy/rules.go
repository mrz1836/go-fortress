package policy

import (
	"regexp"
	"strings"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Policy defines a validation rule for workflows.
type Policy struct {
	// ID uniquely identifies the policy (e.g., "sha-pinned-actions").
	ID string

	// Severity is the default severity when this policy is violated.
	Severity validator.Severity

	// Description explains what this policy checks.
	Description string

	// HelpURL links to documentation about this policy.
	HelpURL string

	// Tags enable filtering policies (e.g., ["security", "required"]).
	Tags []string

	// Check is the function that validates a workflow.
	Check CheckFunc
}

// CheckFunc is the signature for policy check functions.
type CheckFunc func(workflow *Workflow) []validator.Finding

// SHAPinnedActions requires all actions to be pinned to full commit SHA.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var SHAPinnedActions = &Policy{
	ID:          "sha-pinned-actions",
	Severity:    validator.SeverityError,
	Description: "All actions must be pinned to full commit SHA",
	HelpURL:     "https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions",
	Tags:        []string{"security", "required"},
	Check:       checkSHAPinnedActions,
}

// ExplicitPermissions requires workflows to declare explicit permissions.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var ExplicitPermissions = &Policy{
	ID:          "explicit-permissions",
	Severity:    validator.SeverityWarning,
	Description: "Workflows should declare explicit permissions",
	HelpURL:     "https://docs.github.com/en/actions/security-guides/automatic-token-authentication#modifying-the-permissions-for-the-github_token",
	Tags:        []string{"security", "recommended"},
	Check:       checkExplicitPermissions,
}

// NoDangerousWorkflows prevents dangerous pull_request_target patterns.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var NoDangerousWorkflows = &Policy{
	ID:          "no-dangerous-workflows",
	Severity:    validator.SeverityError,
	Description: "Workflows must not use dangerous pull_request_target patterns",
	HelpURL:     "https://securitylab.github.com/research/github-actions-preventing-pwn-requests/",
	Tags:        []string{"security", "required"},
	Check:       checkNoDangerousWorkflows,
}

// NoSecretLogging prevents secrets from being logged.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var NoSecretLogging = &Policy{
	ID:          "no-secret-logging",
	Severity:    validator.SeverityError,
	Description: "Secrets must not be logged or exposed in output",
	HelpURL:     "https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions",
	Tags:        []string{"security", "required"},
	Check:       checkNoSecretLogging,
}

// ConcurrencyDefined recommends concurrency configuration.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var ConcurrencyDefined = &Policy{
	ID:          "concurrency-defined",
	Severity:    validator.SeverityWarning,
	Description: "Workflows should define concurrency groups to prevent duplicate runs",
	HelpURL:     "https://docs.github.com/en/actions/using-jobs/using-concurrency",
	Tags:        []string{"efficiency", "recommended"},
	Check:       checkConcurrencyDefined,
}

// MinimalPermissions recommends least-privilege permissions.
//
//nolint:gochecknoglobals // policy definitions are intentionally global
var MinimalPermissions = &Policy{
	ID:          "minimal-permissions",
	Severity:    validator.SeverityWarning,
	Description: "Use least-privilege permissions for GITHUB_TOKEN",
	HelpURL:     "https://docs.github.com/en/actions/security-guides/automatic-token-authentication",
	Tags:        []string{"security", "recommended"},
	Check:       checkMinimalPermissions,
}

// SHA pattern matches a 40-character hex string (full commit SHA).
var shaPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)

// Action reference pattern: owner/repo@ref or owner/repo/path@ref.
var actionRefPattern = regexp.MustCompile(`^([^@]+)@(.+)$`)

func checkSHAPinnedActions(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	for jobName, job := range w.Jobs {
		for i, step := range job.Steps {
			if step.Uses == "" {
				continue
			}

			// Skip local actions (./path)
			if strings.HasPrefix(step.Uses, "./") {
				continue
			}

			// Skip Docker actions (docker://)
			if strings.HasPrefix(step.Uses, "docker://") {
				continue
			}

			matches := actionRefPattern.FindStringSubmatch(step.Uses)
			if len(matches) != 3 {
				continue
			}

			ref := matches[2]
			if !shaPattern.MatchString(ref) {
				findings = append(findings, validator.Finding{
					RuleID:     "policy/sha-pinned-actions",
					Severity:   validator.SeverityError,
					Message:    "action is not pinned to a full commit SHA",
					File:       w.Path,
					Line:       step.Line,
					Source:     validator.SourcePolicy,
					Suggestion: "Pin to a full 40-character SHA, e.g., actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
				})
				_ = jobName // referenced for context
				_ = i
			}
		}
	}

	return findings
}

func checkExplicitPermissions(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	// Check if workflow-level permissions are defined
	if w.Permissions == nil {
		// Check if any job has permissions defined
		hasJobPermissions := false
		for _, job := range w.Jobs {
			if job.Permissions != nil {
				hasJobPermissions = true
				break
			}
		}

		if !hasJobPermissions {
			findings = append(findings, validator.Finding{
				RuleID:     "policy/explicit-permissions",
				Severity:   validator.SeverityWarning,
				Message:    "workflow does not declare explicit permissions",
				File:       w.Path,
				Line:       1,
				Source:     validator.SourcePolicy,
				Suggestion: "Add 'permissions: {}' at workflow or job level to restrict GITHUB_TOKEN",
			})
		}
	}

	return findings
}

func checkNoDangerousWorkflows(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	// Check for pull_request_target with write permissions
	if w.On != nil && w.On.PullRequestTarget != nil {
		// First, check for the actually dangerous pattern: checking out PR head
		// This is the real security issue - executing untrusted code
		hasDangerousCheckout := false
		for _, job := range w.Jobs {
			for _, step := range job.Steps {
				if strings.Contains(step.Uses, "actions/checkout") {
					// Check if it's checking out PR head
					if ref, ok := step.With["ref"].(string); ok {
						if strings.Contains(ref, "pull_request") ||
							strings.Contains(ref, "github.event.pull_request") {
							hasDangerousCheckout = true
							findings = append(findings, validator.Finding{
								RuleID:     "policy/no-dangerous-workflows",
								Severity:   validator.SeverityError,
								Message:    "pull_request_target checking out PR head is dangerous",
								File:       w.Path,
								Line:       step.Line,
								Source:     validator.SourcePolicy,
								Suggestion: "Do not checkout PR head in pull_request_target workflows",
							})
						}
					}
				}
			}
		}

		// Only warn about pull_request_target + write permissions if there's no
		// explicit dangerous checkout. The two-workflow pattern (separate workflows
		// for same-repo and fork PRs) is legitimate when no PR code is executed.
		if w.HasWritePermissions() && !hasDangerousCheckout {
			findings = append(findings, validator.Finding{
				RuleID:     "policy/no-dangerous-workflows",
				Severity:   validator.SeverityWarning,
				Message:    "pull_request_target with write permissions requires careful review",
				File:       w.Path,
				Line:       1,
				Source:     validator.SourcePolicy,
				Suggestion: "Ensure no untrusted code from the PR is checked out or executed",
			})
		}
	}

	return findings
}

// secretPatterns matches common patterns for logging secrets.
//
//nolint:gochecknoglobals // compiled regex patterns for reuse
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`echo.*\$\{\{.*secrets\.`),
	regexp.MustCompile(`echo.*\$secrets\.`),
	regexp.MustCompile(`printf.*\$\{\{.*secrets\.`),
	regexp.MustCompile(`cat.*secrets\.`),
	regexp.MustCompile(`env:.*\$\{\{.*secrets\..*\}\}.*#.*debug`),
}

func checkNoSecretLogging(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			if step.Run == "" {
				continue
			}

			for _, pattern := range secretPatterns {
				if pattern.MatchString(step.Run) {
					findings = append(findings, validator.Finding{
						RuleID:     "policy/no-secret-logging",
						Severity:   validator.SeverityError,
						Message:    "potential secret logging detected",
						File:       w.Path,
						Line:       step.Line,
						Source:     validator.SourcePolicy,
						Suggestion: "Do not log secrets; use add-mask for dynamic values",
					})
					break
				}
			}
		}
	}

	return findings
}

func checkConcurrencyDefined(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	// Skip if workflow already has concurrency
	if w.Concurrency != nil {
		return findings
	}

	// Check if any job has concurrency
	for _, job := range w.Jobs {
		if job.Concurrency != nil {
			return findings
		}
	}

	// Only warn for workflows triggered by push or pull_request
	if w.On != nil && (w.On.Push != nil || w.On.PullRequest != nil) {
		findings = append(findings, validator.Finding{
			RuleID:     "policy/concurrency-defined",
			Severity:   validator.SeverityWarning,
			Message:    "workflow does not define concurrency group",
			File:       w.Path,
			Line:       1,
			Source:     validator.SourcePolicy,
			Suggestion: "Add concurrency group to cancel outdated runs: concurrency: { group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true }",
		})
	}

	return findings
}

func checkMinimalPermissions(w *Workflow) []validator.Finding {
	var findings []validator.Finding

	// Check for write-all or broad permissions
	if w.Permissions != nil {
		if w.Permissions.All == "write-all" {
			findings = append(findings, validator.Finding{
				RuleID:     "policy/minimal-permissions",
				Severity:   validator.SeverityWarning,
				Message:    "workflow uses 'write-all' permissions",
				File:       w.Path,
				Line:       1,
				Source:     validator.SourcePolicy,
				Suggestion: "Specify only the permissions needed for each job",
			})
		}
	}

	return findings
}
