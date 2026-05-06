package scenarios

import "time"

// Fork safety scenarios test fork PR handling and secret protection.

func registerForkScenarios(r *Registry) {
	// FORK-001: Fork detection mechanism
	r.Register(&Scenario{
		ID:          "FORK-001",
		Category:    CategoryForkSafety,
		Description: "Fork detection mechanism validation",
		FixturePath: ".github/ci-tester/fixtures/fork-test",
		Workflow:    workflowCI,
		EventFile:   "testdata/guardian/events/fork-pr.json",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"fork.*detected|is_fork.*true"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fork", tagSecurity},
	})

	// FORK-002: Secret protection on fork PRs
	r.Register(&Scenario{
		ID:          "FORK-002",
		Category:    CategoryForkSafety,
		Description: "Secret protection on fork PRs",
		FixturePath: ".github/ci-tester/fixtures/fork-test",
		Workflow:    workflowCI,
		Job:         "secrets-test",
		EventFile:   "testdata/guardian/events/fork-pr.json",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"skipping.*secret|fork.*no.*secret"},
			ExcludePatterns: []string{
				"SECRET_VALUE", // Actual secret should never appear
			},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fork", tagSecurity},
	})

	// FORK-003: Job skipping for fork PRs
	r.Register(&Scenario{
		ID:          "FORK-003",
		Category:    CategoryForkSafety,
		Description: "Job skipping for fork PRs",
		FixturePath: ".github/ci-tester/fixtures/fork-test",
		Workflow:    workflowCI,
		Job:         "fork-unsafe",
		EventFile:   "testdata/guardian/events/fork-pr.json",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"skipped|skipping.*fork"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fork", tagSecurity},
	})
}
