package scenarios_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/scenarios"
)

// TestScenarioRegistry_NewRegistry tests creating a new registry.
func TestScenarioRegistry_NewRegistry(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()
	require.NotNil(t, reg)

	// Empty registry should return empty counts
	assert.Equal(t, 0, reg.Count())
	assert.Equal(t, 0, reg.EnabledCount())
	assert.Empty(t, reg.All())
}

// TestScenarioRegistry_RegisterAndGet tests registering and retrieving scenarios.
func TestScenarioRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scenarios []*scenarios.Scenario
		getIDs    []string
		expectOK  []bool
	}{
		{
			name: "register single scenario",
			scenarios: []*scenarios.Scenario{
				{ID: "TEST-001", Description: "Test scenario"},
			},
			getIDs:   []string{"TEST-001", "TEST-002"},
			expectOK: []bool{true, false},
		},
		{
			name: "register multiple scenarios",
			scenarios: []*scenarios.Scenario{
				{ID: "LINT-001"},
				{ID: "LINT-002"},
				{ID: "SEC-001"},
			},
			getIDs:   []string{"LINT-001", "LINT-002", "SEC-001", "SEC-002"},
			expectOK: []bool{true, true, true, false},
		},
		{
			name: "re-register same scenario replaces it",
			scenarios: []*scenarios.Scenario{
				{ID: "TEST-001", Description: "Original"},
				{ID: "TEST-001", Description: "Replaced"},
			},
			getIDs:   []string{"TEST-001"},
			expectOK: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := scenarios.NewRegistry()

			for _, s := range tt.scenarios {
				reg.Register(s)
			}

			for i, id := range tt.getIDs {
				s, ok := reg.Get(id)
				assert.Equal(t, tt.expectOK[i], ok, "Get(%q) ok mismatch", id)
				if ok {
					assert.NotNil(t, s)
					assert.Equal(t, id, s.ID)
				} else {
					assert.Nil(t, s)
				}
			}
		})
	}
}

// TestScenarioRegistry_All tests that All returns only enabled scenarios in order.
func TestScenarioRegistry_All(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	// Register scenarios in specific order
	reg.Register(&scenarios.Scenario{ID: "A-001"})
	reg.Register(&scenarios.Scenario{ID: "B-001", Disabled: true}) // Disabled
	reg.Register(&scenarios.Scenario{ID: "C-001"})

	all := reg.All()
	require.Len(t, all, 2) // Only enabled scenarios

	// Verify order is preserved
	assert.Equal(t, "A-001", all[0].ID)
	assert.Equal(t, "C-001", all[1].ID)
}

// TestScenarioRegistry_All_ReregisterPreservesOrder tests that re-registering doesn't change order.
func TestScenarioRegistry_All_ReregisterPreservesOrder(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	reg.Register(&scenarios.Scenario{ID: "a"})
	reg.Register(&scenarios.Scenario{ID: "b"})
	reg.Register(&scenarios.Scenario{ID: "c"})
	reg.Register(&scenarios.Scenario{ID: "b", Description: "updated"}) // Re-register b

	all := reg.All()
	require.Len(t, all, 3)

	// Order should still be a, b, c (not a, c, b)
	assert.Equal(t, "a", all[0].ID)
	assert.Equal(t, "b", all[1].ID)
	assert.Equal(t, "c", all[2].ID)
	assert.Equal(t, "updated", all[1].Description)
}

// TestScenarioRegistry_ByCategory tests filtering by category.
func TestScenarioRegistry_ByCategory(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	reg.Register(&scenarios.Scenario{ID: "QUAL-001", Category: scenarios.CategoryQuality})
	reg.Register(&scenarios.Scenario{ID: "QUAL-002", Category: scenarios.CategoryQuality})
	reg.Register(&scenarios.Scenario{ID: "SEC-001", Category: scenarios.CategorySecurity})
	reg.Register(&scenarios.Scenario{ID: "QUAL-003", Category: scenarios.CategoryQuality, Disabled: true})

	// Get quality scenarios
	quality := reg.ByCategory(scenarios.CategoryQuality)
	require.Len(t, quality, 2) // QUAL-003 is disabled
	assert.Equal(t, "QUAL-001", quality[0].ID)
	assert.Equal(t, "QUAL-002", quality[1].ID)

	// Get security scenarios
	security := reg.ByCategory(scenarios.CategorySecurity)
	require.Len(t, security, 1)
	assert.Equal(t, "SEC-001", security[0].ID)

	// Get empty category
	testingCategory := reg.ByCategory(scenarios.CategoryTesting)
	assert.Empty(t, testingCategory)
}

// TestScenarioRegistry_ByTags tests filtering by tags.
func TestScenarioRegistry_ByTags(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	reg.Register(&scenarios.Scenario{ID: "S-001", Tags: []string{"fast", "security"}})
	reg.Register(&scenarios.Scenario{ID: "S-002", Tags: []string{"slow", "security"}})
	reg.Register(&scenarios.Scenario{ID: "S-003", Tags: []string{"fast", "quality"}})
	reg.Register(&scenarios.Scenario{ID: "S-004", Tags: []string{"fast", "security"}, Disabled: true})

	tests := []struct {
		name        string
		tags        []string
		expectedIDs []string
	}{
		{
			name:        "single tag matches multiple",
			tags:        []string{"fast"},
			expectedIDs: []string{"S-001", "S-003"}, // S-004 is disabled
		},
		{
			name:        "multiple tags narrows results",
			tags:        []string{"fast", "security"},
			expectedIDs: []string{"S-001"}, // S-004 is disabled
		},
		{
			name:        "empty tags returns all enabled",
			tags:        []string{},
			expectedIDs: []string{"S-001", "S-002", "S-003"},
		},
		{
			name:        "non-matching tag returns empty",
			tags:        []string{"nonexistent"},
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := reg.ByTags(tt.tags)
			require.Len(t, result, len(tt.expectedIDs))

			for i, expected := range tt.expectedIDs {
				assert.Equal(t, expected, result[i].ID)
			}
		})
	}
}

// TestScenarioRegistry_Count tests counting scenarios.
func TestScenarioRegistry_Count(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	assert.Equal(t, 0, reg.Count())

	reg.Register(&scenarios.Scenario{ID: "S-001"})
	assert.Equal(t, 1, reg.Count())

	reg.Register(&scenarios.Scenario{ID: "S-002", Disabled: true})
	assert.Equal(t, 2, reg.Count()) // Count includes disabled

	reg.Register(&scenarios.Scenario{ID: "S-003"})
	assert.Equal(t, 3, reg.Count())
}

// TestScenarioRegistry_EnabledCount tests counting enabled scenarios.
func TestScenarioRegistry_EnabledCount(t *testing.T) {
	t.Parallel()

	reg := scenarios.NewRegistry()

	assert.Equal(t, 0, reg.EnabledCount())

	reg.Register(&scenarios.Scenario{ID: "S-001"})
	assert.Equal(t, 1, reg.EnabledCount())

	reg.Register(&scenarios.Scenario{ID: "S-002", Disabled: true})
	assert.Equal(t, 1, reg.EnabledCount()) // Disabled not counted

	reg.Register(&scenarios.Scenario{ID: "S-003"})
	assert.Equal(t, 2, reg.EnabledCount())

	reg.Register(&scenarios.Scenario{ID: "S-004", Disabled: true})
	assert.Equal(t, 2, reg.EnabledCount())
}

// TestCategories tests the Categories function.
func TestCategories(t *testing.T) {
	t.Parallel()

	categories := scenarios.Categories()
	require.NotEmpty(t, categories)

	// Verify all expected categories are present
	expectedCategories := []scenarios.Category{
		scenarios.CategoryQuality,
		scenarios.CategoryTesting,
		scenarios.CategorySecurity,
		scenarios.CategoryCoverage,
		scenarios.CategoryForkSafety,
		scenarios.CategoryConfig,
		scenarios.CategoryServices,
		scenarios.CategoryTooling,
		scenarios.CategoryArtifacts,
		scenarios.CategorySupplyChain,
	}

	assert.Len(t, categories, len(expectedCategories))
	for i, expected := range expectedCategories {
		assert.Equal(t, expected, categories[i])
	}
}

// TestHasAllTags tests the tag matching logic through ByTags.
func TestHasAllTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scenarioTags []string
		requiredTags []string
		expectMatch  bool
	}{
		{
			name:         "empty required matches any",
			scenarioTags: []string{"a", "b"},
			requiredTags: []string{},
			expectMatch:  true,
		},
		{
			name:         "exact match",
			scenarioTags: []string{"a", "b"},
			requiredTags: []string{"a", "b"},
			expectMatch:  true,
		},
		{
			name:         "subset required matches superset scenario",
			scenarioTags: []string{"a", "b", "c"},
			requiredTags: []string{"a", "b"},
			expectMatch:  true,
		},
		{
			name:         "superset required doesn't match subset scenario",
			scenarioTags: []string{"a"},
			requiredTags: []string{"a", "b"},
			expectMatch:  false,
		},
		{
			name:         "no common tags",
			scenarioTags: []string{"x", "y"},
			requiredTags: []string{"a", "b"},
			expectMatch:  false,
		},
		{
			name:         "partial match fails",
			scenarioTags: []string{"a", "c"},
			requiredTags: []string{"a", "b"},
			expectMatch:  false,
		},
		{
			name:         "empty scenario tags with required fails",
			scenarioTags: []string{},
			requiredTags: []string{"a"},
			expectMatch:  false,
		},
		{
			name:         "both empty matches",
			scenarioTags: []string{},
			requiredTags: []string{},
			expectMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := scenarios.NewRegistry()
			reg.Register(&scenarios.Scenario{
				ID:   "TEST",
				Tags: tt.scenarioTags,
			})

			result := reg.ByTags(tt.requiredTags)
			if tt.expectMatch {
				require.Len(t, result, 1)
				assert.Equal(t, "TEST", result[0].ID)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}
