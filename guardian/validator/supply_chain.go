package validator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SourceSupplyChain identifies findings from supply chain validation.
const SourceSupplyChain FindingSource = "supply-chain"

// SupplyChainValidator validates supply chain security requirements.
// It checks for SLSA compliance, provenance attestations, and build isolation.
type SupplyChainValidator struct{}

// NewSupplyChainValidator creates a new supply chain validator.
func NewSupplyChainValidator() *SupplyChainValidator {
	return &SupplyChainValidator{}
}

// Name returns the validator identifier.
func (v *SupplyChainValidator) Name() string {
	return "supply-chain"
}

// Validate checks a workflow file for supply chain security issues.
func (v *SupplyChainValidator) Validate(ctx context.Context, path string) ([]Finding, error) {
	_ = ctx // reserved for future async validation

	findings := make([]Finding, 0, 4) //nolint:mnd // preallocate for typical number of checks

	// Check if this is a release/build workflow
	if !v.isReleaseWorkflow(path) {
		return findings, nil
	}

	content, err := os.ReadFile(path) //nolint:gosec // path from trusted workflow discovery
	if err != nil {
		return nil, err
	}

	// Check for provenance attestation
	provenanceFindings := v.checkProvenanceAttestation(path, string(content))
	findings = append(findings, provenanceFindings...)

	// Check for build isolation settings
	isolationFindings := v.checkBuildIsolation(path, string(content))
	findings = append(findings, isolationFindings...)

	// Check for pinned dependencies
	pinnedFindings := v.checkPinnedDependencies(path, string(content))
	findings = append(findings, pinnedFindings...)

	// Check for SBOM generation
	sbomFindings := v.checkSBOMGeneration(path, string(content))
	findings = append(findings, sbomFindings...)

	return findings, nil
}

// ValidateSBOMFile checks an SBOM file for compliance.
func (v *SupplyChainValidator) ValidateSBOMFile(path string) ([]Finding, error) {
	findings := make([]Finding, 0, 1) //nolint:mnd // preallocate for typical findings

	content, err := os.ReadFile(path) //nolint:gosec // path from SBOM validation API
	if err != nil {
		return nil, err
	}

	// Detect format
	if strings.HasSuffix(path, ".spdx.json") || strings.Contains(string(content), "spdxVersion") {
		return v.validateSPDX(path, content)
	}

	if strings.HasSuffix(path, ".cdx.json") || strings.Contains(string(content), "bomFormat") {
		return v.validateCycloneDX(path, content)
	}

	findings = append(findings, Finding{
		RuleID:   "supply-chain/sbom-format",
		Severity: SeverityWarning,
		Message:  "unrecognized SBOM format",
		File:     path,
		Line:     1,
		Source:   SourceSupplyChain,
	})

	return findings, nil
}

// isReleaseWorkflow checks if the workflow is related to releases/builds.
func (v *SupplyChainValidator) isReleaseWorkflow(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	releasePatterns := []string{"release", "build", "deploy", "publish", "artifact"}

	for _, pattern := range releasePatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

// checkProvenanceAttestation verifies provenance attestation configuration.
func (v *SupplyChainValidator) checkProvenanceAttestation(path, content string) []Finding {
	var findings []Finding

	// Check for SLSA provenance action
	slsaPatterns := []string{
		"slsa-framework/slsa-github-generator",
		"actions/attest-build-provenance",
		"sigstore/cosign",
	}

	hasProvenance := false

	for _, pattern := range slsaPatterns {
		if strings.Contains(content, pattern) {
			hasProvenance = true

			break
		}
	}

	if !hasProvenance {
		findings = append(findings, Finding{
			RuleID:     "supply-chain/provenance-attestation",
			Severity:   SeverityWarning,
			Message:    "release workflow does not generate provenance attestation",
			File:       path,
			Line:       1,
			Source:     SourceSupplyChain,
			Suggestion: "Add SLSA provenance generation using slsa-framework/slsa-github-generator or actions/attest-build-provenance",
		})
	}

	return findings
}

// checkBuildIsolation verifies build isolation settings.
func (v *SupplyChainValidator) checkBuildIsolation(path, content string) []Finding {
	var findings []Finding

	// Check for hermetic build indicators
	hermeticPatterns := []string{
		"network: none",
		"--network=none",
		"GOPROXY=off",
		"GOFLAGS=-mod=vendor",
	}

	hasIsolation := false

	for _, pattern := range hermeticPatterns {
		if strings.Contains(content, pattern) {
			hasIsolation = true

			break
		}
	}

	// Also check for container-based builds which provide isolation
	containerPatterns := []string{
		"container:",
		"docker build",
		"uses: docker/build-push-action",
	}

	for _, pattern := range containerPatterns {
		if strings.Contains(content, pattern) {
			hasIsolation = true

			break
		}
	}

	if !hasIsolation {
		findings = append(findings, Finding{
			RuleID:     "supply-chain/build-isolation",
			Severity:   SeverityNote,
			Message:    "build workflow may not have network isolation",
			File:       path,
			Line:       1,
			Source:     SourceSupplyChain,
			Suggestion: "Consider using hermetic builds with network isolation for SLSA Level 3 compliance",
		})
	}

	return findings
}

// checkPinnedDependencies verifies dependencies are pinned to digests.
func (v *SupplyChainValidator) checkPinnedDependencies(path, content string) []Finding {
	var findings []Finding

	// Check for Docker image digests
	dockerPullPattern := regexp.MustCompile(`docker pull ([^\s]+)`)
	matches := dockerPullPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			image := match[1]
			// Check if image is pinned to digest
			if !strings.Contains(image, "@sha256:") {
				findings = append(findings, Finding{
					RuleID:     "supply-chain/pinned-dependencies",
					Severity:   SeverityWarning,
					Message:    "Docker image not pinned to digest: " + image,
					File:       path,
					Line:       1,
					Source:     SourceSupplyChain,
					Suggestion: "Pin Docker images to SHA256 digest for reproducible builds",
				})
			}
		}
	}

	// Check for Go module verification
	if strings.Contains(content, "go build") || strings.Contains(content, "go install") {
		if !strings.Contains(content, "GOSUMDB") && !strings.Contains(content, "go.sum") {
			findings = append(findings, Finding{
				RuleID:     "supply-chain/pinned-dependencies",
				Severity:   SeverityNote,
				Message:    "Go build may not verify module checksums",
				File:       path,
				Line:       1,
				Source:     SourceSupplyChain,
				Suggestion: "Ensure go.sum is committed and GOSUMDB is enabled for module verification",
			})
		}
	}

	return findings
}

// checkSBOMGeneration verifies SBOM generation configuration.
func (v *SupplyChainValidator) checkSBOMGeneration(path, content string) []Finding {
	var findings []Finding

	// Check for SBOM generation actions/tools
	sbomPatterns := []string{
		"anchore/sbom-action",
		"cyclonedx",
		"spdx",
		"syft",
		"trivy",
		"sbom",
	}

	hasSBOM := false

	for _, pattern := range sbomPatterns {
		if strings.Contains(strings.ToLower(content), pattern) {
			hasSBOM = true

			break
		}
	}

	if !hasSBOM {
		findings = append(findings, Finding{
			RuleID:     "supply-chain/sbom-generation",
			Severity:   SeverityNote,
			Message:    "release workflow does not generate SBOM",
			File:       path,
			Line:       1,
			Source:     SourceSupplyChain,
			Suggestion: "Consider generating SBOM using anchore/sbom-action or similar tool for supply chain transparency",
		})
	}

	return findings
}

// validateSPDX checks SPDX format compliance.
func (v *SupplyChainValidator) validateSPDX(path string, content []byte) ([]Finding, error) {
	var findings []Finding

	var spdx struct {
		SPDXVersion       string `json:"spdxVersion"`
		DataLicense       string `json:"dataLicense"`
		Name              string `json:"name"`
		DocumentNamespace string `json:"documentNamespace"`
	}

	if err := json.Unmarshal(content, &spdx); err != nil {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityError,
			Message:  "invalid SPDX JSON format",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
		// Return findings with error converted to finding - intentionally not returning err
		return findings, nil //nolint:nilerr // error converted to finding
	}

	// Check required fields
	if spdx.SPDXVersion == "" {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityError,
			Message:  "SPDX document missing spdxVersion",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
	}

	if spdx.DataLicense == "" {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityWarning,
			Message:  "SPDX document missing dataLicense",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
	}

	return findings, nil
}

// validateCycloneDX checks CycloneDX format compliance.
func (v *SupplyChainValidator) validateCycloneDX(path string, content []byte) ([]Finding, error) {
	var findings []Finding

	var cdx struct {
		BomFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
		Version     int    `json:"version"`
	}

	if err := json.Unmarshal(content, &cdx); err != nil {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityError,
			Message:  "invalid CycloneDX JSON format",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
		// Return findings with error converted to finding - intentionally not returning err
		return findings, nil //nolint:nilerr // error converted to finding
	}

	// Check required fields
	if cdx.BomFormat != "CycloneDX" {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityError,
			Message:  "CycloneDX document has invalid bomFormat",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
	}

	if cdx.SpecVersion == "" {
		findings = append(findings, Finding{
			RuleID:   "supply-chain/sbom-format",
			Severity: SeverityError,
			Message:  "CycloneDX document missing specVersion",
			File:     path,
			Line:     1,
			Source:   SourceSupplyChain,
		})
	}

	return findings, nil
}
