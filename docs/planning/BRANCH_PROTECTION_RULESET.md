# Branch Protection Ruleset — 2026-06-27

## Overview

This document describes the branch protection strategy for Atheon-Enhanced, covering all permanent branches, required status checks, review requirements, and known gaps.

**Status**: Documented — some items require GitHub admin UI action.

---

## Permanent Branches

| Branch | Purpose | Protection Level |
|---|---|---|
| `main` | Production-ready; all PRs target here | **Full** — required checks + code owner |
| `stable/clean` | Tracks upstream HoraDomu/Atheon | **None** (policy-only, read-only sync target) |
| `dev/full-feature` | Comprehensive testing (stale, unused) | **None** |

---

## Main Branch Protection

### Required Status Checks (Branch Protection Rules)

```
CI / Test (Go 1.22)
CI / Test (Go 1.23)
CI / Test (Go 1.24)
CI / Test (Go 1.25)
CI / Lint
CI / Build (ubuntu-latest)
CI / Build (macos-latest)
CI / Build (windows-latest)
CI / Integration Tests
CI / Documentation Check
Security / CodeQL (Go)
Security / Self-Scan (secrets — blocking)
Security / Security Anti-Patterns
Security / Go Vulnerability Check
```

> **Note**: Go 1.21 was removed from CI and branch protection in Wave 13 (PR #129) as it reached EOL.

### Review Requirements

- **1 CodeOwner approval** required
- Stale reviews dismissed on new commits
- Last push approval required
- All conversations must be resolved

### Additional Settings

- **Required linear history**: Enabled (no merge commits on main)
- **Require deployments to succeed before merging**: Disabled (no GitHub Environments configured)
- **Allow force pushes**: Disabled for all
- **Allow deletions**: Disabled
- **Block creation**: Enabled (prevents branch creation via API if protected)
- **Required conversation resolution**: Enabled
- **Lock branch**: Disabled
- **Allow fork syncing**: Disabled
- **Do not allow bypassing the above settings**: **Enabled** — enforces all above rules for administrators and custom roles with "bypass branch protections" permission

---

## Gap Analysis

### HIGH Priority

| # | Gap | Risk | Resolution |
|---|---|---|---|
| 1 | **`enforce_admins: false`** | Admin users can bypass branch protection entirely (push directly to main) | **Requires GitHub admin UI action** — navigate to Settings → Branches → Protection rules → Enable "Include administrators" |

> **Note**: "Do not allow bypassing the above settings" IS enabled — this constrains users with explicit "bypass branch protections" permission, but does NOT address `enforce_admins: false`, which exempts admin users from all branch protection rules entirely.

### MEDIUM Priority

| # | Gap | Risk | Resolution |
|---|---|---|---|
| 2 | **`stable/clean` unprotected** | Accidental force-push or deletion could break sync | Document as policy-only; or add protection rule |
| 3 | **Single owner (`@aliasfoxkde`) for all paths** | Burnout / single point of failure | Add co-maintainers for specific categories |
| 4 | **`dev/full-feature` is stale** | Confuses contributors | Remove from documentation or archive |

### LOW Priority

| # | Gap | Risk | Resolution |
|---|---|---|---|
| 5 | **No CODEOWNERS for security-specific paths** | Security changes lack security reviewer | Add `/community/security-*` owner entry |

---

## enforce_admins Resolution

To enable `enforce_admins: true`:

1. Navigate to: https://github.com/aliasfoxkde/Atheon-Enhanced/settings/branches
2. Click "Edit" on the `main` branch protection rule
3. Check "Include administrators"
4. Save

This cannot be set via GitHub API or CLI without repo admin rights.

---

## CODEOWNERS Structure

```
# Default: maintainer for everything
* @aliasfoxkde

# CI/CD — extra scrutiny
/.github/ @aliasfoxkde
/.goreleaser.yml @aliasfoxkde

# Core product: pattern engine
/core/ @aliasfoxkde
/bundler/ @aliasfoxkde

# Wire protocols (JSON-RPC, JSON, SARIF)
/cmd/ @aliasfoxkde

# Community patterns — co-maintainers welcome
/community/ @aliasfoxkde

# Documentation
/docs/ @aliasfoxkde
```

### Adding a Co-Maintainer

To add a co-maintainer for a specific category:

```gitignore
# In CODEOWNERS, add per-category entries:
/community/secrets/ @co-maintainer-username
/community/web-security/ @security-reviewer
```

---

## Branch Protection API State

Queried via `gh api repos/aliasfoxkde/Atheon-Enhanced/branches/main/protection`:

```json
{
  "enforce_admins": { "enabled": false },
  "required_linear_history": { "enabled": true },
  "allow_force_pushes": { "enabled": false },
  "allow_deletions": { "enabled": false },
  "block_creations": { "enabled": false },
  "required_conversation_resolution": { "enabled": true },
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true,
    "require_last_push_approval": true,
    "required_approving_review_count": 1
  },
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "CI / Test (Go 1.22)",
      "CI / Test (Go 1.23)",
      "CI / Test (Go 1.24)",
      "CI / Test (Go 1.25)",
      "CI / Lint",
      "CI / Build (ubuntu-latest)",
      "CI / Build (macos-latest)",
      "CI / Build (windows-latest)",
      "CI / Integration Tests",
      "CI / Documentation Check",
      "Security / CodeQL (Go)",
      "Security / Self-Scan (secrets — blocking)",
      "Security / Security Anti-Patterns",
      "Security / Go Vulnerability Check"
    ]
  }
}
```

---

## Relationship to Other Docs

- **SDLC_AUDIT_2026-06-27.md** — Section 4 (Branching Strategy) covers the same ground with CI/CD context
- **CONTRIBUTING.md** — Links to branch protection requirements for contributors
- **AGENTS.md** — AI agents instructed to never force-push or delete branches

---

## Future Improvements

- Add co-maintainers for specific pattern categories
- Archive or remove `dev/full-feature` documentation
- Enable `enforce_admins: true` (requires UI action)
- Add protection rule for `stable/clean` if it becomes a merge target