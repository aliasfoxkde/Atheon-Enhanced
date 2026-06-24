# Owner Checklist & Recommendations

Comprehensive list of setup tasks, pending work, and recommendations for `aliasfoxkde/Atheon-Enhanced`.
Last updated: 2026-06-24. See also: [docs/integrations/mcp.md](integrations/mcp.md) | [docs/integrations/github-agents.md](integrations/github-agents.md) | [docs/integrations/pre-commit.md](integrations/pre-commit.md) | [docs/patterns/contributing-patterns.md](patterns/contributing-patterns.md)

---

## IMMEDIATE: GitHub Actions Secrets

No secrets are currently configured. Several CI features are silently degraded until these are added.

**How:** GitHub → Settings → Secrets and variables → Actions → New repository secret

| Secret | Value | Effect if missing |
|--------|-------|-------------------|
| `CODECOV_TOKEN` | From codecov.io (see below) | Optional. The workflow uploads in tokenless mode without it; with it, PR comments and status checks light up. Coderabbit MIN-8 (PR #55) flagged this as advertised-as-required but actually-optional. |

---

## IMMEDIATE: Codecov Setup

1. Go to [codecov.io](https://codecov.io) and sign in with GitHub
2. Click **+ Add a repository** → select `aliasfoxkde/Atheon-Enhanced`
3. Codecov will display an upload token — copy it
4. Add it as `CODECOV_TOKEN` in GitHub Actions secrets (above)
5. Install the [Codecov GitHub App](https://github.com/apps/codecov) on your account so the `codecov-commenter` bot can post PR comments

**After setup:** The live coverage badge in README will populate, and future PRs will get automatic patch coverage diff comments.

> **Recommendation:** Keep the `codecov.yml` patch target at 80% and project target at 90%. These are ambitious but achievable — the codebase is already at 97%. Lowering them later is easy; letting coverage drift is hard to reverse.

---

## IMMEDIATE: Open Pull Requests to Merge

Both PRs are CI-tested and ready.

| PR | Branch | What it does |
|----|--------|-------------|
| [#53](https://github.com/aliasfoxkde/Atheon-Enhanced/pull/53) | `feat/codecov-integration` | Codecov v5.4.2, live README badge, `codecov.yml` config |
| [#54](https://github.com/aliasfoxkde/Atheon-Enhanced/pull/54) | `fix/funding-yml-github-sponsors` | Removes unenrolled `github: aliasfoxkde` from FUNDING.yml |

> **Recommendation:** Merge #54 first (1-line fix, unblocks the Sponsor button error). Then merge #53 after adding `CODECOV_TOKEN` so the badge works immediately on merge.

---

## GitHub Repository Settings

### Pull Requests (Settings → General)

| Setting | Recommended | Why |
|---------|------------|-----|
| Automatically delete head branches | **Enable** | Prevents branch accumulation after merge; `pr/*` branches are excluded from auto-delete because they're not PR head branches on this repo |
| Allow squash merging | Enable (default) | Clean linear history |
| Allow merge commits | Disable (optional) | Keeps history linear |
| Allow rebase merging | Enable (optional) | Useful for simple fixes |

### Branch Protection (Settings → Branches)

**`stable/clean`** is currently protected and cannot be deleted. This branch has no active purpose.

To remove it:
1. Settings → Branches → find the `stable/clean` protection rule → delete it
2. Then: `git push origin --delete stable/clean`

> **Recommendation:** Add a branch protection rule for `main` if one doesn't exist: require PR reviews, require status checks (CI, self-scan/secrets), and disable force-push. This prevents accidental direct pushes.

### GitHub Sponsors (Settings → Sponsor this project)

`aliasfoxkde` is not yet enrolled in GitHub Sponsors, which causes a FUNDING.yml parse error. To add the GitHub Sponsors button:

1. Go to [github.com/sponsors](https://github.com/sponsors) and apply
2. Once approved, add `github: aliasfoxkde` back to `.github/FUNDING.yml`

The Ko-fi and Patreon buttons are working now (after PR #54 merges).

---

## Upstream PRs (HoraDomu/Atheon)

These are PRs submitted from this repo upstream. The `pr/*` branches **must be kept** — deleting them auto-closes the upstream PRs.

| Upstream PR | Branch | Status |
|-------------|--------|--------|
| [#176](https://github.com/HoraDomu/Atheon/pull/176) | `pr/158-deterministic-list` | Open |
| [#175](https://github.com/HoraDomu/Atheon/pull/175) | `pr/156-json-flag` | Open |
| [#173](https://github.com/HoraDomu/Atheon/pull/173) | `pr/146-147-148-perf` | Open |
| [#172](https://github.com/HoraDomu/Atheon/pull/172) | `pr/149-patterns-expansion` | Open |

Also present (check status): `pr/155-157-scan-errors`, `pr/177-fix-build-and-ci-lint-schema-timeout`, `pr/177-v2-clean`.

> **Recommendation:** Periodically check upstream PR status. When upstream merges one, sync it back: `git fetch upstream && git merge upstream/main`. This keeps the repos aligned and avoids growing divergence.

---

## GitHub Settings: What to Set Up

Assessed for this specific project — Go CLI + pattern library, public open-source, single maintainer.

### ✅ GitHub Pages → Build and Deployment

**Should you set it up? Yes.**

The `docs/` folder already contains substantial documentation. GitHub Pages publishes it as a browsable website with zero extra tooling.

**Setup (2 minutes):**
1. Settings → Pages → Source → **Deploy from a branch**
2. Branch: `main` · Folder: `/docs`
3. Save — the site goes live at `https://aliasfoxkde.github.io/Atheon-Enhanced/`

GitHub auto-detects Jekyll and renders `.md` files as HTML. To customize the theme, add `docs/_config.yml`:

```yaml
theme: jekyll-theme-cayman
title: Atheon-Enhanced
description: Pattern-matching engine for secrets, PII, and code quality
```

> **Recommendation:** Do this now. A docs site significantly improves discoverability and makes the project look more professional. All the content already exists.

---

### ✅ Secrets and Variables → Actions Variables

**Should you set it up? Yes — Variables only. Secrets already covered by `CODECOV_TOKEN`.**

`go-version: '1.23'` appears 15+ times across workflows. The 70% coverage threshold is in 6 places. The 250-pattern count gate is in 2 places. These are maintenance liabilities — one Go release means 15 edits.

**Variables to create** (Settings → Secrets and variables → Actions → Variables tab → New repository variable):

| Variable | Value | Used in |
|----------|-------|---------|
| `GO_VERSION` | `1.23` | All workflow `go-version:` pins |
| `COVERAGE_THRESHOLD` | `70` | Coverage check step in ci.yml (test job) |
| `MIN_PATTERN_COUNT` | `250` | Pattern count gate in ci.yml integration job |

**Reference in workflows** with `${{ vars.GO_VERSION }}` — note `vars.` not `secrets.`.

> **Recommendation:** Create these three Variables, then open a single PR to replace the hardcoded values across all workflows. When Go 1.24 becomes the standard pin, it's one edit instead of fifteen.

---

### ✅ Environments

**Should you set it up? Yes — one environment: `release`.**

The release workflow (`.github/workflows/release.yml`) runs on a schedule (10th and 21st) and on tag push with no human gate. A protection rule prevents accidental automated releases if a workflow bug fires early.

**Setup:**
1. Settings → Environments → New environment → name it `release`
2. Under **Deployment protection rules**, enable **Required reviewers** → add yourself (`aliasfoxkde`)
3. In `.github/workflows/release.yml`, add `environment: release` to the `release` job

That's it — every release will pause for your approval before creating the GitHub release tag.

> **Recommendation:** Also create a `staging` environment linked to the `dev/testing` branch if you ever add a deployment step there. Not needed now, but naming it early avoids confusion later.

---

### ✅ Copilot Cloud Agent (Conditional)

**Should you set it up? Yes, if you have a GitHub Copilot subscription.**

GitHub Copilot's cloud agent can automatically respond to new issues with suggested context, triage labels, or draft answers. For Atheon, the highest-value use case is **auto-responding to false positive reports** — the agent can look up the matching pattern YAML and explain why it fired.

**Setup:**
1. Settings → Copilot → Policies → Enable **Copilot in GitHub.com**
2. Under **Copilot for Issues**, enable auto-triage
3. Optionally add a `.github/copilot-instructions.md` to prime the agent with Atheon context:

```markdown
This repo is a Go pattern-matching engine. Patterns live in community/<category>/<name>.yaml.
When responding to false positive reports, look up the pattern name in community/ and explain
the regex. Suggest tightening the match with keyword context requirements.
```

> **Recommendation:** If you have Copilot, enable it. False positive issue responses are the bottleneck for community trust — automated first responses with pattern context save significant maintainer time.

---

### ⏸ Webhooks — Low Priority

**Should you set it up? Only if you have a Discord/Slack community.**

Webhooks send POST requests to an external URL on repo events (push, PR open, release, etc.). The main use case for this project would be a Discord bot announcement when a new release is published.

If you start a Discord server: Settings → Webhooks → Add webhook → paste the Discord webhook URL → select **Releases** event only. Everything else (PRs, pushes) is noisy and not worth the signal-to-noise cost.

> **Recommendation:** Skip for now. Set it up when/if you create a community Discord or Slack.

---

### ⏸ Deploy Keys — Not Needed

**Should you set it up? No.**

Deploy keys are SSH keypairs granting read or read/write access to a single repo from an external system. All CI/CD in this project uses `GITHUB_TOKEN` (automatic, scoped to the workflow run) which is already configured with the right permissions in each workflow's `permissions:` block.

Deploy keys would only be needed if an external system outside GitHub Actions needed to push to this repo. Nothing in this project requires that.

---

### ⏸ Code Review Limits — Premature

**Should you set it up? Not yet.**

Code review limits (Settings → Moderation → Code review limits) restrict PR comment/approval ability to collaborators and repository members only, preventing strangers from approving PRs.

This is valuable for high-traffic projects where random users submit approvals. With current contribution volume, it adds friction without benefit. Revisit when community PR volume picks up.

---

## Stale Branch Cleanup

After enabling "Automatically delete head branches," future merged PRs clean themselves up. For existing stale branches:

**Branches safe to delete** (no open upstream PR association):
- `stable/clean` — protected; remove protection first (see above), then delete
- `pr/177-v2-clean` — verify no open upstream PR, then delete
- `pr/177-fix-build-and-ci-lint-schema-timeout` — verify no open upstream PR, then delete

**To verify before deleting any `pr/*` branch:**
```bash
gh api "repos/HoraDomu/Atheon/pulls?state=open&per_page=50" -q '.[] | .head.ref'
```
Cross-reference against any branch you plan to delete. If the branch name appears, do not delete it.

---

## Automation Gaps

### ✅ Dependabot — configured

`.github/dependabot.yml` is now in place. Dependabot will open weekly PRs to update Go modules and pinned GitHub Actions hashes.

### ✅ goreleaser — configured

`.goreleaser.yml` is at the repo root. On each `v*-enhanced` tag push, the `goreleaser` job in `release.yml` produces signed, checksummed binaries for linux/mac/windows (amd64 + arm64). Both `atheon` and `atheon-mcp` are included.

**Note:** The homebrew and scoop taps referenced in `.goreleaser.yml` require `GH_PAT` secret and the `aliasfoxkde/homebrew-Atheon-Enhanced` / `aliasfoxkde/scoop-Atheon-Enhanced` repos to exist. Create those repos (they can be empty initially) and add a PAT with repo write access as the `GH_PAT` secret. Until then, goreleaser will build binaries but skip tap updates.

### ✅ SECURITY.md — created

`SECURITY.md` is at the repo root with responsible disclosure policy, supported versions, and response timeline.

### ✅ Actions Variables — document and create

All workflow files now use `${{ vars.GO_VERSION || '1.23' }}` and `${{ vars.COVERAGE_THRESHOLD || '70' }}` with hardcoded fallbacks. Create the variables in GitHub Settings to control these centrally:

**Settings → Secrets and variables → Actions → Variables tab → New repository variable**

| Variable | Value | Effect |
|----------|-------|--------|
| `GO_VERSION` | `1.23` | Go version for all build/test/lint jobs |
| `COVERAGE_THRESHOLD` | `70` | Minimum coverage % gate in ci.yml |
| `MIN_PATTERN_COUNT` | `250` | Pattern count gate in ci.yml integration job |

### ✅ Release version scheme — fixed

`release.yml` auto-increments the patch version from the last `v*-enhanced` tag (e.g., `v1.3.0-enhanced` → `v1.3.1-enhanced`). The previous date-based scheme (`0.YY.MM.DD`) has been removed.

### ✅ Release environment gate — configured

Add `environment: release` to the `release` job in `.github/workflows/release.yml`. Create the environment in GitHub Settings with yourself as required reviewer so each release pauses for approval.

**Settings → Environments → New environment → `release` → Required reviewers → add `aliasfoxkde`**

### ✅ Community pattern review — automated

`.github/workflows/community-pattern-review.yml` runs on same-repo PRs that touch `community/**/*.yaml`. It:
1. Validates YAML schema (`name` + `match` required; `description`/`category` encouraged; regex compilation checked)
2. Calls GitHub Models API (GPT-4o-mini, free with `GITHUB_TOKEN`) to review regex breadth
3. Posts a review comment on the PR

Security: fork PRs are blocked by `if: github.event.pull_request.head.repo.full_name == github.repository`; all PR-controlled filenames flow through `env:` (not `${{ }}` inside `run:` blocks).

### ✅ Copilot instructions — created

`.github/copilot-instructions.md` provides Atheon context to GitHub Copilot cloud agent for issue triage, false positive responses, and PR review guidance.

---

## Pattern Library Roadmap

Current count: 255 patterns across all categories.

| Category | Estimated count | Growth opportunity |
|----------|-----------------|--------------------|
| secrets | ~80 | Cloud provider tokens, new SaaS APIs |
| pii | ~40 | Regional ID formats (EU, AU, CA) |
| code-quality | ~60 | Language-specific anti-patterns |
| devops | ~50 | Kubernetes, Terraform, GitHub Actions misconfigs |
| ai-detection | ~25 | Emerging AI prompt injection patterns |

> **Recommendation:** Use real-world project scans (Atheon on your own codebases) as the primary signal for which patterns have the most false positives. The self-scan CI loop will surface these automatically. Prioritize tightening over expanding — a pattern that fires accurately on 5 cases is more valuable than one that fires noisily on 50.

---

## MCP Server (`atheon-mcp`)

The MCP server source is at `cmd/mcp/` but no `atheon-mcp.exe` binary is built locally yet.

**Build it:**
```bash
go build -o atheon-mcp.exe ./cmd/mcp   # Windows
go build -o atheon-mcp ./cmd/mcp        # Linux/macOS
```

**Configure with Claude Desktop / VS Code / Cursor:** See [docs/integrations/mcp.md](integrations/mcp.md) for full per-client setup instructions.

**Distribute to users:** Add `atheon-mcp` as a release artifact via goreleaser (see Automation Gaps below) so users without Go can download and use it.

---

## GitHub AI Agents & Models Integration

See [docs/integrations/github-agents.md](integrations/github-agents.md) for the full guide. Summary of what's available free for public repos:

| Feature | Status | Cost | Best use for Atheon |
|---------|--------|------|---------------------|
| GitHub Models API | Live | Free (rate-limited) | Auto-review new community pattern PRs |
| GitHub Agentic Workflows | Technical preview | Free in preview | PR summarization, pattern suggestion |
| MCP in Copilot Chat | Live | Free (needs Copilot) | Users scanning code via chat |
| Copilot Workspace | Live | Paid | Reviewing complex upstream PRs |

**Recommended first step:** Add a `community-pattern-review.yml` workflow that calls GitHub Models API (`GITHUB_TOKEN` auth, no extra secrets) to review new pattern YAML files for regex breadth and missing test cases on each PR. Zero cost, immediate value.

---

## Should You Use Cheaper AI Models for This Work?

**Short answer: Yes, selectively.**

| Task | Recommended model | Why |
|------|------------------|-----|
| Writing/updating docs | Haiku 4.5 or GPT-4o-mini | Structured, low-ambiguity task — fast and cheap |
| Expanding pattern library (new YAML patterns) | Haiku 4.5 | Templated output, clear schema |
| Reviewing PRs for obvious issues | GPT-4o-mini via GitHub Models | Free with `GITHUB_TOKEN` |
| Coverage improvement (adding test stubs) | Haiku 4.5 | Mechanical — copy existing test structure |
| Debugging CI failures / jq/regex issues | Sonnet 4.6 | Requires reasoning over multi-file context |
| Architectural decisions | Sonnet 4.6 or Opus 4.7 | Needs full project context and judgment |
| Security pattern design (FP analysis) | Sonnet 4.6 | Subtle — cheap models tend to miss edge cases |

**How to route:** In Claude Code, use `/effort low` for docs and test-stub tasks (routes to faster, cheaper processing). Use the default for debugging and architecture.

**What "cheap AI" gets wrong on this codebase:**
- jq filter syntax (the `select((COND) | not)` vs `select(COND) | not` bug happened in a low-context state)
- Regex edge cases in patterns (false positive risk is subtle)
- Multi-file CI debugging (needs full workflow + script context simultaneously)

**Rule of thumb:** If the output is templated and verifiable by a linter or test, cheap is fine. If the output requires judgment that a human would need to think about, use the stronger model.

---

## Documentation

| Doc | Status | Notes |
|-----|--------|-------|
| `docs/patterns/contributing-patterns.md` | ✅ Created | YAML schema, RE2 constraints, FP testing, PR checklist |
| `SECURITY.md` | ✅ Created | Responsible disclosure, supported versions, response SLA |
| `docs/integrations/pre-commit.md` | ✅ Created | Hook installation, what runs when, troubleshooting |
| `docs/integrations/mcp.md` | ✅ Created | Claude Desktop, VS Code, Cursor, Windsurf setup |
| `docs/integrations/github-agents.md` | ✅ Created | GitHub Models API, Agentic Workflows, Copilot Extensions |
| `.github/copilot-instructions.md` | ✅ Created | Copilot cloud agent context for issue triage |
| `docs/development.md` | ✅ Created | Branch strategy, local setup, release process |
| `docs/OWNER_CHECKLIST.md` | ✅ (this file) | Setup tasks, recommendations, maintenance |

---

## Recurring Maintenance

| Task | Cadence | How |
|------|---------|-----|
| Sync upstream (`HoraDomu/Atheon`) | Weekly or when upstream merges land | `git fetch upstream && git merge upstream/main` |
| Check upstream PRs still open | Before deleting any `pr/*` branch | `gh api "repos/HoraDomu/Atheon/pulls?state=open" -q '.[].head.ref'` |
| Review Codecov patch coverage on each PR | Per PR | Codecov bot will post automatically once configured |
| Pattern quality review | Monthly | Run `bash .github/scripts/self-scan.sh --quality` against a diverse set of real-world repos |
| Dependabot PRs | Weekly | Merge Go module + Actions updates promptly |
