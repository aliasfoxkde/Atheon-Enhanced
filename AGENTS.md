# AGENTS.md — Guidance for AI Agents Working on Atheon-Enhanced

This document is for AI coding agents (Claude Code, Cursor, Windsurf, etc.) that interact with
the Atheon-Enhanced repository. It is the canonical reference for the project's conventions.

**If you are a human contributor, you can ignore this file.**

---

## Project at a Glance

- **Language**: Go 1.21+
- **Layout**:
  - `core/` — pattern engine, scanner, bundle loader, pattern state
  - `cmd/atheon/` — CLI (`atheon`)
  - `cmd/mcp/` — MCP server (`atheon-mcp`)
  - `bundler/` — compiles `community/*.yaml` → `core/patterns.bundle`
  - `community/<category>/*.yaml` — pattern definitions (406 across 39 categories)
  - `docs/` — user-facing documentation
  - `.github/workflows/` — 10 GitHub Actions workflows (consolidation planned)
- **Module path**: `github.com/aliasfoxkde/Atheon`
- **Repository**: https://github.com/aliasfoxkde/Atheon-Enhanced

---

## Non-Negotiable Rules

1. **`-p 1` on every `go test`** invocation. The `core` package has package-level state in
   `init()` that is not concurrency-safe under parallel package execution.

2. **`go:generate` is replaced by `go run ./bundler`**. Run this after adding patterns to
   `community/` to regenerate `core/patterns.bundle`. The bundle is `//go:embed`-ed into the
   binary at build time.

3. **No new non-stdlib dependencies.** `core/` must remain dependency-free. This is enforced
   by the project's policy and by `golangci-lint`.

4. **RE2 regex only.** No PCRE features (no lookahead, lookbehind, or backreferences). Test
   patterns with the bundler before submitting.

5. **Context-aware APIs.** Every public scanner entry point takes `context.Context` as the
   first parameter. Check `ctx.Err()` between iterations.

6. **Sentinel errors over strings.** Use `errors.Is(err, core.ErrPatternNotFound)` etc.
   Existing sentinels: `ErrPatternNotFound`, `ErrBundleDownload`, `ErrBundleParse`,
   `ErrInvalidPattern`.

7. **Structured logging via `log/slog`.** No `fmt.Fprintf(os.Stderr, ...)` in production code.

8. **No force-pushes.** Never use `git push --force`. Branch protection disallows it.

9. **Conventional commits.** `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`,
   `ci:`. Title in present tense imperative.

10. **One PR per logical change.** Don't bundle unrelated changes.

---

## Adding a Pattern

The simplest contribution. Three steps:

```bash
# 1. Create the YAML
cat > community/secrets/my-new-key.yaml <<'EOF'
name: my-new-key
match: '\bMY_[A-Z0-9]{32}\b'
EOF

# 2. Rebuild the bundle
go run ./bundler

# 3. Verify
./atheon list | grep my-new-key
./atheon --categories=secrets . --all | grep my-new-key
```

Submit a PR. CI will:

- Validate the regex compiles (via `go run ./bundler`)
- Run the test suite (must stay ≥90% project / ≥80% patch on Codecov)
- Self-scan (must not flag your YAML as a production secret)

---

## Adding a Category

A category is a directory under `community/`. The category name comes from the directory name
(no separate metadata file). To add a new category:

1. Create `community/<new-category>/`
2. Add at least one pattern YAML
3. Run `go run ./bundler`
4. Update README.md's pattern-distribution table
5. Update `docs/architecture/PATTERN_CATEGORIES.md`
6. Submit PR

---

## Adding a CLI Command

Edit `cmd/atheon/main.go`. Follow the existing pattern:

```go
switch args[0] {
case "your-command":
    // ...
    return 0
}
```

Don't forget:
- A test in `cmd/atheon/*_test.go`
- A help-text line in `printHelp()`
- A README section under "Quick Start"

---

## Adding an MCP Tool

Edit `cmd/mcp/main.go`. Three places:

1. `toolList()` — add the tool schema
2. `handleCall()` — add the switch case
3. Update `docs/api/README.md` MCP tool table

The MCP server uses JSON-RPC over stdio. Rate limiting is already wired in.

---

## Workflow Boundaries

| Change type | File(s) |
|-------------|---------|
| Engine behavior | `core/*.go` (+ tests) |
| CLI behavior | `cmd/atheon/*.go` (+ tests) |
| MCP behavior | `cmd/mcp/*.go` (+ tests) |
| New pattern | `community/<cat>/<name>.yaml` + run bundler |
| New category | `community/<cat>/` + README + PATTERN_CATEGORIES.md |
| Documentation | `docs/**` |
| CI/CD | `.github/workflows/*.yml` |
| Release | `.goreleaser.yml` |
| Hooks | `scripts/hooks/*` |

---

## Testing

```bash
# All tests
go test ./... -p 1 -timeout 15m -coverprofile=coverage.out

# Just one package
go test ./core -p 1 -v

# Race detection
go test ./... -p 1 -race

# Coverage gate
go tool cover -func=coverage.out | grep total:
# Must show ≥70%

# Lint
golangci-lint run --timeout=5m
```

---

## Common Pitfalls

1. **Forgetting `-p 1`** → flaky CI, mysterious bundle-state corruption. Use `make test`
   instead of `go test ./...`.

2. **Adding a PCRE feature to a pattern regex** → silently won't compile in `regexp.Compile`.
   Bundler will skip the pattern with a warning. Test your regex with `go run ./bundler`.

3. **Mutating package-level state from tests** → other tests fail mysteriously. The
   `allPatterns`, `activeScanners`, and `activeCategoryFilter` are package-globals. Use
   `core.EnableAllPatterns()` / `core.SetPatternEnabled(name, false)` to reset.

4. **Forgetting to rebuild the bundle** → new pattern works locally but binary doesn't include
   it. The `//go:embed patterns.bundle` captures the bundle at build time.

5. **Editing files in `.claude/`, `.github/wiki/`, or `docs/planning/`** — these are
   scaffolding files. Don't touch `.claude/` (it's a symlink to the user's global config).
   Don't touch `docs/planning/` (it's marked gitignored in some workflows and contains stale
   template copies slated for removal).

---

## When You're Stuck

1. Read the relevant doc: `docs/architecture/SYSTEM_ARCHITECTURE.md`,
   `docs/api/README.md`, `docs/development/SETUP.md`.
2. Read the recent git log: `git log --oneline -30`.
3. Read the IMPROVEMENT_PLAN.md at `docs/reports/IMPROVEMENT_PLAN.md` — it tracks partial
   execution of past improvements.
4. Read the canonical analysis: `docs/ANALYSIS_REPORT.md`.
5. Open an issue with `gh issue create` if the change is non-trivial.

---

## Author Attribution

The repository owner is `@aliasfoxkde`. PRs are reviewed by the CODEOWNERS file at
`.github/CODEOWNERS`. Don't add unrelated co-authors to commits.

---

*Last updated: 2026-06-23*
