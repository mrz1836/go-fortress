# Feature Specification: Fortress Guardian CI Testing Framework

**Feature Branch**: `001-ci-testing-framework`
**Created**: 2026-01-24
**Status**: Draft
**Input**: User description: "Fortress Guardian: A Go-native CI validation framework that tests the CI system itself, enabling local reproduction of CI failures and validating GitHub Actions workflows before they reach master."

## Executive Overview

Fortress Guardian is the final line of defense for the Go-Fortress ecosystem - a CI testing framework that treats GitHub Actions workflows as testable code. By enabling accurate local reproduction of all CI failure states, Guardian eliminates "commit-and-pray" workflow development and ensures no broken CI configuration ever reaches the `master` branch.

### Stakeholder Value

| Stakeholder | Benefit |
|-------------|---------|
| **Developers** | Instant local feedback on workflow changes before pushing |
| **Maintainers** | Confidence that CI changes won't break production pipelines |
| **Security Teams** | Verified supply chain controls and policy enforcement |
| **Operations** | Predictable CI behavior with documented failure modes |

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Instant Static Validation (Priority: P1)

As a developer modifying GitHub Actions workflows, I want instant validation of my changes locally so that I can catch syntax errors, policy violations, and deprecated actions before pushing code.

**Why this priority**: This is the fastest path to value. Static analysis catches 80% of CI issues in under 2 seconds, requires no Docker, and provides immediate feedback during development. Developers can validate on every save.

**Independent Test**: Can be fully tested by running `magex ci:static` on any repository with workflows. Delivers immediate value by catching actionlint errors, unpinned actions, and policy violations without container overhead.

**Acceptance Scenarios**:

1. **Given** a workflow file with a syntax error, **When** I run static validation, **Then** I see the error location and description within 2 seconds
2. **Given** a workflow using an unpinned action (tag instead of SHA), **When** I run static validation, **Then** I see a warning with the action name and suggested SHA pin
3. **Given** a workflow with deprecated runner labels, **When** I run static validation, **Then** I see deprecation warnings with migration guidance
4. **Given** a workflow missing explicit permissions, **When** I run static validation, **Then** I see a warning recommending permission declarations
5. **Given** a valid workflow with no issues, **When** I run static validation, **Then** I see a success message confirming validation passed

---

### User Story 2 - CI Failure Reproduction (Priority: P1)

As a developer investigating a CI failure, I want to reproduce the exact failure locally so that I can debug the issue without pushing commits and waiting for CI runs.

**Why this priority**: Equal priority with static validation because this is the core value proposition - eliminating the commit-and-pray debugging cycle. Saves hours of developer time per CI issue.

**Independent Test**: Can be fully tested by selecting a known failure scenario (e.g., lint failure) and verifying it produces identical results locally. Delivers value by enabling offline debugging.

**Acceptance Scenarios**:

1. **Given** a fixture repository with a linting violation, **When** I run the lint failure scenario, **Then** I see the same error output as GitHub Actions produces
2. **Given** a fixture with a failing unit test, **When** I run the test failure scenario, **Then** I see the test failure with stack trace matching CI output
3. **Given** a fixture with a data race condition, **When** I run the race detection scenario, **Then** I see the "DATA RACE" warning with goroutine traces
4. **Given** a fixture with a hardcoded secret pattern, **When** I run the security scan scenario, **Then** I see the Gitleaks detection matching CI behavior
5. **Given** any scenario execution, **When** it completes, **Then** the exit code matches what GitHub Actions would return (0 for success, 1 for failure)

---

### User Story 3 - Comprehensive Pre-Merge Validation (Priority: P2)

As a maintainer reviewing a PR that modifies CI configuration, I want comprehensive validation of all CI scenarios so that I have confidence the changes won't break production pipelines.

**Why this priority**: Critical for maintainers but depends on P1 foundation. Runs the full matrix of scenarios to validate complete CI behavior before merge.

**Independent Test**: Can be fully tested by running `magex ci:verify` and confirming all defined scenarios execute with expected outcomes. Delivers confidence through comprehensive coverage.

**Acceptance Scenarios**:

1. **Given** a PR modifying workflow files, **When** I run full verification, **Then** all 35+ scenarios execute with pass/fail results
2. **Given** verification completes, **When** I view results, **Then** I see a summary with passed/failed/skipped counts and execution duration
3. **Given** any scenario fails unexpectedly, **When** verification completes, **Then** the command exits with code 1 and highlights the failure
4. **Given** verification runs in CI, **When** it completes, **Then** a SARIF report is uploaded for GitHub Security integration
5. **Given** verification runs in CI, **When** it completes, **Then** the GitHub Step Summary displays a formatted results table

---

### User Story 4 - Single Scenario Debugging (Priority: P2)

As a developer debugging a specific CI issue, I want to run and inspect a single scenario in isolation so that I can iterate quickly on fixes.

**Why this priority**: Essential for debugging workflow but depends on the execution engine from P1. Enables rapid iteration on specific issues.

**Independent Test**: Can be fully tested by running `magex ci:scenario name=LINT-001 verbose=true` and confirming detailed output. Delivers value through focused debugging.

**Acceptance Scenarios**:

1. **Given** I specify a scenario ID, **When** I run it in verbose mode, **Then** I see detailed execution logs including container output
2. **Given** I run a scenario with the `keep=true` flag, **When** it completes, **Then** the container remains running for manual inspection
3. **Given** I specify a non-existent scenario ID, **When** I run it, **Then** I see a clear error message listing available scenarios
4. **Given** a scenario execution, **When** it times out, **Then** I see a timeout error with the configured limit

---

### User Story 5 - Policy Enforcement (Priority: P2)

As a security engineer, I want automated enforcement of workflow security policies so that dangerous patterns are blocked before reaching production.

**Why this priority**: Critical for security but can be implemented in parallel with execution engine. Provides automated guardrails.

**Independent Test**: Can be fully tested by introducing policy violations and confirming they're detected. Delivers value through automated security controls.

**Acceptance Scenarios**:

1. **Given** a workflow with unpinned actions, **When** policy check runs, **Then** violations are reported as errors (not just warnings)
2. **Given** a workflow using `pull_request_target` with write permissions, **When** policy check runs, **Then** a "dangerous workflow" error is raised
3. **Given** a workflow that logs secrets, **When** policy check runs, **Then** a "secret exposure" error is raised
4. **Given** a workflow missing concurrency groups, **When** policy check runs, **Then** a warning recommends adding concurrency
5. **Given** a policy exception configured in `.github/guardian.yaml`, **When** policy check runs, **Then** the excepted rule is skipped with an audit note

---

### User Story 6 - Supply Chain Validation (Priority: P3)

As a compliance officer, I want validation that our CI produces verifiable artifacts with provenance attestations so that we meet SLSA Level 2+ requirements.

**Why this priority**: Important for enterprise compliance but depends on core execution engine. Can be implemented after MVP core is stable.

**Independent Test**: Can be fully tested by running supply chain scenarios and verifying attestation checks. Delivers value through compliance assurance.

**Acceptance Scenarios**:

1. **Given** a release workflow, **When** I run provenance validation, **Then** I confirm attestation action is present and configured correctly
2. **Given** a release build, **When** supply chain validation runs, **Then** it verifies build isolation settings
3. **Given** release dependencies, **When** supply chain validation runs, **Then** it verifies all dependencies are pinned to digests
4. **Given** SBOM generation in release workflow, **When** validation runs, **Then** it confirms SPDX/CycloneDX format compliance

---

### User Story 7 - Scenario Discovery (Priority: P3)

As a new team member, I want to list and understand available scenarios so that I can learn about CI validation capabilities.

**Why this priority**: Improves discoverability but is not core functionality. Nice-to-have for onboarding.

**Independent Test**: Can be fully tested by running `magex ci:list` and confirming clear output. Delivers value through improved discoverability.

**Acceptance Scenarios**:

1. **Given** I run the list command, **When** it completes, **Then** I see all scenarios with ID, category, and description
2. **Given** I filter by category (e.g., `filter=security`), **When** list runs, **Then** only scenarios in that category appear
3. **Given** the scenario list, **When** I view it, **Then** each scenario indicates whether it's a pass or fail test

---

### Edge Cases

- What happens when Docker is not available? (Graceful degradation to static-only validation with clear messaging)
- What happens when a scenario exceeds timeout? (Clean container termination with timeout error message)
- What happens when Act fails to pull an image? (Retry with exponential backoff, then fail with network guidance)
- How does the system handle parallel scenario failures? (Collect all results before reporting aggregate failure)
- What happens when fixture files are missing or corrupted? (Clear error identifying the missing fixture path)
- How does the system behave with insufficient disk space? (Fail fast with disk space warning before starting)
- What happens when a user runs scenarios on Windows without WSL2? (Clear error with WSL2 installation guidance)
- What happens when container runtime permissions are insufficient? (Clear error with permissions fix guidance)

---

## Requirements *(mandatory)*

### Functional Requirements

#### Core Execution

- **FR-001**: System MUST wrap the `act` tool to execute GitHub Actions workflows locally with deterministic results
- **FR-002**: System MUST run each scenario in complete isolation (fresh container per scenario, no shared state)
- **FR-003**: System MUST inject synthetic GitHub event payloads (Push, Pull Request, Release) into scenario execution
- **FR-004**: System MUST capture and parse workflow output to determine pass/fail status
- **FR-005**: System MUST support running specific jobs within a workflow (not full workflow execution)
- **FR-006**: System MUST provide secret injection via secure file-based mechanism (no environment variable exposure)
- **FR-007**: System MUST support context cancellation and configurable timeouts for all scenario execution

#### Static Validation

- **FR-010**: System MUST validate workflow YAML syntax using actionlint
- **FR-011**: System MUST detect deprecated action versions and runner labels
- **FR-012**: System MUST validate `action.yml` files for composite actions against schema
- **FR-013**: System MUST validate `.env.base` environment variable schema (types, required fields, naming conventions)
- **FR-014**: System MUST complete static validation within 5 seconds for typical repository

#### Policy Engine

- **FR-020**: System MUST enforce SHA-pinning requirement for all third-party actions (configurable severity)
- **FR-021**: System MUST detect and block dangerous workflow patterns (`pull_request_target` with write permissions)
- **FR-022**: System MUST warn on missing explicit permissions declarations
- **FR-023**: System MUST warn on missing concurrency group definitions
- **FR-024**: System MUST support policy exceptions via declarative configuration file with expiration dates
- **FR-025**: System MUST audit all policy exceptions in output reports

#### Scenario Framework

- **FR-030**: System MUST support defining scenarios as Go structs with type safety
- **FR-031**: System MUST maintain fixture repositories for each failure type (lint, test, race, security, coverage)
- **FR-032**: System MUST validate scenario expected outcomes (exit code, log patterns, specific outputs)
- **FR-033**: System MUST support scenario categorization (Quality, Testing, Security, Coverage, Fork Safety, etc.)
- **FR-034**: System MUST provide 35+ pre-defined scenarios covering all documented failure modes

#### Reporting

- **FR-040**: System MUST generate JSONL output compatible with existing CI results format
- **FR-041**: System MUST generate SARIF 2.1.0 output for GitHub Security integration
- **FR-042**: System MUST generate GitHub annotations for inline PR feedback
- **FR-043**: System MUST generate GitHub Step Summary markdown when running in CI environment
- **FR-044**: System MUST provide terminal output with pass/fail highlighting for local development

#### MAGE-X Integration

- **FR-050**: System MUST expose `ci:test` command for quick developer validation
- **FR-051**: System MUST expose `ci:verify` command for comprehensive pre-merge validation
- **FR-052**: System MUST expose `ci:static` command for fast static-only analysis
- **FR-053**: System MUST expose `ci:scenario` command for single scenario execution with debug options
- **FR-054**: System MUST expose `ci:list` command for scenario discovery with filtering
- **FR-055**: System MUST integrate with `deps:install` to install required tools (act, actionlint)

#### Fork Safety

- **FR-060**: System MUST validate fork detection mechanism in workflows
- **FR-061**: System MUST verify secret protection on fork pull requests
- **FR-062**: System MUST validate appropriate job skipping for fork PRs

### Key Entities

- **Scenario**: A single test case with ID, category, description, fixture path, event type, expected result (status, log patterns, outputs), and timeout configuration
- **Fixture**: A pre-configured repository state designed to fail in a specific, predictable way (located in `.github/ci-tester/fixtures/`)
- **Policy**: A rule with ID, severity, description, and check function that validates workflow content against security/quality standards
- **Finding**: A validation result with severity, location (file, line), message, and rule ID
- **Report**: Aggregated results containing run metadata, scenario results, findings, and performance metrics

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Static validation (`ci:static`) completes in under 2 seconds for repositories with up to 20 workflow files
- **SC-002**: Quick test suite (`ci:test`) completes in under 60 seconds with all fast scenarios
- **SC-003**: Full verification (`ci:verify`) completes in under 5 minutes with all 35+ scenarios
- **SC-004**: Individual scenario execution completes in under 30 seconds (excluding container pull time)
- **SC-005**: Zero false positives - every failure scenario fails for the expected reason (verified by log pattern matching)
- **SC-006**: Zero false negatives - every passing scenario passes with expected outputs
- **SC-007**: Local execution produces identical results to CI execution for all scenarios
- **SC-008**: All policy violations are detected with actionable fix suggestions
- **SC-009**: 100% of defined scenarios are independently runnable and debuggable
- **SC-010**: System gracefully handles Docker unavailability by falling back to static-only validation with clear messaging

### Acceptance Threshold

The MVP is complete when:
- All P1 user stories pass their acceptance scenarios
- At least 80% of defined scenarios (28/35) are implemented and passing
- Static validation catches all actionlint issues and policy violations
- Documentation exists for all MAGE-X commands
- System runs successfully on macOS (Apple Silicon), macOS (Intel), and Linux (x86_64)

---

## Assumptions

1. **Docker Availability**: Docker or Podman is available for scenario execution (static validation works without)
2. **MAGE-X Ecosystem**: MAGE-X build system is already integrated and functional
3. **Go Version**: Go 1.24+ is available (current go.mod specifies 1.24)
4. **Network Access**: Internet access available for initial tool/image downloads (offline mode supported after setup)
5. **Disk Space**: At least 10 GB available for container images and workspace
6. **Single Module**: Guardian is a package within the existing go-fortress module (not a separate module)
7. **Act Limitations**: Reusable workflows (`workflow_call`) cannot be directly tested via Act; schema validation and caller-workflow testing provide coverage
8. **Fixture Strategy**: Pre-broken fixtures are version-controlled and maintained as valid Go modules that fail in predictable ways

---

## Scope Boundaries

### In Scope

- Static validation (actionlint, schema, deprecation, policy)
- Scenario execution via Act
- Policy-as-code engine
- JSONL/SARIF/Markdown reporting
- MAGE-X command integration
- Fork safety validation
- Supply chain validation (SLSA basics)
- macOS and Linux support

### Out of Scope (Future Roadmap)

- AI-assisted flaky test detection
- Ephemeral preview environments
- Kubernetes executor
- OPA/Rego policy language
- OpenTelemetry integration
- Windows native support (WSL2 required)
- Full OIDC token mocking
- Actions cache simulation beyond basic hit/miss

---

## Dependencies

- **nektos/act**: Local GitHub Actions runner (Go-installable)
- **actionlint**: Static analyzer for GitHub Actions (Go-native binary)
- **Docker/Podman**: Container runtime for scenario execution
- **MAGE-X**: Existing build system for command integration
- **go-fortress module**: Parent module for package integration
