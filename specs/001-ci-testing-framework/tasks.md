# Tasks: Fortress Guardian CI Testing Framework

**Input**: Design documents from `/specs/001-ci-testing-framework/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/guardian-api.go, quickstart.md

**Tests**: Tests are NOT explicitly requested in the feature specification. Test tasks are omitted per project guidelines.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, etc.)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- **Guardian package**: `guardian/` at repository root
- **Subpackages**: `guardian/runner/`, `guardian/validator/`, `guardian/policy/`, `guardian/reporter/`, `guardian/scenarios/`
- **Golden files**: `testdata/guardian/`
- **Fixtures**: `.github/ci-tester/fixtures/`
- **MAGE-X integration**: `magefiles/ci.go`

---

## Phase 1: Setup (Project Infrastructure)

**Purpose**: Initialize project structure, dependencies, and configuration

- [X] T001 Update go.mod to add required dependencies (actionlint, go-sarif) in go.mod
- [X] T002 Create guardian package directory structure per plan.md
- [X] T003 [P] Create guardian/guardian.go with package entry point and public API stub
- [X] T004 [P] Create guardian/config.go with Config struct and DefaultConfig() function
- [X] T005 [P] Add Guardian environment variables to .env.base (ENABLE_CI_GUARDIAN, tool versions, settings)
- [X] T006 [P] Create .github/guardian.yaml with empty exceptions config structure

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 Implement config loading from environment in guardian/config.go (parse GUARDIAN_* env vars)
- [X] T008 [P] Create guardian/validator/finding.go with Finding struct and Severity constants
- [X] T009 [P] Create guardian/validator/validator.go with Validator interface and ValidatorRegistry
- [X] T010 [P] Create guardian/reporter/report.go with Report, StaticResults, ScenarioResult structs
- [X] T011 [P] Create guardian/reporter/reporter.go with Reporter interface and ReporterRegistry
- [X] T012 Create guardian/policy/workflow.go with Workflow, Job, Step, Permissions structs for parsed workflows
- [X] T013 Create MAGE-X ci namespace skeleton in magefiles/ci.go with Ci mg.Namespace type
- [X] T014 [P] Create testdata/guardian/events/ directory with sample push.json event payload
- [X] T015 [P] Create testdata/guardian/workflows/ directory with sample valid and invalid workflow files

**Checkpoint**: Foundation ready - user story implementation can now begin

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 3: User Story 1 - Instant Static Validation (Priority: P1)

**Goal**: Enable instant validation of workflow changes locally to catch syntax errors, policy violations, and deprecated actions in under 2 seconds

**Independent Test**: Run `magex ci:static` on any repository with workflows. Should complete < 2s and report actionlint errors, unpinned actions, and policy violations.

### Implementation for User Story 1

- [X] T016 [US1] Implement actionlint wrapper in guardian/validator/actionlint.go using Go API
- [X] T017 [US1] Implement error type mapping from actionlint.Error to Finding in guardian/validator/actionlint.go
- [X] T018 [P] [US1] Implement action.yml schema validation in guardian/validator/schema.go
- [X] T019 [P] [US1] Implement deprecated action/runner detection in guardian/validator/deprecation.go
- [X] T020 [US1] Create guardian/policy/engine.go with PolicyEngine interface and basic implementation
- [X] T021 [US1] Create guardian/policy/rules.go with Policy struct and PolicyCheckFunc type
- [X] T022 [US1] Implement sha-pinned-actions policy rule in guardian/policy/rules.go
- [X] T023 [P] [US1] Implement explicit-permissions policy rule in guardian/policy/rules.go
- [X] T024 [P] [US1] Implement no-dangerous-workflows policy rule (pull_request_target detection) in guardian/policy/rules.go
- [X] T025 [P] [US1] Implement no-secret-logging policy rule in guardian/policy/rules.go
- [X] T026 [P] [US1] Implement concurrency-defined policy rule in guardian/policy/rules.go
- [X] T027 [US1] Create guardian/policy/exceptions.go with Exception struct and ExceptionConfig loading
- [X] T028 [US1] Implement exception matching logic in guardian/policy/exceptions.go (glob patterns, expiration)
- [X] T029 [US1] Implement RunStatic method in guardian/guardian.go orchestrating validators and policy engine
- [X] T030 [US1] Implement terminal reporter for static results in guardian/reporter/terminal.go
- [X] T031 [US1] Implement ci:static MAGE-X command in magefiles/ci.go calling Guardian.RunStatic
- [X] T121 [US1] Implement .env.base schema validator in guardian/validator/env.go (type validation, required fields, naming conventions per FR-013)

**Checkpoint**: User Story 1 complete - `magex ci:static` should validate workflows in < 2s

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 4: User Story 2 - CI Failure Reproduction (Priority: P1)

**Goal**: Enable local reproduction of CI failures to debug issues without pushing commits

**Independent Test**: Run a known failure scenario (e.g., lint failure) and verify it produces identical results locally. Should enable offline debugging.

### Implementation for User Story 2

- [X] T032 [US2] Create guardian/runner/runner.go with Runner interface
- [X] T033 [US2] Implement act availability check in guardian/runner/act.go (CheckAvailable method)
- [X] T034 [US2] Implement act CLI wrapper in guardian/runner/act.go (Run method with subprocess invocation)
- [X] T035 [US2] Implement RunOptions and RunResult types in guardian/runner/runner.go
- [X] T036 [US2] Create guardian/runner/events.go for GitHub event payload injection helpers
- [X] T037 [US2] Create guardian/scenarios/scenarios.go with Scenario struct and Category constants
- [X] T038 [US2] Create guardian/scenarios/registry.go with scenario registry and lookup functions
- [X] T039 [US2] Create guardian/runner/scenario.go with scenario execution logic and result validation
- [X] T040 [US2] Implement log pattern matching for ExpectedResult validation in guardian/runner/scenario.go
- [X] T041 [P] [US2] Create fixture .github/ci-tester/fixtures/lint-fail/ with go.mod, main.go (unused var), ci.yml
- [X] T042 [P] [US2] Create fixture .github/ci-tester/fixtures/test-fail/ with go.mod, main.go, main_test.go (failing assert), ci.yml
- [X] T043 [P] [US2] Create fixture .github/ci-tester/fixtures/race-fail/ with go.mod, main.go (concurrent map), main_test.go, ci.yml
- [X] T044 [P] [US2] Create fixture .github/ci-tester/fixtures/sec-fail/ with go.mod, main.go (AWS key pattern), ci.yml
- [X] T045 [P] [US2] Create fixture .github/ci-tester/fixtures/vuln-fail/ with go.mod (CVE dep), go.sum, main.go, ci.yml
- [X] T046 [P] [US2] Create fixture .github/ci-tester/fixtures/cov-fail/ with go.mod, main.go, main_test.go (50% coverage), ci.yml
- [X] T047 [US2] Define LINT-001 scenario (unused variable detection) in guardian/scenarios/quality.go
- [X] T048 [P] [US2] Define LINT-002 scenario (gofmt violation) in guardian/scenarios/quality.go
- [X] T049 [P] [US2] Define LINT-003 scenario (golangci-lint violation) in guardian/scenarios/quality.go
- [X] T050 [US2] Define TEST-001 scenario (failing unit test) in guardian/scenarios/quality.go
- [X] T051 [P] [US2] Define TEST-002 scenario (test panic) in guardian/scenarios/quality.go
- [X] T052 [US2] Define RACE-001 scenario (data race detection) in guardian/scenarios/quality.go
- [X] T053 [US2] Define SEC-001 scenario (hardcoded AWS key) in guardian/scenarios/security.go
- [X] T054 [P] [US2] Define SEC-002 scenario (private key detection) in guardian/scenarios/security.go
- [X] T055 [US2] Define VULN-001 scenario (vulnerable dependency) in guardian/scenarios/security.go
- [X] T056 [US2] Define COV-001 scenario (low coverage threshold) in guardian/scenarios/quality.go

**Checkpoint**: User Story 2 complete - Individual scenarios can be executed locally and reproduce CI failures

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 5: User Story 3 - Comprehensive Pre-Merge Validation (Priority: P2)

**Goal**: Enable comprehensive validation of all CI scenarios for pre-merge confidence

**Independent Test**: Run `magex ci:verify` and confirm all 35+ scenarios execute with expected outcomes in < 5 minutes

### Implementation for User Story 3

- [X] T057 [US3] Implement parallel scenario execution in guardian/runner/scenario.go (respecting ParallelScenarios config)
- [X] T058 [US3] Implement RunTest method in guardian/guardian.go (static + fast scenarios)
- [X] T059 [US3] Implement RunVerify method in guardian/guardian.go (static + all scenarios)
- [X] T060 [US3] Implement JSONL reporter in guardian/reporter/jsonl.go with JSONL output format
- [X] T061 [US3] Implement SARIF reporter in guardian/reporter/sarif.go using go-sarif library
- [X] T062 [US3] Implement GitHub annotations reporter in guardian/reporter/annotations.go
- [X] T063 [US3] Implement GitHub Step Summary generation in guardian/reporter/summary.go (markdown table)
- [X] T064 [US3] Implement ReportSummary calculation in guardian/reporter/report.go
- [X] T065 [US3] Implement ci:test MAGE-X command in magefiles/ci.go calling Guardian.RunTest
- [X] T066 [US3] Implement ci:verify MAGE-X command in magefiles/ci.go calling Guardian.RunVerify
- [X] T067 [P] [US3] Create .github/workflows/fortress-guardian.yml workflow for Guardian CI integration
- [X] T068 [P] [US3] Add SARIF upload step to fortress-guardian.yml for GitHub Security tab integration

**Checkpoint**: User Story 3 complete - `magex ci:verify` runs all scenarios with full reporting

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 6: User Story 4 - Single Scenario Debugging (Priority: P2)

**Goal**: Enable running and inspecting a single scenario in isolation for rapid debugging

**Independent Test**: Run `magex ci:scenario name=LINT-001 verbose=true` and confirm detailed output

### Implementation for User Story 4

- [X] T069 [US4] Implement RunScenario method in guardian/guardian.go with ScenarioOptions support
- [X] T070 [US4] Add verbose output mode to scenario execution in guardian/runner/scenario.go
- [X] T071 [US4] Add keep-container support for debugging in guardian/runner/act.go
- [X] T072 [US4] Add timeout override support in guardian/runner/scenario.go
- [X] T073 [US4] Implement ci:scenario MAGE-X command in magefiles/ci.go with name, verbose, keep, timeout params

**Checkpoint**: User Story 4 complete - Single scenarios can be debugged with verbose output and container preservation

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 7: User Story 5 - Policy Enforcement (Priority: P2)

**Goal**: Automate enforcement of workflow security policies to block dangerous patterns

**Independent Test**: Introduce policy violations (unpinned actions, dangerous workflows) and confirm they're detected as errors

### Implementation for User Story 5

- [X] T074 [US5] Implement policy severity escalation (warnings to errors for specific rules) in guardian/policy/engine.go
- [X] T075 [US5] Add pull_request_target with write permissions detection in guardian/policy/rules.go (detailed check)
- [X] T076 [US5] Add secret logging pattern detection in guardian/policy/rules.go (comprehensive patterns)
- [X] T077 [US5] Implement minimal-permissions policy rule in guardian/policy/rules.go
- [X] T078 [US5] Implement exception audit logging in guardian/policy/exceptions.go
- [X] T079 [US5] Add expiration checking for exceptions in guardian/policy/exceptions.go

**Checkpoint**: User Story 5 complete - Policies are enforced with configurable severity and auditable exceptions

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 8: User Story 6 - Supply Chain Validation (Priority: P3)

**Goal**: Validate that CI produces verifiable artifacts with provenance attestations for SLSA compliance

**Independent Test**: Run supply chain scenarios and verify attestation checks report correctly

### Implementation for User Story 6

- [X] T080 [US6] Define SLSA-001 scenario (provenance attestation present) in guardian/scenarios/supply_chain.go
- [X] T081 [P] [US6] Define SLSA-002 scenario (build isolation settings) in guardian/scenarios/supply_chain.go
- [X] T082 [P] [US6] Define SLSA-003 scenario (dependencies pinned to digests) in guardian/scenarios/supply_chain.go
- [X] T083 [US6] Define SBOM-001 scenario (SPDX/CycloneDX format compliance) in guardian/scenarios/supply_chain.go
- [X] T084 [US6] Implement supply chain validation logic in guardian/validator/supply_chain.go

**Checkpoint**: User Story 6 complete - Supply chain compliance can be validated

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 9: User Story 7 - Scenario Discovery (Priority: P3)

**Goal**: Allow users to list and understand available scenarios for learning CI validation capabilities

**Independent Test**: Run `magex ci:list` and confirm clear output with all scenarios and categories

### Implementation for User Story 7

- [X] T085 [US7] Implement ListScenarios method in guardian/guardian.go with ScenarioFilter support
- [X] T086 [US7] Add category filtering to ListScenarios in guardian/guardian.go
- [X] T087 [US7] Add tag filtering to ListScenarios in guardian/guardian.go
- [X] T088 [US7] Implement ci:list MAGE-X command in magefiles/ci.go with filter parameter
- [X] T089 [US7] Implement formatted scenario list output in guardian/reporter/terminal.go

**Checkpoint**: User Story 7 complete - `magex ci:list` displays all scenarios with filtering

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 10: Additional Scenarios & Fork Safety

**Purpose**: Complete remaining scenarios to reach 35+ total

### Fork Safety Scenarios

- [X] T090 [P] Define FORK-001 scenario (fork detection mechanism validation) in guardian/scenarios/fork.go
- [X] T091 [P] Define FORK-002 scenario (secret protection on fork PRs) in guardian/scenarios/fork.go
- [X] T092 [P] Define FORK-003 scenario (job skipping for fork PRs) in guardian/scenarios/fork.go

### Configuration Scenarios

- [X] T093 [P] Define MATRIX-001 scenario (matrix expansion validation) in guardian/scenarios/config.go
- [X] T094 [P] Define ENV-001 scenario (.env.base schema validation: types, required fields, naming conventions per FR-013) in guardian/scenarios/config.go
- [X] T095 [P] Define CACHE-001 scenario (cache hit mode) in guardian/scenarios/config.go
- [X] T096 [P] Define CACHE-002 scenario (cache miss mode) in guardian/scenarios/config.go

### Additional Security Scenarios

- [X] T097 [P] Define GITLEAKS-001 scenario (Gitleaks integration) in guardian/scenarios/security.go
- [X] T098 [P] Define GITLEAKS-002 scenario (custom patterns) in guardian/scenarios/security.go

### Additional Quality Scenarios

- [X] T099 [P] Define BENCH-001 scenario (benchmark execution) in guardian/scenarios/quality.go
- [X] T100 [P] Define FUZZ-001 scenario (fuzz testing) in guardian/scenarios/quality.go

### Pass Scenarios (Success Cases)

- [X] T101 [P] Create fixture .github/ci-tester/fixtures/pass-basic/ with passing code
- [X] T102 [P] Define PASS-001 scenario (clean code passes all checks) in guardian/scenarios/quality.go
- [X] T103 [P] Define PASS-002 scenario (full CI pipeline passes) in guardian/scenarios/quality.go

### Additional Quality Scenarios (reaching 35+ total)

- [X] T113 [P] Define LINT-004 scenario (staticcheck violations) in guardian/scenarios/quality.go
- [X] T114 [P] Define TEST-003 scenario (test timeout exceeded) in guardian/scenarios/quality.go
- [X] T115 [P] Define COV-002 scenario (coverage threshold met - success case) in guardian/scenarios/quality.go

### Additional Security Scenarios (reaching 35+ total)

- [X] T116 [P] Define SEC-003 scenario (govulncheck findings) in guardian/scenarios/security.go
- [X] T117 [P] Define VULN-002 scenario (nancy CVE detection) in guardian/scenarios/security.go

### Workflow Validation Scenarios

- [X] T118 [P] Define WORKFLOW-001 scenario (invalid workflow YAML syntax) in guardian/scenarios/config.go
- [X] T119 [P] Define WORKFLOW-002 scenario (deprecated runner labels detection) in guardian/scenarios/config.go
- [X] T120 [P] Define ACTION-001 scenario (unpinned action static detection) in guardian/scenarios/config.go
- [X] T122 [P] Define ACTION-002 scenario (action.yml schema validation errors) in guardian/scenarios/config.go

**Checkpoint**: All 35+ scenarios defined and registered (37 total: 10 Phase 4 + 4 Phase 8 + 23 Phase 10)

**Quality Gate**: Run `go-pre-commit --all-files && magex format:fix && magex lint` after this phase

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Improvements affecting multiple components

- [X] T104 Implement graceful degradation when Docker is unavailable in guardian/guardian.go
- [X] T105 Add disk space check before scenario execution in guardian/runner/act.go
- [X] T106 Implement retry with exponential backoff for image pulls in guardian/runner/act.go
- [X] T107 [P] Add CI environment detection (GITHUB_ACTIONS) in guardian/reporter/reporter.go
- [X] T108 [P] Create testdata/guardian/reports/ with expected golden files for reporter tests
- [X] T109 Create guardian/README.md documenting package overview, installation, usage, API reference
- [X] T110 Update deps:install MAGE-X target to install act and actionlint with pinned versions
- [X] T111 Run quickstart.md validation - verify all documented commands work
- [X] T112 Performance validation - ensure ci:static < 2s, ci:test < 60s, ci:verify < 5m

**Quality Gate**: Final `go-pre-commit --all-files && magex format:fix && magex lint`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational - Static validation core
- **User Story 2 (Phase 4)**: Depends on Foundational - Scenario execution (can parallel with US1 models)
- **User Story 3 (Phase 5)**: Depends on US1 + US2 - Combines static + scenario execution
- **User Story 4 (Phase 6)**: Depends on US2 - Single scenario debugging
- **User Story 5 (Phase 7)**: Depends on US1 - Policy enhancement
- **User Story 6 (Phase 8)**: Depends on US2 - Supply chain scenarios
- **User Story 7 (Phase 9)**: Depends on US2 - Scenario listing
- **Additional Scenarios (Phase 10)**: Depends on US2 - More scenarios
- **Polish (Phase 11)**: Depends on all previous phases

### User Story Dependencies

```text
Setup (Phase 1)
    ↓
Foundational (Phase 2)
    ↓
    ├──→ US1: Static Validation (P1) ──→ US5: Policy Enforcement (P2)
    │
    └──→ US2: CI Reproduction (P1) ──┬──→ US3: Full Verification (P2)
                                     ├──→ US4: Single Scenario (P2)
                                     ├──→ US6: Supply Chain (P3)
                                     ├──→ US7: Discovery (P3)
                                     └──→ Additional Scenarios (Phase 10)
```

### Parallel Opportunities

**Within Phase 1 (Setup)**:
- T003, T004, T005, T006 can run in parallel

**Within Phase 2 (Foundational)**:
- T008, T009, T010, T011 can run in parallel (different files)
- T014, T015 can run in parallel (test data)

**Within Phase 3 (US1)**:
- T018, T019 can run in parallel (different validators)
- T023, T024, T025, T026 can run in parallel (different policy rules)

**Within Phase 4 (US2)**:
- T041, T042, T043, T044, T045, T046 can run in parallel (different fixtures)
- T048, T049, T051, T054 can run in parallel (different scenarios)

**Cross-Phase Parallelism**:
- Once Foundational completes, US1 and US2 can progress in parallel
- US3 requires both US1 and US2 complete
- US4, US5, US6, US7 can progress once their dependencies are met

---

## Parallel Examples

### Phase 2: Foundational Parallelism

```bash
# Launch all foundational entity files together:
Task: "Create guardian/validator/finding.go with Finding struct"
Task: "Create guardian/validator/validator.go with Validator interface"
Task: "Create guardian/reporter/report.go with Report structs"
Task: "Create guardian/reporter/reporter.go with Reporter interface"
```

### Phase 4: Fixture Creation Parallelism

```bash
# Launch all fixture directories together:
Task: "Create fixture .github/ci-tester/fixtures/lint-fail/"
Task: "Create fixture .github/ci-tester/fixtures/test-fail/"
Task: "Create fixture .github/ci-tester/fixtures/race-fail/"
Task: "Create fixture .github/ci-tester/fixtures/sec-fail/"
Task: "Create fixture .github/ci-tester/fixtures/vuln-fail/"
Task: "Create fixture .github/ci-tester/fixtures/cov-fail/"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Static Validation)
4. Complete Phase 4: User Story 2 (CI Reproduction)
5. **STOP and VALIDATE**: Test `magex ci:static` and individual scenarios independently
6. Deploy/demo if ready - this is the core value proposition

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Static) → Test independently → `magex ci:static` works (MVP start!)
3. Add US2 (Scenarios) → Test independently → Individual scenarios work (MVP complete!)
4. Add US3 (Full Verify) → Test independently → `magex ci:verify` works
5. Add US4-US7 → Test independently → Full feature set
6. Add remaining scenarios → 35+ scenarios complete

### Definition of Done per Phase

Each phase completion requires:
1. All tasks marked complete
2. Quality gate passes: `go-pre-commit --all-files && magex format:fix && magex lint`
3. Independent testability verified per checkpoint description

---

## Summary

| Phase | User Story | Priority | Task Count | Key Deliverable |
|-------|------------|----------|------------|-----------------|
| 1 | Setup | - | 6 | Project structure |
| 2 | Foundational | - | 9 | Core interfaces |
| 3 | US1 Static Validation | P1 | 17 | `magex ci:static` |
| 4 | US2 CI Reproduction | P1 | 25 | Scenario execution |
| 5 | US3 Full Verification | P2 | 12 | `magex ci:verify` |
| 6 | US4 Single Scenario | P2 | 5 | Debug mode |
| 7 | US5 Policy Enforcement | P2 | 6 | Security policies |
| 8 | US6 Supply Chain | P3 | 5 | SLSA compliance |
| 9 | US7 Discovery | P3 | 5 | `magex ci:list` |
| 10 | Additional Scenarios | - | 23 | 37 scenarios |
| 11 | Polish | - | 9 | Documentation, validation |

**Total Tasks**: 122
**MVP Scope**: Phases 1-4 (57 tasks) - Static validation + CI reproduction
**Full Scope**: All phases (122 tasks) - Complete Guardian framework

---

## Notes

- [P] tasks = different files, no dependencies within same phase
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Quality gate required after each phase: `go-pre-commit --all-files && magex format:fix && magex lint`
