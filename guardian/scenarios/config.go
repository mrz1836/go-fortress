// Package scenarios provides CI test scenario definitions.
package scenarios

import "time"

// registerConfigScenarios adds config validation test scenarios.

func registerConfigScenarios(r *Registry) {
	// MATRIX-001: Matrix expansion validation
	r.Register(&Scenario{
		ID:          "MATRIX-001",
		Category:    CategoryConfig,
		Description: "Matrix expansion validation",
		FixturePath: ".github/ci-tester/fixtures/matrix-test",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"matrix|os|version"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"config", "matrix"},
	})

	// ENV-001: .env.base schema validation
	r.Register(&Scenario{
		ID:          "ENV-001",
		Category:    CategoryConfig,
		Description: "modular env schema validation (types, required fields, naming conventions)",
		FixturePath: ".github/ci-tester/fixtures/env-test",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"env.*valid|schema.*pass"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"config", "env", "fast"},
	})

	// CACHE-001: Cache hit mode
	r.Register(&Scenario{
		ID:          "CACHE-001",
		Category:    CategoryConfig,
		Description: "Cache hit mode validation",
		FixturePath: ".github/ci-tester/fixtures/cache-test",
		Workflow:    "ci.yml",
		Job:         "cache-hit",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"cache.*hit|restored"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"config", "cache"},
	})

	// CACHE-002: Cache miss mode
	r.Register(&Scenario{
		ID:          "CACHE-002",
		Category:    CategoryConfig,
		Description: "Cache miss mode validation",
		FixturePath: ".github/ci-tester/fixtures/cache-test",
		Workflow:    "ci.yml",
		Job:         "cache-miss",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"cache.*miss|not.*found|saving"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"config", "cache"},
	})

	// WORKFLOW-001: Invalid workflow YAML syntax
	r.Register(&Scenario{
		ID:          "WORKFLOW-001",
		Category:    CategoryConfig,
		Description: "Invalid workflow YAML syntax detection",
		FixturePath: ".github/ci-tester/fixtures/workflow-invalid",
		Workflow:    "invalid.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"yaml|syntax|parse|invalid"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"config", "workflow", "fast"},
	})

	// WORKFLOW-002: Deprecated runner labels
	r.Register(&Scenario{
		ID:          "WORKFLOW-002",
		Category:    CategoryConfig,
		Description: "Deprecated runner labels detection",
		FixturePath: ".github/ci-tester/fixtures/workflow-deprecated",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"deprecated|ubuntu-18|runner"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"config", "workflow"},
	})

	// ACTION-001: Unpinned action detection
	r.Register(&Scenario{
		ID:          "ACTION-001",
		Category:    CategoryConfig,
		Description: "Unpinned action static detection",
		FixturePath: ".github/ci-tester/fixtures/action-unpinned",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"unpinned|not.*pinned|SHA"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"config", "action", "security"},
	})

	// ACTION-002: action.yml schema validation errors
	r.Register(&Scenario{
		ID:          "ACTION-002",
		Category:    CategoryConfig,
		Description: "action.yml schema validation errors",
		FixturePath: ".github/ci-tester/fixtures/action-invalid",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"action.yml|schema|invalid|required"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"config", "action"},
	})
}
