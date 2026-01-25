# Fortress Guardian: CI Testing Specification

> **Status:** Draft Implementation Spec
> **Target:** `v1.x.x` (MVP)
> **Philosophy:** Pure Go. Zero False Positives. Enterprise Quality.

---

## 1. Executive Summary

**Fortress Guardian** is the final line of defense for the Go-Fortress ecosystem. It is a **Go-native CI validation framework** designed to test the CI system itself.

By treating GitHub Actions workflows as compile-able, test-able code, Guardian ensures that no broken configuration ever reaches the `master` branch. It eliminates "commit-and-pray" workflow development by enabling accurate local reproduction of all CI failure states.

### Value Proposition

| Stakeholder | Benefit |
|-------------|---------|
| **Developers** | Instant local feedback on workflow changes before pushing |
| **Maintainers** | Confidence that CI changes won't break production pipelines |
| **Security Teams** | Verified supply chain controls and policy enforcement |
| **Operations** | Predictable CI behavior with documented failure modes |

### Scope

Guardian validates the entire Go-Fortress CI ecosystem:
- 16+ GitHub Actions workflows
- 17 composite actions
- 390+ environment variables
- Fork safety mechanisms
- Supply chain security controls

---

## 2. Architecture Principles

### 2.1 The "Pure Go" Promise

We reject Python, Ruby, and Node.js dependencies. Every component of Guardian must be manageable via the standard Go toolchain.

| Component | Tool | Installation |
|-----------|------|--------------|
| Execution | `nektos/act` | `go install` |
| Linting | `actionlint` | Go-native binary |
| Orchestration | `MAGE-X` | Go-native build system |
| Validation | Custom Go | Internal packages |

### 2.2 Context-First Design

All internal components must adhere to the [Go Essentials](../tech-conventions/go-essentials.md) "Context-First" principle. Every function call that performs I/O or long-running tasks must accept `context.Context` as its first argument to ensure graceful cancellation/timeout support during parallel test execution.

```go
// Correct: Context-first signature
func RunScenario(ctx context.Context, id string, opts ...Option) (*Result, error)

// Incorrect: Missing context
func RunScenario(id string) (*Result, error)
```

### 2.3 MAGE-X Integration

Guardian is not a standalone tool; it is a seamless extension of the MAGE-X ecosystem. It introduces a new `ci:` namespace to the existing 150+ commands.

### 2.4 GitOps & Declarative Configuration

All Guardian behavior is defined declaratively in version-controlled files:
- Scenario definitions in YAML/Go structs
- Policy rules as Go code (not external DSL)
- No runtime configuration outside of environment variables

### 2.5 Hermetic Execution

Each test scenario runs in complete isolation:
- Fresh container per scenario (when using Act)
- No shared state between scenarios
- Deterministic ordering with parallel execution support

---

## 3. Supply Chain Security

Guardian enforces and validates supply chain security controls required for enterprise 2026 compliance.

### 3.1 SLSA Compliance

**Target:** SLSA Level 2+ for all artifacts

| SLSA Requirement | Guardian Validation |
|------------------|---------------------|
| Source integrity | Verify branch protections in workflow |
| Build isolation | Validate container isolation settings |
| Provenance | Check attestation generation in release workflow |
| Dependency management | Validate lockfile presence and freshness |

**Scenario Coverage:**
- `SLSA-001`: Verify provenance attestation is generated on release
- `SLSA-002`: Validate build runs in isolated environment
- `SLSA-003`: Check that dependencies are pinned to digests

### 3.2 SBOM Generation

Guardian validates Software Bill of Materials generation for all releases.

| Format | Validation |
|--------|------------|
| SPDX 2.3 | Verify completeness, required fields |
| CycloneDX 1.4+ | Verify component enumeration |

**Validation Checks:**
- All direct dependencies enumerated
- Transitive dependencies resolved
- License information populated
- PURL format compliance

### 3.3 Provenance Attestation

Integration with GitHub's artifact attestations:

```yaml
# Expected in release workflow
- uses: actions/attest-build-provenance@1c608d11d69870c2092266b3f9a6f3abbf17002c # v1.4.3
  with:
    subject-path: dist/*.tar.gz
```

**Guardian validates:**
- Attestation action is present in release workflow
- Attestation is generated for all release artifacts
- Attestation can be verified via `gh attestation verify`

### 3.4 Scorecard Integration

Guardian validates alignment with OpenSSF Scorecard checks:

| Check | Guardian Validation |
|-------|---------------------|
| Token-Permissions | Verify minimal permissions in workflows |
| Pinned-Dependencies | Verify SHA-pinned actions |
| Branch-Protection | Validate branch rules (via API mock) |
| Dangerous-Workflow | No `pull_request_target` with write perms |

---

## 4. System Design

### 4.1 Directory Structure

Guardian is a package in the main go-fortress module, making it part of the showcase alongside `fortress.go`. It is importable as `github.com/mrz1836/go-fortress/guardian` and tested/linted/covered like all other Go code.

```
go-fortress/
├── fortress.go                 # Showcase: simple Go code
├── fortress_test.go            # Showcase: testing patterns
├── guardian/                   # CI testing framework
│   ├── runner/                 # Act wrapper, job isolation
│   │   ├── act.go
│   │   ├── scenario.go
│   │   └── events.go
│   ├── validator/              # Static analysis
│   │   ├── schema.go
│   │   ├── actionlint.go
│   │   └── deprecation.go
│   ├── policy/                 # Policy-as-code engine
│   │   ├── engine.go
│   │   └── rules.go
│   ├── reporter/               # Output formatting
│   │   ├── jsonl.go
│   │   ├── sarif.go
│   │   └── annotations.go
│   ├── scenarios.go            # Scenario definitions
│   └── guardian.go             # Package entry point
├── testdata/
│   └── guardian/               # Golden files for Guardian tests
└── .github/
    ├── ci-tester/
    │   └── fixtures/           # Broken repos for Act execution
    │       ├── lint-fail/      # Fixture: unused variable
    │       ├── test-fail/      # Fixture: failing test
    │       ├── race-fail/      # Fixture: data race
    │       ├── sec-fail/       # Fixture: hardcoded secret
    │       ├── vuln-fail/      # Fixture: CVE dependency
    │       └── cov-fail/       # Fixture: low coverage
    └── workflows/
        └── fortress-guardian.yml  # Runs Guardian on PR
```

**Key Points:**
- Go code in `guardian/` (standard location, linted/tested/covered)
- Act fixtures in `.github/ci-tester/fixtures/` (near workflows)
- Golden files in `testdata/guardian/` (Go convention)
- Single module (no separate `go.mod`)

### 4.2 Core Components

#### Static Validation Layer (`guardian/validator/`)

Before running any containers, Guardian performs instant static analysis:

| Validator | Purpose | Speed |
|-----------|---------|-------|
| Schema | Validate `action.yml` and `.env.base` structure | < 100ms |
| Deprecation | Scan for outdated action versions and runner labels | < 200ms |
| Actionlint | Robust syntax checking via Go wrapper | < 500ms |
| Policy | Enforce security policies (see Section 13) | < 100ms |

#### Execution Engine (`guardian/runner/`)

Wraps `act` to provide a deterministic, isolated environment.

| Capability | Implementation |
|------------|----------------|
| Job Isolation | Run specific jobs (e.g., `pre-commit`, `security`) in isolation |
| Event Injection | Inject synthetic GitHub event payloads (Push, PR, Release) |
| Artifact Mocking | Stub GitHub Actions artifact APIs to filesystem paths |
| Service Containers | Mock Redis, PostgreSQL via Docker Compose |
| Secret Injection | Provide test secrets via `--secret-file` |

#### Scenario System (`guardian/scenarios.go`)

Scenario definitions are kept in Go code for type safety. Test fixtures (pre-broken repositories) live in `.github/ci-tester/fixtures/` near the workflows they test. Each fixture is a valid Go module designed to fail in a specific, predictable way.

| Fixture Type | Location | Purpose |
|--------------|----------|---------|
| `lint-fail` | `.github/ci-tester/fixtures/lint-fail/` | Code with formatting errors |
| `vuln-fail` | `.github/ci-tester/fixtures/vuln-fail/` | Code with known CVE dependencies |
| `sec-fail` | `.github/ci-tester/fixtures/sec-fail/` | Code containing mock secrets |
| `race-fail` | `.github/ci-tester/fixtures/race-fail/` | Code with data race conditions |
| `cov-fail` | `.github/ci-tester/fixtures/cov-fail/` | Code with insufficient coverage |

#### Reporter (`guardian/reporter/`)

Generates structured output in multiple formats:
- JSONL for machine consumption
- SARIF for GitHub Security integration
- Markdown for PR comments
- Terminal for local development

#### Policy Engine (`guardian/policy/`)

Enforces workflow policies as Go code. Policies are defined in `guardian/policy/rules.go` and executed by `guardian/policy/engine.go`.

### 4.3 Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        magex ci:verify                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Static Validation                           │
│  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌────────────────┐ │
│  │ Schema   │  │ Actionlint │  │ Policy   │  │ Deprecation    │ │
│  └──────────┘  └────────────┘  └──────────┘  └────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Scenario Execution                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ For each scenario:                                       │   │
│  │   1. Load fixture (broken repo)                          │   │
│  │   2. Inject event payload                                │   │
│  │   3. Execute via Act                                     │   │
│  │   4. Capture output + exit code                          │   │
│  │   5. Validate against success criteria                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Result Aggregation                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────────┐  │
│  │ JSONL    │  │ SARIF    │  │ Markdown │  │ GitHub Summary  │  │
│  └──────────┘  └──────────┘  └──────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Act Compatibility Matrix

`nektos/act` has significant limitations that Guardian must account for.

### 5.1 Feature Support Matrix

| GitHub Actions Feature | Act Support | Guardian Strategy |
|------------------------|-------------|-------------------|
| `workflow_call` (reusable) | No | Test caller workflow; validate callee schema only |
| `workflow_dispatch` | Yes | Full support with input validation |
| `services:` containers | Partial | Use Docker Compose mock for complex services |
| `actions/cache` | Limited | Local filesystem mock with hit/miss simulation |
| Matrix builds | Yes | Full support |
| Composite actions | Yes | Full support |
| Container actions | Yes | Full support |
| JavaScript actions | Yes | Full support (Node.js in container) |
| `GITHUB_TOKEN` | Partial | Mock token with limited permissions |
| Secrets | Yes | Inject via `--secret-file` |
| Environment files | Yes | Full support |
| Job outputs | Yes | Full support |
| Artifacts upload/download | Partial | Local filesystem stub |
| `github.event` context | Yes | Inject via event JSON |
| `runner.os` / `runner.arch` | Yes | Based on container image |

### 5.2 Unsupported Features & Workarounds

#### `workflow_call` (Reusable Workflows)

**Problem:** Act cannot invoke reusable workflows.

**Strategy:**
1. Validate reusable workflow schema statically
2. Test the calling workflow with mocked outputs
3. Create integration tests that inline the reusable workflow logic

#### GitHub-Hosted Runner Features

**Problem:** Some features only work on GitHub-hosted runners.

| Feature | Workaround |
|---------|------------|
| `actions/cache` | Mock cache with local directory |
| OIDC tokens | Mock token endpoint |
| Larger runners | Verify configuration only |

### 5.3 Local vs CI Parity

Guardian ensures scenarios produce identical results locally and in CI:

| Aspect | Local (Act) | CI (GitHub) | Parity Strategy |
|--------|-------------|-------------|-----------------|
| Container runtime | Docker/Podman | Docker | Same base images |
| Event payloads | Synthetic JSON | Real events | Captured from CI runs |
| Secrets | `.secrets` file | Repository secrets | Same test values |
| Caching | Disabled/mocked | Enabled | Test both modes |

---

## 6. MAGE-X Interface

Guardian is controlled entirely through `magex` commands.

### 6.1 Command Reference

#### `magex ci:test` (Primary)

The "do everything" command for developers.

```bash
magex ci:test
```

**Behavior:**
1. Runs full static analysis
2. Executes fast failure scenarios (< 10s each)
3. Reports summary with pass/fail counts

**Exit Codes:**
- `0`: All scenarios passed
- `1`: One or more scenarios failed unexpectedly
- `2`: Configuration or setup error

#### `magex ci:verify`

Deep verification suite for PRs.

```bash
magex ci:verify
```

**Behavior:**
1. Runs exhaustive matrix of all failure and success scenarios
2. Generates detailed JSONL report
3. Produces SARIF output for security findings
4. Updates GitHub Step Summary (in CI)

#### `magex ci:static`

Fast static analysis only.

```bash
magex ci:static
```

**Behavior:**
- Schema validation
- Actionlint execution
- Deprecation checks
- Policy enforcement

**Target:** < 2s execution

#### `magex ci:scenario [ID]`

Debug a specific scenario locally.

```bash
magex ci:scenario name=LINT-001
magex ci:scenario name=LINT-001 verbose=true
magex ci:scenario name=LINT-001 keep=true  # Keep container for debugging
```

#### `magex ci:list`

List all available scenarios.

```bash
magex ci:list
magex ci:list filter=security  # Filter by category
```

### 6.2 Integration with Existing Namespaces

Guardian commands integrate with existing MAGE-X namespaces:

| Existing Command | Guardian Integration |
|------------------|----------------------|
| `magex deps:install` | Installs `act` and `actionlint` |
| `magex pre-commit` | Runs `ci:static` as part of pre-commit |
| `magex check:all` | Includes `ci:test` in comprehensive check |

---

## 7. Test Scenarios

Guardian verifies strict adherence to the **Multi-Stage Defense System** defined in `README.md`.

### 7.1 Failure Scenarios (Must Fail Gracefully)

#### Quality Gate Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `LINT-001` | Quality | Unused variable in `main.go` | Job `failure` + linter error in logs |
| `LINT-002` | Quality | Gofmt formatting violation | Job `failure` + `gofmt` diff in output |
| `LINT-003` | Quality | golangci-lint rule violation | Job `failure` + specific linter name |

#### Testing Gate Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `TEST-001` | Testing | Failing unit test assertion | Job `failure` + `FAIL:` in output |
| `TEST-002` | Testing | Test panic (nil pointer) | Job `failure` + stack trace in logs |
| `RACE-001` | Testing | Concurrent map write | Job `failure` + `WARNING: DATA RACE` |
| `TIMEOUT-001` | Testing | Test exceeding 30m timeout | Job `failure` + timeout message |

#### Security Gate Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `SEC-001` | Security | Hardcoded AWS key pattern | Job `failure` + Gitleaks detection |
| `SEC-002` | Security | Private key in repository | Job `failure` + Gitleaks detection |
| `VULN-001` | Security | Dependency with CVSS > 7.0 | Job `failure` + Nancy/govulncheck report |
| `VULN-002` | Security | Go version with known CVE | Job `failure` + version check failure |
| `GITLEAKS-001` | Security | API token in config file | Job `failure` + Gitleaks finding |

#### Coverage Gate Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `COV-001` | Coverage | Coverage 50% (threshold 80%) | Job `failure` + coverage report |
| `COV-002` | Coverage | Untested new function | Job `failure` + diff coverage check |

#### Cache Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `CACHE-001` | Caching | Cache key mismatch (forces rebuild) | Cache miss logged, build succeeds |
| `CACHE-002` | Caching | Cache corruption detection | Graceful fallback to rebuild |

#### Fork Safety Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `FORK-001` | Fork Safety | Fork PR detection | `is-fork-pr` output = `true` |
| `FORK-002` | Fork Safety | Secret protection on fork | FORK-UNSAFE jobs skipped |
| `FORK-003` | Fork Safety | Label application on fork | Fork PR labels applied correctly |

#### Matrix & Configuration Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `MATRIX-001` | Config | Empty matrix generation | Workflow fails with clear error |
| `ENV-001` | Config | Missing required secret | Job `failure` + secret name in error |
| `ENV-002` | Config | Invalid env variable value | Validation failure message |

#### Service Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `REDIS-001` | Services | Redis container startup failure | Graceful error handling |
| `REDIS-002` | Services | Redis connection timeout | Retry with backoff, then fail |

#### Tooling Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `MAGEX-001` | Tooling | Binary cache miss cold start | Tool installation logged |
| `MAGEX-002` | Tooling | Invalid mage target | Clear error message |

#### Artifact Failures

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `ARTIFACT-001` | Artifacts | Upload failure with retry | Retry logged, eventual failure |
| `ARTIFACT-002` | Artifacts | Download missing artifact | Clear error, job failure |

### 7.2 Success Scenarios (Must Pass)

| ID | Gate | Description | Success Criteria |
|----|------|-------------|------------------|
| `PASS-001` | Full | Pristine Go module | All jobs `success` |
| `PASS-002` | Cache | Verified cache restore | Cache hit reported in logs |
| `PASS-003` | Matrix | Valid matrix expansion | All matrix jobs complete |
| `PASS-004` | Fork | Safe fork PR (read-only) | Appropriate jobs run |
| `PASS-005` | Release | Tag-triggered release | Release artifacts generated |

### 7.3 Scenario Definition Format

Each scenario is defined as a Go struct:

```go
type Scenario struct {
    ID          string            `yaml:"id"`
    Category    string            `yaml:"category"`
    Description string            `yaml:"description"`
    Fixture     string            `yaml:"fixture"`     // Path to broken repo
    Event       string            `yaml:"event"`       // Path to event JSON
    Workflow    string            `yaml:"workflow"`    // Workflow file to run
    Job         string            `yaml:"job"`         // Specific job (optional)
    Expected    ExpectedResult    `yaml:"expected"`
    Timeout     time.Duration     `yaml:"timeout"`
}

type ExpectedResult struct {
    Status      string   `yaml:"status"`      // "success" or "failure"
    LogPatterns []string `yaml:"log_patterns"` // Regex patterns to match
    Outputs     map[string]string `yaml:"outputs"` // Expected outputs
}
```

---

## 8. Configuration Validation

### 8.1 `.env.base` Schema Validation

Guardian validates all 390+ environment variables in `.env.base`:

| Validation | Description |
|------------|-------------|
| Type checking | Boolean, integer, string, duration, path |
| Required fields | Variables marked as required must be present |
| Default values | Verify defaults are sensible |
| Deprecated check | Flag removed/deprecated variables |
| Naming convention | `UPPER_SNAKE_CASE` enforcement |

**Schema Definition:**

```go
type EnvVarSchema struct {
    Name        string      `yaml:"name"`
    Type        string      `yaml:"type"`     // bool, int, string, duration, path
    Required    bool        `yaml:"required"`
    Default     interface{} `yaml:"default"`
    Deprecated  bool        `yaml:"deprecated"`
    Description string      `yaml:"description"`
    ValidValues []string    `yaml:"valid_values"` // For enums
}
```

### 8.2 `.env.custom` Override Testing

| Scenario | Validation |
|----------|------------|
| Precedence | Custom values override base values |
| Unknown vars | Warning on undefined variables |
| Type mismatch | Error on type violations |
| Empty file | Graceful handling of missing/empty file |

### 8.3 `load-env` Action Validation

Validate the composite action that loads environment:

| Check | Description |
|-------|-------------|
| Output completeness | All required vars exported |
| Secret masking | Sensitive vars properly masked |
| Path expansion | `~` and `$HOME` expanded correctly |

---

## 9. Output Specification

### 9.1 JSONL Format (CI Results)

Compatible with existing `.mage-x/ci-results.jsonl`:

```jsonl
{"type":"run_start","timestamp":"2026-01-23T10:00:00Z","version":"1.0.0"}
{"type":"scenario","id":"LINT-001","status":"pass","duration_ms":1234,"logs_path":"/tmp/lint-001.log"}
{"type":"scenario","id":"TEST-001","status":"fail","duration_ms":5678,"error":"unexpected success"}
{"type":"run_end","timestamp":"2026-01-23T10:05:00Z","passed":19,"failed":1,"skipped":0}
```

**Field Definitions:**

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Event type: `run_start`, `scenario`, `run_end` |
| `id` | string | Scenario ID (e.g., `LINT-001`) |
| `status` | string | `pass`, `fail`, `skip`, `error` |
| `duration_ms` | int | Execution time in milliseconds |
| `error` | string | Error message (on failure) |
| `logs_path` | string | Path to detailed logs |

### 9.2 SARIF Format (Security Findings)

Standard SARIF 2.1.0 for GitHub Security tab integration:

```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [{
    "tool": {
      "driver": {
        "name": "Fortress Guardian",
        "version": "1.0.0",
        "rules": [...]
      }
    },
    "results": [...]
  }]
}
```

**Mapped Findings:**

| Guardian Finding | SARIF Level | Rule ID |
|------------------|-------------|---------|
| Unpinned action | warning | `guardian/unpinned-action` |
| Exposed secret | error | `guardian/exposed-secret` |
| Dangerous workflow | error | `guardian/dangerous-workflow` |
| Missing permission | note | `guardian/missing-permission` |

### 9.3 GitHub Annotations

Format for inline PR feedback:

```
::error file=.github/workflows/ci.yml,line=42::Action 'actions/checkout' should be SHA-pinned
::warning file=.github/workflows/ci.yml,line=15::Missing explicit permissions declaration
```

### 9.4 GitHub Step Summary

Markdown template for `$GITHUB_STEP_SUMMARY`:

```markdown
## Fortress Guardian Results

### Summary
- **Passed:** 19/20 scenarios
- **Failed:** 1 scenario
- **Duration:** 4m 32s

### Failed Scenarios

| ID | Description | Error |
|----|-------------|-------|
| `TEST-001` | Failing test assertion | Expected failure, got success |

### Coverage

All quality gates validated. See [detailed report](link).
```

---

## 10. Workflow Integration

### 10.1 Primary Integration Points

Guardian integrates with the 16+ existing workflows:

| Workflow | Integration Type | Validation |
|----------|------------------|------------|
| `fortress.yml` | Orchestrator | Validate `needs:` dependency chains |
| `fortress-setup-config.yml` | Reusable | Validate matrix output schema |
| `fortress-security-scans.yml` | Reusable | Mock vulnerability injection |
| `fortress-format-lint.yml` | Reusable | Validate linter configuration |
| `fortress-testing.yml` | Reusable | Validate test matrix |

### 10.2 Composite Action Unit Testing

Guardian provides unit tests for all 17 composite actions:

| Action | Test Coverage |
|--------|---------------|
| `load-env` | Output completeness, secret masking |
| `determine-changes` | Path filter accuracy |
| `setup-go` | Version resolution, caching |
| `setup-magex` | Binary caching, fallback |
| `format-lint` | Linter invocation, output parsing |
| `security-scans` | Tool installation, finding format |
| `notify-slack` | Message formatting (mock) |

### 10.3 New Workflow: `fortress-guardian.yml`

```yaml
name: "Guardian: CI Tester"

on:
  pull_request:
    paths:
      - '.github/**'
      - 'magefiles/**'
      - '.env.base'
  workflow_dispatch:
    inputs:
      scenario:
        description: 'Specific scenario to run (leave empty for all)'
        required: false
        type: string

permissions:
  contents: read
  security-events: write  # For SARIF upload

concurrency:
  group: guardian-${{ github.ref }}
  cancel-in-progress: true

jobs:
  guardian:
    name: Validate CI Configuration
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2

      - name: Setup Go
        uses: ./.github/actions/setup-go

      - name: Setup MAGE-X
        uses: ./.github/actions/setup-magex

      - name: Run Guardian Verification
        run: magex ci:verify

      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@48ab28a6f5dbc2a99bf1e0131198dd8f1df78169 # v3.28.0
        with:
          sarif_file: .mage-x/guardian.sarif
```

---

## 11. Performance Requirements

### 11.1 Execution Time Targets

| Command | Target | Hard Limit | Failure Action |
|---------|--------|------------|----------------|
| `ci:static` | < 2s | < 5s | Optimize validators |
| `ci:test` | < 60s | < 120s | Parallelize scenarios |
| `ci:verify` | < 5m | < 10m | Add caching layer |
| Individual scenario | < 30s | < 60s | Flag as slow test |

### 11.2 Resource Constraints

| Resource | Limit | Rationale |
|----------|-------|-----------|
| Memory per scenario | 2 GB | Match GitHub runner limits |
| Disk per scenario | 10 GB | Container image + workspace |
| Parallel scenarios | 4 | Balance speed vs resources |
| Total disk | 20 GB | CI runner constraint |

### 11.3 Performance Monitoring

Guardian tracks execution metrics:

```jsonl
{"type":"perf","scenario":"LINT-001","cpu_ms":450,"mem_peak_mb":128,"disk_mb":52}
```

Metrics are aggregated for trend analysis and regression detection.

---

## 12. Observability

### 12.1 Metrics Collection

Guardian collects operational metrics during execution:

| Metric | Type | Description |
|--------|------|-------------|
| `guardian_scenario_duration_seconds` | histogram | Time per scenario |
| `guardian_scenario_status` | counter | Pass/fail/skip counts |
| `guardian_cache_hit_ratio` | gauge | Act image cache efficiency |
| `guardian_validator_duration_seconds` | histogram | Static validator times |

### 12.2 Structured Logging

All logs follow structured format:

```json
{
  "level": "info",
  "ts": "2026-01-23T10:00:00Z",
  "msg": "scenario completed",
  "scenario_id": "LINT-001",
  "status": "pass",
  "duration_ms": 1234
}
```

### 12.3 Future: OpenTelemetry Integration

Planned for v2.0:

| Feature | Description |
|---------|-------------|
| Distributed tracing | Span per workflow job |
| Trace context | Propagate through Act execution |
| Export | OTLP to observability backend |
| Dashboards | Pre-built Grafana templates |

---

## 13. Policy-as-Code

Guardian enforces workflow policies automatically.

### 13.1 Enforced Policies

| Policy | Severity | Description |
|--------|----------|-------------|
| `sha-pinned-actions` | error | All actions MUST be SHA-pinned |
| `explicit-permissions` | warning | Permissions MUST be explicitly declared |
| `no-dangerous-workflows` | error | No `pull_request_target` with write perms |
| `no-secret-logging` | error | Secrets MUST NOT be logged |
| `concurrency-defined` | warning | Concurrency groups MUST be defined |
| `minimal-permissions` | warning | Use least-privilege permissions |

### 13.2 Policy Definition

Policies are defined as Go functions:

```go
type Policy struct {
    ID          string
    Severity    Severity  // Error, Warning, Note
    Description string
    Check       func(workflow *Workflow) []Finding
}

var SHAPinnedActions = Policy{
    ID:          "sha-pinned-actions",
    Severity:    Error,
    Description: "All actions must be pinned to a full commit SHA",
    Check: func(w *Workflow) []Finding {
        var findings []Finding
        for _, job := range w.Jobs {
            for _, step := range job.Steps {
                if step.Uses != "" && !isSHAPinned(step.Uses) {
                    findings = append(findings, Finding{
                        Line:    step.Line,
                        Message: fmt.Sprintf("Action %q not SHA-pinned", step.Uses),
                    })
                }
            }
        }
        return findings
    },
}
```

### 13.3 Policy Exceptions

Exceptions are declared in `.github/guardian.yaml`:

```yaml
exceptions:
  - policy: sha-pinned-actions
    path: .github/workflows/test.yml
    reason: "Testing unpinned action behavior"
    expires: 2026-06-01
```

---

## 14. Implementation Plan

### Phase 1: Foundation

**Goals:** Establish package structure, static validation layer.

| Task | Description | Deliverable |
|------|-------------|-------------|
| Package setup | Create `guardian/` package in existing module | Directory structure, `guardian.go` |
| Tool installation | Add `act`, `actionlint` to `deps:install` | Updated mage targets |
| Schema validator | Implement `action.yml` schema validation | `guardian/validator/schema.go` |
| Policy engine | Implement policy-as-code framework | `guardian/policy/engine.go` |
| Static commands | Wire up `ci:static` mage target | Working static analysis |

### Phase 2: Execution Core

**Goals:** Implement Act wrapper, basic scenarios.

| Task | Description | Deliverable |
|------|-------------|-------------|
| Runner package | Implement Act wrapper with isolation | `guardian/runner/act.go` |
| Context support | Ensure all runner methods accept context | Timeout/cancellation support |
| Event injection | Implement event payload injection | `guardian/runner/events.go` |
| Basic fixtures | Create LINT-001, TEST-001, SEC-001 fixtures | `.github/ci-tester/fixtures/` |
| Scenario runner | Implement scenario execution logic | `guardian/runner/scenario.go` |

### Phase 3: Integration

**Goals:** Full MAGE-X integration, CI workflow.

| Task | Description | Deliverable |
|------|-------------|-------------|
| MAGE-X targets | Wire up all `ci:*` commands | `magefiles/ci.go` |
| CI workflow | Create `fortress-guardian.yml` | `.github/workflows/fortress-guardian.yml` |
| Reporter | Implement JSONL, SARIF, annotations | `guardian/reporter/*.go` |
| Fork scenarios | Implement FORK-001, FORK-002, FORK-003 | Fork safety validation |
| Self-test | Verify Guardian catches regressions | Passing CI |

### Phase 4: Hardening

**Goals:** Complete scenario coverage, performance optimization.

| Task | Description | Deliverable |
|------|-------------|-------------|
| All scenarios | Implement remaining 15+ scenarios | Full scenario suite |
| Performance | Optimize for < 5m `ci:verify` | Performance benchmarks |
| Documentation | Complete user documentation | `docs/guardian.md` |
| Observability | Add metrics collection | Prometheus-compatible output |
| Supply chain | Implement SLSA validation | Supply chain scenarios |

---

## 15. Requirements & Constraints

### 15.1 Runtime Requirements

| Requirement | Details |
|-------------|---------|
| Docker | Required for Act execution; graceful skip if unavailable |
| Podman | Supported as Docker alternative |
| Go 1.22+ | Required for module and MAGE-X |
| Disk space | Minimum 10 GB for container images |
| Memory | Minimum 4 GB for parallel execution |

### 15.2 Platform Parity

All scenarios must produce identical results across platforms:

| Platform | Support Level | Notes |
|----------|---------------|-------|
| macOS (Apple Silicon) | Full | Primary development platform |
| macOS (Intel) | Full | CI runner support |
| Linux (x86_64) | Full | Primary CI platform |
| Linux (arm64) | Full | Emerging CI platform |
| Windows | Limited | WSL2 required for Act |

### 15.3 Docker Requirements

| Requirement | Specification |
|-------------|---------------|
| Docker Engine | 20.10+ |
| Docker Compose | 2.0+ (for service mocking) |
| Podman | 4.0+ (alternative) |
| Rootless mode | Supported |
| Docker-in-Docker | Not required |

### 15.4 Network Requirements

| Requirement | Details |
|-------------|---------|
| Internet access | Required for first-run image pulls |
| Offline mode | Supported after initial setup |
| Proxy support | HTTP_PROXY/HTTPS_PROXY honored |

---

## 16. Future Roadmap

### 16.1 AI-Assisted CI (v2.0)

| Feature | Description |
|---------|-------------|
| Flaky test detection | Analyze historical JSONL for flaky patterns |
| Test selection | Run only tests affected by changed files |
| Failure prediction | Predict failures from commit patterns |
| Auto-remediation | Suggest fixes for common failures |

### 16.2 Ephemeral Environments (v2.0)

| Feature | Description |
|---------|-------------|
| Preview environments | Spin up environment per PR |
| Integration testing | Test against real services |
| Cost optimization | Auto-teardown after merge |

### 16.3 Advanced Container Strategies (v2.0)

| Strategy | Description |
|----------|-------------|
| Podman rootless | Full support without Docker daemon |
| Kubernetes executor | Run scenarios in K8s pods |
| Image caching | Pre-warm container registry |
| Multi-arch builds | Validate arm64 + amd64 |

### 16.4 Extended Policy Engine (v2.0)

| Feature | Description |
|---------|-------------|
| Rego policies | OPA-compatible policy language |
| Policy inheritance | Org-level policy templates |
| Compliance reports | SOC2, HIPAA mapping |

---

## Appendix A: Scenario Quick Reference

| ID | Category | Type | Description |
|----|----------|------|-------------|
| LINT-001 | Quality | Fail | Unused variable |
| LINT-002 | Quality | Fail | Gofmt violation |
| LINT-003 | Quality | Fail | golangci-lint error |
| TEST-001 | Testing | Fail | Failing assertion |
| TEST-002 | Testing | Fail | Test panic |
| RACE-001 | Testing | Fail | Data race |
| TIMEOUT-001 | Testing | Fail | Timeout exceeded |
| SEC-001 | Security | Fail | AWS key pattern |
| SEC-002 | Security | Fail | Private key |
| VULN-001 | Security | Fail | High CVSS dependency |
| VULN-002 | Security | Fail | Vulnerable Go version |
| GITLEAKS-001 | Security | Fail | API token |
| COV-001 | Coverage | Fail | Below threshold |
| COV-002 | Coverage | Fail | Untested function |
| CACHE-001 | Caching | Info | Cache miss |
| CACHE-002 | Caching | Fail | Cache corruption |
| FORK-001 | Fork | Info | Fork detection |
| FORK-002 | Fork | Pass | Secret protection |
| FORK-003 | Fork | Pass | Label application |
| MATRIX-001 | Config | Fail | Empty matrix |
| ENV-001 | Config | Fail | Missing secret |
| ENV-002 | Config | Fail | Invalid value |
| REDIS-001 | Services | Fail | Startup failure |
| REDIS-002 | Services | Fail | Connection timeout |
| MAGEX-001 | Tooling | Info | Cache miss |
| MAGEX-002 | Tooling | Fail | Invalid target |
| ARTIFACT-001 | Artifacts | Fail | Upload failure |
| ARTIFACT-002 | Artifacts | Fail | Download missing |
| SLSA-001 | Supply Chain | Pass | Provenance generated |
| SLSA-002 | Supply Chain | Pass | Isolated build |
| SLSA-003 | Supply Chain | Pass | Pinned deps |
| PASS-001 | Full | Pass | Pristine module |
| PASS-002 | Cache | Pass | Cache hit |
| PASS-003 | Matrix | Pass | Valid expansion |
| PASS-004 | Fork | Pass | Safe fork PR |
| PASS-005 | Release | Pass | Tag release |

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **Act** | `nektos/act` - Local GitHub Actions runner |
| **Actionlint** | Static analyzer for GitHub Actions workflows |
| **Fixture** | Pre-configured repository state for testing |
| **Guardian** | This CI testing framework |
| **MAGE-X** | Go-native build system used by Go-Fortress |
| **Scenario** | A single test case with expected outcome |
| **SARIF** | Static Analysis Results Interchange Format |
| **SBOM** | Software Bill of Materials |
| **SLSA** | Supply-chain Levels for Software Artifacts |
