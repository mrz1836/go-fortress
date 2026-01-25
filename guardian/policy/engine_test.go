package policy_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/policy"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestNewEngine tests creating a new policy engine.
func TestNewEngine(t *testing.T) {
	t.Parallel()

	engine, err := policy.NewEngine()
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Should have default policies registered
	policies := engine.Policies()
	assert.NotEmpty(t, policies, "should have default policies")

	// Verify known default policies are present
	policyIDs := make(map[string]bool)
	for _, p := range policies {
		policyIDs[p.ID] = true
	}

	expectedPolicies := []string{
		"sha-pinned-actions",
		"explicit-permissions",
		"no-dangerous-workflows",
		"no-secret-logging",
		"concurrency-defined",
		"minimal-permissions",
	}

	for _, expected := range expectedPolicies {
		assert.True(t, policyIDs[expected], "expected policy %q to be registered", expected)
	}
}

// TestEngine_EscalateToError tests severity escalation.
func TestEngine_EscalateToError(t *testing.T) {
	t.Parallel()

	engine, err := policy.NewEngine()
	require.NoError(t, err)

	// Before escalation
	assert.False(t, engine.IsEscalated("explicit-permissions"))

	// Escalate a warning policy to error
	engine.EscalateToError("explicit-permissions")

	// After escalation
	assert.True(t, engine.IsEscalated("explicit-permissions"))
}

// TestEngine_SetSeverity tests custom severity overrides.
func TestEngine_SetSeverity(t *testing.T) {
	t.Parallel()

	engine, err := policy.NewEngine()
	require.NoError(t, err)

	// Set custom severity
	engine.SetSeverity("sha-pinned-actions", validator.SeverityWarning)

	// Create a workflow with unpinned action to trigger the policy
	workflow := &policy.Workflow{
		Path: "test.yml",
		Jobs: map[string]*policy.Job{
			"build": {
				Steps: []*policy.Step{
					{Uses: "actions/checkout@v4", Line: 10},
				},
			},
		},
	}

	ctx := context.Background()
	findings, err := engine.Evaluate(ctx, workflow)
	require.NoError(t, err)

	// Find the sha-pinned-actions finding
	var found *validator.Finding
	for i := range findings {
		if findings[i].RuleID == "policy/sha-pinned-actions" {
			found = &findings[i]
			break
		}
	}

	require.NotNil(t, found, "expected to find sha-pinned-actions violation")
	assert.Equal(t, validator.SeverityWarning, found.Severity,
		"severity should be overridden to warning")
}

// TestEngine_IsEscalated tests checking escalation status.
func TestEngine_IsEscalated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		escalate  []string
		checkID   string
		expectVal bool
	}{
		{
			name:      "not escalated",
			escalate:  []string{},
			checkID:   "some-policy",
			expectVal: false,
		},
		{
			name:      "escalated policy returns true",
			escalate:  []string{"my-policy"},
			checkID:   "my-policy",
			expectVal: true,
		},
		{
			name:      "other policies not affected",
			escalate:  []string{"policy-a"},
			checkID:   "policy-b",
			expectVal: false,
		},
		{
			name:      "multiple escalations",
			escalate:  []string{"policy-a", "policy-b", "policy-c"},
			checkID:   "policy-b",
			expectVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine, err := policy.NewEngine()
			require.NoError(t, err)

			for _, id := range tt.escalate {
				engine.EscalateToError(id)
			}

			result := engine.IsEscalated(tt.checkID)
			assert.Equal(t, tt.expectVal, result)
		})
	}
}

// TestEngine_Evaluate tests evaluating policies against workflows.
func TestEngine_Evaluate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		workflow         *policy.Workflow
		expectedFindings int
		expectRuleIDs    []string
	}{
		{
			name: "compliant workflow has no findings",
			workflow: &policy.Workflow{
				Path: "test.yml",
				Permissions: &policy.Permissions{
					Contents: "read",
				},
				Concurrency: &policy.Concurrency{
					Group: "test-group",
				},
				Jobs: map[string]*policy.Job{
					"build": {
						Steps: []*policy.Step{
							{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
						},
					},
				},
			},
			expectedFindings: 0,
		},
		{
			name: "unpinned action triggers violation",
			workflow: &policy.Workflow{
				Path:        "test.yml",
				Permissions: &policy.Permissions{Contents: "read"},
				Jobs: map[string]*policy.Job{
					"build": {
						Steps: []*policy.Step{
							{Uses: "actions/checkout@v4", Line: 10},
						},
					},
				},
			},
			expectedFindings: 1,
			expectRuleIDs:    []string{"policy/sha-pinned-actions"},
		},
		{
			name: "missing permissions triggers warning",
			workflow: &policy.Workflow{
				Path: "test.yml",
				On: &policy.WorkflowTriggers{
					Push: &policy.PushTrigger{},
				},
				Jobs: map[string]*policy.Job{
					"build": {
						Steps: []*policy.Step{
							{Uses: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd"},
						},
					},
				},
			},
			expectedFindings: 2, // explicit-permissions + concurrency-defined
			expectRuleIDs:    []string{"policy/explicit-permissions", "policy/concurrency-defined"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine, err := policy.NewEngine()
			require.NoError(t, err)

			ctx := context.Background()
			findings, err := engine.Evaluate(ctx, tt.workflow)
			require.NoError(t, err)

			assert.Len(t, findings, tt.expectedFindings)

			if len(tt.expectRuleIDs) > 0 {
				foundRuleIDs := make(map[string]bool)
				for _, f := range findings {
					foundRuleIDs[f.RuleID] = true
				}

				for _, expectedID := range tt.expectRuleIDs {
					assert.True(t, foundRuleIDs[expectedID],
						"expected rule %q to be in findings", expectedID)
				}
			}
		})
	}
}

// TestEngine_RegisterPolicy tests adding custom policies.
func TestEngine_RegisterPolicy(t *testing.T) {
	t.Parallel()

	engine, err := policy.NewEngine()
	require.NoError(t, err)

	initialCount := len(engine.Policies())

	// Register a custom policy
	customPolicy := &policy.Policy{
		ID:          "custom-test-policy",
		Severity:    validator.SeverityWarning,
		Description: "Test policy for unit tests",
		Tags:        []string{"test"},
		Check: func(_ *policy.Workflow) []validator.Finding {
			return []validator.Finding{
				{
					RuleID:   "policy/custom-test-policy",
					Severity: validator.SeverityWarning,
					Message:  "custom check triggered",
				},
			}
		},
	}

	engine.RegisterPolicy(customPolicy)

	// Verify policy count increased
	assert.Len(t, engine.Policies(), initialCount+1)

	// Verify custom policy is in the list
	policies := engine.Policies()
	var found bool
	for _, p := range policies {
		if p.ID == "custom-test-policy" {
			found = true
			assert.Equal(t, "warning", p.Severity)
			assert.Equal(t, "Test policy for unit tests", p.Description)
			assert.Contains(t, p.Tags, "test")
			break
		}
	}
	assert.True(t, found, "custom policy should be registered")

	// Verify custom policy is evaluated
	ctx := context.Background()
	findings, err := engine.Evaluate(ctx, &policy.Workflow{Path: "test.yml"})
	require.NoError(t, err)

	var foundFinding bool
	for _, f := range findings {
		if f.RuleID == "policy/custom-test-policy" {
			foundFinding = true
			break
		}
	}
	assert.True(t, foundFinding, "custom policy should produce findings")
}

// TestEngine_LoadExceptions tests loading exceptions from file.
func TestEngine_LoadExceptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expectError bool
	}{
		{
			name: "valid exceptions file",
			yaml: `
exceptions:
  - policy: sha-pinned-actions
    path: test.yml
    reason: "Test exception"
    created_at: 2025-01-01T00:00:00Z
`,
			expectError: false,
		},
		{
			name:        "missing file returns no error",
			yaml:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "guardian.yaml")

			if tt.yaml != "" {
				require.NoError(t, os.WriteFile(configPath, []byte(tt.yaml), 0o600))
			}

			engine, err := policy.NewEngine()
			require.NoError(t, err)

			err = engine.LoadExceptions(context.Background(), configPath)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestEngine_Policies tests listing policy metadata.
func TestEngine_Policies(t *testing.T) {
	t.Parallel()

	engine, err := policy.NewEngine()
	require.NoError(t, err)

	policies := engine.Policies()
	require.NotEmpty(t, policies)

	// Verify policy info fields are populated
	for _, p := range policies {
		assert.NotEmpty(t, p.ID, "policy ID should not be empty")
		assert.NotEmpty(t, p.Severity, "policy severity should not be empty")
		assert.NotEmpty(t, p.Description, "policy description should not be empty")
	}

	// Check specific known policy
	var shaPinned *policy.Info
	for i := range policies {
		if policies[i].ID == "sha-pinned-actions" {
			shaPinned = &policies[i]
			break
		}
	}

	require.NotNil(t, shaPinned)
	assert.Equal(t, "error", shaPinned.Severity)
	assert.Contains(t, shaPinned.Description, "SHA")
	assert.NotEmpty(t, shaPinned.HelpURL)
	assert.Contains(t, shaPinned.Tags, "security")
}
