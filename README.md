<div align="center">

# 🏰&nbsp;&nbsp;go-fortress

**Enterprise-grade CI/CD fortress for Go applications**

<br/>

<a href="https://github.com/mrz1836/go-fortress/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/go-fortress?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-fortress?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/go-fortress/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/go-fortress?style=flat-square&color=blue" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-fortress/actions"><img src="https://img.shields.io/github/actions/workflow/status/mrz1836/go-fortress/fortress.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/mrz1836/go-fortress/actions"><img src="https://img.shields.io/github/last-commit/mrz1836/go-fortress?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://goreportcard.com/report/github.com/mrz1836/go-fortress"><img src="https://goreportcard.com/badge/github.com/mrz1836/go-fortress?style=flat-square" alt="Go Report"></a>
       <a href="https://codecov.io/gh/mrz1836/go-fortress"><img src="https://codecov.io/gh/mrz1836/go-fortress/branch/master/graph/badge.svg?style=flat-square" alt="Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/mrz1836/go-fortress"><img src="https://api.scorecard.dev/projects/github.com/mrz1836/go-fortress/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-fortress/graphs/contributors"><img src="https://img.shields.io/github/contributors/mrz1836/go-fortress?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://mrz1818.com/"><img src="https://img.shields.io/badge/donate-bitcoin-ff9900?style=flat-square&logo=bitcoin" alt="Bitcoin"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="33%">
       🚀&nbsp;<a href="#-quick-start"><code>Quick&nbsp;Start</code></a>
    </td>
    <td align="center" width="33%">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
    <td align="center" width="33%">
       🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       👥&nbsp;<a href="#-maintainers"><code>Maintainers</code></a>
    </td>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
  </tr>
  <tr>
    <td align="center" colspan="3">
       ⚖️&nbsp;<a href="#-license"><code>License</code></a>
    </td>
  </tr>
</table>
<br/>

## 🏰 The GoFortress CI/CD System

> **Built Strong. Tested Harder.** Enterprise-grade CI/CD that transforms your Go development pipeline into an impenetrable fortress of quality through multi-stage verification, pure Go toolchain integration, and zero-dependency architecture.

<br/>

### ⚡ Performance Metrics

Speed isn't just a feature—it's our religion. Watch your entire CI/CD pipeline execute faster than your coffee break, with intelligent parallelization that would make a Swiss watchmaker jealous.

<table>
  <thead>
    <tr>
      <th>Metric</th>
      <th>Value</th>
      <th>Details</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>🚀 Full Pipeline</strong></td>
      <td><code>~2-3 min</code></td>
      <td>Complete CI/CD execution from push to report</td>
    </tr>
    <tr>
      <td><strong>⚙️ Parallel Jobs</strong></td>
      <td><code>14+ concurrent</code></td>
      <td>Security, quality, testing run simultaneously</td>
    </tr>
    <tr>
      <td><strong>🎯 Setup Time</strong></td>
      <td><code>3 seconds</code></td>
      <td>Environment loading and matrix generation</td>
    </tr>
    <tr>
      <td><strong>🧪 Test Execution</strong></td>
      <td><code>32 seconds</code></td>
      <td>Full test suite with coverage + race detection + Redis services</td>
    </tr>
    <tr>
      <td><strong>🔐 Security Scans</strong></td>
      <td><code>5-15 seconds</code></td>
      <td>Nancy, Govulncheck, Gitleaks combined</td>
    </tr>
    <tr>
      <td><strong>📊 Coverage Deploy</strong></td>
      <td><code>21 seconds</code></td>
      <td>Full report generation and GitHub Pages deploy</td>
    </tr>
  </tbody>
</table>

<br/>

### 🛠️ Pure Go Tool Integration

Say goodbye to Python dependencies and hello to pure Go bliss. These battle-tested tools speak your language natively, delivering enterprise-grade automation without the interpreter overhead or dependency hell.

<table>
  <thead>
    <tr>
      <th>Tool</th>
      <th>Purpose</th>
      <th>Integration</th>
      <th>Key Features</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong><a href="https://github.com/mrz1836/mage-x">MAGE-X</a></strong> <code>v1.x.x</code></td>
      <td>Build Automation</td>
      <td>Zero-config commands</td>
      <td>• 190+ built-in targets<br>• Auto-discovery<br>• Cross-platform<br>• Enterprise-ready</td>
    </tr>
    <tr>
      <td><strong><a href="https://github.com/mrz1836/go-coverage">go-coverage</a></strong> <code>v1.x.x</code></td>
      <td>Coverage Intelligence</td>
      <td>Self-hosted reports</td>
      <td>• GitHub Pages deploy<br>• History tracking (90 days)<br>• SVG badges<br>• PR comments</td>
    </tr>
    <tr>
      <td><strong><a href="https://github.com/mrz1836/go-pre-commit">go-pre-commit</a></strong> <code>v1.x.x</code></td>
      <td>Git Hooks</td>
      <td>17x faster validation</td>
      <td>• Pure Go (no Python)<br>• Parallel execution<br>• 8+ configurable checks<br>• Auto-fix support</td>
    </tr>
  </tbody>
</table>

<br/>

### 🏗️ Multi-Stage Defense System

Like a medieval fortress with multiple walls, each layer of our defense system catches what the previous might miss. From security vulnerabilities to race conditions, nothing gets past this gauntlet of quality gates.

<table>
  <thead>
    <tr>
      <th>Stage</th>
      <th>Components</th>
      <th>Capabilities</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>🛡️ Security Layer</strong></td>
      <td>Nancy, Govulncheck, Gitleaks <code>v8.x.x</code></td>
      <td>• Dependency vulnerability scanning<br>• Go-specific CVE detection<br>• Secret leak prevention<br>• Fork-safe execution</td>
    </tr>
    <tr>
      <td><strong>📊 Quality Gates</strong></td>
      <td>golangci-lint <code>v2.x.x</code>, go vet, yamlfmt</td>
      <td>• 66 linters enabled<br>• Static analysis<br>• YAML/JSON validation<br>• Triple-layer caching</td>
    </tr>
    <tr>
      <td><strong>🧪 Testing Arsenal</strong></td>
      <td>Unit, Fuzz, Race, Benchmarks, Redis</td>
      <td>• Multi-OS matrix (Ubuntu, macOS)<br>• Race condition detection<br>• Fuzz testing (5m timeout)<br>• Benchmark modes (quick/normal/full)<br>• Conditional Redis service</td>
    </tr>
    <tr>
      <td><strong>📈 Coverage Intelligence</strong></td>
      <td>go-coverage <code>v1.x.x</code> or Codecov</td>
      <td>• Switchable providers<br>• History tracking (90 days)<br>• Trend analysis<br>• Badge generation</td>
    </tr>
    <tr>
      <td><strong>🚀 Release Automation</strong></td>
      <td>GoReleaser, GoDocs</td>
      <td>• Semantic versioning<br>• Binary distribution<br>• Changelog generation<br>• pkg.go.dev syndication</td>
    </tr>
  </tbody>
</table>

<br/>

### 🛡️ Fortress Guardian - Local CI Validation

Test your CI pipelines before pushing. Guardian provides a Go-native framework for validating GitHub Actions workflows locally.

<table>
  <thead>
    <tr>
      <th>Feature</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>🔬 Static Validation</strong></td>
      <td>Workflow syntax and policy checks without Docker (< 2s)</td>
    </tr>
    <tr>
      <td><strong>🐳 Local Execution</strong></td>
      <td>Run GitHub Actions locally via <code>nektos/act</code></td>
    </tr>
    <tr>
      <td><strong>📋 35+ Scenarios</strong></td>
      <td>Pre-built test cases for quality, security, fork safety, and supply chain</td>
    </tr>
    <tr>
      <td><strong>🔒 Policy Engine</strong></td>
      <td>Enforce SHA-pinned actions, explicit permissions, and security best practices</td>
    </tr>
  </tbody>
</table>

**Quick Commands:**
```bash
magex ci:static    # Static validation (no Docker required)
magex ci:test      # Quick test with fast scenarios
magex ci:verify    # Full 35+ scenario verification
```

See **[guardian/README.md](guardian/README.md)** for complete documentation.

<br/>

### 🎯 Configuration Power

With 225+ knobs to turn and switches to flip, you're the architect of your own CI/CD destiny. Fine-tune every aspect from test matrices to runner selection—because one size never fits all in the real world.

<table>
  <thead>
    <tr>
      <th>Feature</th>
      <th>Configuration</th>
      <th>Flexibility</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>Environment System</strong></td>
      <td><a href=".github/env/README.md">Modular env files</a></td>
      <td>225+ parameters in domain-specific files</td>
    </tr>
    <tr>
      <td><strong>Test Matrices</strong></td>
      <td>Dynamic generation</td>
      <td>Multi-OS, multi-Go version with auto-deduplication</td>
    </tr>
    <tr>
      <td><strong>Feature Flags</strong></td>
      <td>17+ granular toggles</td>
      <td>Enable/disable any component individually</td>
    </tr>
    <tr>
      <td><strong>Provider Switching</strong></td>
      <td>Runtime selection</td>
      <td>Internal go-coverage or external Codecov</td>
    </tr>
    <tr>
      <td><strong>Fork Security</strong></td>
      <td>Automatic detection</td>
      <td>Safe job routing for fork PRs</td>
    </tr>
    <tr>
      <td><strong>Cost Optimization</strong></td>
      <td>Runner selection</td>
      <td>Linux (1x) / macOS (10x) cost awareness</td>
    </tr>
  </tbody>
</table>

<br/>

### 📐 Workflow Architecture

Witness the symphony of 18 specialized workflows performing in perfect harmony. Each component knows its role, executes with precision, and passes the baton seamlessly—like a well-oiled machine, but cooler.

```
                          🏰 GoFortress Pipeline
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│  [Load Env] ──► [Setup Config] ──► [MAGE-X Verify] ──► [Cache Warm]     │
│                                                                         │
│                          ┌── Parallel Execution ──┐                     │
│                          │                        │                     │
│    ┌─────────────┬───────┴────┬───────────┬───────┴─────┐               │
│    │             │            │           │             │               │
│    ▼             ▼            ▼           ▼             ▼               │
│ [Security]  [Pre-commit] [Quality]  [Test Suite]  [Benchmarks]          │
│ • Nancy     • Format     • Lint     • Matrix      • Quick/Full          │
│ • Govuln    • Imports    • Vet      • Fuzz        • Performance         │
│ • Gitleaks  • Checks     • YAML     • Race                              │
│                                     • Coverage                          │
│    └─────────────┴────────────┴───────────┴─────────────┘               │
│                                    │                                    │
│                                    ▼                                    │
│                             [Status Check]                              │
│                                    │                                    │
│                    ┌───────────────┴───────────────┐                    │
│                    │                               │                    │
│                    ▼                               ▼                    │
│              [Release]                   [Completion Report]            │
│             (tags only)                  • Statistics                   │
│                                          • Timing Analysis              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

<details>
<summary><strong><code>Detailed Workflow Architecture</code></strong></summary>
<br/>

**18 Specialized Workflows:**

```
fortress.yml (Main Orchestrator v1.6.0)
│
├── load-env ─────────────────────► Loads modular env files (.github/env/)
│
├── setup ────────────────────────► fortress-setup-config.yml
│   ├── Fork PR detection          • Generates test matrices
│   ├── Go version parsing         • 35+ configuration outputs
│   └── Matrix generation          • Feature flag evaluation
│
├── test-magex ───────────────────► fortress-test-magex.yml
│                                   • Verifies MAGE-X installation
│
├── warm-cache ───────────────────► fortress-warm-cache.yml
│                                   • Pre-warms Go module cache
│                                   • Optional Redis image cache
│
├── security ─────────────────────► fortress-security-scans.yml
│   ├── ask-nancy                   • Dependency vulnerabilities
│   ├── govulncheck                 • Go CVE detection
│   └── gitleaks                    • Secret scanning (fork-safe)
│
├── pre-commit ───────────────────► fortress-pre-commit.yml
│                                   • 8+ parallel checks like:
│                                   • gomt, golangci-lint, gitleaks
│                                   • mod-tidy, whitespace, eof, fumpt
│
├── code-quality ─────────────────► fortress-code-quality.yml
│   ├── govet                       • Static analysis
│   ├── lint                        • 60+ golangci-lint rules
│   └── yaml-format                 • YAML/JSON validation
│
├── test-suite ───────────────────► fortress-test-suite.yml
│   ├── execute-test-matrix ──────► fortress-test-matrix.yml
│   │                               • Multi-OS/Go version matrix
│   ├── execute-fuzz-tests ───────► fortress-test-fuzz.yml
│   │                               • Fuzz testing (5m timeout)
│   ├── validate-test-results ────► fortress-test-validation.yml
│   │                               • JSONL aggregation
│   └── process-coverage ─────────► fortress-coverage.yml
│       ├── check-provider          • Provider selection
│       ├── process-coverage        • go-coverage (internal)
│       └── process-codecov         • Codecov (external)
│
├── benchmarks ───────────────────► fortress-benchmarks.yml
│                                   • Modes: quick/normal/full
│                                   • Non-blocking (optional)
│
├── status-check ─────────────────► Final validation (always runs)
│
├── release ──────────────────────► fortress-release.yml
│                                   • Tag-triggered only (v*)
│                                   • GoReleaser integration
│
└── completion-report ────────────► fortress-completion-report.yml
    ├── fortress-completion-tests.yml
    ├── fortress-completion-statistics.yml
    └── fortress-completion-finalize.yml
```

</details>

<br/>

### 🚦 Quick Start

```bash
# Clone and setup
git clone https://github.com/mrz1836/go-fortress
cd go-fortress

# Install the pure Go toolchain
go install github.com/mrz1836/mage-x/cmd/magex@latest
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest

# Configure pre-commit hooks (17x faster than Python)
go-pre-commit install

# Create your own fortress
cp -r .github/workflows/fortress-* your-project/.github/workflows/
cp -r .github/env your-project/.github/
```

<br/>

## 📚 Documentation

> **Good to know:** As a CI workflow system, `go-fortress` leverages **14 external GitHub Actions**
> (official + security/coverage), integrates with **5+ external services** (Codecov, OSS Index,
> pkg.go.dev, Docker Hub, GitHub Pages), and downloads **pure Go tools** at runtime (MAGE-X ecosystem,
> security scanners). The Go code itself uses only `testify` for tests.

<br/>

<details>
<summary><strong>Repository Features</strong></summary>
<br/>

This repository includes 25+ built-in features covering CI/CD, security, code quality, developer experience, and community tooling.

**[View the full Repository Features list →](.github/docs/repository-features.md)**

</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

This project uses [goreleaser](https://github.com/goreleaser/goreleaser) for streamlined binary and library deployment to GitHub. To get started, install it via:

```bash
brew install goreleaser
```

The release process is defined in the [.goreleaser.yml](.goreleaser.yml) configuration file.


Then create and push a new Git tag using:

```bash
magex version:bump push=true bump=patch branch=master
```

This process ensures consistent, repeatable releases with properly versioned artifacts and citation metadata.

</details>

<details>
<summary><strong><code>Pre-commit Hooks</code></strong></summary>
<br/>

Set up the Go-Pre-commit System to run the same formatting, linting, and tests defined in [pre-commit.md](.github/tech-conventions/pre-commit.md) before every commit:

```bash
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
go-pre-commit install
```

The system is configured via the [modular env system](.github/env/README.md) and provides 17x faster execution than traditional Python-based pre-commit hooks. See the [complete documentation](http://github.com/mrz1836/go-pre-commit) for details.

</details>

<details>
<summary><strong>GitHub Workflows</strong></summary>
<br/>

All workflows are driven by modular configuration in [`.github/env/`](.github/env/README.md) — no YAML editing required.

**[View all workflows and the control center →](.github/docs/workflows.md)**

</details>

<details>
<summary><strong><code>Updating Dependencies</code></strong></summary>
<br/>

To update all dependencies (Go modules, linters, and related tools), run:

```bash
magex deps:update
```

This command ensures all dependencies are brought up to date in a single step, including Go modules and any tools managed by [MAGE-X](https://github.com/mrz1836/mage-x). It is the recommended way to keep your development environment and CI in sync with the latest versions.

</details>

<details>
<summary><strong><code>Build Commands</code></strong></summary>
<br/>

View all build commands

```bash script
magex help
```

</details>

<details>
<summary><strong><code>Fork PR Security</code></strong></summary>
<br/>

GoFortress automatically detects fork PRs and routes jobs safely:

**Fork-Safe Jobs** (always run):
- `setup`, `test-magex`, `warm-cache` — Infrastructure setup
- `code-quality`, `pre-commit` — Linting and formatting
- `benchmarks` — Performance testing

**Fork-Unsafe Jobs** (skipped on forks):
- `security` — Requires `OSSI_TOKEN`, `GITLEAKS_LICENSE`
- `test-suite` coverage upload — Requires `CODECOV_TOKEN`
- `release` — Tag-triggered only

Fork contributors see a clear message in the workflow summary explaining which jobs ran and why some were skipped. Maintainers can manually trigger security scans after reviewing fork PRs.

</details>

<details>
<summary><strong><code>Dual Coverage Providers</code></strong></summary>
<br/>

Switch between coverage providers via `90-project.env`:

```bash
# Option 1: Internal go-coverage (default)
GO_COVERAGE_PROVIDER=internal

# Option 2: External Codecov (set in .github/env/90-project.env)
GO_COVERAGE_PROVIDER=codecov
CODECOV_TOKEN_REQUIRED=true
```

**Internal Provider (go-coverage):**
- Self-hosted GitHub Pages deployment
- 90-day history tracking
- SVG badge generation
- PR comments with coverage diff
- HTML reports with github-dark theme

**Codecov Provider:**
- External service integration
- Requires `CODECOV_TOKEN` secret
- Automatic PR annotations

</details>

<details>
<summary><strong><code>Coverage Threshold Enforcement</code></strong></summary>
<br/>

GoFortress enforces minimum coverage requirements in CI. Builds fail if coverage drops below the configured threshold.

```bash
# In .github/env/10-coverage.env or 90-project.env
GO_COVERAGE_THRESHOLD=65.0  # Minimum coverage percentage (0-100)
```

**How it works:**
1. Tests generate coverage data with configured exclusions applied
2. After coverage processing, the threshold is checked
3. If coverage < threshold, the build fails with actionable guidance

**Coverage Exclusions** (applied during test generation):
```bash
GO_COVERAGE_EXCLUDE_PATHS=test/,vendor/,testdata/  # Excluded directories
GO_COVERAGE_EXCLUDE_FILES=*_test.go,*.pb.go        # Excluded file patterns
GO_COVERAGE_EXCLUDE_TESTS=true                      # Exclude test files
GO_COVERAGE_EXCLUDE_GENERATED=true                  # Exclude generated code
```

**Disable threshold enforcement:**
```bash
GO_COVERAGE_THRESHOLD=0  # Set to 0 to disable
```

> **Note:** Exclusions are applied when generating coverage, not when checking the threshold. The threshold check uses the already-filtered coverage data.

</details>

<details>
<summary><strong><code>Benchmark Modes</code></strong></summary>
<br/>

Three benchmark execution modes via `BENCHMARK_MODE`:

| Mode | Duration | Use Case |
|------|----------|----------|
| `quick` | 50ms | Fast CI feedback |
| `normal` | 100ms | Default CI runs |
| `full` | 5s | Comprehensive analysis |

```bash
# In .github/env/00-core.env or 90-project.env
BENCHMARK_MODE=quick  # Fast CI
BENCHMARK_MODE=full   # Detailed benchmarks
```

Benchmarks are **non-blocking** — failures won't fail the pipeline. Results include ns/op, B/op, and allocs/op metrics with 7-day artifact retention.

</details>

<details>
<summary><strong><code>Redis Service Configuration</code></strong></summary>
<br/>

Optional Redis service containers for tests and benchmarks:

```bash
# In .github/env/20-redis.env or 90-project.env
REDIS_SERVICE_MODE=auto    # Auto-detect from go.mod
REDIS_SERVICE_MODE=always  # Always start Redis
REDIS_SERVICE_MODE=never   # Never start Redis (default)

REDIS_VERSION=7-alpine     # Docker image version
REDIS_PORT=6379            # Connection port
```

**Auto-detection** searches `go.mod` for:
- `github.com/redis/go-redis`
- `github.com/go-redis/redis`
- `github.com/garyburd/redigo`
- `github.com/gomodule/redigo`

**Health checks** are configurable:
```bash
REDIS_HEALTH_CHECK_RETRIES=10
REDIS_HEALTH_CHECK_INTERVAL=10
REDIS_HEALTH_CHECK_TIMEOUT=5
```

</details>

<details>
<summary><strong><code>Multi-Module Support</code></strong></summary>
<br/>

Enable `go.work` workspace testing for monorepos:

```bash
# In .github/env/00-core.env or 90-project.env
ENABLE_MULTI_MODULE_TESTING=true
GO_SUM_FILE=go.sum  # Or path to specific module
```

When enabled:
- Tests run from repository root
- All modules in `go.work` are included
- Coverage aggregated across modules
- Matrix testing applies to all modules

</details>

<details>
<summary><strong><code>Completion Report Analytics</code></strong></summary>
<br/>

Automatic workflow analytics when `ENABLE_COMPLETION_REPORT=true`:

**Statistics Collected:**
- Cache hit/miss rates per job
- Benchmark performance metrics
- Coverage percentages and trends
- Lines of code analysis

**Timing Analysis:**
- Per-job execution duration
- Bottleneck identification
- Parallel vs sequential breakdown

**Report Generation:**
- `fortress-completion-tests.yml` — Test analysis
- `fortress-completion-statistics.yml` — Metrics aggregation
- `fortress-completion-finalize.yml` — Final assembly

Reports appear in the GitHub Actions workflow summary with expandable sections.

</details>

<br/>

## 🛠️ Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines
Read the [AI Usage & Assistant Guidelines](.github/tech-conventions/ai-compliance.md) for details on how AI is used in this project and how to interact with the AI assistants.

<br/>

## 👥 Maintainers
| [<img src="https://github.com/mrz1836.png" height="50" width="50" alt="MrZ" />](https://github.com/mrz1836) |
|:-----------------------------------------------------------------------------------------------------------:|
|                                      [MrZ](https://github.com/mrz1836)                                      |

<br/>

## 🤝 Contributing
View the [contributing guidelines](.github/CONTRIBUTING.md) and please follow the [code of conduct](.github/CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap:
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&utm_source=github&utm_medium=sponsor-link&utm_campaign=go-fortress&utm_term=go-fortress&utm_content=go-fortress) to ensure this journey continues indefinitely! :rocket:

[![Stars](https://img.shields.io/github/stars/mrz1836/go-fortress?label=Please%20like%20us&style=social&v=1)](https://github.com/mrz1836/go-fortress/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/go-fortress.svg?style=flat&v=1)](LICENSE)
