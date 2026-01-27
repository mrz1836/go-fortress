# Fortress Guardian

A Go-native CI validation framework that treats GitHub Actions workflows as testable code.

## Overview

Fortress Guardian enables:
- **Local CI reproduction**: Run GitHub Actions locally via `nektos/act`
- **Static validation**: Check workflow syntax and policies without Docker
- **Security policies**: Enforce best practices as code
- **Scenario testing**: Validate CI behaviors with predefined test cases

## Installation

Guardian is part of the go-fortress module. Install dependencies:

```bash
# Install MAGE-X
go install github.com/mrz1836/mage-x/cmd/magex@latest

# Install Guardian dependencies
magex deps:install

# Or manually:
go install github.com/nektos/act@v0.2.84
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.10
```

## Quick Start

### Static Validation (< 2s, no Docker)

```bash
magex ci:static
```

Validates all workflow files for:
- YAML syntax errors (via actionlint)
- Policy violations (SHA-pinned actions, explicit permissions, etc.)
- Deprecated actions and runners
- Schema compliance

### Quick Test (< 60s)

```bash
magex ci:test
```

Runs static validation plus fast failure scenarios.

### Full Verification (< 5min)

```bash
magex ci:verify
```

Executes all 35+ scenarios for comprehensive pre-merge validation.

### List Scenarios

```bash
magex ci:list                   # All scenarios
magex ci:list filter=security   # Filter by category
```

### Run Single Scenario

```bash
magex ci:scenario name=LINT-001              # Basic
magex ci:scenario name=LINT-001 verbose=true # Verbose output
magex ci:scenario name=SEC-001 keep=true     # Keep container for debugging
```

## Package Structure

```text
guardian/
├── guardian.go       # Main entry point, public API
├── config.go         # Configuration loading
├── runner/           # Act wrapper for workflow execution
├── validator/        # Static analysis validators
├── policy/           # Security policies as code
├── reporter/         # Output formatters (terminal, JSONL, SARIF)
└── scenarios/        # Test scenario definitions
```

## API Reference

### Creating a Guardian Instance

```go
import "github.com/mrz1836/go-fortress/guardian"

// With default configuration
g, err := guardian.New(ctx, guardian.DefaultConfig())

// Or load from environment
cfg := guardian.LoadFromEnv()
g, err := guardian.New(ctx, cfg)
```

### Running Validation

```go
ctx := context.Background()

// Static validation only
results, err := g.RunStatic(ctx)

// Quick test (static + fast scenarios)
report, err := g.RunTest(ctx)

// Full verification
report, err := g.RunVerify(ctx)

// Single scenario
result, err := g.RunScenario(ctx, "LINT-001", guardian.ScenarioOptions{
    Verbose: true,
})
```

### Listing Scenarios

```go
scenarios, err := g.ListScenarios(ctx, guardian.ScenarioFilter{
    Category: "security",
    Tags:     []string{"fast"},
})
```

## Configuration

### Environment Variables

```bash
# Execution settings
GUARDIAN_SCENARIO_TIMEOUT=30s
GUARDIAN_STATIC_TIMEOUT=5s
GUARDIAN_PARALLEL_SCENARIOS=4

# Output
GUARDIAN_OUTPUT_DIR=.mage-x
GUARDIAN_SARIF_OUTPUT=guardian.sarif
GUARDIAN_JSONL_OUTPUT=ci-results.jsonl

# Debug options
GUARDIAN_VERBOSE=false
GUARDIAN_DRY_RUN=false
GUARDIAN_KEEP_CONTAINERS=false
GUARDIAN_POLICY_STRICT=false
```

### Policy Exceptions

Create `.github/guardian.yaml` to exempt specific rules:

```yaml
exceptions:
  - policy: sha-pinned-actions
    path: .github/workflows/test.yml
    reason: "Testing unpinned action behavior"
    expires: 2026-06-01
    approved_by: "@username"
    created_at: 2026-01-24
```

## Built-in Policies

| Policy | Severity | Description |
|--------|----------|-------------|
| `sha-pinned-actions` | error | All actions must be SHA-pinned |
| `explicit-permissions` | warning | Workflows should declare permissions |
| `no-dangerous-workflows` | error | No pull_request_target with write |
| `no-secret-logging` | error | Secrets must not be logged |
| `concurrency-defined` | warning | Concurrency groups recommended |
| `minimal-permissions` | warning | Use least-privilege permissions |

## Scenario Categories

- **Quality**: LINT-001 to LINT-004, TEST-001 to TEST-003, RACE-001, COV-001/002, BENCH-001, FUZZ-001
- **Security**: SEC-001 to SEC-003, VULN-001/002, GITLEAKS-001/002
- **Fork Safety**: FORK-001 to FORK-003
- **Config**: MATRIX-001, ENV-001, CACHE-001/002, WORKFLOW-001/002, ACTION-001/002
- **Supply Chain**: SLSA-001 to SLSA-003
- **Pass Cases**: PASS-001/002

## Creating Custom Scenarios

1. Create a fixture directory:
```bash
mkdir -p .github/ci-tester/fixtures/my-fail
```

2. Add a minimal Go module with the failure condition

3. Add a workflow that triggers the failure

4. Register the scenario in `guardian/scenarios/`:
```go
var MyScenario = &Scenario{
    ID:          "MY-001",
    Category:    CategoryQuality,
    Description: "My custom failure scenario",
    FixturePath: ".github/ci-tester/fixtures/my-fail",
    Workflow:    "ci.yml",
    Expected: ExpectedResult{
        Status:      StatusFailure,
        LogPatterns: []string{"expected error pattern"},
    },
}
```

## Report Formats

### Terminal
Human-readable colored output to stdout.

### JSONL
Machine-readable line-delimited JSON at `.mage-x/ci-results.jsonl`.

### SARIF
GitHub Security integration at `.mage-x/guardian.sarif`. Upload with:
```bash
gh api repos/{owner}/{repo}/code-scanning/sarifs -X POST -F sarif=@.mage-x/guardian.sarif
```

### GitHub Annotations
When running in GitHub Actions, findings appear as PR annotations.

## Requirements

- Go 1.24
- Docker 20.10+ (for scenario execution)
- 10 GB disk space (for container images)
- 4 GB memory (for parallel execution)

## License

MIT - See [LICENSE](../LICENSE) for details.
