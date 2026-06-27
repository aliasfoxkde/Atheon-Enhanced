# SDLC Audit — 2026-06-27

## Overview

This document captures the systematic audit of the Atheon-Enhanced SDLC covering:
CI/CD pipelines, git hooks, commit conventions, PR/merge workflow, and branching strategy.

**Auditors**: Claude (automated agents + manual verification)
**Date**: 2026-06-27
**Status**: Gaps identified, roadmap documented

---

## 1. CI/CD Pipelines

### Workflows Present

| Workflow | File | Purpose |
|---|---|---|
| **CI** | `ci.yml` | Go tests (5 versions), lint, build (3 OSes), integration, benchmarks, docs check, coverage |
| **Security** | `security.yml` | CodeQL, self-scan (secrets blocking), anti-patterns, govulncheck |
| **Release** | `release.yml` | Tag-based GoReleaser publishing, scheduled (10th/21st) |
| **Community Pattern Review** | `community-pattern-review.yml` | AI review of YAML pattern changes via GitHub Models API |
| **Dev Testing** | `dev-testing.yml` | Relaxed gates for `dev/testing` branch |
| **Sync** | `sync.yml` | **NON-FUNCTIONAL** — blocked by branch protection on `stable/clean` |
| **Wiki** | `wiki.yml` | Publishes `.github/wiki/*.md` to GitHub Wiki |
| **Auto Merge** | `auto-merge.yml` | Enables squash-merge + conflict reporting on PRs |

### Strengths

- Multi-version Go testing (1.21–1.25)
- Cross-platform builds (Ubuntu/macOS/Windows)
- `-race` flag enabled in CI
- Coverage gate at 70%
- Self-scan for secrets/PII blocking in CI
- AI-assisted pattern review for community contributions
- JUnit XML + Codecov integration
- Benchmark tracking with artifact upload

### Gaps (Priority Order)

#### HIGH Priority

| # | Gap | Risk | Fix |
|---|---|---|---|
| 1 | **Release workflow skips lint/coverage/integration** | A release can tag with failing lint or coverage | Add lint + integration steps to release validation, or gate on `ci.yml` status |
| 2 | **Auto-merge enables without verifying required checks** | Auto-merge fires even when CI is still running/failing | Add status-check gate before enabling auto-merge |
| 3 | **No commit message linting** | Conventional commits documented but not enforced | Add `commitlint` to pre-commit or CI |

#### MEDIUM Priority

| # | Gap | Risk | Fix |
|---|---|---|---|
| 4 | **`sync.yml` is non-functional** | Confusing for contributors | Remove the workflow or document resolution path |
| 5 | **No benchmark regression tracking** | No alerting on performance regressions | Consider benchmark-action or trend storage |
| 6 | **No Dependabot for Go module updates** | Outdated dependencies not auto-PR'd | Add dependabot config for Go modules |
| 7 | **No stale PR/issue cleanup** | Orphaned PRs accumulate | Add stale PR workflow |
| 8 | **`enforce_admins: false`** | Admin users can bypass branch protection | Enable `enforce_admins: true` |
| 9 | **`stable/clean` has no branch protection** | Only policy-enforced, not technically enforced | Add protection rules or document why not needed |

#### LOW Priority

| # | Gap | Risk | Fix |
|---|---|---|---|
| 10 | No auto-label PRs by path | Reviewer routing manual | Add `actions/labeler` |
| 11 | Dev branch tests don't upload artifacts | Debugging harder | Add `actions/upload-artifact` to dev-testing.yml |
| 12 | `go.sum` vulnerability scanning only via govulncheck | Deps not auto-updated | Consider adding `dependabot` for Go modules |
| 13 | Release uses hardcoded Go 1.24 instead of `vars.GO_VERSION` | Version drift | Use `${{ vars.GO_VERSION }}` in release.yml |

---

## 2. Git Hooks

### Hooks Present

| Hook | Location | Purpose |
|---|---|---|
| `pre-commit` | `scripts/hooks/pre-commit` | gofmt, goimports, go vet, selective tests, coverage gate, bundle rebuild |
| `pre-push` | `scripts/hooks/pre-push` | Full build + full test suite |

### Strengths

- Smart — only runs relevant checks based on staged files
- Bundle auto-regeneration when community YAML files change
- Selective test execution (full suite only when `core/` touched)
- Coverage gate enforced locally before push

### Gaps

| # | Gap | Severity | Fix |
|---|---|---|---|
| 1 | **No commit message linting** | HIGH | Add `commitlint` to pre-commit hook |
| 2 | **No commit template** | MEDIUM | Add `commit.template` git config |
| 3 | **`Makefile setup` points to wrong path** | LOW | Fix `setup` target: `.githooks` → `scripts/hooks` |
| 4 | **Hooks not wired by default** | LOW | Add `git config core.hooksPath scripts/hooks` to setup or README |
| 5 | **No secrets pre-commit scan** | LOW | Wire `scripts/self-scan.sh` into pre-commit |

---

## 3. Commit Conventions

### Documented In

- `AGENTS.md` — Rule 9: conventional commits with prefixes (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`)
- `CONTRIBUTING.md` — Development setup mentions hook auto-formatting
- `PULL_REQUEST_TEMPLATE.md` — PR title format: `type(scope): short description`

### Enforced By

- **Nothing** — No `commitlint`, no `commit-msg` hook, no CI check

### Gaps

| # | Gap | Severity |
|---|---|---|
| 1 | **Conventional commits not enforced** | HIGH |
| 2 | **No Co-Authored-By convention documentation** | MEDIUM |
| 3 | **No PR/issue reference convention in commits** | LOW |

---

## 4. Branching Strategy

### Permanent Branches

| Branch | Purpose | Protection |
|---|---|---|
| `main` | Production-ready; all PRs target here | Full (required checks + 1 code owner) |
| `stable/clean` | Tracks upstream HoraDomu/Atheon | **None** (policy only) |
| `dev/full-feature` | Comprehensive testing | **None** |

### Feature Branch Conventions

- `feature/` — new features
- `fix/` — bug fixes
- `docs/` — documentation
- `test/` — test improvements
- `refactor/` — code refactoring

### Merge Strategy

- **Squash merge** only (enforced by `auto-merge.yml`)
- `--delete-branch` after merge
- PR title becomes commit subject

### Gaps

| # | Gap | Severity |
|---|---|---|
| 1 | **Single owner (`@aliasfoxkde`) for all paths** | MEDIUM |
| 2 | **`enforce_admins: false`** | HIGH |
| 3 | **`stable/clean` unprotected** | MEDIUM |
| 4 | **`dev/full-feature` is stale but still documented** | LOW |

---

## 5. PR Workflow

### Requirements (per CONTRIBUTING.md)

1. Branch from `main` → PR to `main`
2. Pre-commit hook must pass
3. Local validation: `go vet ./...`, `go test ./... -p 1`, `go build ./...`, `golangci-lint run .`
4. Pattern changes require pattern count verification
5. Pattern YAML changes require test case in `core/bundle_test.go`
6. Update `CHANGELOG.md` under `[Unreleased]`
7. PR title format: `type(scope): short description`

### Required Status Checks (Branch Protection)

```
CI / Test (Go 1.21, 1.22, 1.23, 1.24, 1.25)
CI / Lint
CI / Build (ubuntu-latest, macos-latest, windows-latest)
CI / Integration Tests
CI / Documentation Check
Security / CodeQL (Go)
Security / Self-Scan (secrets — blocking)
Security / Security Anti-Patterns
Security / Go Vulnerability Check
```

### Review Requirements

- 1 CodeOwner approval required
- Stale reviews dismissed on new commits
- Last push approval required
- All conversations must be resolved

### Gaps

| # | Gap | Severity |
|---|---|---|
| 1 | **No CHANGELOG enforcement in CI** | MEDIUM |
| 2 | **No ADR enforcement in CI** | LOW |
| 3 | **No CODEOWNERS for security-specific paths** | MEDIUM |

---

## 6. Deployment & Releases

### Release Pipeline

1. Tag push (`v*`) or scheduled (10th/21st) or manual dispatch
2. `release.yml` validates: `go build ./...` + `go test -p 1 -race -timeout 15m`
3. **MISSING**: lint, coverage, integration, pattern validation gates
4. GoReleaser builds + publishes to GitHub Releases

### Gaps

| # | Gap | Severity |
|---|---|---|
| 1 | **Release validation incomplete** — skips lint/coverage/integration | HIGH |
| 2 | **No SLSA provenance** | MEDIUM (release.yml has `--prov` flag but not confirmed working) |
| 3 | **No release artifact signing** | MEDIUM |

---

## 7. Improvement Roadmap

### Phase 1: Quick Wins (1 PR)

- [ ] Fix `Makefile setup` target path
- [ ] Add `commitlint` to pre-commit hook
- [ ] Enable `enforce_admins: true` on branch protection
- [ ] Fix `sync.yml` or remove it

### Phase 2: CI/CD Gaps (1-2 PRs)

- [ ] Add lint/coverage/integration to release validation gate
- [ ] Add auto-merge status-check gate
- [ ] Add stale PR cleanup workflow
- [ ] Add dependabot for Go module updates

### Phase 3: Documentation (1 PR)

- [ ] Create `docs/planning/SDLC.md` with full system documentation
- [ ] Create `docs/planning/COMMIT_CONVENTIONS.md`
- [ ] Update `CONTRIBUTING.md` with all SDLC details

### Phase 4: Observability (Future)

- [ ] Benchmark regression tracking
- [ ] Coverage trend dashboards
- [ ] Self-scan trend tracking

---

## 8. Testing End-to-End

### What Was Tested This Session

| Component | Result |
|---|---|
| `go test ./...` all packages | ✅ PASS |
| `go vet ./...` | ✅ PASS |
| `gofmt -l .` | ✅ PASS |
| `python3 scripts/validate-patterns.py community/` | ✅ PASS (272 patterns) |
| `go build ./bundler/` | ✅ PASS |
| `bundler` produces valid bundle | ✅ PASS (272 patterns) |
| `go build ./cmd/atheon/` | ✅ PASS |
| `atheon list` | ✅ PASS (272 patterns listed) |
| `gh pr create/merge` workflow | ✅ PASS (PRs #109-112 merged) |
| Branch protection enforcement | ✅ CONFIRMED |

### What Needs Testing

- [ ] Self-scan workflow end-to-end
- [ ] Release workflow (requires tag push — manual)
- [ ] Benchmark comparison across commits
- [ ] MCP server JSON-RPC roundtrip

---

## Appendix: File Locations

| File | Purpose |
|---|---|
| `.github/workflows/ci.yml` | Main CI pipeline |
| `.github/workflows/security.yml` | Security scanning |
| `.github/workflows/release.yml` | Release publishing |
| `.github/workflows/auto-merge.yml` | Auto-merge automation |
| `scripts/hooks/pre-commit` | Local pre-commit hook |
| `scripts/hooks/pre-push` | Local pre-push hook |
| `scripts/self-scan.sh` | Secrets/PII scanner script |
| `docs/integrations/pre-commit.md` | Hook documentation |
| `docs/BRANCH_STRATEGY.md` | Branching documentation |
| `CONTRIBUTING.md` | Contributor guidelines |
| `AGENTS.md` | AI agent conventions |
