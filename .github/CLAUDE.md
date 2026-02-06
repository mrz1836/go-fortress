# CLAUDE.md

## 🤖 Welcome, Claude

This repository showcases **GoFortress** - an enterprise-grade CI/CD fortress system that transforms Go development pipelines into impenetrable quality fortresses through multi-stage verification, pure Go toolchain integration, and zero-dependency architecture.

## 🏰 System Architecture

### Core Components
GoFortress orchestrates **16+ specialized workflows** that work in concert:

| Component | Workflows | Purpose |
|-----------|-----------|---------|
| **Setup & Config** | `fortress.yml`, `fortress-setup-config.yml` | Environment loading, matrix generation |
| **Quality Gates** | `fortress-code-quality.yml`, `fortress-pre-commit.yml` | Linting, formatting, static analysis |
| **Security Layer** | `fortress-security-scans.yml` | Nancy, Govulncheck, Gitleaks |
| **Testing Arsenal** | `fortress-test-*.yml` (6 files) | Unit, fuzz, race, benchmarks, Redis services |
| **Coverage System** | `fortress-coverage.yml` | Internal or Codecov reporting |
| **Release Pipeline** | `fortress-release.yml` | GoReleaser automation |
| **Reporting** | `fortress-completion-*.yml` (3 files) | Statistics, summaries, artifacts |

### Configuration System
- **Modular Env Files**: `.github/env/` - 225+ parameters in domain-specific files
- **Project Overrides**: `90-project.env` - Project-specific settings (e.g., coverage provider)
- **Dynamic Loading**: `load-env.sh` loads all env files in sorted order

## 🚀 Integrated Pure Go Tools

### Tool Symphony
| Tool | Version | Integration | Key Features |
|------|---------|-------------|--------------|
| **MAGE-X** | v1.13.x | Build automation | 190+ zero-config commands, auto-discovery |
| **go-coverage** | v1.1.x | Coverage tracking | GitHub Pages deploy, history, badges |
| **go-pre-commit** | v1.4.x | Git hooks | 17x faster than Python, parallel execution |

### Performance Metrics
Based on actual CI runs:
- **Full Pipeline**: ~2-3 minutes
- **Parallel Jobs**: 14+ concurrent
- **Setup Time**: 3 seconds
- **Test Execution**: 32 seconds (with coverage + race)
- **Security Scans**: 5-15 seconds each

## 📋 Essential Commands

### Development Workflow
```bash
# Install pre-commit hooks (17x faster than Python alternatives)
go-pre-commit install

# Update all dependencies
magex deps:update

# Run tests with coverage
magex test:coverage

# Create a release
magex version:bump push=true bump=patch
```

### CI Inspection
```bash
# View recent CI runs
gh run list --repo mrz1836/go-fortress --limit 5

# Check workflow timing
gh run view <run-id> --json jobs | jq '.jobs[] | {name, startedAt, completedAt}'

# Download artifacts
gh run download <run-id>
```

## 🔧 Configuration Guide

### Key Environment Variables
```bash
# Go Versions (supports matrix testing)
GO_PRIMARY_VERSION=1.24.x
GO_SECONDARY_VERSION=1.24.x

# Runner Selection (cost optimization)
PRIMARY_RUNNER=ubuntu-24.04  # 10x cheaper than macOS
SECONDARY_RUNNER=ubuntu-24.04

# Coverage Provider Switch
GO_COVERAGE_PROVIDER=internal  # or 'codecov' with token

# Feature Flags
ENABLE_BENCHMARKS=true
ENABLE_CODE_COVERAGE=true
ENABLE_FUZZ_TESTING=true
ENABLE_RACE_DETECTION=true
ENABLE_REDIS_SERVICE=false

# Redis Service Configuration
REDIS_SERVICE_MODE=never  # Options: auto, always, never
REDIS_VERSION=7-alpine
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TRUST_SERVICE_HEALTH=true
```

### Custom Configuration Example
`90-project.env` overrides for specific needs:
```bash
# Use Codecov instead of internal coverage (in .github/env/90-project.env)
GO_COVERAGE_PROVIDER=codecov
CODECOV_TOKEN_REQUIRED=true

# Custom exclusions
GO_COVERAGE_EXCLUDE_PATHS=test/,vendor/,examples/
```

## 🎯 Workflow Orchestration

### Execution Flow
1. **Load Environment** (3s) - Load modular env files from .github/env/
2. **Setup Config** (4s) - Generate test matrices
3. **MAGE-X Verification** (6s) - Ensure build system ready
4. **Cache Warming** (17s) - Parallel Go module caching
5. **Parallel Execution**:
   - Security Scans (5-15s)
   - Code Quality (31s)
   - Pre-commit Checks (26s)
   - Test Suite (32s) with Redis services
   - Benchmarks (24s) with Redis services
6. **Coverage Processing** (21s)
7. **Completion Report** (27s)

### Smart Features
- **Conditional Workflows**: Skip expensive steps via feature flags
- **Matrix Testing**: Multi-OS, multi-version automatically
- **Cache Optimization**: Intelligent dependency caching
- **Artifact Management**: Compressed test outputs with retention
- **Failure Detection**: Smart test failure analysis and reporting

## 🛠️ Debugging Tips

### Common Issues
1. **Coverage Provider Switching**
   - Check `90-project.env` for `GO_COVERAGE_PROVIDER`
   - Ensure `CODECOV_TOKEN` secret is set if using Codecov

2. **Test Timeouts**
   - Adjust `TEST_TIMEOUT` variables in `00-core.env`
   - Race+Coverage tests need more time (`TEST_TIMEOUT_RACE_COVER=30m`)

3. **Cache Misses**
   - Verify `go.sum` hasn't changed unexpectedly
   - Check runner OS matches cache key
   - Redis image cache may need warming

4. **Redis Service Issues**
   - Check `ENABLE_REDIS_SERVICE` and `REDIS_SERVICE_MODE` settings
   - Verify service container health checks are passing
   - Consider disabling `REDIS_TRUST_SERVICE_HEALTH` for debugging

5. **Pre-commit Failures**
   - Run `magex format:fix` locally first
   - Ensure `go-pre-commit` is installed

## 📚 Additional Resources

### Core Documentation
- **Standards**: `AGENTS.md` - The source of truth for all conventions
- **Tech Specs**: `.github/tech-conventions/` - Detailed technical guidelines
- **Workflows**: `.github/workflows/fortress-*.yml` - Implementation details

### Quick Checklist for Claude
1. ✅ **Study `AGENTS.md`** - All conventions originate here
2. ✅ **Respect Configuration Hierarchy** - Later env files override earlier ones
3. ✅ **Use MAGE-X Commands** - Never manually run go build/test
4. ✅ **Follow Branch Standards** - They trigger auto-labeling
5. ✅ **Never Tag Releases** - Use `magex version:bump` instead
6. ✅ **Never commit code** - unless you are told to do so

Happy fortress building! 🏰
