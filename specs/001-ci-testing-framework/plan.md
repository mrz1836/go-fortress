# Implementation Plan: Fortress Guardian CI Testing Framework

**Branch**: `001-ci-testing-framework` | **Date**: 2026-01-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-ci-testing-framework/spec.md`

## Summary

Fortress Guardian is a Go-native CI validation framework that treats GitHub Actions workflows as testable code. The framework enables local reproduction of CI failures via `nektos/act`, performs static validation using `actionlint`, enforces security policies as Go code, and integrates with MAGE-X through a new `ci:` command namespace. All code lives in a `guardian/` package within the existing go-fortress module.

## Technical Context

**Language/Version**: Go 1.24+ (current go.mod specifies 1.24)
**Primary Dependencies**:
- `nektos/act` - Local GitHub Actions runner (go-installable)
- `actionlint` - Static analyzer for workflows (Go-native)
- `gopkg.in/yaml.v3` - YAML parsing (already in go.sum)
- `github.com/stretchr/testify` - Testing (already in go.mod)

**Storage**: Filesystem-based (JSONL reports in `.mage-x/`, fixtures in `.github/ci-tester/fixtures/`)
**Testing**: Standard Go testing with `go test`, race detection, fuzz testing via MAGE-X
**Target Platform**: macOS (Apple Silicon, Intel), Linux (x86_64, arm64); Windows via WSL2
**Project Type**: Single module - Guardian is a package within go-fortress
**Performance Goals**:
- `ci:static` < 2s
- `ci:test` < 60s
- `ci:verify` < 5m
- Individual scenario < 30s

**Constraints**:
- Docker/Podman required for scenario execution (graceful degradation to static-only)
- 10 GB disk for container images
- 4 GB memory for parallel execution
- No Python/Ruby/Node.js dependencies

**Scale/Scope**: 16+ workflows, 17 composite actions, 390+ env vars, 35+ test scenarios

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Pure Go Philosophy
- [x] All build/CI tools are pure Go (MAGE-X, go-pre-commit, go-coverage)
- [x] No Python, Ruby, or Node.js dependencies introduced
- [x] New tools are CGO-free for cross-platform compatibility
  - ✅ `nektos/act` is pure Go, installable via `go install`
  - ✅ `actionlint` is pure Go, installable via `go install`
  - ✅ Guardian package is standard Go without CGO

### II. Multi-Stage Defense System
- [x] Security scanning preserved (Nancy, Govulncheck, Gitleaks)
- [x] Quality gates maintained (golangci-lint, go vet)
- [x] Testing arsenal intact (unit, fuzz, race, benchmarks)
  - ✅ Guardian validates that security scanning tools are configured in workflows
  - ✅ Guardian adds validation layer, does not replace existing gates
  - ✅ Guardian tests itself via standard Go test conventions

### III. Configuration-Driven Architecture
- [x] Changes use `.env.base`/`.env.custom` (no hardcoded values in YAML)
- [x] New features have corresponding `ENABLE_*` flags
- [x] Tool versions are pinnable via configuration
  - ✅ Will add `ENABLE_CI_GUARDIAN=true`, `ACT_VERSION`, `ACTIONLINT_VERSION` to `.env.base`
  - ✅ Scenario configuration stored in Go code (type-safe)
  - ✅ Policy exceptions in `.github/guardian.yaml` (declarative)

### IV. Fork-Safe Security Model
- [x] Fork PR handling preserved (fork-safe vs fork-unsafe job separation)
- [x] No secrets exposed in fork-safe jobs
  - ✅ Guardian workflow validates fork detection mechanisms
  - ✅ Guardian test scenarios include FORK-001, FORK-002, FORK-003
  - ✅ Guardian itself runs in fork-safe mode (static validation only)

### V. Go Development Standards
- [x] Context-first design (`context.Context` as first parameter)
- [x] No global state (dependency injection used)
- [x] No `init()` functions (explicit constructors)
- [x] Proper error handling (checked, wrapped with `%w`)
  - ✅ All Guardian APIs follow `func Xxx(ctx context.Context, ...) error` pattern
  - ✅ Runner, Validator, Policy, Reporter use constructor injection
  - ✅ Configuration loaded via explicit `Load()` functions

### VI. Performance-First Execution
- [x] Parallel execution maximized where possible
- [x] Cache warming utilized
- [x] No unnecessary blocking dependencies between jobs
  - ✅ Static validation runs independently (no Docker)
  - ✅ Scenarios run in parallel (up to 4 concurrent)
  - ✅ Act image caching supported

### VII. Release Automation Excellence
- [x] Releases triggered by semantic version tags
- [x] `magex version:bump` used (not manual tagging)
  - ✅ Guardian is part of go-fortress module, follows existing release process
  - ✅ No separate versioning required

## Project Structure

### Documentation (this feature)

```text
specs/001-ci-testing-framework/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
go-fortress/
├── fortress.go                 # Existing showcase code
├── fortress_test.go            # Existing tests
├── go.mod                      # Single module (add guardian deps)
├── guardian/                   # CI testing framework package
│   ├── guardian.go             # Package entry point, public API
│   ├── guardian_test.go        # Package-level integration tests
│   ├── config.go               # Configuration loading
│   ├── config_test.go
│   ├── runner/                 # Act wrapper, job isolation
│   │   ├── runner.go           # Runner interface and factory
│   │   ├── act.go              # Act-based implementation
│   │   ├── act_test.go
│   │   ├── scenario.go         # Scenario execution logic
│   │   ├── scenario_test.go
│   │   ├── events.go           # GitHub event payload injection
│   │   └── events_test.go
│   ├── validator/              # Static analysis
│   │   ├── validator.go        # Validator interface
│   │   ├── schema.go           # action.yml schema validation
│   │   ├── schema_test.go
│   │   ├── actionlint.go       # Actionlint wrapper
│   │   ├── actionlint_test.go
│   │   ├── deprecation.go      # Deprecated action/runner detection
│   │   └── deprecation_test.go
│   ├── policy/                 # Policy-as-code engine
│   │   ├── engine.go           # Policy execution engine
│   │   ├── engine_test.go
│   │   ├── rules.go            # Built-in policy rules
│   │   ├── rules_test.go
│   │   ├── exceptions.go       # Exception handling
│   │   └── exceptions_test.go
│   ├── reporter/               # Output formatting
│   │   ├── reporter.go         # Reporter interface
│   │   ├── jsonl.go            # JSONL output
│   │   ├── jsonl_test.go
│   │   ├── sarif.go            # SARIF 2.1.0 output
│   │   ├── sarif_test.go
│   │   ├── terminal.go         # Terminal output with colors
│   │   ├── terminal_test.go
│   │   ├── annotations.go      # GitHub annotations
│   │   └── annotations_test.go
│   └── scenarios/              # Scenario definitions
│       ├── scenarios.go        # Scenario registry
│       ├── scenarios_test.go
│       ├── quality.go          # LINT-*, TEST-*, RACE-*, COV-*
│       ├── security.go         # SEC-*, VULN-*, GITLEAKS-*
│       ├── fork.go             # FORK-*
│       ├── config.go           # MATRIX-*, ENV-*, CACHE-*
│       └── supply_chain.go     # SLSA-*, PASS-*
├── testdata/
│   └── guardian/               # Golden files for Guardian tests
│       ├── events/             # Sample event payloads
│       ├── reports/            # Expected report outputs
│       └── workflows/          # Sample workflow files
├── .github/
│   ├── ci-tester/
│   │   └── fixtures/           # Broken repos for Act execution
│   │       ├── lint-fail/      # go.mod + main.go with unused var
│   │       ├── test-fail/      # go.mod + failing test
│   │       ├── race-fail/      # go.mod + data race
│   │       ├── sec-fail/       # go.mod + hardcoded secret pattern
│   │       ├── vuln-fail/      # go.mod + CVE dependency
│   │       └── cov-fail/       # go.mod + low coverage
│   ├── guardian.yaml           # Policy exceptions configuration
│   └── workflows/
│       └── ci-tester.yml       # Guardian CI workflow
└── magefiles/                  # MAGE-X integration (if custom targets needed)
    └── ci.go                   # ci:* namespace commands
```

**Structure Decision**: Single module, Guardian as internal package at `guardian/`. Fixtures near workflows in `.github/ci-tester/fixtures/`. Golden files in `testdata/guardian/` per Go convention.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

---

## Phase Completion Status

### Phase 0: Research ✅ Complete

- **Output**: [research.md](./research.md)
- **Key Decisions**:
  - Act integration via CLI subprocess (not Go API)
  - Actionlint via direct Go API
  - SARIF output via `go-sarif` library
  - MAGE-X integration via `Ci mg.Namespace` pattern
  - Fixtures as minimal Go modules per failure type
  - Policies as Go functions (not OPA/Rego)

### Phase 1: Design & Contracts ✅ Complete

- **Outputs**:
  - [data-model.md](./data-model.md) - Entity definitions and relationships
  - [contracts/guardian-api.go](./contracts/guardian-api.go) - Interface contracts
  - [quickstart.md](./quickstart.md) - Usage documentation
- **Agent Context**: Updated via `update-agent-context.sh claude`

### Constitution Re-Check (Post-Design)

All constitution principles remain satisfied:

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Pure Go Philosophy | ✅ Pass | All tools Go-native, no Python/Ruby/Node |
| II. Multi-Stage Defense | ✅ Pass | Adds validation layer, preserves existing |
| III. Configuration-Driven | ✅ Pass | Uses .env.base, ENABLE_* flags |
| IV. Fork-Safe Security | ✅ Pass | Static-only mode for forks |
| V. Go Development Standards | ✅ Pass | Context-first, DI, no init() |
| VI. Performance-First | ✅ Pass | Parallel scenarios, < 5m target |
| VII. Release Automation | ✅ Pass | Part of module, no separate versioning |

### Phase 2: Task Generation

Ready for `/speckit.tasks` command to generate implementation tasks.

---

## Dependencies (go.mod additions)

```go
require (
    github.com/rhysd/actionlint v1.6.27
    github.com/owenrumney/go-sarif/v2 v2.4.0
)
```

## Tool Version Pinning Standards

Following go-fortress CI patterns:

### Go Tools (via go install)
- Always use explicit versions: `go install tool@vX.Y.Z`
- Version controlled in `.env.base`:
  - `GUARDIAN_ACT_VERSION=v0.2.84`
  - `GUARDIAN_ACTIONLINT_VERSION=v1.6.27`

### GitHub Actions
- All third-party actions must be SHA-pinned
- Format: `uses: owner/repo@<full-sha> # vX.Y.Z`
- Example: `uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2`

### Go Dependencies
- All module dependencies pinned in go.mod with exact versions
- No floating versions or ranges

---

## Code Quality Gates

After significant Go code changes, run the full quality pipeline:

```bash
# Full quality check (required before commits)
go-pre-commit --all-files && magex format:fix && magex lint
```

This ensures:
1. **go-pre-commit** - Pre-commit hooks pass (formatting, goimports, etc.)
2. **magex format:fix** - All formatting issues auto-fixed
3. **magex lint** - golangci-lint passes with no errors

Add this as a task requirement for all implementation tasks.

---

## Final Polish Step: Guardian README

Create `guardian/README.md` documenting:

1. **Package Overview** - What Guardian does
2. **Installation** - How to install dependencies (act, actionlint)
3. **Quick Start** - Basic usage via MAGE-X commands
4. **API Reference** - Link to GoDoc
5. **Configuration** - Environment variables and guardian.yaml
6. **Scenarios** - List of included test scenarios
7. **Contributing** - How to add new scenarios

This README is the primary documentation for users consuming the guardian package.

---

## .env.base Additions (Complete)

```bash
# ================================================================================================
# 🛡️ GUARDIAN CI TESTING FRAMEWORK
# ================================================================================================

# Feature Toggle
ENABLE_CI_GUARDIAN=true

# Tool Versions (pinned)
GUARDIAN_ACT_VERSION=v0.2.84
GUARDIAN_ACTIONLINT_VERSION=v1.6.27
GUARDIAN_GO_SARIF_VERSION=v2.4.0

# Execution Settings
GUARDIAN_SCENARIO_TIMEOUT=30s
GUARDIAN_STATIC_TIMEOUT=5s
GUARDIAN_PARALLEL_SCENARIOS=4

# Output Configuration
GUARDIAN_OUTPUT_DIR=.mage-x
GUARDIAN_SARIF_OUTPUT=guardian.sarif
GUARDIAN_JSONL_OUTPUT=ci-results.jsonl

# Policy Configuration
GUARDIAN_EXCEPTIONS_FILE=.github/guardian.yaml
GUARDIAN_POLICY_STRICT=true

# Debug Settings
GUARDIAN_VERBOSE=false
GUARDIAN_DRY_RUN=false
GUARDIAN_KEEP_CONTAINERS=false
```

All settings can be overridden in `.env.custom` per repository.
