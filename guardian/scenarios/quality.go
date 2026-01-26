package scenarios

import "time"

// Quality scenarios test code quality checks like linting, testing, and coverage.

func registerQualityScenarios(r *Registry) {
	// LINT-001: Unused variable detection
	r.Register(&Scenario{
		ID:          "LINT-001",
		Category:    CategoryQuality,
		Description: "Unused variable detection",
		FixturePath: ".github/ci-tester/fixtures/lint-fail",
		Workflow:    "ci.yml",
		Job:         "lint",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"declared and not used", "unusedVar"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fast", "lint", "p1"},
	})

	// LINT-002: Gofmt formatting violation
	r.Register(&Scenario{
		ID:          "LINT-002",
		Category:    CategoryQuality,
		Description: "Gofmt formatting violation",
		FixturePath: ".github/ci-tester/fixtures/lint-fail",
		Workflow:    "ci.yml",
		Job:         "format",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"gofmt|File is not.*formatted"},
		},
		Timeout: 30 * time.Second,
		Tags:    []string{"lint"},
	})

	// LINT-003: golangci-lint violation
	r.Register(&Scenario{
		ID:          "LINT-003",
		Category:    CategoryQuality,
		Description: "golangci-lint rule violation",
		FixturePath: ".github/ci-tester/fixtures/lint-fail",
		Workflow:    "ci.yml",
		Job:         "lint",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"golangci-lint|staticcheck|govet"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"lint"},
	})

	// LINT-004: staticcheck violations
	r.Register(&Scenario{
		ID:          "LINT-004",
		Category:    CategoryQuality,
		Description: "staticcheck violations",
		FixturePath: ".github/ci-tester/fixtures/lint-fail",
		Workflow:    "ci.yml",
		Job:         "staticcheck",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"SA|ST|S1|QF|U1000"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"lint"},
	})

	// TEST-001: Failing unit test
	r.Register(&Scenario{
		ID:          "TEST-001",
		Category:    CategoryTesting,
		Description: "Failing unit test assertion",
		FixturePath: ".github/ci-tester/fixtures/test-fail",
		Workflow:    "ci.yml",
		Job:         "test",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"FAIL|assert|Error"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fast", "test", "p1"},
	})

	// TEST-002: Test panic
	r.Register(&Scenario{
		ID:          "TEST-002",
		Category:    CategoryTesting,
		Description: "Test panic (nil pointer)",
		FixturePath: ".github/ci-tester/fixtures/test-fail",
		Workflow:    "ci.yml",
		Job:         "test-panic",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"panic|nil pointer|runtime error"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"test"},
	})

	// TEST-003: Test timeout exceeded
	r.Register(&Scenario{
		ID:          "TEST-003",
		Category:    CategoryTesting,
		Description: "Test timeout exceeded",
		FixturePath: ".github/ci-tester/fixtures/test-fail",
		Workflow:    "ci.yml",
		Job:         "test-timeout",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"test.*timeout|timed out"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"test", "slow"},
	})

	// RACE-001: Data race detection
	r.Register(&Scenario{
		ID:          "RACE-001",
		Category:    CategoryQuality,
		Description: "Data race detection",
		FixturePath: ".github/ci-tester/fixtures/race-fail",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"DATA RACE|race detected|concurrent.*write"},
		},
		Timeout: 90 * time.Second,
		Tags:    []string{"race", "p1"},
	})

	// COV-001: Low coverage threshold
	r.Register(&Scenario{
		ID:          "COV-001",
		Category:    CategoryCoverage,
		Description: "Coverage below threshold",
		FixturePath: ".github/ci-tester/fixtures/cov-fail",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"coverage.*below|threshold|insufficient"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"coverage"},
	})

	// COV-002: Coverage threshold met (success case)
	r.Register(&Scenario{
		ID:          "COV-002",
		Category:    CategoryCoverage,
		Description: "Coverage threshold met",
		FixturePath: ".github/ci-tester/fixtures/pass-basic",
		Workflow:    "ci.yml",
		Job:         "coverage",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"coverage|ok"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"coverage"},
	})

	// BENCH-001: Benchmark execution
	r.Register(&Scenario{
		ID:          "BENCH-001",
		Category:    CategoryQuality,
		Description: "Benchmark execution",
		FixturePath: ".github/ci-tester/fixtures/pass-basic",
		Workflow:    "ci.yml",
		Job:         "bench",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"Benchmark|ns/op"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"bench", "slow"},
	})

	// FUZZ-001: Fuzz testing
	r.Register(&Scenario{
		ID:          "FUZZ-001",
		Category:    CategoryQuality,
		Description: "Fuzz testing execution",
		FixturePath: ".github/ci-tester/fixtures/pass-basic",
		Workflow:    "ci.yml",
		Job:         "fuzz",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"Fuzz|fuzz"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"fuzz", "slow"},
	})

	// PASS-001: Clean code passes all checks
	r.Register(&Scenario{
		ID:          "PASS-001",
		Category:    CategoryQuality,
		Description: "Clean code passes all checks",
		FixturePath: ".github/ci-tester/fixtures/pass-basic",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"ok|PASS|success"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"pass"},
	})

	// PASS-002: Full CI pipeline passes
	r.Register(&Scenario{
		ID:          "PASS-002",
		Category:    CategoryQuality,
		Description: "Full CI pipeline passes",
		FixturePath: ".github/ci-tester/fixtures/pass-basic",
		Workflow:    "ci.yml",
		Expected: ExpectedResult{
			Status: StatusSuccess,
		},
		Timeout: 180 * time.Second,
		Tags:    []string{"pass", "slow"},
	})
}
