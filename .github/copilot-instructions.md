# Copilot Instructions for Atheon-Enhanced

## What this project does

Atheon is a Go pattern-matching engine for detecting secrets, PII, and code quality issues. It scans files, directories, environment variables, and strings against a library of YAML-defined regex patterns. It outputs plain text, JSON, or SARIF format.

Key binaries:
- `atheon` — CLI scanner (`cmd/atheon/`)
- `atheon-mcp` — MCP server for AI assistant integration (`cmd/mcp/`)

## Repository layout

```text
core/           — pattern loading, scanning engine, matcher
cmd/atheon/     — CLI entrypoint, flag parsing, output formatting
cmd/mcp/        — MCP server (stdio JSON-RPC)
community/      — YAML pattern library, organized by category:
  secrets/      — API keys, tokens, credentials
  pii/          — Personal identifiable information
  code-quality/ — Anti-patterns, insecure coding practices
  devops/       — Infrastructure misconfigs
  ai-detection/ — AI prompt injection patterns
bundler/        — Compiles community/ YAML into core/patterns.bundle
docs/           — Documentation (published via GitHub Pages)
```

## Pattern YAML format

Each pattern file follows this schema:

```yaml
name: my-pattern-name
description: What this pattern detects
severity: low|medium|high|critical
category: secrets|pii|code-quality|devops|ai-detection
enabled: true
pattern: "regex here"
keywords: ["optional", "context", "keywords"]
true_positives:
  - "example that should match"
false_positives:
  - "example that should NOT match"
```

## Common issue types

**False positive reports:** The user is reporting that Atheon flagged something it shouldn't have. Look up the pattern name in `community/<category>/<name>.yaml`. Explain why the regex matched and whether narrowing with `keywords` would help.

**False negative reports:** Atheon missed something it should have caught. Check if an existing pattern covers the case or if a new pattern is needed.

**Build failures:** Run `go build ./...` from the repo root. The bundler (`go run ./bundler`) must be run before building to regenerate `core/patterns.bundle` if community patterns changed.

**Test failures:** Run `go test ./... -timeout 15m -p 1`. Tests use real pattern files — a failing test usually means a pattern regex changed without updating test fixtures. The `-p 1` flag is MANDATORY: core has package-level state in init() that is not safe under parallel package execution.

## CI workflows

- `ci.yml` — main build, test, coverage gate (70%), lint, cross-platform builds, integration tests, benchmarks, docs check, test-results reporting. The consolidated workflow that replaced the previous ci/quality-assurance/scheduled-release/publish/comprehensive-ci split.
- `release.yml` — tag-driven GoReleaser pipeline + scheduled (10th/21st) auto-versioned release with `release` environment approval gate
- `security.yml` — govulncheck, CodeQL, Trivy, secret-scan
- `sync.yml` — keep our `main` rebased onto `upstream/HoraDomu:main` so we track upstream security patches
- `auto-merge.yml` — auto-merge Dependabot PRs once checks pass
- `community-pattern-review.yml` — schema validation + AI (GPT-4o-mini) review of community pattern PRs (aliasfoxkde/Atheon-Enhanced only — see fork-token notes below)
- `dev-testing.yml` — relaxed gates for `dev/testing` branch (50% coverage, non-blocking)

## Contribution guidelines

- Pattern PRs go to `community/` — must include `true_positives` and `false_positives`
- Code PRs must pass `go vet ./...` and `go test ./... -timeout 15m -p 1`
- Coverage must not drop below 70% on `main`
- See `docs/patterns/contributing-patterns.md` for the full pattern contribution guide
