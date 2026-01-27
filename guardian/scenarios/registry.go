package scenarios

// Registry manages available scenarios.
type Registry struct {
	scenarios map[string]*Scenario
	order     []string // maintains registration order
}

// NewRegistry creates a new scenario registry.
func NewRegistry() *Registry {
	return &Registry{
		scenarios: make(map[string]*Scenario),
		order:     []string{},
	}
}

// Register adds a scenario to the registry.
func (r *Registry) Register(s *Scenario) {
	if _, exists := r.scenarios[s.ID]; !exists {
		r.order = append(r.order, s.ID)
	}
	r.scenarios[s.ID] = s
}

// Get returns a scenario by ID.
func (r *Registry) Get(id string) (*Scenario, bool) {
	s, ok := r.scenarios[id]
	return s, ok
}

// All returns all registered scenarios in registration order.
func (r *Registry) All() []*Scenario {
	result := make([]*Scenario, 0, len(r.order))
	for _, id := range r.order {
		if s := r.scenarios[id]; !s.Disabled {
			result = append(result, s)
		}
	}
	return result
}

// ByCategory returns scenarios matching the given category.
func (r *Registry) ByCategory(category Category) []*Scenario {
	var result []*Scenario
	for _, id := range r.order {
		s := r.scenarios[id]
		if !s.Disabled && s.Category == category {
			result = append(result, s)
		}
	}
	return result
}

// ByTags returns scenarios that have all the specified tags.
func (r *Registry) ByTags(tags []string) []*Scenario {
	var result []*Scenario
	for _, id := range r.order {
		s := r.scenarios[id]
		if !s.Disabled && hasAllTags(s.Tags, tags) {
			result = append(result, s)
		}
	}
	return result
}

// Count returns the number of registered scenarios.
func (r *Registry) Count() int {
	return len(r.scenarios)
}

// EnabledCount returns the number of enabled scenarios.
func (r *Registry) EnabledCount() int {
	count := 0
	for _, s := range r.scenarios {
		if !s.Disabled {
			count++
		}
	}
	return count
}

// hasAllTags checks if scenario has all required tags.
func hasAllTags(scenarioTags, requiredTags []string) bool {
	if len(requiredTags) == 0 {
		return true
	}

	tagSet := make(map[string]bool)
	for _, t := range scenarioTags {
		tagSet[t] = true
	}

	for _, t := range requiredTags {
		if !tagSet[t] {
			return false
		}
	}

	return true
}

// RegisterAll registers all built-in scenarios.
func RegisterAll(r *Registry) {
	// Quality scenarios
	registerQualityScenarios(r)

	// Security scenarios
	registerSecurityScenarios(r)

	// Fork safety scenarios
	registerForkScenarios(r)

	// Config scenarios
	registerConfigScenarios(r)

	// Supply chain scenarios
	registerSupplyChainScenarios(r)
}

// Categories returns all available categories.
func Categories() []Category {
	return []Category{
		CategoryQuality,
		CategoryTesting,
		CategorySecurity,
		CategoryCoverage,
		CategoryForkSafety,
		CategoryConfig,
		CategoryServices,
		CategoryTooling,
		CategoryArtifacts,
		CategorySupplyChain,
	}
}
