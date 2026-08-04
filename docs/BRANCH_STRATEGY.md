# Branch Strategy

**Last Updated**: 2026-06-25

This document describes the branching strategy for the Atheon-Enhanced project. Three permanent branches carry the project's lifecycle (stable/clean, main, dev/full-feature), and short-lived feature branches land work via pull requests. Configuration profiles under `config/profiles/` tune scanner behavior per environment.

---

## Core Branches

### `main` — Production Build Branch

**Purpose**: Production-ready build with all validated enhancements.

**Characteristics**:
- ✅ All validated PRs merged
- ✅ 406 patterns across 29 categories
- ✅ MCP server (`atheon-mcp`)
- ✅ Severity wired end-to-end into SARIF
- ✅ Fuzz-tested bundle loader
- ✅ Multi-version Go CI matrix

**Installation**:
```bash
go install github.com/aliasfoxkde/Atheon@latest
```

**When to use**:
- User-facing installation
- Production deployment
- Release tagging
- Feature integration testing

### `stable/clean` — Upstream Tracking Branch

**Purpose**: Reference point tracking upstream HoraDomu/Atheon.

**Characteristics**:
- Clean upstream code without modifications
- Updated periodically via manual sync (no automated daily job)
- Reference point for upstream changes
- Maintainer-only access

**When to use**:
- Checking upstream changes
- Resolving merge conflicts
- Verifying upstream compatibility

> **Note**: Despite the legacy sync-script reference (`.github/scripts/sync-stable-clean.sh`), there is no automated daily sync — the script is preserved for manual runs. Synchronization with upstream is opportunistic, not scheduled. See the [deferred items in TASKS.md](./TASKS.md#deferred--open) for related cleanup.

### `dev/full-feature` — Development Branch

**Purpose**: Comprehensive testing with ALL patterns enabled.

**Characteristics**:
- All 406 patterns enabled
- Experimental features active
- Self-scanning validation enabled
- Performance benchmarking enabled

> **Note**: This branch is documented for completeness but is **not actively maintained** at the moment. The work that previously lived here has been absorbed into the main branch's hardening waves. See the [open question in PLAN.md](./PLAN.md#open-questions) on whether to revive it.

---

## Workflow

1. **Feature Development**: Create feature branches from `main` (e.g. `feature/wave-6-docs-fill-and-cleanup`).
2. **Verify locally**: `go test ./... -p 1`, `gofmt -l .`, manual smoke.
3. **Open PR**: Target `main`. Open via REST (`gh api -X POST repos/.../pulls --field ...`) — `gh pr create` GraphQL is flaky on this repo.
4. **Resolve review threads**: CodeRabbit threads must be resolved before merge.
5. **Squash-merge**: Via `gh pr merge <num> --squash --delete-branch`.
6. **Cleanup**: Update `MEMORY.md` with a wave summary if the PR is part of a wave cluster.

---

## Feature Branch Naming

```
feature/   # New features
fix/       # Bug fixes
docs/      # Documentation updates
test/      # Test improvements
refactor/  # Code refactoring
```

Examples from shipped waves: `feature/wave-6-docs-fill-and-cleanup`, `feature/severity-wiring`, `fix/--json--version-ordering`, `docs/fill-plan-tasks`.

---

## Branch Protection Rules

| Branch | Push access | PR required | CI required |
|--------|-------------|-------------|-------------|
| `main` | Maintainer | Yes | Yes |
| `stable/clean` | Maintainer only | Yes | — |
| `dev/full-feature` | Maintainer | Yes | — |

`main` is the only branch that auto-publishes (via GoReleaser on tagged releases).

---

## Configuration Profiles

Profiles under `config/profiles/` tune scanner behavior per environment:

| Profile | Use case | Patterns | Notable |
|---------|----------|----------|---------|
| `development.json` | Local dev | All enabled | Debug logging, self-scan friendly |
| `mcp-integration.json` | MCP server | All | Streaming output, AI-assistant-friendly |
| `pipeline.json` | CI/CD | Security + quality | JSON output, fast mode |
| `production.json` | Production | Curated | Conservative defaults |

The CI grep check in `.github/workflows/ci.yml` verifies that any new profile under `config/profiles/` is also documented here.

---

## Branch-Specific Configurations

### `main`
- **Profile**: `config/profiles/production.json`
- **Patterns**: All 406 patterns, per-category enablement
- **Testing**: Multi-version Go matrix + integration tests
- **Features**: MCP server, SARIF severity, fuzz coverage

### `stable/clean`
- **Profile**: Default upstream settings
- **Patterns**: Whatever upstream ships at last sync
- **Testing**: Smoke-only

### `dev/full-feature`
- **Profile**: `config/profiles/development.json`
- **Patterns**: All 406 enabled regardless of category
- **Testing**: Comprehensive + self-scanning

---

## Branch Comparison

| Feature | stable/clean | main | dev/full-feature |
|---------|--------------|------|------------------|
| Upstream Sync | ✅ Manual | ⚠️ Opportunistic | ❌ None |
| Pattern Count | Upstream default | 406 | 406 (all enabled) |
| MCP Integration | ❌ None | ✅ Yes | ✅ Yes |
| Severity Wiring | ❌ None | ✅ Yes | ✅ Yes |
| Fuzz Coverage | ❌ None | ✅ Yes | ✅ Yes |
| Experimental Features | ❌ None | ❌ No | ✅ Yes |
| Self-Scanning | ❌ No | ⚠️ Optional | ✅ Yes |
| User Installation | ❌ No | ✅ Yes | ⚠️ Testing only |

---

## Decision Tree

```
Starting work?
│
├─ Need clean upstream baseline?      → stable/clean
├─ Adding a new feature?              → feature/* off main  →  PR to main
├─ Fixing a bug?                      → fix/* off main      →  PR to main
├─ Pattern YAML only?                → docs/patterns/* off main → PR to main
├─ Need every pattern enabled for testing? → dev/full-feature
└─ Production deployment / release tag?   → main
```

---

## Update & Sync Workflow

### Periodic upstream sync (manual)

```bash
git checkout stable/clean
git fetch upstream
git merge upstream/main
git push origin stable/clean
```

### Feature branch rebases

```bash
# While working on a feature branch, keep it current with main:
git checkout feature/my-feature
git merge --no-ff main
# Resolve conflicts, continue.
```

---

## Branch Maintenance

**Weekly**:
- Review and merge completed PRs to main
- Clean up stale remote branches (`git fetch --prune`)
- Verify CI green on the latest main

**Monthly**:
- Audit merged PRs for CHANGELOG coverage
- Review dependabot PRs (grouped: GitHub Actions weekly, Go minor weekly, Go patch daily)
- Spot-check that all `config/profiles/*.json` are still referenced here

**Quarterly**:
- Branch strategy review — is `dev/full-feature` worth reviving?
- Sync with upstream (`stable/clean`) and assess any large upstream changes
- Profile audit (drop unused profiles, add new ones for new features)

---

## Quick Start

### For users

```bash
# Install the latest released version
go install github.com/aliasfoxkde/Atheon@latest

# Scan a directory
atheon ./my-project

# Output JSON for a CI artifact
atheon --json ./my-project > findings.json
```

### For contributors

```bash
# Fork on GitHub, clone your fork
git clone git@github.com:<you>/Atheon-Enhanced.git
cd Atheon-Enhanced

# Branch off main
git checkout -b feature/my-feature main

# Make changes, then
go test ./... -p 1
gofmt -l .

# Push and open a PR targeting main
git push -u origin feature/my-feature
gh api -X POST repos/aliasfoxkde/Atheon-Enhanced/pulls \
  --field title='feat: my feature' \
  --field head='<you>:feature/my-feature' \
  --field base='main'
```

### For maintainers

```bash
# Review PRs
gh pr list --state open

# Merge with squash + auto-delete branch
gh pr merge <pr-number> --squash --delete-branch

# Verify CI is green on main
gh workflow view ci --yaml | head -20
```

---

## References

- [PLAN.md](./PLAN.md) — overall project plan and wave timeline
- [TASKS.md](./TASKS.md) — task ledger (open and completed work)
- [RELEASE.md](./RELEASE.md) — release runbook (tag format, bundle regen, publishing)
- [.github/workflows/ci.yml](../.github/workflows/ci.yml) — the CI grep check that enforces this doc
