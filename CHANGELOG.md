# Changelog - Atheon Enhanced

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `cmd/mcp/main.go` `$/cancelRequest` notification handler: stores
  active request IDs in a `sync.Map` and returns `-32802` (Request
  Cancelled) when a subsequent `tools/call` arrives with a matching
  ID. This lets AI agents cancel long-running scans without killing
  the MCP server.
- `cmd/mcp/main.go` `safeError(err)` helper: maps `os.IsNotExist` →
  `"file not found"`, `os.IsPermission` → `"permission denied"`,
  everything else → `"internal error"`. Replaces bare `err.Error()`
  calls in `handleScanFile`, `handleScanDir`, and `handleUpdateBundle`,
  preventing the MCP server from leaking raw filesystem paths and
  syscall strings to AI agents that parse error messages (HIGH risk
  finding from Wave 9 audit).
- `core/bundle.go` ETag-based stale-bundle detection: `DownloadBundle`
  now accepts a `force` parameter. When `force=false` (the default),
  a HEAD-less 24-hour freshness window skips the network round-trip
  if the upstream bundle hasn't changed. The `update_bundle` MCP
  tool gains a `force: bool` parameter to bypass the cache.
- `core/pattern_state.go` `PatternState` gains `BundleETag` and
  `BundleLastChecked` fields for stale-bundle detection.
- `core/bundle.go` `fetchBundleData` now returns the `ETag` header
  alongside the bundle bytes; `recordBundleETag` persists it to the
  state file on a successful download.
- `cmd/mcp/run_pr97_test.go` `TestCancelRequestHandler`,
  `TestCancelRequestNoID`, `TestCancelRequestUnknownMethod`: cancel
  protocol tests.
- `cmd/mcp/mcp_error_sanitization_test.go`: `TestSafeError{NotExist,
  Permission,Generic,Nil}` verify the `safeError` contract.
- `core/bundle_etag_test.go`: `TestShouldSkipDownload{Fresh,Stale,
  NoState,EmptyETag}`, `TestRecordBundleETag`, `TestLoadBundleETag`
  cover the ETag tracking and stale-bundle logic.
- `core.ScanOpts` struct with `NoFollowSymlinks` and `MaxFileSize`
  fields; `ScanDir` now takes the opts as a third argument. The CLI
  defaults to follow-symlinks and the package-level `maxFileSize`
  (preserves historical behaviour); the MCP server defaults to
  no-follow-symlinks so agents scanning untrusted trees can't escape
  via planted symlinks (e.g. `repo/leak -> /etc/passwd`).
- `--no-follow-symlinks` CLI flag for `atheon <dir>` directory scans.
  Off by default (the historical behaviour); explicit opt-in for
  environments where the operator wants the same safe default the MCP
  server uses.
- `core.readFileCapped(path, maxBytes)` helper, factored out of
  `ScanFile` so ScanDir workers can share the size guard. Without it,
  a single 10 GiB file could OOM the scanner on `atheon ./big-tree`
  because the per-file goroutines bypassed the cap entirely.
- `core.ErrFileTooLarge` sentinel: callers can `errors.Is(err,
  ErrFileTooLarge)` to distinguish size-skips (benign) from
  permission errors (operator-actionable) in `Stats.Errors`.
- NUL-byte content sniff (8 KiB head) in ScanDir workers. The
  extension allowlist misses extensionless binaries; the sniff catches
  them before any per-pattern matching runs.
- Walk-error capture in ScanDir: previously dropped via
  `//nolint:nilerr` with no surface to the caller. Now collected into
  `Stats.Errors` so the CLI can exit non-zero and the MCP server can
  return a useful count.
- `slog.Warn("combined regex compile failed; falling back to
  per-pattern matching", ...)` in `rebuildActiveScanners`. The
  historical silent drop meant a broken pattern lingered unnoticed —
  matches just stopped appearing for its category.
- Skip-dirs expanded to include `.idea`, `.vscode`, `.svn`, `.hg`,
  `.cache` (IDE and VCS metadata noise that bloats scans without
  yielding findings).
- `core.TestReadFileCapped{UnderCap,Boundary,OverCap,ZeroBytes,Unreadable}`
  and `core.TestScanDirSizeCapSurfacesError` pin the new size-cap
  contract.
- `core.TestScanDir{FollowsSymlinksByDefault,NoFollowSymlinks,
  NoFollowSymlinksLoop,NoFollowSymlinksDangling}` pin the symlink
  guard: default follows, opt-in doesn't (including the dangling-link
  and loop cases).
- `core.TestCombinedRegexCompileErrorLogged` pins the new
  slog.Warn-on-compile-failure behaviour using a per-test
  `slog.Handler` capture (no global slog mutation).
- `cmd/mcp/main.go` panic recovery: `dispatchRequest` wraps the
  per-request switch in `defer recover()`, returning `-32603`
  (internal error) on panic. Pre-PR-#97 a single panicking tool
  handler tore down the whole server and forced every connected
  client to reconnect; the panic response is now a recoverable
  per-request error.
- `cmd/mcp/main.go` 64 MiB global request cap (`mcpMaxRequestBytes`)
  via `bufio.Scanner` (Buffer: 64 KiB initial, 64 MiB max). The
  prior 1 MiB cap silently errored past the buffer size with no
  hard memory ceiling; the new cap lets legitimate large scan
  requests through while bounding a single malicious client.
- `cmd/mcp/main.go` 30-second per-request timeout
  (`mcpRequestTimeout`). Each request runs in a child ctx derived
  via `context.WithTimeout`, so a stuck handler can't wedge the
  server indefinitely. Cancelled parent ctx (e.g. SIGTERM via
  `signal.NotifyContext`) still terminates the loop because the
  child inherits parent cancellation.
- `cmd/mcp/main.go` 32 MiB content cap for `scan_string` and
  100-entry cap for `scan_env.categories`. Both surface as
  `-32602 invalid params` with a message naming the cap.
- `cmd/mcp/main.go` JSON-RPC version validation: requests with
  `jsonrpc != "2.0"` now return `-32600 Invalid Request`. Pre-PR-#97
  they fell through to the unknown-method branch and returned the
  misleading `-32601 method not found`.
- `cmd/mcp/main.go` notification handling: requests with `id == nil`
  no longer emit a response (the loop `continue`s after dispatch).
  Pre-PR-#97 the loop wrote a response with a null ID, confusing
  well-behaved clients into treating it as a reply to a later
  request.
- `cmd/mcp/main.go` encode-error handling: `enc.Encode(resp)`
  errors are now checked; on failure the loop logs and exits 1
  instead of silently retrying against a vanished peer.
- `cmd/mcp/run_pr97_test.go` with 11 new tests:
  `TestMCPPanicRecovered`, `TestMCPDispatchRecoverIsolated`,
  `TestMCPRateLimitAppliesToInitialize`,
  `TestMCPLargeRequestRejected`, `TestMCPScanStringSizeCap`,
  `TestMCPScanEnvCategoriesCap`, `TestMCPInvalidJSONRPCVersion`,
  `TestMCPNotificationNoResponse`, `TestMCPPerRequestTimeout`.
- `cmd/mcp/mcp_internal_test.go` `TestHandleCallDoesNotRateLimit`
  documents the rate-limit-at-run-level contract (replaces the
  pre-PR-#97 `TestHandleCallRateLimited` which checked the old
  handleCall-level rate limit that was removed).

### Changed
- `core.ScanDir` signature: now `(ctx, root, opts ScanOpts)` instead of
  `(ctx, root)`. All in-tree callers updated.
- `cmd/mcp` rate-limit location: moved from `handleCall` (which only
  covered `tools/call`) to the top of `run()` so `initialize`,
  `tools/list`, and notifications all share the same token bucket.
  Pre-PR-#97 an attacker could DoS the server by spamming
  `initialize`, which never consumed a token.
- `cmd/mcp` JSON-RPC read loop: replaced the prior `bufio.Scanner`
  with 1 MiB buffer and the proposed `json.NewDecoder(io.LimitReader)`
  with `bufio.Scanner` at 64 MiB max (`bufio.ErrTooLong` surfaces as
  `slog.Error`). The `json.Decoder` path was tried first but its
  internal buffer leaks across `Decode()` calls — a single 70 MiB
  truncated payload generated 25+ parse errors before EOF, and the
  Decoder never returned `io.EOF` cleanly on a buffered parse failure.
  The new Scanner path gives well-defined line-by-line behavior.
- `docs/RELEASE.md` (maintainer runbook): tag format (`v0.YY.MM.DD[-rcN]`),
  pre-release checklist, GoReleaser publishing, hotfix workflow, ldflags,
  bundle regeneration, troubleshooting.
- `docs/PLAN.md` filled with project-specific content reflecting the
  multi-wave hardening cycle (274 patterns, 19 categories, MCP server,
  wave-by-wave hardening). Replaces the prior `{{...}}` template.
- `docs/TASKS.md` filled with the actual task ledger (waves 1–6 marked
  completed, deferred items including the `pattern_state` mutex work).
- `docs/BRANCH_STRATEGY.md` consolidated with the richer Quick Start,
  Decision Tree, Branch Comparison table, and per-branch configuration
  sections from the previous `docs/reports/BRANCH_STRATEGY.md`.

### Removed
- `docs/reports/BRANCH_STRATEGY.md` (duplicate of the canonical
  `docs/BRANCH_STRATEGY.md`). Its unique content was merged into the
  canonical copy and the duplicate was deleted to satisfy the project's
  "no duplicate implementations" rule.

## [0.6.0] - 2026-06-25

### Added
- Pattern severity wired end-to-end: each `community/*.yaml` declares
  `severity: low|medium|high|critical`. Defaults applied by category
  (secrets/pii/web-security/compliance/security-hardening = high;
  code-quality/performance/accessibility/web-development = low; all others = medium).
  Severity flows through `Pattern → bundlePattern → Finding → SARIF`.
- SARIF severity mapping: CVSS-like scores (9.5/7.5/5.0/2.5) and levels
  (error/warning/note) derived per-pattern, plus a `security-severity-label`
  property for human readability. Replaces hard-coded
  `security-severity: "High"` / `level: "error"` for every result.
- `Pattern.Severity()` interface method and `normalizeSeverity()` helper that
  coerces empty/typo'd values to `medium` so downstream code stays safe.
- Hot-path benchmarks: `BenchmarkLoadBundle` (gzip decode + 274-pattern
  compile), `BenchmarkCompileIgnoreFile` and `BenchmarkIgnoreMatcherMatch`
  (recursive regex on `.atheonignore`), and `BenchmarkRedact` (per-finding
  redact on the `--json` output path). Run with `go test -bench=. ./core/`
  and `./cmd/atheon/`.
- `scanErrorsPresent()` helper: bumps exit code when a scan silently dropped
  files (permission denied, unreadable). Closes a data-loss gap where a
  partial failure was reported as success.
- Pre-commit hook now surfaces bundler warnings (per-file skip reasons) on
  the same line as the success message, so contributors see when a pattern
  was dropped without a manual re-run.
- `atheon list --category=<bogus>` now errors with the known-category list,
  instead of silently filtering to zero matches.
- `slog.Info` line emitted when `loadBundle` flips all patterns to enabled
  (legacy compatibility path). Without this log, a contributor bundle
  that's accidentally all-`enabled: false` looks identical at runtime —
  every pattern silently comes on. Surface the path so it stays observable.
- Regression tests `TestLoadBundleLegacyDefaultFlip` and
  `TestLoadBundleNoFlipWhenAnyEnabled` in `core/bundle_legacy_default_test.go`
  guard the legacy-flip behavior and the log-on/only-on gate.
- `TestVersionFlagWithJSON` subtests for `["--version"]`,
  `["--json", "--version"]`, and `["--sarif", "--version"]` flag orderings.
- CI JSON-RPC roundtrip integration test in `.github/workflows/ci.yml`:
  replaces the prior smoke step (which only verified clean exit on empty
  stdin) with a real `initialize` + `tools/list` roundtrip and `jq`
  assertions on `protocolVersion`, `capabilities`, tool count, and tool
  names. Catches framing and discovery regressions that the smoke test
  missed.

### Changed
- Bundler (`go run ./bundler`) no longer aborts on broken pattern files.
  Malformed YAML, missing fields, whitespace in pattern names, duplicate
  names, and invalid regex are logged to stderr and the file is skipped.
  This mirrors `loadBundle`'s runtime tolerance.
- 67 community patterns had pre-existing regex corruption (severity text
  embedded in the `match:` value); these were repaired so all 274 patterns
  ship cleanly.
- `atheon --json --version` (and `--sarif --version`) now print the version
  cleanly. Previously the `--version` check ran before the `--json`/`--sarif`
  strip, so `atheon --json --version` fell into the default branch and
  errored with `path not found: --version`. Flag order is now forgiving.

### Fixed
- `community-pattern-review` workflow SIGPIPE: `git diff ... | head -10`
  exited 141 under `set -euo pipefail` when `head` closed the pipe early.
  Disabled pipefail around that pipeline only — captured files unchanged.
- `gofmt` alignment in `cmd/atheon/main.go` (the SARIF map literal had a
  misaligned key after the severity wiring change).

### Removed
- Dead scripts: `scripts/doc-validate.sh`, `scripts/doc-exemptions.sh`,
  `scripts/doc-categorize.sh`. No callers existed anywhere in the repo;
  deleting them removes a code-maintenance violation.

## [0.5.0] - 2026-06-25

### Added
- New patterns in PII category: national-id, dob-format, gender-field, health-record-id, tax-id-ein
- New patterns in Secrets category: cloudflare-token, okta-api-token, pagerduty-api-key, heroku-api-key, travis-ci-token, circleci-token, sonarqube-token, artifactory-token, firebase-api-key, vercel-token
- New patterns in Cloud-native category: aws-arn, gcp-project-id, azure-connection-string, k8s-imagepullsecret, helm-secret-value
- New patterns in Code-quality category: sleep-in-test, fmt-println-prod, panic-in-handler, direct-sql-query, global-variable, unused-import-comment
- New `compliance` category: gdpr-personal-data-comment, hipaa-phi-field, pci-cardholder-data, data-retention-comment
- New `git-hygiene` category: merge-conflict-marker, fixup-commit-message, rebase-todo-leftover, git-rerere-conflict
- `scripts/pattern-count.sh` — single source of truth for pattern counts (replaces
  hardcoded numbers scattered across docs). Supports `--json`, `--total`, `--table`,
  `--help`. Confirmed: **274 patterns / 19 categories**.
- `docs/architecture/decisions/` directory for Architecture Decision Records (ADRs)
  (planned)

### Changed
- Structured logging via `log/slog` for consistency and flexibility
- `ValidatePattern()` helper in core for reusable pattern validation
- **CI consolidation**: 10 GitHub Actions workflows → 5 (ci, security, release,
  sync, auto-merge). Removed duplicate test/lint/build, self-scan, and CodeQL
  workflows. Consolidated into a coherent set with single-responsibility jobs.
- `gofmt` check now uses the standard `gofmt -l .` idiom (replaced a non-standard
  `--debug-level` pattern from docs)
- Codecov upload no longer has `continue-on-error` (was silently masking failures)
- Scheduled release tag format changed from `0.4.YYMMDD` to `v0.YY.MM.DD` for
  consistency with manual tags
- All documented `go test ./...` invocations now include `-p 1` (the flag is
  mandatory because `core/` has package-level state in `init()` that breaks under
  parallel package execution). Updated: `.pre-commit-config.yaml`,
  `scripts/coverage.sh`, `.github/wiki/TROUBLESHOOTING.md`,
  `docs/guides/TROUBLESHOOTING.md`, `docs/PATTERN_FORMAT.md`,
  `docs/reports/REPOSITORY_RENAME_PLAN.md`, `docs/reports/BRANCH_STRATEGY.md`,
  `docs/reports/FEATURE_COMPARISON.md`, `docs/architecture/SYSTEM_ARCHITECTURE.md`,
  `docs/self-scan.md`. README CI badge updated to `ci.yml`.
- `docs/reports/BRANCH_STRATEGY.md` is now a redirect stub pointing to canonical
  `docs/BRANCH_STRATEGY.md`
- `docs/architecture/SYSTEM_ARCHITECTURE.md` "Code Organization" section rewritten
  to reflect actual file inventory (removed phantom `core/streaming.go`,
  `core/quality_enforcement.go`, `config/defaults/` entries)
- Pattern counts in 6 doc files updated to reflect actual 274 patterns / 19
  categories (was stale: 177 / 225 / 255 / 190 depending on file)

### Fixed
- Finding.Line guard ensures 1-indexed line numbers (0 becomes 1)
- README CI badge pointed at non-existent `comprehensive-ci.yml` — now points at
  the consolidated `ci.yml`

## [0.4.0] - 2026-06-22

### Added
- 223+ patterns across 19 categories
- `atheon update` command for downloading latest pattern bundle
- `atheon list --enabled/--disabled/--category=` filtering
- `atheon --json` JSON output mode
- `atheon --env` for scanning environment variables
- `atheon --stdin` for scanning piped content
- MCP server (`atheon-mcp`) for IDE integration

### Changed
- Bundle format: gzip-compressed JSON for smaller size and faster loading
- Pattern enable/disable persists across runs via `~/.atheon/pattern_state.json`

### Security
- SHA-pinned GitHub Actions
- govulncheck in CI
- JUnit test reporting in CI

---

## [0.3.0] - 2026-06-20

### Added
- `.atheonignore` file support
- Context cancellation support for all scan operations
- `atehon:ignore` inline directive

### Changed
- Improved performance with combined regex per category

---

## [0.2.0] - 2026-06-17

### Added
- Initial release with core pattern categories
- Secrets and PII detection
- Multiple output formats (text, JSON)

---

[Unreleased]: https://github.com/aliasfoxkde/Atheon-Enhanced/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/aliasfoxkde/Atheon-Enhanced/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/aliasfoxkde/Atheon-Enhanced/releases/tag/v0.4.0
[0.3.0]: https://github.com/aliasfoxkde/Atheon-Enhanced/releases/tag/v0.3.0
[0.2.0]: https://github.com/aliasfoxkde/Atheon-Enhanced/releases/tag/v0.2.0