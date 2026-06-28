# Development Guide

## Branch Strategy

| Branch | Purpose | Quality Gates |
|--------|---------|---------------|
| `main` | Production-ready code — all PRs target here | Strict: 70% coverage, no lint errors |
| `dev/testing` | Integration staging — merge features here before main | Relaxed: 50% coverage, lint is informational |
| `fix/*` | Bug fixes — branch from main, PR to main | Strict |
| `feat/*` | New features — branch from main, PR to main | Strict |

### `dev/testing` Branch

The `dev/testing` branch is a long-lived integration branch for experimental work, combining multiple feature branches, and testing changes that aren't yet ready for main.

**CI on dev/testing differs from main:**
- Coverage threshold is 50% (vs 70% on main)
- All lint checks are informational (`continue-on-error: true`)
- Self-scan runs across all categories but never blocks the build
- Build matrix (linux/mac/windows) still runs so cross-platform regressions are caught early
- Race detector is disabled for faster iteration

**When to use dev/testing:**
- Combining multiple in-progress feature branches for integration testing
- Experimenting with patterns that aren't yet stable enough for main
- Testing configuration changes across environments before locking them in
- Staging work that needs review across multiple contributors before a main PR

**Merging to main:**
All code must pass main's full CI suite before a PR to main. The dev/testing branch is a scratchpad — never merge dev/testing → main directly; instead open a PR so the full gate runs.

## Local Development

### Prerequisites

- Go 1.22+
- `jq` (for self-scan scripts)

### Running Tests

```bash
# Full test suite
go test ./... -p 1

# With coverage
go test ./... -p 1 -coverprofile=coverage.out
go tool cover -html=coverage.out   # opens browser

# Race detector (catches concurrent map writes, etc.)
go test ./... -p 1 -race
```

### Building

```bash
go build -o atheon ./cmd/atheon
go build -o atheon-mcp ./cmd/mcp

./atheon --help
./atheon list
./atheon list categories
```

### Running Self-Scan Locally

```bash
# Same filter as CI — excludes test files, testdata, community/, wiki, issue templates
bash .github/scripts/self-scan.sh --secrets
bash .github/scripts/self-scan.sh --quality
```

### Configuration Profiles

Profiles live in `config/profiles/`. Pass with `--config`:

```bash
./atheon --config config/profiles/development.json .
./atheon --config config/profiles/production.json .
```

| Profile | Use case |
|---------|---------|
| `development.json` | All patterns, verbose output, debug mode, experimental features |
| `production.json` | Secrets + PII only, strict mode, optimized performance |
| `pipeline.json` | CI/CD — JSON output, exit-on-findings |
| `mcp-integration.json` | MCP server mode |

The `development.json` profile enables all categories including experimental patterns and produces verbose output suitable for debugging false positives.

## Pattern Development

New patterns go in `community/<category>/<name>.yaml`. The schema:

```yaml
id: category/pattern-name
name: Human-Readable Name
description: What this detects and why it matters.
category: secrets        # secrets | pii | code-quality | devops | ai-detection
severity: high           # critical | high | medium | low | info
match: 'regex here'
test:
  true_positives:
    - 'example that SHOULD match'
  false_positives:
    - 'example that should NOT match'
```

**False positive risk checklist before submitting a pattern:**
1. Does the regex require context (surrounding keywords) or just match bare tokens?
2. Does it match common Go/Python/JS identifiers of the same length?
3. Run `bash .github/scripts/self-scan.sh --quality` — if it fires on Atheon's own source, it's too broad.

## Release Process

Releases are tagged from main as `v<major>.<minor>.<patch>-enhanced`.

1. Ensure all target PRs are merged to main and CI is green
2. Check `git log --oneline v<last-tag>..HEAD` to review what's in scope
3. Create the release: `gh release create v<version> --repo aliasfoxkde/Atheon-Enhanced --generate-notes`
4. Edit the generated notes to highlight key features

Current release cadence: 10th and 21st of each month (automated via `.github/workflows/scheduled-release.yml`).
