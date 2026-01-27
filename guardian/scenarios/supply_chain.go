package scenarios

import "time"

// Supply chain scenarios test SLSA compliance and artifact verification.

func registerSupplyChainScenarios(r *Registry) {
	// SLSA-001: Provenance attestation present
	r.Register(&Scenario{
		ID:          "SLSA-001",
		Category:    CategorySupplyChain,
		Description: "Provenance attestation present",
		FixturePath: ".github/ci-tester/fixtures/slsa-test",
		Workflow:    "release.yml",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"provenance|attestation|slsa"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"supply-chain", "slsa"},
	})

	// SLSA-002: Build isolation settings
	r.Register(&Scenario{
		ID:          "SLSA-002",
		Category:    CategorySupplyChain,
		Description: "Build isolation settings validation",
		FixturePath: ".github/ci-tester/fixtures/slsa-test",
		Workflow:    "release.yml",
		Job:         "build",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"isolated|hermetic|reproducible"},
		},
		Timeout: 120 * time.Second,
		Tags:    []string{"supply-chain", "slsa"},
	})

	// SLSA-003: Dependencies pinned to digests
	r.Register(&Scenario{
		ID:          "SLSA-003",
		Category:    CategorySupplyChain,
		Description: "Dependencies pinned to digests",
		FixturePath: ".github/ci-tester/fixtures/slsa-test",
		Workflow:    "release.yml",
		Job:         "verify-deps",
		Expected: ExpectedResult{
			Status:      StatusSuccess,
			LogPatterns: []string{"pinned|digest|sha256"},
		},
		Timeout: 60 * time.Second,
		Tags:    []string{"supply-chain", "slsa"},
	})
}
