package scenarios

import "time"

// Security scenarios test security scanning and secret detection.

func registerSecurityScenarios(r *Registry) {
	// SEC-001: Hardcoded AWS key pattern
	r.Register(&Scenario{
		ID:          "SEC-001",
		Category:    CategorySecurity,
		Description: "Hardcoded AWS key pattern detection",
		FixturePath: ".github/ci-tester/fixtures/sec-fail",
		Workflow:    workflowCI,
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"AKIA|AWS.*key|secret.*detected|gitleaks"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"fast", tagSecurity, "p1"},
	})

	// SEC-002: Private key detection
	r.Register(&Scenario{
		ID:          "SEC-002",
		Category:    CategorySecurity,
		Description: "Private key in repository detection",
		FixturePath: ".github/ci-tester/fixtures/sec-fail",
		Workflow:    workflowCI,
		Job:         "secrets-scan",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"private.*key|RSA|BEGIN.*PRIVATE"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{tagSecurity},
	})

	// SEC-003: govulncheck findings
	r.Register(&Scenario{
		ID:          "SEC-003",
		Category:    CategorySecurity,
		Description: "govulncheck vulnerability detection",
		FixturePath: ".github/ci-tester/fixtures/vuln-fail",
		Workflow:    workflowCI,
		Job:         "govulncheck",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"GO-|vulnerability|CVE-"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{tagSecurity, "vuln"},
	})

	// VULN-001: Vulnerable dependency
	r.Register(&Scenario{
		ID:          "VULN-001",
		Category:    CategorySecurity,
		Description: "Vulnerable dependency detection",
		FixturePath: ".github/ci-tester/fixtures/vuln-fail",
		Workflow:    workflowCI,
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"CVE-|vulnerability|vulnerable|nancy"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{tagSecurity, "vuln", "p1"},
	})

	// VULN-002: Nancy CVE detection
	r.Register(&Scenario{
		ID:          "VULN-002",
		Category:    CategorySecurity,
		Description: "Nancy CVE detection",
		FixturePath: ".github/ci-tester/fixtures/vuln-fail",
		Workflow:    workflowCI,
		Job:         "nancy",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"CVE-|vulnerable|nancy"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{tagSecurity, "vuln"},
	})

	// GITLEAKS-001: Gitleaks integration
	r.Register(&Scenario{
		ID:          "GITLEAKS-001",
		Category:    CategorySecurity,
		Description: "Gitleaks secret scanning",
		FixturePath: ".github/ci-tester/fixtures/sec-fail",
		Workflow:    workflowCI,
		Job:         "gitleaks",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"gitleaks|leaks? found|secret"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{tagSecurity, "gitleaks"},
	})

	// GITLEAKS-002: Custom patterns
	r.Register(&Scenario{
		ID:          "GITLEAKS-002",
		Category:    CategorySecurity,
		Description: "Gitleaks custom pattern detection",
		FixturePath: ".github/ci-tester/fixtures/sec-fail",
		Workflow:    workflowCI,
		Job:         "gitleaks-custom",
		Expected: ExpectedResult{
			Status:      StatusFailure,
			LogPatterns: []string{"custom.*pattern|secret|leak"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{tagSecurity, "gitleaks"},
	})
}
