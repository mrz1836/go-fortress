package policy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/policy"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// helper to create engine and evaluate workflow
func evaluateWorkflow(t *testing.T, workflow *policy.Workflow) []validator.Finding {
	t.Helper()

	engine, err := policy.NewEngine()
	require.NoError(t, err)

	ctx := context.Background()
	findings, err := engine.Evaluate(ctx, workflow)
	require.NoError(t, err)

	return findings
}

// filterByRuleID returns findings matching the given rule ID prefix.
func filterByRuleID(findings []validator.Finding, rulePrefix string) []validator.Finding {
	var filtered []validator.Finding
	for _, f := range findings {
		if f.RuleID == rulePrefix {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// TestCheckSHAPinnedActions tests the SHA pinned actions policy.
func TestCheckSHAPinnedActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		steps          []*policy.Step
		expectFindings int
	}{
		{
			name: "SHA pinned action passes",
			steps: []*policy.Step{
				{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
			},
			expectFindings: 0,
		},
		{
			name: "tag reference triggers violation",
			steps: []*policy.Step{
				{Uses: "actions/checkout@v4", Line: 10},
			},
			expectFindings: 1,
		},
		{
			name: "branch reference triggers violation",
			steps: []*policy.Step{
				{Uses: "actions/checkout@main", Line: 15},
			},
			expectFindings: 1,
		},
		{
			name: "local action is allowed",
			steps: []*policy.Step{
				{Uses: "./local-action"},
			},
			expectFindings: 0,
		},
		{
			name: "docker action is allowed",
			steps: []*policy.Step{
				{Uses: "docker://alpine:latest"},
			},
			expectFindings: 0,
		},
		{
			name: "run step without uses is ignored",
			steps: []*policy.Step{
				{Run: "echo hello"},
			},
			expectFindings: 0,
		},
		{
			name: "action with path is checked",
			steps: []*policy.Step{
				{Uses: "owner/repo/path/to/action@v1", Line: 20},
			},
			expectFindings: 1,
		},
		{
			name: "multiple unpinned actions",
			steps: []*policy.Step{
				{Uses: "actions/checkout@v4", Line: 10},
				{Uses: "actions/setup-go@v5", Line: 15},
				{Uses: "actions/cache@v3", Line: 20},
			},
			expectFindings: 3,
		},
		{
			name: "mixed pinned and unpinned",
			steps: []*policy.Step{
				{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
				{Uses: "actions/setup-go@v5", Line: 15},
				{Uses: "actions/cache@abc123def456789012345678901234567890abcd"},
			},
			expectFindings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := &policy.Workflow{
				Path:        "test.yml",
				Permissions: &policy.Permissions{Contents: "read"},
				Jobs: map[string]*policy.Job{
					"test": {Steps: tt.steps},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			shaFindings := filterByRuleID(findings, "policy/sha-pinned-actions")
			assert.Len(t, shaFindings, tt.expectFindings)
		})
	}
}

// TestCheckExplicitPermissions tests the explicit permissions policy.
func TestCheckExplicitPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		workflowPerms  *policy.Permissions
		jobPerms       *policy.Permissions
		expectFindings int
	}{
		{
			name:           "workflow-level permissions pass",
			workflowPerms:  &policy.Permissions{Contents: "read"},
			expectFindings: 0,
		},
		{
			name:           "job-level permissions pass",
			workflowPerms:  nil,
			jobPerms:       &policy.Permissions{Contents: "read"},
			expectFindings: 0,
		},
		{
			name:           "no permissions triggers warning",
			workflowPerms:  nil,
			jobPerms:       nil,
			expectFindings: 1,
		},
		{
			name:           "empty workflow permissions count",
			workflowPerms:  &policy.Permissions{},
			expectFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := &policy.Workflow{
				Path:        "test.yml",
				Permissions: tt.workflowPerms,
				Jobs: map[string]*policy.Job{
					"test": {
						Permissions: tt.jobPerms,
						Steps: []*policy.Step{
							{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
						},
					},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			permFindings := filterByRuleID(findings, "policy/explicit-permissions")
			assert.Len(t, permFindings, tt.expectFindings)
		})
	}
}

// TestCheckNoDangerousWorkflows tests the dangerous workflows policy.
func TestCheckNoDangerousWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		triggers       *policy.WorkflowTriggers
		permissions    *policy.Permissions
		steps          []*policy.Step
		expectFindings int
		expectSeverity validator.Severity
	}{
		{
			name:           "regular push workflow is safe",
			triggers:       &policy.WorkflowTriggers{Push: &policy.PushTrigger{}},
			permissions:    &policy.Permissions{Contents: "write"},
			expectFindings: 0,
		},
		{
			name:           "pull_request workflow is safe",
			triggers:       &policy.WorkflowTriggers{PullRequest: &policy.PullRequestTrigger{}},
			permissions:    &policy.Permissions{Contents: "write"},
			expectFindings: 0,
		},
		{
			name:        "pull_request_target with write perms triggers warning",
			triggers:    &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions: &policy.Permissions{Contents: "write"},
			steps: []*policy.Step{
				{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
			},
			expectFindings: 1,
			expectSeverity: validator.SeverityWarning,
		},
		{
			name:        "pull_request_target checking out PR head is dangerous",
			triggers:    &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions: &policy.Permissions{Contents: "write"},
			steps: []*policy.Step{
				{
					Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
					With: map[string]interface{}{
						"ref": "${{ github.event.pull_request.head.sha }}",
					},
					Line: 20,
				},
			},
			expectFindings: 1,
			expectSeverity: validator.SeverityError,
		},
		{
			name:           "pull_request_target with read perms is ok",
			triggers:       &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions:    &policy.Permissions{Contents: "read"},
			expectFindings: 0,
		},
		{
			name:        "pull_request_target checking out head.ref is dangerous",
			triggers:    &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions: &policy.Permissions{Contents: "write"},
			steps: []*policy.Step{
				{
					Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
					With: map[string]interface{}{
						"ref": "${{ github.event.pull_request.head.ref }}",
					},
					Line: 25,
				},
			},
			expectFindings: 1,
			expectSeverity: validator.SeverityError,
		},
		{
			name:        "pull_request_target with refs/pull/*/head is dangerous",
			triggers:    &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions: &policy.Permissions{Contents: "write"},
			steps: []*policy.Step{
				{
					Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
					With: map[string]interface{}{
						"ref": "refs/pull/${{ github.event.number }}/head",
					},
					Line: 30,
				},
			},
			expectFindings: 1,
			expectSeverity: validator.SeverityError,
		},
		{
			name:        "pull_request_target with refs/pull/*/merge is dangerous",
			triggers:    &policy.WorkflowTriggers{PullRequestTarget: &policy.PullRequestTrigger{}},
			permissions: &policy.Permissions{Contents: "write"},
			steps: []*policy.Step{
				{
					Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
					With: map[string]interface{}{
						"ref": "refs/pull/123/merge",
					},
					Line: 35,
				},
			},
			expectFindings: 1,
			expectSeverity: validator.SeverityError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := &policy.Workflow{
				Path:        "test.yml",
				On:          tt.triggers,
				Permissions: tt.permissions,
				Concurrency: &policy.Concurrency{Group: "test"},
				Jobs: map[string]*policy.Job{
					"test": {
						Steps: tt.steps,
					},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			dangerousFindings := filterByRuleID(findings, "policy/no-dangerous-workflows")
			assert.Len(t, dangerousFindings, tt.expectFindings)

			if tt.expectFindings > 0 && tt.expectSeverity != "" {
				assert.Equal(t, tt.expectSeverity, dangerousFindings[0].Severity)
			}
		})
	}
}

// TestCheckNoSecretLogging tests the secret logging detection policy.
func TestCheckNoSecretLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		runScript      string
		expectFindings int
	}{
		{
			name:           "normal echo is safe",
			runScript:      "echo Hello World",
			expectFindings: 0,
		},
		{
			name:           "echo with secret expression is detected",
			runScript:      `echo ${{ secrets.MY_SECRET }}`,
			expectFindings: 1,
		},
		{
			name:           "printf with secret is detected",
			runScript:      `printf "Token: ${{ secrets.TOKEN }}\n"`,
			expectFindings: 1,
		},
		{
			name:           "cat with secrets interpolation is detected",
			runScript:      `cat ${{ secrets.CONFIG_FILE }}`,
			expectFindings: 1,
		},
		{
			name:           "cat secrets.txt (file path) is NOT flagged - no false positive",
			runScript:      "cat secrets.txt",
			expectFindings: 0,
		},
		{
			name:           "cat .secrets/config (directory path) is NOT flagged - no false positive",
			runScript:      "cat .secrets/config",
			expectFindings: 0,
		},
		{
			name:           "safe secret usage is allowed",
			runScript:      "MY_TOKEN=${{ secrets.TOKEN }} ./deploy.sh",
			expectFindings: 0,
		},
		{
			name:           "step without run is safe",
			runScript:      "",
			expectFindings: 0,
		},
		{
			name:           "console.log with secret is detected",
			runScript:      `node -e "console.log(${{ secrets.API_KEY }})"`,
			expectFindings: 1,
		},
		{
			name:           "base64 encoding secret is detected",
			runScript:      `echo ${{ secrets.TOKEN }} | base64`,
			expectFindings: 1,
		},
		{
			name:           "print command with secret is detected",
			runScript:      `print "Debug: ${{ secrets.PASSWORD }}"`,
			expectFindings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			steps := []*policy.Step{}
			if tt.runScript != "" {
				steps = append(steps, &policy.Step{Run: tt.runScript, Line: 10})
			} else {
				steps = append(steps, &policy.Step{
					Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
				})
			}

			workflow := &policy.Workflow{
				Path:        "test.yml",
				Permissions: &policy.Permissions{Contents: "read"},
				Jobs: map[string]*policy.Job{
					"test": {Steps: steps},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			secretFindings := filterByRuleID(findings, "policy/no-secret-logging")
			assert.Len(t, secretFindings, tt.expectFindings)
		})
	}
}

// TestCheckConcurrencyDefined tests the concurrency policy.
func TestCheckConcurrencyDefined(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		triggers            *policy.WorkflowTriggers
		workflowConcurrency *policy.Concurrency
		jobConcurrency      *policy.Concurrency
		expectFindings      int
	}{
		{
			name:                "workflow concurrency defined passes",
			triggers:            &policy.WorkflowTriggers{Push: &policy.PushTrigger{}},
			workflowConcurrency: &policy.Concurrency{Group: "${{ github.workflow }}"},
			expectFindings:      0,
		},
		{
			name:           "job concurrency defined passes",
			triggers:       &policy.WorkflowTriggers{Push: &policy.PushTrigger{}},
			jobConcurrency: &policy.Concurrency{Group: "my-group"},
			expectFindings: 0,
		},
		{
			name:           "push without concurrency triggers warning",
			triggers:       &policy.WorkflowTriggers{Push: &policy.PushTrigger{}},
			expectFindings: 1,
		},
		{
			name:           "pull_request without concurrency triggers warning",
			triggers:       &policy.WorkflowTriggers{PullRequest: &policy.PullRequestTrigger{}},
			expectFindings: 1,
		},
		{
			name:           "workflow_dispatch without concurrency is ok",
			triggers:       &policy.WorkflowTriggers{WorkflowDispatch: &policy.WorkflowDispatchTrigger{}},
			expectFindings: 0,
		},
		{
			name:           "no triggers without concurrency is ok",
			triggers:       nil,
			expectFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := &policy.Workflow{
				Path:        "test.yml",
				On:          tt.triggers,
				Permissions: &policy.Permissions{Contents: "read"},
				Concurrency: tt.workflowConcurrency,
				Jobs: map[string]*policy.Job{
					"test": {
						Concurrency: tt.jobConcurrency,
						Steps: []*policy.Step{
							{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
						},
					},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			concurrencyFindings := filterByRuleID(findings, "policy/concurrency-defined")
			assert.Len(t, concurrencyFindings, tt.expectFindings)
		})
	}
}

// TestCheckMinimalPermissions tests the minimal permissions policy.
func TestCheckMinimalPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		permissions    *policy.Permissions
		expectFindings int
	}{
		{
			name:           "specific permissions pass",
			permissions:    &policy.Permissions{Contents: "read", Actions: "read"},
			expectFindings: 0,
		},
		{
			name:           "write-all triggers warning",
			permissions:    &policy.Permissions{All: "write-all"},
			expectFindings: 1,
		},
		{
			name:           "read-all is acceptable",
			permissions:    &policy.Permissions{All: "read-all"},
			expectFindings: 0,
		},
		{
			name:           "no permissions triggers explicit-permissions instead",
			permissions:    nil,
			expectFindings: 0, // This triggers explicit-permissions, not minimal-permissions
		},
		{
			name:           "empty permissions object is ok",
			permissions:    &policy.Permissions{},
			expectFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := &policy.Workflow{
				Path:        "test.yml",
				Permissions: tt.permissions,
				Jobs: map[string]*policy.Job{
					"test": {
						Steps: []*policy.Step{
							{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
						},
					},
				},
			}

			findings := evaluateWorkflow(t, workflow)
			minimalFindings := filterByRuleID(findings, "policy/minimal-permissions")
			assert.Len(t, minimalFindings, tt.expectFindings)
		})
	}
}
