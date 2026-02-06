<!--
  ============================================================================
  SYNC IMPACT REPORT
  ============================================================================
  Version change: 0.0.0 → 1.0.0 (MAJOR - initial constitution)

  Modified principles: N/A (initial creation)

  Added sections:
  - Core Principles (7 principles)
  - Pure Go Ecosystem section
  - CI/CD Excellence section
  - Governance section

  Removed sections: N/A

  Templates requiring updates:
  - .specify/templates/plan-template.md: ✅ updated - added Constitution Check gates
  - .specify/templates/spec-template.md: ✅ compatible (no changes needed)
  - .specify/templates/tasks-template.md: ✅ compatible (no changes needed)

  Follow-up TODOs: None
  ============================================================================
-->

# GoFortress Constitution

## Core Principles

### I. Pure Go Philosophy

Every tool in the CI/CD pipeline MUST be written in pure Go without CGO dependencies.
This ensures single-binary distribution, cross-platform compatibility, and zero external
runtime requirements.

**Non-Negotiable Rules:**
- Build automation MUST use MAGE-X (pure Go, 190+ commands, auto-discovery)
- Pre-commit hooks MUST use go-pre-commit (17x faster than Python alternatives)
- Coverage reporting MUST use go-coverage (self-hosted GitHub Pages, no external SaaS)
- Security scanning tools (govulncheck, nancy, gitleaks) MUST be Go-native
- No Python, Ruby, or Node.js dependencies in the build pipeline

**Rationale:** Pure Go tooling eliminates dependency conflicts, provides consistent
behavior across all platforms, and allows developers to debug and extend tools using
the same language as the project itself.

### II. Multi-Stage Defense System

Code quality MUST be enforced through layered verification stages that execute in
parallel where possible. Each stage catches what previous stages might miss.

**Defense Layers (in execution order):**
1. **Security Layer**: Nancy (dependency CVEs), Govulncheck (Go-specific vulnerabilities),
   Gitleaks (secret scanning)
2. **Quality Gates**: golangci-lint (66+ linters), go vet (static analysis),
   yamlfmt (YAML validation)
3. **Testing Arsenal**: Unit tests, fuzz tests, race detection, benchmarks,
   multi-OS matrix testing
4. **Coverage Intelligence**: Threshold enforcement, trend tracking, PR comments

**Rationale:** Multiple verification layers ensure comprehensive quality assurance.
No single tool catches everything; layered defense provides defense-in-depth.

### III. Configuration-Driven Architecture

All CI/CD behavior MUST be controllable via environment configuration files without
modifying workflow YAML. This enables zero-touch customization across projects.

**Configuration Hierarchy:**
- `.github/env/` contains modular env files loaded in sorted order
- `00-core.env` through `20-*.env` contain default parameters
- `90-project.env` contains project-specific overrides (takes precedence)
- Feature flags enable/disable any component individually
- Runtime provider switching (internal vs external coverage, Redis modes)

**Non-Negotiable Rules:**
- Workflows MUST NOT contain hardcoded project-specific values
- All toggleable features MUST have corresponding `ENABLE_*` flags
- Tool versions MUST be pinnable via configuration
- Matrix testing MUST be dynamically generated from configuration

**Rationale:** Configuration-driven design allows one workflow system to serve
unlimited projects with different requirements, without fork divergence.

### IV. Fork-Safe Security Model

CI/CD pipelines MUST intelligently handle fork PRs by detecting fork status and
conditionally skipping jobs that require repository secrets.

**Fork-Safe Jobs (always run):**
- setup, test-magex, warm-cache — Infrastructure setup
- code-quality, pre-commit — Linting and formatting
- benchmarks — Performance testing

**Fork-Unsafe Jobs (skipped on forks):**
- security — Requires OSSI_TOKEN, GITLEAKS_LICENSE
- test-suite coverage upload — Requires CODECOV_TOKEN
- release — Tag-triggered only, never for forks

**Rationale:** Fork contributors can receive quality feedback without exposing
secrets. Maintainers can manually trigger security scans after review.

### V. Go Development Standards

All Go code MUST follow idiomatic patterns and the project's established conventions
as defined in AGENTS.md and tech-conventions.

**Non-Negotiable Rules:**
- Context-first design: `context.Context` MUST be the first parameter
- No global state: Use dependency injection, not package-level variables
- No `init()` functions: Use explicit constructors (`NewXxx()`)
- Error handling: Always check errors, wrap with context using `%w`
- Interface design: Accept interfaces, return concrete types; keep interfaces small

**Testing Standards:**
- Unit tests MUST accompany all new functionality
- Race detection MUST be enabled in CI (`-race` flag)
- Coverage threshold MUST be maintained (configurable, default 65%)
- Fuzz tests SHOULD be added for parsing and input validation

**Rationale:** Consistent coding standards enable maintainability, testability,
and make code accessible to both human developers and AI assistants.

### VI. Performance-First Execution

CI/CD pipelines MUST maximize parallel execution and minimize total wall-clock time.
Performance is not optional—it directly impacts developer productivity.

**Performance Targets:**
- Full pipeline completion: ~2-3 minutes
- Parallel job execution: 14+ concurrent jobs
- Setup time: <5 seconds
- Test suite with coverage + race: <60 seconds
- Security scans combined: <20 seconds

**Optimization Requirements:**
- Cache warming MUST be enabled by default
- Module caches MUST be shared across matrix jobs
- Artifact compression MUST be used for test outputs
- Conditional job execution MUST skip unnecessary work

**Rationale:** Fast feedback loops encourage more frequent commits and faster
iteration. Slow CI is a tax on every developer's productivity.

### VII. Release Automation Excellence

Releases MUST be fully automated, triggered by semantic version tags, and include
comprehensive artifact generation.

**Release Pipeline:**
- Tag pattern: `v*` (e.g., v1.0.0, v2.1.3)
- Binary distribution via GoReleaser
- Automatic changelog generation
- pkg.go.dev syndication on release
- Optional notification broadcasts (Slack, Discord, Twitter)

**Version Bumping:**
- MUST use `magex version:bump` (never manual tagging)
- Semantic versioning (MAJOR.MINOR.PATCH) strictly enforced
- Pre-release tags supported (v1.0.0-beta.1)

**Rationale:** Automated releases eliminate human error, ensure consistency,
and enable rapid iteration with confidence.

## Pure Go Ecosystem

### Tool Symphony

GoFortress integrates three core pure Go tools that form a cohesive ecosystem:

| Tool | Purpose | Key Advantage |
|------|---------|---------------|
| **MAGE-X** | Build automation | 190+ zero-config commands, Go-native syntax |
| **go-coverage** | Coverage reporting | Self-hosted GitHub Pages, no external SaaS |
| **go-pre-commit** | Git hooks | 17x faster than Python, parallel execution |

### Why Pure Go Matters

1. **Single Binary Distribution**: Download and run—no interpreters, no virtual
   environments, no dependency resolution
2. **Cross-Platform Consistency**: Same behavior on Linux, macOS, Windows without
   platform-specific workarounds
3. **Developer Familiarity**: Build logic written in Go, debuggable with standard
   Go tools, extendable by Go developers
4. **Security Posture**: Fewer dependencies mean smaller attack surface and faster
   vulnerability scanning

## CI/CD Excellence

### Workflow Architecture

The fortress comprises 18 specialized workflows orchestrated by `fortress.yml`:

```
fortress.yml (Main Orchestrator)
├── load-env ─────────► Load modular env files from .github/env/
├── setup ────────────► Generate matrices, detect forks
├── test-magex ───────► Verify MAGE-X installation
├── warm-cache ───────► Pre-warm Go module caches
├── security ─────────► Nancy + Govulncheck + Gitleaks (parallel)
├── pre-commit ───────► 8+ parallel checks via go-pre-commit
├── code-quality ─────► golangci-lint + go vet + yamlfmt
├── test-suite ───────► Matrix tests + fuzz + race + coverage
├── benchmarks ───────► Quick/normal/full modes
├── status-check ─────► Final validation gate
├── release ──────────► GoReleaser (tags only)
└── completion-report ► Statistics and timing analysis
```

### Completion Reporting

When enabled, the completion report provides:
- Cache hit/miss rates per job
- Benchmark performance metrics
- Coverage percentages and trends
- Per-job execution duration
- Parallel vs sequential timing breakdown

## Governance

### Amendment Process

1. Proposed changes MUST be documented in a PR with rationale
2. Constitution Check MUST be updated in plan templates if principles change
3. All dependent templates MUST be reviewed for compatibility
4. Version MUST be incremented according to semantic versioning:
   - MAJOR: Principle removal or backward-incompatible redefinition
   - MINOR: New principle added or material guidance expansion
   - PATCH: Clarifications, wording refinements, typo fixes

### Compliance Review

- All PRs MUST verify alignment with constitution principles
- AI assistants MUST reference this constitution for architectural decisions
- Complexity beyond constitution limits MUST be explicitly justified
- Violations require documented rationale and approval

### Runtime Guidance

For day-to-day development guidance, refer to:
- `AGENTS.md` — Comprehensive standards for all contributors
- `.github/tech-conventions/` — Detailed technical guidelines
- `.github/CLAUDE.md` — AI assistant quick reference

**Version**: 1.0.0 | **Ratified**: 2026-01-24 | **Last Amended**: 2026-01-24
