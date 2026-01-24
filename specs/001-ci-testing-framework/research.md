# Research: Fortress Guardian CI Testing Framework

**Date**: 2026-01-24
**Branch**: `001-ci-testing-framework`

## Overview

This document consolidates research findings for implementing the Fortress Guardian CI testing framework. All "NEEDS CLARIFICATION" items from the Technical Context have been resolved.

---

## 1. nektos/act Integration

### Decision: CLI Subprocess Invocation

**Rationale**: The nektos/act project is fundamentally designed as a CLI tool, not a library. While Go packages exist (`pkg/runner`, `pkg/model`), they are internal APIs subject to change without notice. The most stable and maintainable approach is subprocess invocation.

**Alternatives Considered**:
- **Direct Go API (pkg/runner)**: Rejected - Internal packages may change between versions; pre-v1.0 stability not guaranteed
- **act-js Node.js wrapper**: Rejected - Violates Pure Go Philosophy (Constitution Principle I)
- **Custom GitHub Actions runner**: Rejected - Massive scope increase; act already provides this

### Implementation Pattern

```go
// runner/act.go
func (r *ActRunner) Run(ctx context.Context, opts RunOptions) (*Result, error) {
    args := []string{
        "--workflows", opts.WorkflowFile,
        "--job", opts.Job,
        "--eventpath", opts.EventPath,
        "--secret-file", opts.SecretsFile,
        "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest",
    }

    cmd := exec.CommandContext(ctx, "act", args...)
    // Capture stdout/stderr, parse exit code
}
```

### Key Capabilities

| Feature | Act Support | Implementation Strategy |
|---------|-------------|------------------------|
| Single job execution | Yes | `--job <job-name>` flag |
| Event payload injection | Yes | `--eventpath <json-file>` flag |
| Secret injection | Yes | `--secret-file <path>` flag |
| Container isolation | Yes (default) | Fresh container per scenario |
| Exit code capture | Yes | Process exit code |
| Output capture | Yes | Stdout/stderr piping |
| Timeout control | Yes | Context cancellation |

### Limitations to Work Around

1. **workflow_call unsupported**: Validate reusable workflow schemas statically; test calling workflows with mocked outputs
2. **Cache simulation limited**: Mock cache with local directory; test both hit/miss modes
3. **IPv6 not supported**: No workaround needed for Guardian use cases
4. **Log format**: Parse stderr; use regex patterns for validation

---

## 2. actionlint Integration

### Decision: Direct Go API

**Rationale**: actionlint provides a well-documented Go API (`github.com/rhysd/actionlint`) with stable types for programmatic linting. Unlike act, this is designed for library consumption.

**Alternatives Considered**:
- **CLI subprocess**: Rejected - Unnecessary overhead; Go API is mature and documented
- **Custom parser**: Rejected - Duplicates effort; actionlint already handles edge cases

### Implementation Pattern

```go
// validator/actionlint.go
func (v *ActionlintValidator) Validate(ctx context.Context, workflowPath string) ([]Finding, error) {
    opts := &actionlint.LinterOptions{
        Verbose: false,
        Color:   false,
    }
    linter, err := actionlint.NewLinter(io.Discard, opts)
    if err != nil {
        return nil, fmt.Errorf("creating linter: %w", err)
    }

    project, err := actionlint.NewProject(filepath.Dir(workflowPath))
    if err != nil {
        return nil, fmt.Errorf("detecting project: %w", err)
    }

    errs, err := linter.LintFile(workflowPath, project)
    if err != nil {
        return nil, fmt.Errorf("linting file: %w", err)
    }

    // Convert actionlint.Error to guardian.Finding
    return convertErrors(errs), nil
}
```

### Error Type Mapping

| actionlint Kind | Guardian Severity | Guardian Category |
|-----------------|-------------------|-------------------|
| `expression` | error | syntax |
| `permissions` | warning | security |
| `deprecated-command` | warning | deprecation |
| `shellcheck` | warning | quality |
| `step-id` | error | syntax |
| `job-needs` | error | config |

---

## 3. SARIF 2.1.0 Output

### Decision: Use go-sarif Library

**Rationale**: The `github.com/owenrumney/go-sarif` library provides full SARIF 2.1.0 compliance with GitHub Security tab compatibility. No need for custom JSON serialization.

**Alternatives Considered**:
- **Manual JSON serialization**: Rejected - Error-prone; schema is complex
- **Other Go SARIF libraries**: Rejected - go-sarif is most complete and maintained

### Implementation Pattern

```go
// reporter/sarif.go
import "github.com/owenrumney/go-sarif/sarif"

func (r *SARIFReporter) Write(ctx context.Context, findings []Finding) error {
    report, err := sarif.New(sarif.Version210)
    if err != nil {
        return fmt.Errorf("creating sarif report: %w", err)
    }

    run := sarif.NewRun("Fortress Guardian", "https://github.com/mrz1836/go-fortress")
    run.Tool.Driver.Version = "1.0.0"

    for _, f := range findings {
        result := sarif.NewResult(f.RuleID, sarifLevel(f.Severity), f.Message)
        result.WithLocation(sarif.NewLocation(f.File, f.Line, f.Column))
        run.AddResult(result)
    }

    report.AddRun(run)
    return report.WriteFile(r.OutputPath)
}
```

### GitHub Requirements

| Field | Required | Guardian Value |
|-------|----------|----------------|
| `$schema` | Yes | SARIF 2.1.0 schema URL |
| `version` | Yes | "2.1.0" |
| `tool.driver.name` | Yes | "Fortress Guardian" |
| `results[].ruleId` | Yes | From policy rule ID |
| `results[].level` | Yes | error/warning/note |
| `results[].message.text` | Yes | Finding description |
| `results[].locations` | Yes | File, line, column |
| `results[].partialFingerprints` | Recommended | Hash for deduplication |

### Severity Mapping

| Guardian Severity | SARIF Level | Security Score |
|-------------------|-------------|----------------|
| critical | error | 9.0-10.0 |
| high | error | 7.0-8.9 |
| medium | warning | 4.0-6.9 |
| low | note | 0.1-3.9 |
| info | none | N/A |

---

## 4. MAGE-X Integration

### Decision: Custom Namespace via magefile.go

**Rationale**: MAGE-X supports custom namespaces via type-based patterns. Creating a `ci:` namespace is straightforward and follows existing MAGE-X conventions.

**Implementation Pattern**:

```go
// magefile.go (repository root)
//go:build mage

package main

import "github.com/magefile/mage/mg"

type Ci mg.Namespace

func (c Ci) Test() error    { /* ... */ }
func (c Ci) Verify() error  { /* ... */ }
func (c Ci) Static() error  { /* ... */ }
func (c Ci) Scenario() error { /* ... */ }
func (c Ci) List() error    { /* ... */ }
```

### Parameter Handling

Parameters passed via `MAGE_ARGS` environment variable:

```bash
magex ci:scenario name=LINT-001 verbose=true
# Sets MAGE_ARGS="name=LINT-001 verbose=true"
```

```go
func getMageArgs() []string {
    if args := os.Getenv("MAGE_ARGS"); args != "" {
        return strings.Fields(args)
    }
    return nil
}
```

### Command Reference

| Command | Purpose | Parameters |
|---------|---------|------------|
| `ci:test` | Quick validation | None |
| `ci:verify` | Full verification | None |
| `ci:static` | Static analysis only | None |
| `ci:scenario` | Single scenario | `name=<ID>`, `verbose=bool`, `keep=bool` |
| `ci:list` | List scenarios | `filter=<category>` |

---

## 5. Fixture Design

### Decision: Minimal Go Modules Per Failure Type

**Rationale**: Each fixture is a valid Go module that fails in exactly one predictable way. This ensures scenario isolation and reproducible failures.

### Fixture Structure

```text
.github/ci-tester/fixtures/
├── lint-fail/
│   ├── go.mod          # module fixture-lint-fail
│   ├── main.go         # Contains unused variable
│   └── .github/
│       └── workflows/
│           └── ci.yml  # Standard lint workflow
├── test-fail/
│   ├── go.mod
│   ├── main.go
│   ├── main_test.go    # Contains failing test
│   └── .github/workflows/ci.yml
├── race-fail/
│   ├── go.mod
│   ├── main.go         # Contains concurrent map write
│   ├── main_test.go    # Triggers race with -race flag
│   └── .github/workflows/ci.yml
├── sec-fail/
│   ├── go.mod
│   ├── main.go         # Contains AWS key pattern
│   └── .github/workflows/ci.yml
├── vuln-fail/
│   ├── go.mod          # Depends on package with known CVE
│   ├── go.sum
│   ├── main.go
│   └── .github/workflows/ci.yml
└── cov-fail/
    ├── go.mod
    ├── main.go         # Has functions
    ├── main_test.go    # Tests only 50% of functions
    └── .github/workflows/ci.yml
```

### Fixture Guidelines

1. **Single Failure Mode**: Each fixture fails for exactly one reason
2. **Minimal Dependencies**: Only what's needed to trigger the failure
3. **Standard Workflow**: Use workflows matching go-fortress patterns
4. **Version Controlled**: Fixtures are committed to repository
5. **Documented**: README in each fixture explains the expected failure

---

## 6. Policy Engine Design

### Decision: Go Functions as Policies

**Rationale**: Following Constitution Principle II (Multi-Stage Defense) and the spec requirement for "policy-as-code", policies are implemented as Go functions. This provides type safety, IDE support, and testability.

**Alternatives Considered**:
- **OPA/Rego**: Rejected - Future roadmap item; adds external dependency
- **YAML configuration**: Rejected - Less expressive; harder to test
- **JSON Schema**: Rejected - Limited to structure validation

### Policy Structure

```go
// policy/rules.go
type Policy struct {
    ID          string
    Severity    Severity
    Description string
    Check       func(workflow *Workflow) []Finding
}

var SHAPinnedActions = Policy{
    ID:          "sha-pinned-actions",
    Severity:    Error,
    Description: "All actions must be pinned to full commit SHA",
    Check: func(w *Workflow) []Finding {
        // Implementation
    },
}

var ExplicitPermissions = Policy{
    ID:          "explicit-permissions",
    Severity:    Warning,
    Description: "Workflows should declare explicit permissions",
    Check: func(w *Workflow) []Finding {
        // Implementation
    },
}
```

### Built-in Policies

| Policy ID | Severity | Description |
|-----------|----------|-------------|
| `sha-pinned-actions` | error | All actions must be SHA-pinned |
| `explicit-permissions` | warning | Permissions must be declared |
| `no-dangerous-workflows` | error | No pull_request_target with write |
| `no-secret-logging` | error | Secrets must not be logged |
| `concurrency-defined` | warning | Concurrency groups recommended |
| `minimal-permissions` | warning | Use least-privilege permissions |

---

## 7. Dependencies Summary

### Required Dependencies

| Dependency | Purpose | Version | Install Method |
|------------|---------|---------|----------------|
| `nektos/act` | Workflow execution | v0.2.84+ | `go install` |
| `rhysd/actionlint` | Static analysis | v1.6.27+ | `go install` / Go import |
| `owenrumney/go-sarif` | SARIF output | v2.4+ | Go import |
| `gopkg.in/yaml.v3` | YAML parsing | v3.0+ | Go import (existing) |

### go.mod Additions

```go
require (
    github.com/rhysd/actionlint v1.6.27
    github.com/owenrumney/go-sarif/v2 v2.4.0
)
```

### Tool Installation (deps:install)

```bash
# Add to deps:install target
go install github.com/nektos/act@v0.2.84
go install github.com/rhysd/actionlint/cmd/actionlint@v1.6.27
```

---

## 8. Environment Variables

### New .env.base Additions

```bash
# Guardian CI Testing Framework
ENABLE_CI_GUARDIAN=true
GUARDIAN_ACT_VERSION=v0.2.84
GUARDIAN_ACTIONLINT_VERSION=v1.6.27
GUARDIAN_SCENARIO_TIMEOUT=30s
GUARDIAN_PARALLEL_SCENARIOS=4
GUARDIAN_STATIC_TIMEOUT=5s
GUARDIAN_OUTPUT_DIR=.mage-x
```

---

## 9. CI Security Best Practices (from go-fortress)

### 1. Explicit Permissions

```yaml
# Default to no permissions
permissions: {}

jobs:
  job-name:
    permissions:
      contents: read  # Only what's needed
      actions: write  # If cancellation needed
```

### 2. Fork Safety Detection

- Check `github.event.pull_request.head.repo.full_name` vs `github.repository`
- Skip secret-sensitive operations on fork PRs
- Use separate workflows for fork-safe and fork-unsafe jobs

### 3. SHA-Pinned Actions

- Never use tag references (`@v4`) in production workflows
- Always use full commit SHA with comment showing version
- Format: `uses: owner/repo@<full-sha> # vX.Y.Z`
- Rationale: Tags can be moved, SHAs are immutable

### 4. Secret Handling

- Never log secrets (use `add-mask` for dynamic values)
- Use file-based secret injection (not environment variables)
- Secrets not available to fork PR workflows

### 5. Input Validation

- Validate all `workflow_call` inputs
- Use explicit types (string, boolean, number)
- Provide defaults for optional inputs

---

## Sources

- [nektos/act GitHub Repository](https://github.com/nektos/act)
- [act User Guide](https://nektosact.com/)
- [rhysd/actionlint GitHub Repository](https://github.com/rhysd/actionlint)
- [actionlint Go API Documentation](https://pkg.go.dev/github.com/rhysd/actionlint)
- [SARIF 2.1.0 Specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
- [GitHub SARIF Support](https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning)
- [go-sarif Library](https://github.com/owenrumney/go-sarif)
- [MAGE-X Documentation](https://github.com/mrz1836/mage-x)
