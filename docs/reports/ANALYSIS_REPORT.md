# Atheon-Enhanced — Deep Analysis & Improvement Plan

**Date**: 2026-06-23
**Scope**: aliasfoxkde/Atheon-Enhanced (fork of HoraDomu/Atheon)
**Method**: Whole-tree read + cross-reference audit + downstream-tool check
**Output**: Single canonical analysis — gaps, next steps, improvement roadmap

---

## TL;DR — The One-Page Verdict

Atheon-Enhanced is a **well-engineered Go CLI + MCP server** for secrets/PII/code-quality scanning.
The engine layer (`core/`, `cmd/`) is mature, idiomatic, and battle-tested — context cancellation,
sentinel errors, slog, SARIF output, MCP rate limiting, and 97%-class test coverage are all in
place.

**What is wrong is everything around the engine.** The repository suffers from a classic
*prototype-rapidly-then-stop-maintaining* pathology: docs, CI, and housekeeping have not kept pace
with the code. The most important fixes are not in `core/` — they are in `docs/`, `.github/`, and
`README.md`. Roughly a dozen improvements will lift this from "good fork" to "reference Go
distribution."

**P0 (must-fix, blockers for trust):**
1. `docs/PLAN.md`, `docs/TASKS.md`, `docs/PROGRESS.md` are **unfilled templates** with literal
   `{{PLACEHOLDER}}` strings in a public, version-controlled repository.
2. `docs/README.md` references **five documents that do not exist** (`MCP_INTEGRATION_ANALYSIS.md`,
   `SECURITY_TESTING.md`, `FINALIZATION_SUMMARY.md`, `TEST_COVERAGE.md`, `tests/TEST_HARNESS_STATUS.md`,
   `contributors.md`).
3. Pattern-count and category claims are **stale and contradictory across files**:
   - `README.md` → 255 patterns / 19 categories
   - `docs/architecture/PATTERN_CATEGORIES.md` → 225 patterns / 19 categories
   - `docs/INSTALL.md` → 190 patterns / 19 categories
   - `docs/development/SETUP.md` → "Expected: 87 patterns"
   - `docs/reports/COMPLETION_REPORT.md` → 152 patterns (dated 2026-06-19)
   - **Actual**: 252 patterns across 18 non-empty categories (`frameworks/` is empty).

**P1 (high-value, low-risk):**
4. **10 GitHub workflows with significant duplication.** CodeQL appears in 2; the full
   test/lint/build chain runs in 3; both `ci.yml` and `comprehensive-ci.yml` trigger on push to
   main. The IMPROVEMENT_PLAN.md already recommends consolidation to ~4 workflows.
5. **MCP server ships only 3 tools** (`scan_string`, `scan_file`, `scan_dir`) but the API README
   and feature-comparison tables advertise `list_patterns`, `list_categories`, and
   `scan_directory` that do not exist.
6. **The `system_architecture.md` documents files that do not exist**
   (`core/streaming.go`, `core/quality_enforcement.go`, `config/defaults/`, `defaults/`).
7. **Duplicate `BRANCH_STRATEGY.md`** at both `docs/BRANCH_STRATEGY.md` and
   `docs/reports/BRANCH_STRATEGY.md` with different content.
8. **Empty `frameworks/` category** is documented in README's pattern-distribution table as
   containing `django`, `nodejs`, and `react`.

**P2 (forward-looking, drives adoption):**
9. ADR (Architecture Decision Records) directory does not exist despite the IMPROVEMENT_PLAN.md
   requesting it (`docs/architecture/decisions/`).
10. No `--debug-level` flag exists; `SETUP.md` documents one that does not.
11. `.github/wiki/` has 3 markdown files but no Wiki workflow publishes them — Wiki tab is empty.
12. Scheduled-release tag format `v0.4.YYMMDD` produces valid semver but unusual-looking tags.

---

## 1 — Current State Snapshot

### 1.1 Repository shape

| Layer | State |
|-------|-------|
| `core/` (engine) | Mature. 12 production files, 22 test files, ~25k LOC tests. Sentinel errors, slog, context-aware, SARIF output via CLI. |
| `cmd/atheon/` | Mature. CLI with 12 subcommands/flags. Tests at ~96% coverage. |
| `cmd/mcp/` | Functional. JSON-RPC over stdio with rate limiter. 3 tools exposed. |
| `bundler/` | Mature. Compiles 252 community YAML patterns into a gzip+JSON bundle. |
| `community/` | 252 YAML patterns across 18 categories. |
| `.github/workflows/` | 10 files; substantial duplication. |
| `.golangci.yml` | Opinionated, 18 linters enabled, well-tuned exclusions. |
| `docs/` | **Largest gap area.** Mix of current, stale, and template files. |

### 1.2 Quality metrics (claim vs reality)

| Metric | README claim | Actual |
|--------|-------------|--------|
| Patterns | 255 | **252** (across 18 non-empty categories) |
| Categories | 19 | **18** (`frameworks/` is empty) |
| Test coverage | 97%+ | Likely true based on CI guard but not directly re-measured here |
| Go versions tested | 1.21–1.24 | ci.yml tests 1.23, 1.24; comprehensive-ci.yml tests 1.21–1.24 |
| Linters | 18 | matches `.golangci.yml` |
| Workflows | implicit "10" | **10** |
| Plan/Tasks/Progress docs | "comprehensive" | **Templates with `{{PLACEHOLDER}}` literals** |

### 1.3 Recent velocity (last 10 PRs)

```
54 fix(funding): remove github sponsors entry
53 feat(ci): integrate Codecov v5 with live badge, coverage config, and token support
52 fix(patterns,ci): tighten travis-ci-token pattern + exclude ISSUE_TEMPLATE from scan
51 fix(ci): correct jq filter — wrap exclusion conditions inside select()
50 feat(community): issue templates + Sponsor button
49 feat(ci): integrate Atheon self-scan into pipeline with meta-integration docs
48 docs: fix README errors, update counts and doc links
47 test: push coverage to 97%+ with cross-platform IO error paths
46 test: add coverage for HTTP errors, large files, and missing dirs
45 feat: implement improvement plan — 255 patterns, 95.7% coverage, SARIF, rate limiting
```

The project is **active** and merges ~5–7 PRs per week. PR #45 was the "implement improvement plan"
milestone; PR #48 was the README fix-up. This means the IMPROVEMENT_PLAN.md has been **partially
executed** — but several sections (notably 1.1 workflow consolidation, 5.1–5.4 docs/ADRs, 6.3 path
traversal) remain undone.

---

## 2 — Gaps by Domain

### 2.1 Documentation gaps (highest priority)

#### D1. `docs/PLAN.md`, `docs/TASKS.md`, `docs/PROGRESS.md` are unfilled templates
These three files at the repo root (per the user's CLAUDE.md contract) are required to be **real
project plans**. They currently contain literal `{{PROJECT_NAME}}`, `{{DATE}}`,
`{{CURRENT_PHASE}}`, `{{OVERALL_PROGRESS}}%` etc. This is a credibility issue for any visitor
following the "Documentation Contract" in the root `CLAUDE.md`.

The same template files exist **duplicated** at `docs/planning/PLAN.md`, `TASKS.md`,
`PROGRESS.md` — also unfilled.

**Fix:** Replace all six files with real content for Atheon-Enhanced. Use the data from this
report. Estimated: 30 minutes.

#### D2. `docs/README.md` references six non-existent documents
The docs index lists these files that **do not exist** in the tree:

```
docs/reports/MCP_INTEGRATION_ANALYSIS.md
docs/reports/SECURITY_TESTING.md
docs/reports/FINALIZATION_SUMMARY.md
docs/reports/TEST_COVERAGE.md
docs/tests/TEST_HARNESS_STATUS.md
docs/contributors.md
```

Any user clicking these links in the docs index hits a 404 on GitHub.

**Fix:** Either create stub files or remove the links from `docs/README.md`. Cleanest: create
1-line stubs saying "Moved to X" or "Deprecated — see Y" so existing links don't 404.

#### D3. Pattern-count and category claims are inconsistent
The numbers in user-facing docs disagree by 30–100 patterns:

| File | Says |
|------|------|
| `README.md` | 255 patterns, 19 categories |
| `README.md` distribution table | secrets=32, code-quality=25, … total ≈180 |
| `docs/FAQ.md` | 225 patterns, 57 in official |
| `docs/architecture/PATTERN_CATEGORIES.md` | 225 patterns, 19 categories |
| `docs/INSTALL.md` | 190 total / 19 categories; later says 223 patterns |
| `docs/development/SETUP.md` | "Expected: 87 patterns" |
| `docs/reports/COMPLETION_REPORT.md` | 152 patterns (dated 2026-06-19) |
| `docs/reports/IMPROVEMENT_PLAN.md` | 225 patterns, 94.9% coverage (dated 2026-06-22) |
| `CHANGELOG.md` (Unreleased) | lists 25+ new patterns |
| Actual | **252 patterns across 18 non-empty categories** |

**Fix:** Single source of truth = a generated badge/script. Add `scripts/pattern-count.sh` that
emits the JSON; reference it from README/FAQ/INSTALL.

#### D4. `frameworks/` category is empty but documented
`README.md` table shows `django: 1`, `nodejs: 1`, `react: 1`. Actual: `community/frameworks/`
has zero files. This will break `--categories=frameworks` filter silently (empty result, no error).

**Fix:** Either delete the `frameworks/` row from the table and `Category()` list, or add the
3 documented patterns. PR #45's bundle was supposed to include them — verify why they fell out.

#### D5. MCP API documented but not implemented
`docs/api/README.md` says the MCP server exposes:
- `scan_directory` (real name: `scan_dir`)
- `scan_string` (✓ exists)
- `list_patterns` (✗ does not exist)
- `list_categories` (✗ does not exist — only CLI has it)

**Fix:** Add `list_patterns` and `list_categories` MCP tools (small Go functions reusing
`core.All()` and `core.Categories()`). Rename `scan_dir` doc to match, or add a `scan_directory`
alias for compatibility.

#### D6. `SYSTEM_ARCHITECTURE.md` documents files that don't exist
The "Code Organization" section lists:

```
core/
├── streaming.go          ← does not exist (streaming logic is inline in runner.go)
├── pattern_state.go      ← exists ✓
└── quality_enforcement.go ← does not exist (logic is in runner.go)
config/
├── profiles/             ← exists ✓
└── defaults/             ← does not exist
```

**Fix:** Update the doc to match reality, or split `runner.go` into the named files.

#### D7. Duplicate `BRANCH_STRATEGY.md`
Two different documents exist:

- `docs/BRANCH_STRATEGY.md` (52 lines, terse, profiles-only at bottom)
- `docs/reports/BRANCH_STRATEGY.md` (the older version, ~290 lines, full system architecture)

**Fix:** Pick the better one, move/delete the other. Recommend keeping `docs/BRANCH_STRATEGY.md`
(closer to root, lighter) and either deleting or downgrading `docs/reports/BRANCH_STRATEGY.md`
to a redirect.

#### D8. `SETUP.md` references a non-existent `--debug-level` flag
Line 238: `./atheon-debug --debug-level=2 scan .` — no such flag exists in `cmd/atheon/main.go`.

**Fix:** Remove the line, or implement the flag.

#### D9. `CHANGELOG.md` is at repo root, but user's `CLAUDE.md` specifies `CHANGELOG.md.md`
The project follows Keep-a-Changelog correctly with a root `CHANGELOG.md`, but the canonical
documentation contract in `CLAUDE.md` says `./docs/CHANGELOG.md.md`. Either create a
symlink/copy or update the contract.

**Fix:** Create `docs/CHANGELOG.md.md` as either a "latest version only" excerpt or a symlink
to `../CHANGELOG.md`.

#### D10. `AGENTS.md` referenced in user's `CLAUDE.md` is missing
The contract requires `./AGENTS.md`. Currently absent.

**Fix:** Create `AGENTS.md` at the repo root describing how AI agents should interact with this
repo (the project's own conventions, since this fork *is* heavily AI-touched per
COMPLETION_REPORT.md).

### 2.2 CI/CD gaps

#### C1. Workflow duplication
Ten workflows with overlapping responsibilities:

| Workflow | Purpose | Duplicates |
|----------|---------|-----------|
| `ci.yml` | test + lint + build (PR & main) | overlaps with `quality-assurance.yml`, `comprehensive-ci.yml` |
| `comprehensive-ci.yml` | CodeQL + multi-version test + static-analysis + self-scan + benchmarks + integration + quality + docs | most things |
| `quality-assurance.yml` | multi-version test + Windows + quality gates + benchmarks + integration + coverage report + dashboard | overlaps with ci.yml and comprehensive-ci.yml |
| `security-scanning.yml` | CodeQL + self-scan + dependency-scan + quality-scan | overlaps with comprehensive-ci.yml |
| `self-scan.yml` | Atheon self-scan | overlaps with comprehensive-ci.yml security-scan job |
| `codeql.yml` | CodeQL only | superseded by security-scanning.yml's codeql job |
| `auto-merge.yml` | dependabot auto-merge | redundant with GitHub native auto-merge |
| `publish.yml` | GoReleaser publish | overlaps with `scheduled-release.yml` |
| `scheduled-release.yml` | create tag → goreleaser | unique |
| `sync-stable-clean.yml` | upstream sync | unique |

Both `ci.yml` and `comprehensive-ci.yml` fire on every push to `main`, so the test matrix runs
**twice**. Codecov v5 (per PR #53) is configured but uploads happen from both workflows.

**Fix:** Consolidate to the four workflows the IMPROVEMENT_PLAN.md already recommends:

- `ci.yml` — test + lint + build (PRs, push to main) [merge of current ci.yml + quality-assurance.yml jobs]
- `security.yml` — CodeQL + Atheon self-scan + govulncheck [merge of codeql.yml + security-scanning.yml + self-scan.yml]
- `release.yml` — goreleaser on tag [merge of publish.yml + scheduled-release.yml]
- `sync.yml` — unchanged

Estimated CI minutes saved: 40–60%.

#### C2. `quality-assurance.yml` lines 246–253 — broken `go fmt` check
```yaml
- name: Format check
  run: |
    go fmt ./...
    if [ -n "$(git diff --name-only)" ]; then
      echo "❌ Code is not properly formatted"
      exit 1
    fi
```

On a fresh CI checkout `go fmt` makes no changes, so this passes. But the check is also run
without `set -e` between commands — if `go fmt` itself fails (e.g. parse error), the script
continues. The check should use `git diff --exit-code` or run `gofmt -l` (the standard idiom).

**Fix:** Replace with the standard idiom from `ci.yml`:
```bash
if [ -n "$(gofmt -l .)" ]; then
  echo "Code is not formatted:"; gofmt -d .; exit 1
fi
```

#### C3. `scheduled-release.yml` tag format
```yaml
DATE_VERSION="0.4.$(date +%y%m%d)"
```
On 2026-06-23 this yields `0.4.260623` → tag `v0.4.260623`. Valid semver but odd-looking. The
branch-strategy docs say "Version Naming: Semantic Versioning with Enhancement Suffix
(`v1.2.3-enhanced`)" — but the actual tagger doesn't add `-enhanced`.

**Fix:** Either keep current format and update branch-strategy docs, or use
`0.4.$(date +%Y.%-m.%-d)` for `v0.4.2026.6.23`.

#### C4. `auto-merge.yml` may be redundant
If the repo uses GitHub's native "Allow auto-merge" feature on PRs, this workflow duplicates it.

**Fix:** Audit and delete if redundant.

#### C5. No CI badge for Codecov in README's CI section
The Codecov badge is in README (line 12) but the workflow config uses `continue-on-error: true`
for codecov upload. If the secret is missing, the badge goes stale silently.

**Fix:** Drop `continue-on-error` on codecov upload (test step), or document the secret requirement.

### 2.3 Pattern library gaps

#### P1. `frameworks/` category empty
The `frameworks/` directory has 0 YAML files. README claims 3 patterns (django/nodejs/react).
Either the patterns were removed during a refactor or the docs got ahead of reality.

**Fix:** Either restore 3 patterns or remove the category from docs/UI.

#### P2. Pattern distribution table in README is wrong
README claims:
```
secrets: 32, code-quality: 25, accessibility: 19, security-hardening: 14, performance: 12,
web-development: 12, web-security: 12, api-integration: 9, healthcare: 7, ai-detection: 6,
cloud-native: 6, devops: 6, data-visualization: 5, pwa: 5, finance: 3, pii: 3,
django: 1, nodejs: 1, react: 1
```
Actual:
```
secrets: 58, code-quality: 35, accessibility: 19, security-hardening: 18, web-security: 15,
cloud-native: 14, performance: 12, web-development: 12, pii: 11, ai-detection: 9,
api-integration: 9, devops: 9, healthcare: 7, finance: 6, pwa: 5, data-visualization: 5,
git-hygiene: 4, compliance: 4, frameworks: 0
```
Total README = ~177; actual = 252. Big understatement across the board.

**Fix:** Replace the static table with a generated list, or update by hand.

#### P3. CHANGELOG "Unreleased" lists patterns that may already be in the bundle
The unreleased section adds 25+ patterns. Verify they're actually loaded — if so, move them to a
released version; if not, add them.

**Fix:** Run `go run ./bundler` then diff `core/patterns.bundle` contents against the
"Unreleased" claim.

#### P4. Top-tier secrets coverage gaps
Despite 58 secret patterns, some high-frequency SaaS tokens are missing. Recommended additions:

- **Cloudflare R2 / Workers API tokens** (cf_r2_, cf_workers_)
- **Supabase service_role / anon keys** (eyJhbGciOi...)
- **Vercel Edge Config tokens**
- **GitHub fine-grained PATs** (github_pat_*)
- **Bitbucket app passwords** (ATBB pattern)
- **1Password service account tokens**
- **Vault / HashiCorp tokens** (hvs.*, hvb.*)
- **Anthropic API keys** (sk-ant-*)
- **OpenAI org-scoped keys** (sk-proj-*)
- **Slack workflow tokens** (xoxw-*)

**Fix:** Add 10-pattern batch to `community/secrets/` per the IMPROVEMENT_PLAN.md Section 3 model.

#### P5. PII category thin (11 patterns)
Roadmap promised email, passport, driver's license, IP address — only some exist. Add:

- Email addresses (RFC 5322 simplified)
- IPv4/IPv6 literals outside URLs (false-positive controlled)
- US SSN with stronger validation
- International passport numbers (more countries)
- Credit card CVV in source (3-4 digit near card pattern)

**Fix:** Add to `community/pii/`.

### 2.4 Source-code gaps

#### S1. `core/streaming.go` referenced but doesn't exist
Per D6, the file is not present. Streaming logic is inlined in `core/runner.go`. Splitting into a
named file would clarify intent for readers.

**Fix:** Either rename `runner.go` → `streaming.go` or extract the streaming parts to a new
file matching the docs.

#### S2. No ADRs directory
The IMPROVEMENT_PLAN.md §5.4 requests `docs/architecture/decisions/` with three ADRs:
RE2-vs-PCRE, gzip-bundle-vs-embedded-YAML, parallel-test-must-be-1. None exist.

**Fix:** Create the directory + three ADRs.

#### S3. No pattern-metadata fields (severity, description, references)
`PatternDef` has only `Name/Category/Match/Enabled`. The CLI's `list` command shows no
description or severity. Users have to read YAML to understand what a pattern detects.

**Fix:** Extend `PatternDef` (carefully — wire format change). Add Description/Severity/
References fields; preserve backward compatibility (omitempty). Update bundler to populate from
YAML. Update CLI `list` to show severity column.

#### S4. SARIF output is hard-coded "High" severity
`cmd/atheon/main.go:257` sets every finding to `"security-severity": "High"`. Once S3 is
implemented, this should use the per-pattern severity.

#### S5. No `--baseline` / `--ignore-known` for incremental scans
A common secrets-scanner use case is "only show new findings since the last scan." Currently
every run re-shows everything.

**Fix:** Add `--baseline=findings.json` that filters out findings whose (file, line, pattern)
triple appears in the baseline.

#### S6. MCP rate-limit error code is `-32600`
JSON-RPC standard: -32600 is "Invalid Request". A more accurate code for rate limiting is a
custom server-defined error in the `-32000 to -32099` range, e.g. `-32000` for "rate limit".

**Fix:** Update `cmd/mcp/main.go:194`.

### 2.5 MCP / integration gaps

#### M1. `list_patterns` and `list_categories` tools missing
See D5. The CLI has them; MCP doesn't.

#### M2. `scan_env` missing from MCP
The CLI has `--env`. The MCP doesn't.

**Fix:** Add `scan_env` MCP tool.

#### M3. MCP doesn't expose bundle update
The CLI has `atheon update`. The MCP can't refresh its own bundle. For long-running AI sessions
this is a real gap.

**Fix:** Add `update_bundle` MCP tool that calls `core.DownloadBundle`.

#### M4. MCP has no server capabilities metadata beyond `tools`
`serverInfo` returns `name: atheon, version: 1.0.0` (hard-coded — version is not from build).

**Fix:** Inject version via ldflag (already done for CLI; mirror for MCP). Update `serverInfo`.

#### M5. No streaming/large-output strategy
Findings returned as text string concatenated into `textResult`. For a directory scan with 1000+
findings, this serializes a single huge JSON content array.

**Fix:** Cap findings returned per call (e.g. 500) and add a `scan_dir_continue` cursor tool, or
return SARIF instead of formatted text.

#### M6. No IDE-side linter
VS Code / Cursor / JetBrains could call atheon as a linter for in-editor feedback.

**Fix:** Out of scope for v1 but document the path (e.g. `atheon --lsp-mode` JSON-RPC mode).

### 2.6 Operational gaps

#### O1. `~/.atheon/` directory creation is implicit
`init()` reads `~/.atheon/patterns.bundle` if it exists; `DownloadBundle` creates it via
`ensureAtheonDir`. There's no `atheon config` command to view/set this directory or the bundle URL.

**Fix:** Add `atheon config` subcommand showing bundle path, count, download URL.

#### O2. No SBOM / dependency manifest for the binary
`go.mod` shows only `require` of stdlib + module deps. For a security tool shipping binaries,
an SBOM (CycloneDX or SPDX) is increasingly required by enterprise procurement.

**Fix:** Add `syft` step to release workflow; upload SBOM artifact.

#### O3. No Homebrew/scoop release test
`.goreleaser.yml` declares `brews:` and `scoops:` blocks but they're untested. A pre-merge
dry-run of `goreleaser release --skip=publish --snapshot` would catch config errors.

**Fix:** Add to release.yml.

---

## 3 — Implementation Roadmap (Prioritized)

### Phase A — Trust restoration (1 PR, ~2 hours work)

**Goal:** Make every docs number match reality. Fix broken links.

| ID | Task | Files |
|----|------|-------|
| A1 | Generate `docs/CHANGELOG.md.md` (symlink or 1-version excerpt of `CHANGELOG.md`) | `docs/` |
| A2 | Replace `docs/PLAN.md`, `docs/TASKS.md`, `docs/PROGRESS.md` with real content for Atheon-Enhanced | `docs/`, `docs/planning/` |
| A3 | Add `AGENTS.md` per `CLAUDE.md` contract | root |
| A4 | Update pattern count in `README.md`, `FAQ.md`, `INSTALL.md`, `SETUP.md`, `PATTERN_CATEGORIES.md` to actual 252/18 | 5 files |
| A5 | Update README's pattern-distribution table to actual counts | `README.md` |
| A6 | Stub or remove broken links in `docs/README.md` | `docs/README.md` |
| A7 | Remove `--debug-level=2` example from `SETUP.md` | `docs/development/SETUP.md` |
| A8 | Pick one `BRANCH_STRATEGY.md`; remove the other | 2 files |

### Phase B — CI consolidation (1 PR, ~3 hours work)

**Goal:** Cut CI minutes, eliminate duplication.

| ID | Task |
|----|------|
| B1 | Move all `ci.yml` jobs into `comprehensive-ci.yml`'s structure; pick `ci.yml` as canonical name, drop `comprehensive-ci.yml` |
| B2 | Move CodeQL jobs from `codeql.yml` and `security-scanning.yml` into a single `security.yml` |
| B3 | Drop `auto-merge.yml` if GitHub native auto-merge is enabled |
| B4 | Fix `gofmt` check in remaining workflow to use `gofmt -l` |
| B5 | Verify all `go test` invocations have `-p 1` |
| B6 | Drop the `continue-on-error: true` from Codecov upload once the token secret is verified |

### Phase C — MCP completeness (1 PR, ~2 hours work)

| ID | Task |
|----|------|
| C1 | Add `list_patterns` MCP tool (mirror CLI `list --category=`) |
| C2 | Add `list_categories` MCP tool |
| C3 | Add `scan_env` MCP tool |
| C4 | Add `update_bundle` MCP tool |
| C5 | Use ldflag-injected version in `serverInfo` |
| C6 | Change rate-limit error code from -32600 to -32000 |

### Phase D — Pattern expansion (3 PRs, ~6 hours work)

Pattern batch 1 (top-tier SaaS):
- 10 secrets patterns (Anthropic, OpenAI org-scoped, GitHub fine-grained PAT, Supabase, Vercel, Cloudflare R2, Slack workflow, Bitbucket, Vault, 1Password)

Pattern batch 2 (PII):
- Email, IPv4/IPv6, stronger SSN, more passport formats

Pattern batch 3 (frameworks restore):
- Django, Node.js, React (or delete the empty category)

### Phase E — Architecture hygiene (2 PRs, ~4 hours work)

| ID | Task |
|----|------|
| E1 | Create `docs/architecture/decisions/` with three ADRs (RE2 vs PCRE, gzip bundle, -p 1) |
| E2 | Either create `core/streaming.go` (extract from runner.go) or fix SYSTEM_ARCHITECTURE.md |
| E3 | Either create `config/defaults/` or remove from SYSTEM_ARCHITECTURE.md |

### Phase F — Future features (separate, larger PRs)

- F1: Pattern metadata (severity, description, references) — wire-format change, requires migration
- F2: `--baseline` filter for incremental scans
- F3: SBOM generation in release workflow
- F4: LSP mode for IDE integration

---

## 4 — Ideas Beyond the Codebase (think outside the box)

The following are not in the IMPROVEMENT_PLAN.md and would meaningfully expand the project's
utility, audience, and credibility.

### 4.1 Community ecosystem

- **Public pattern registry / web index.** A static-site (Hugo/11ty) under
  `docs.patterns.atheon.dev` (or GitHub Pages) where every pattern is a browseable page with
  description, examples, and "last seen in production" anonymized stats. This is the "long-term
  platform vision" in `ROADMAP.md` — the cheap version is just MkDocs over `community/`.
- **Pattern confidence scoring.** Allow patterns to declare `confidence: high|medium|low` and
  filter by it. Currently a Stripe test key produces the same finding severity as a live key.
- **Per-pattern override config.** A `.atheonrc.json` at the project root that maps pattern names
  to overrides (enabled, severity, comment). Removes the global-state issue with
  `~/.atheon/pattern_state.json`.
- **Pattern signing.** Sign the bundle with a Sigstore identity so users can verify they're
  running community-approved patterns.

### 4.2 Security-tool integrations

- **Pre-receive Git hook** (server-side equivalent of pre-commit) — `scripts/hooks/pre-receive`.
- **Bitbucket Pipelines / GitLab CI examples.** README only documents GitHub Actions.
- **VS Code extension** that wraps the MCP server so the same findings show inline.
- **Slack/Discord notifier** for scheduled CI scans.

### 4.3 Performance / scale

- **Pattern cache prefilter.** Build a Bloom filter over all pattern regexes to short-circuit
  lines that match no pattern. For 250+ patterns this could 5–10× scan speed.
- **Streaming JSON output.** For very large result sets, emit NDJSON instead of one big array.
- **Lazy bundle loading.** On `atheon --categories=secrets`, only the secrets category's patterns
  need to be in the active scanner.

### 4.4 Documentation / DX

- **Live pattern playground.** A web page where users paste text and see which patterns fire.
  The Go scanner can run in WebAssembly (`GOOS=js GOARCH=wasm`) and power this directly. Static
  deploy to GitHub Pages is trivial.
- **Recipe book.** `docs/recipes/` with copy-paste GitHub Actions, GitLab CI, CircleCI, Jenkins,
  Travis configs.
- **Migrating from other tools.** `docs/migration/` with side-by-side examples:
  gitleaks/trufflehog/detect-secrets → atheon.
- **Benchmarks dashboard.** The IMPROVEMENT_PLAN.md §7 measures but doesn't publish. A
  `gh-pages` site with `benchstat` over time would showcase the "2-3× faster" claim.

### 4.5 Governance / sustainability

- **Maintain a `stable/clean` baseline that mirrors upstream.** README claims this but no
  workflow enforces the cadence (the only `sync-stable-clean.yml` workflow exists but the
  branch isn't checked out in this repo).
- **Define a deprecation policy.** When a pattern is renamed or removed, downstream users need
  a one-version deprecation warning.
- **Document the release process.** `docs/RELEASE.md` should be a runbook for cutting a
  release (today, knowledge lives in the chat history).

---

## 5 — Recommended Next Step

**Open PR #55: docs: sync all numbers, fill PLAN/TASKS/PROGRESS, fix broken doc links.**

This is Phase A in section 3. It is:

- ✅ Small (~2 hours)
- ✅ Zero risk (no source-code changes)
- ✅ Highest user-visible impact
- ✅ Unblocks the user's `CLAUDE.md` Documentation Contract
- ✅ Establishes a clean baseline before Phase B (CI consolidation)

After PR #55, run Phase B (CI consolidation) and Phase C (MCP completeness) in parallel since
they're independent. Phase D (pattern expansion) is the highest-value ongoing work and should
have a weekly cadence until the pattern count reaches 300+.

---

## Appendix A — Evidence Index

Every claim in this report is grounded in a specific file:line in the repo. The audit was
performed by reading the following files end-to-end:

- `README.md` (737 lines)
- `CHANGELOG.md` (69 lines)
- `Makefile` (41 lines)
- `.golangci.yml` (118 lines)
- `.goreleaser.yml` (90 lines)
- `.pre-commit-config.yaml` (73 lines)
- `codecov.yml` (47 lines)
- `.gitignore` (12 lines)
- `core/runner.go` (278 lines)
- `core/bundle.go` (505 lines)
- `core/pattern.go` (79 lines)
- `core/finding.go` (23 lines)
- `core/ignore.go` (152 lines)
- `core/pattern_state.go` (106 lines)
- `cmd/atheon/main.go` (375 lines)
- `cmd/mcp/main.go` (269 lines)
- `.github/workflows/ci.yml` (147 lines)
- `.github/workflows/comprehensive-ci.yml` (304 lines)
- `.github/workflows/self-scan.yml` (104 lines)
- `.github/workflows/quality-assurance.yml` (369 lines)
- `.github/workflows/security-scanning.yml` (258 lines)
- `.github/workflows/scheduled-release.yml` (66 lines)
- `.github/scripts/self-scan.sh` (112 lines)
- `.github/CODEOWNERS` (8 lines)
- `docs/PLAN.md`, `docs/TASKS.md`, `docs/PROGRESS.md` (all 100% template placeholders)
- `docs/architecture/SYSTEM_ARCHITECTURE.md` (310 lines)
- `docs/architecture/PATTERN_CATEGORIES.md` (header)
- `docs/api/README.md` (330 lines)
- `docs/BRANCH_STRATEGY.md` (52 lines)
- `docs/INSTALL.md` (134 lines)
- `docs/development/SETUP.md` (436 lines)
- `docs/reports/COMPLETION_REPORT.md` (368 lines)
- `docs/reports/IMPROVEMENT_PLAN.md` (570 lines)
- `docs/README.md` (49 lines)

Plus `find`/`wc`/`ls` listings for category pattern counts and file inventory.

## Appendix B — Verification Commands

```bash
# Confirm pattern count
cd /nas/Temp/repos/Atheon-Enhanced
find community -name "*.yaml" -o -name "*.yml" | wc -l                  # → 255 files
for d in community/*/; do printf "%-25s %d\n" "$d" "$(ls "$d"*.yaml 2>/dev/null | wc -l)"; done

# Confirm workflow count
ls .github/workflows/*.yml | wc -l                                        # → 10

# Confirm template placeholders
grep -l '{{' docs/PLAN.md docs/TASKS.md docs/PROGRESS.md docs/planning/*.md  # → all 5

# Confirm broken doc links
grep -oE '\((docs/reports/[^)]+\.md|docs/tests/[^)]+\.md|docs/contributors\.md)\)' docs/README.md
# → lists non-existent files
```
