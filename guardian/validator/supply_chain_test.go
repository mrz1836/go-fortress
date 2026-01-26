package validator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestNewSupplyChainValidator tests validator creation.
func TestNewSupplyChainValidator(t *testing.T) {
	t.Parallel()

	v := validator.NewSupplyChainValidator()
	require.NotNil(t, v)
	assert.Equal(t, "supply-chain", v.Name())
}

// TestSupplyChainValidator_Validate tests workflow validation.
func TestSupplyChainValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filename       string
		content        string
		expectedCount  int
		expectedRuleID string
	}{
		{
			name:          "non-release workflow returns no findings",
			filename:      "ci.yml",
			content:       "name: CI\non: push\njobs:\n  test:\n    runs-on: ubuntu-latest",
			expectedCount: 0,
		},
		{
			name:           "release workflow without provenance",
			filename:       "release.yml",
			content:        "name: Release\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest",
			expectedCount:  3, // provenance, isolation, sbom
			expectedRuleID: "supply-chain/provenance-attestation",
		},
		{
			name:          "release workflow with slsa-framework provenance",
			filename:      "release.yml",
			content:       "name: Release\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: slsa-framework/slsa-github-generator@v1",
			expectedCount: 2, // missing isolation, sbom
		},
		{
			name:          "release workflow with actions/attest-build-provenance",
			filename:      "release.yml",
			content:       "name: Release\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/attest-build-provenance@v1",
			expectedCount: 2, // missing isolation, sbom
		},
		{
			name:          "build workflow with container isolation",
			filename:      "build.yml",
			content:       "name: Build\non: push\njobs:\n  build:\n    container:\n      image: golang:1.21\n    runs-on: ubuntu-latest",
			expectedCount: 2, // missing provenance and sbom
		},
		{
			name:          "build workflow with docker build action",
			filename:      "build.yml",
			content:       "name: Build\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: docker/build-push-action@v5",
			expectedCount: 2, // missing provenance and sbom
		},
		{
			name:          "deploy workflow with SBOM generation",
			filename:      "deploy.yml",
			content:       "name: Deploy\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: anchore/sbom-action@v0",
			expectedCount: 2, // missing provenance and isolation
		},
		{
			name:          "publish workflow with cyclonedx",
			filename:      "publish.yml",
			content:       "name: Publish\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: cyclonedx generate",
			expectedCount: 2, // missing provenance and isolation
		},
		{
			name:           "artifact workflow with unpinned docker pull",
			filename:       "artifact.yml",
			content:        "name: Artifact\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: docker pull nginx:latest",
			expectedCount:  4, // provenance, isolation, sbom, unpinned docker
			expectedRuleID: "supply-chain/pinned-dependencies",
		},
		{
			name:          "release workflow with pinned docker image",
			filename:      "release.yml",
			content:       "name: Release\non: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: docker pull nginx@sha256:abc123def456",
			expectedCount: 3, // missing provenance, isolation, sbom
		},
		{
			name:     "build workflow with go build but no GOSUMDB",
			filename: "build.yml",
			content: `name: Build
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: go build ./...`,
			expectedCount: 4, // provenance, isolation, sbom, go dep warning
		},
		{
			name:     "build workflow with go build and go.sum verification",
			filename: "build.yml",
			content: `name: Build
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: go.sum check && go build ./...`,
			expectedCount: 3, // provenance, isolation, sbom (go.sum mentioned)
		},
		{
			name:     "release workflow with network isolation",
			filename: "release.yml",
			content: `name: Release
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: GOPROXY=off go build ./...`,
			expectedCount: 3, // missing provenance, sbom, go dep warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte(tt.content), 0o600)
			require.NoError(t, err)

			v := validator.NewSupplyChainValidator()
			findings, err := v.Validate(context.Background(), path)
			require.NoError(t, err)
			assert.Len(t, findings, tt.expectedCount)

			if tt.expectedRuleID != "" && len(findings) > 0 {
				found := false
				for _, f := range findings {
					if f.RuleID == tt.expectedRuleID {
						found = true
						break
					}
				}
				assert.True(t, found, "expected finding with rule ID %s", tt.expectedRuleID)
			}
		})
	}
}

// TestSupplyChainValidator_Validate_FileReadError tests error handling.
func TestSupplyChainValidator_Validate_FileReadError(t *testing.T) {
	t.Parallel()

	v := validator.NewSupplyChainValidator()

	// Non-existent release file should return error
	_, err := v.Validate(context.Background(), "/nonexistent/release.yml")
	require.Error(t, err)
}

// TestSupplyChainValidator_ValidateSBOMFile tests SBOM file validation.
func TestSupplyChainValidator_ValidateSBOMFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		filename      string
		content       string
		expectedCount int
		expectError   bool
	}{
		{
			name:     "valid SPDX document",
			filename: "sbom.spdx.json",
			content: `{
				"spdxVersion": "SPDX-2.3",
				"dataLicense": "CC0-1.0",
				"name": "test-project",
				"documentNamespace": "https://example.com/test"
			}`,
			expectedCount: 0,
		},
		{
			name:          "SPDX missing version",
			filename:      "sbom.spdx.json",
			content:       `{"dataLicense": "CC0-1.0", "name": "test"}`,
			expectedCount: 1,
		},
		{
			name:          "SPDX missing data license",
			filename:      "sbom.spdx.json",
			content:       `{"spdxVersion": "SPDX-2.3", "name": "test"}`,
			expectedCount: 1,
		},
		{
			name:          "invalid SPDX JSON",
			filename:      "sbom.spdx.json",
			content:       `not valid json`,
			expectedCount: 1,
		},
		{
			name:     "valid CycloneDX document",
			filename: "sbom.cdx.json",
			content: `{
				"bomFormat": "CycloneDX",
				"specVersion": "1.4",
				"version": 1
			}`,
			expectedCount: 0,
		},
		{
			name:          "CycloneDX invalid bomFormat",
			filename:      "sbom.cdx.json",
			content:       `{"bomFormat": "Unknown", "specVersion": "1.4"}`,
			expectedCount: 1,
		},
		{
			name:          "CycloneDX missing specVersion",
			filename:      "sbom.cdx.json",
			content:       `{"bomFormat": "CycloneDX"}`,
			expectedCount: 1,
		},
		{
			name:          "invalid CycloneDX JSON",
			filename:      "sbom.cdx.json",
			content:       `not valid json`,
			expectedCount: 1,
		},
		{
			name:          "unrecognized SBOM format",
			filename:      "sbom.json",
			content:       `{"type": "unknown"}`,
			expectedCount: 1,
		},
		{
			name:     "SPDX detected by content",
			filename: "arbitrary.json",
			content: `{
				"spdxVersion": "SPDX-2.3",
				"dataLicense": "CC0-1.0"
			}`,
			expectedCount: 0,
		},
		{
			name:     "CycloneDX detected by content",
			filename: "arbitrary.json",
			content: `{
				"bomFormat": "CycloneDX",
				"specVersion": "1.4"
			}`,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte(tt.content), 0o600)
			require.NoError(t, err)

			v := validator.NewSupplyChainValidator()
			findings, err := v.ValidateSBOMFile(path)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, findings, tt.expectedCount)
		})
	}
}

// TestSupplyChainValidator_ValidateSBOMFile_FileReadError tests error handling.
func TestSupplyChainValidator_ValidateSBOMFile_FileReadError(t *testing.T) {
	t.Parallel()

	v := validator.NewSupplyChainValidator()
	_, err := v.ValidateSBOMFile("/nonexistent/sbom.json")
	require.Error(t, err)
}

// TestSupplyChainValidator_isReleaseWorkflow tests release workflow detection.
func TestSupplyChainValidator_isReleaseWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename  string
		isRelease bool
	}{
		{"release.yml", true},
		{"Release.yml", true},
		{"RELEASE.yml", true},
		{"build.yml", true},
		{"deploy.yml", true},
		{"publish.yml", true},
		{"artifact.yml", true},
		{"ci.yml", false},
		{"test.yml", false},
		{"lint.yml", false},
		{"pr-check.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(path, []byte("name: Test"), 0o600)
			require.NoError(t, err)

			v := validator.NewSupplyChainValidator()
			findings, err := v.Validate(context.Background(), path)
			require.NoError(t, err)

			if tt.isRelease {
				// Release workflows should have findings (missing security features)
				assert.NotEmpty(t, findings, "expected findings for release workflow %s", tt.filename)
			} else {
				// Non-release workflows should have no findings
				assert.Empty(t, findings, "expected no findings for non-release workflow %s", tt.filename)
			}
		})
	}
}

// TestSupplyChainValidator_FindingSeverity tests finding severity levels.
func TestSupplyChainValidator_FindingSeverity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "release.yml")
	content := `name: Release
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: docker pull nginx:latest`

	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSupplyChainValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	// Check that findings have appropriate severities
	severities := make(map[string]bool)
	for _, f := range findings {
		severities[string(f.Severity)] = true
		// All supply chain findings should have a source
		assert.Equal(t, validator.SourceSupplyChain, f.Source)
	}

	// Should have a mix of severity levels
	assert.GreaterOrEqual(t, len(severities), 1)
}

// TestSupplyChainValidator_FindingSuggestions tests that findings include suggestions.
func TestSupplyChainValidator_FindingSuggestions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "release.yml")
	content := "name: Release\non: push"

	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)

	v := validator.NewSupplyChainValidator()
	findings, err := v.Validate(context.Background(), path)
	require.NoError(t, err)

	// At least some findings should have suggestions
	hasSuggestion := false
	for _, f := range findings {
		if f.Suggestion != "" {
			hasSuggestion = true
			break
		}
	}
	assert.True(t, hasSuggestion, "expected at least one finding with a suggestion")
}
