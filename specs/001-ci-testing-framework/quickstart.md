# Quickstart: Fortress Guardian CI Testing Framework

**Date**: 2026-01-24
**Branch**: `001-ci-testing-framework`

---

## Overview

Fortress Guardian is a Go-native CI validation framework that tests GitHub Actions workflows as code. This guide covers installation, basic usage, and common workflows.

---

## Prerequisites

| Requirement | Version | Check Command |
|-------------|---------|---------------|
| Go | 1.24+ | `go version` |
| Docker | 20.10+ | `docker --version` |
| MAGE-X | latest | `magex --version` |

**Optional** (installed automatically):
- `nektos/act` - Local GitHub Actions runner
- `actionlint` - Static analyzer for workflows

---

## Installation

Guardian is part of the go-fortress module. No separate installation is required.

### Ensure Tools Are Available

```bash
# Install MAGE-X (if not already installed)
go install github.com/mrz1836/mage-x/cmd/magex@latest

# Install Guardian dependencies
magex deps:install

# Verify act installation
act --version

# Verify actionlint installation
actionlint --version
```

### Tool Version Pinning

Guardian uses pinned tool versions for reproducibility:

```bash
# Tools installed via go install with explicit versions
go install github.com/nektos/act@v0.2.84
go install github.com/rhysd/actionlint/cmd/actionlint@v1.6.27
```

Version configuration in `.env.base`:
- `GUARDIAN_ACT_VERSION=v0.2.84`
- `GUARDIAN_ACTIONLINT_VERSION=v1.6.27`

Override versions in `.env.custom` when needed for testing newer releases.

---

## Basic Usage

### Quick Validation (Recommended for Development)

Run static analysis and fast scenarios:

```bash
magex ci:test
```

**What it does**:
1. Validates workflow YAML syntax
2. Checks for policy violations
3. Runs fast failure scenarios (LINT-001, TEST-001, SEC-001)

**Output**:
```
🧪 CI Test Suite (Fast)
✅ Static validation passed (1.2s)
✅ LINT-001 passed (4.5s)
✅ TEST-001 passed (6.8s)
✅ SEC-001 passed (3.4s)
All scenarios passed (15.9s total)
```

### Static Analysis Only

Fast validation without Docker:

```bash
magex ci:static
```

**What it does**:
1. Runs actionlint on all workflow files
2. Validates action.yml schemas
3. Checks policy compliance
4. Reports deprecated actions/runners

**Target**: < 2 seconds

### Comprehensive Verification (Pre-Merge)

Run all 35+ scenarios:

```bash
magex ci:verify
```

**What it does**:
1. Full static analysis
2. Executes all failure scenarios
3. Executes all success scenarios
4. Generates JSONL and SARIF reports
5. Writes GitHub Step Summary (in CI)

**Target**: < 5 minutes

---

## Debugging Scenarios

### List Available Scenarios

```bash
# List all scenarios
magex ci:list

# Filter by category
magex ci:list filter=security
magex ci:list filter=quality
```

**Output**:
```
Available CI Scenarios (35 total)

Quality (6):
  LINT-001 - Unused variable detection
  LINT-002 - Gofmt formatting violation
  LINT-003 - golangci-lint rule violation
  ...

Security (5):
  SEC-001 - Hardcoded AWS key pattern
  SEC-002 - Private key in repository
  ...

Testing (4):
  TEST-001 - Failing unit test assertion
  TEST-002 - Test panic (nil pointer)
  ...
```

### Run Single Scenario

```bash
# Basic execution
magex ci:scenario name=LINT-001

# With verbose output
magex ci:scenario name=LINT-001 verbose=true

# Keep container for debugging
magex ci:scenario name=SEC-001 keep=true
```

**Debugging a failed scenario**:
```bash
# Run with container preservation
magex ci:scenario name=RACE-001 keep=true verbose=true

# Exec into the preserved container
docker exec -it <container-id> /bin/bash
```

---

## Report Formats

### Terminal Output

All commands produce colored terminal output by default.

### JSONL Report

Machine-readable report at `.mage-x/ci-results.jsonl`:

```jsonl
{"type":"run_start","timestamp":"2026-01-24T10:00:00Z","version":"1.0.0"}
{"type":"scenario","id":"LINT-001","status":"pass","duration_ms":4523}
{"type":"run_end","passed":34,"failed":1,"skipped":0}
```

### SARIF Report

GitHub Security integration at `.mage-x/guardian.sarif`:

```bash
# View in GitHub Security tab after uploading
gh api repos/{owner}/{repo}/code-scanning/sarifs -X POST -F sarif=@.mage-x/guardian.sarif
```

### GitHub Annotations

When running in CI, findings appear as inline annotations on the PR.

---

## Configuration

### Environment Variables

Add to `.env.custom` to override defaults:

```bash
# Disable Guardian in CI
ENABLE_CI_GUARDIAN=false

# Increase scenario timeout
GUARDIAN_SCENARIO_TIMEOUT=60s

# Reduce parallelism
GUARDIAN_PARALLEL_SCENARIOS=2
```

### Policy Exceptions

Create `.github/guardian.yaml` to exempt specific rules:

```yaml
exceptions:
  - policy: sha-pinned-actions
    path: .github/workflows/test.yml
    reason: "Testing unpinned action behavior"
    expires: 2026-06-01
    approved_by: "@mrz1836"
    created_at: 2026-01-24
```

---

## Common Workflows

### Pre-Push Hook Integration

Add to your Git hooks:

```bash
# .git/hooks/pre-push
#!/bin/bash
magex ci:static
```

### CI Pipeline Integration

The `ci-tester.yml` workflow runs Guardian on PRs:

```yaml
# .github/workflows/ci-tester.yml
name: CI Tester (Guardian)

on:
  pull_request:
    paths:
      - '.github/**'
      - 'magefiles/**'
      - '.env.base'

jobs:
  guardian:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ./.github/actions/setup-go
      - uses: ./.github/actions/setup-magex
      - run: magex ci:verify
      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: .mage-x/guardian.sarif
```

### Creating Custom Scenarios

1. Create fixture directory:
```bash
mkdir -p .github/ci-tester/fixtures/my-fail
```

2. Add minimal Go module:
```go
// .github/ci-tester/fixtures/my-fail/go.mod
module fixture-my-fail

go 1.24
```

3. Add failing code:
```go
// .github/ci-tester/fixtures/my-fail/main.go
package main

func main() {
    unusedVar := "this triggers lint"
}
```

4. Add workflow:
```yaml
# .github/ci-tester/fixtures/my-fail/.github/workflows/ci.yml
name: CI
on: push
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go vet ./...
```

5. Register scenario in `guardian/scenarios/quality.go`:
```go
var MyFailScenario = Scenario{
    ID:          "MY-001",
    Category:    CategoryQuality,
    Description: "My custom failure scenario",
    FixturePath: ".github/ci-tester/fixtures/my-fail",
    Workflow:    "ci.yml",
    Expected: ExpectedResult{
        Status:      StatusFailure,
        LogPatterns: []string{"unusedVar declared"},
    },
}
```

---

## Troubleshooting

### "Docker is not running"

Guardian requires Docker for scenario execution:

```bash
# Start Docker
open -a Docker  # macOS
sudo systemctl start docker  # Linux

# Verify
docker ps
```

**Workaround**: Use `magex ci:static` for validation without Docker.

### "act not found"

Install act:

```bash
go install github.com/nektos/act@latest

# Or via deps:install
magex deps:install
```

### Scenario Timeout

Increase timeout for slow scenarios:

```bash
magex ci:scenario name=SLOW-001 timeout=120s
```

Or configure globally in `.env.custom`:

```bash
GUARDIAN_SCENARIO_TIMEOUT=120s
```

### "unexpected success" Error

This means a failure scenario passed when it should have failed. Check:

1. Fixture code actually triggers the expected failure
2. Workflow runs the correct checks
3. Expected log patterns match actual output

```bash
# Debug with verbose output
magex ci:scenario name=FAILING-001 verbose=true
```

### Cache Issues

Clear act's container cache:

```bash
docker system prune -a
```

---

## Command Reference

| Command | Description | Target Time |
|---------|-------------|-------------|
| `magex ci:test` | Quick validation | < 60s |
| `magex ci:verify` | Full verification | < 5m |
| `magex ci:static` | Static analysis only | < 2s |
| `magex ci:scenario name=X` | Run single scenario | < 30s |
| `magex ci:list` | List scenarios | Instant |

### Parameters

| Parameter | Commands | Description |
|-----------|----------|-------------|
| `name=X` | ci:scenario | Scenario ID to run |
| `verbose=true` | ci:scenario | Show detailed output |
| `keep=true` | ci:scenario | Preserve container |
| `timeout=Xs` | ci:scenario | Override timeout |
| `filter=X` | ci:list | Filter by category |

---

## Code Quality

Before committing Go code changes, run the full quality pipeline:

```bash
# Full quality check (required before commits)
go-pre-commit --all-files && magex format:fix && magex lint
```

This ensures:
1. **go-pre-commit** - Pre-commit hooks pass (formatting, goimports, etc.)
2. **magex format:fix** - All formatting issues auto-fixed
3. **magex lint** - golangci-lint passes with no errors

---

## Next Steps

1. Run `magex ci:test` to validate your current workflows
2. Review any findings and fix policy violations
3. Add `ci-tester.yml` to your CI pipeline
4. Create custom scenarios for project-specific validations

For detailed API documentation, see [contracts/guardian-api.go](./contracts/guardian-api.go).
For data model details, see [data-model.md](./data-model.md).
For package documentation, see [guardian/README.md](../../guardian/README.md) (created during implementation).
