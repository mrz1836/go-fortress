# Modular Environment Configuration

This directory contains the modular environment configuration system for go-fortress CI workflows.

## Overview

Instead of a single monolithic `.env.base` file, configuration is split into domain-specific files that are loaded in numeric order. Later files override earlier ones (last wins).

## File Structure

```
.github/env/
├── load-env.sh           # Shell loader script
├── README.md             # This file
│
├── 00-core.env           # Go versions, runners, feature flags, timeouts
├── 10-mage-x.env         # MAGE-X build system configuration
├── 10-coverage.env       # go-coverage settings
├── 10-pre-commit.env     # go-pre-commit settings
├── 10-security.env       # Gitleaks, Nancy, Govulncheck
├── 10-broadcast.env      # go-broadcast sync settings
├── 20-redis.env          # Redis service configuration
├── 20-workflows.env      # Stale, labels, dependabot, PR management
├── 20-guardian.env       # CI Guardian framework
├── 90-project.env        # Project-specific overrides (not synced)
└── 99-local.env          # Local development (gitignored)
```

## Naming Convention

Files are named with numeric prefixes to control load order:

| Prefix | Purpose | Examples |
|--------|---------|----------|
| `00-` | Core/foundation | Go versions, runners, feature flags |
| `10-` | Tool configuration | mage-x, coverage, pre-commit, security |
| `20-` | Services & workflows | Redis, workflow automation |
| `90-` | Project overrides | Project-specific settings |
| `99-` | Local development | Machine-specific (gitignored) |

## Override Behavior

Files are loaded in sorted order. Variables defined in later files override earlier ones:

```bash
# 00-core.env
GO_PRIMARY_VERSION=1.24.x

# 90-project.env (loaded later, wins)
GO_PRIMARY_VERSION=1.23.x  # This value is used
```

## Usage

### In GitHub Actions

The loader is called automatically by the `load-env` composite action:

```yaml
- uses: ./.github/actions/load-env
  id: load-env
```

### Local Development

Source the loader script directly:

```bash
source .github/env/load-env.sh

# With verbose output
source .github/env/load-env.sh --verbose

# Or via environment variable
ENV_LOADER_VERBOSE=1 source .github/env/load-env.sh
```

### Verifying Configuration

```bash
# Check a specific variable
source .github/env/load-env.sh && echo $GO_PRIMARY_VERSION

# List all exported variables (from env files)
source .github/env/load-env.sh && env | grep -E '^[A-Z_]+'
```

## Adding New Variables

1. **Identify the domain** — Which file does this variable belong in?
2. **Add with documentation** — Include a comment explaining the variable
3. **Test locally** — Source the loader and verify
4. **Commit** — Changes sync to other repos via go-broadcast (if configured)

Example:

```bash
# In 10-mage-x.env
# Maximum parallel workers for mage builds
MAGE_X_MAX_WORKERS=4
```

## Project Overrides (90-project.env)

Use `90-project.env` for settings specific to this repository that shouldn't be synced to other repos:

```bash
# 90-project.env - Project-specific overrides
# These settings are NOT synced via go-broadcast

# Override coverage threshold for this repo
GO_COVERAGE_THRESHOLD=80.0

# Enable Redis for this repo's tests
ENABLE_REDIS_SERVICE=true
```

## Local Development (99-local.env)

Create `99-local.env` for machine-specific settings (it's gitignored):

```bash
# 99-local.env - Local development overrides
# This file is gitignored and not committed

# Use local go-coverage binary during development
GO_COVERAGE_USE_LOCAL=true
GO_COVERAGE_LOCAL_PATH=/Users/me/projects/go-coverage/go-coverage

# Verbose output for debugging
MAGE_X_VERBOSE=true
```

## CI Behavior

- In CI (`CI=true`), the loader skips `99-local.env`
- Variables are exported so downstream workflow steps can access them
- The composite action converts exported vars to JSON for workflow compatibility

## Migration from .env.base

The old `.env.base` file is deprecated. All variables have been split into domain files. During the transition period, both systems work:

1. New loader (`load-env.sh`) is tried first
2. Falls back to `.env.base` if loader fails
3. Once validated, `.env.base` will be removed

## Troubleshooting

### Variables not available in workflow steps

Ensure `set -a` is enabled in the loader (exports all sourced variables).

### Wrong value being used

Check load order — later files override earlier ones. Use `--verbose` to see which files are loaded.

### Local overrides not working

Make sure `99-local.env` exists and you're not running in CI mode.
