# Project Plan — Atheon-Enhanced

**Version**: 0.7.0+ (post-Wave 11)
**Last Updated**: 2026-06-27
**Status**: Active — Wave 11 complete; yaml.v3 migration done

---

## Problem & Audience

**Problem.** Repositories leak secrets, PII, and security-sensitive code patterns. Off-the-shelf scanners (TruffleHog, GitLeaks, Semgrep) each cover a slice. We want a single fast scanner that handles **all three** (secrets + PII + code-quality) with a **transparent, community-editable pattern catalog** and an **MCP server** so IDEs can query it.

**Audience.** Open-source maintainers, security-conscious teams, AI-coding-assistant users who want inline pattern feedback inside their editor.

**Differentiators.**
- 274 patterns across 19 categories (secrets, PII, web-security, compliance, security-hardening, cloud-native, devops, healthcare, finance, ai-detection, code-quality, performance, accessibility, web-development, data-visualization, api-integration, git-hygiene, pwa, frameworks).
- Pattern catalog is plain YAML in `community/` — anyone can add a pattern with a PR. No proprietary DSL.
- Per-pattern severity (low/medium/high/critical) flows end-to-end into SARIF for IDE integration.
- MCP server (`atheon-mcp`) exposes the same scanner to Claude/Cursor/etc. via JSON-RPC.

## Architecture

```
                ┌──────────────────────┐
                │ community/*.yaml     │  274 patterns, declarative
                │ (severity, regex)    │
                └──────────┬───────────┘
                           │  go run ./bundler
                           ▼
                ┌──────────────────────┐
                │ core/patterns.bundle  │  gzipped JSON, embedded
                └──────────┬───────────┘
                           │  go:embed
                           ▼
   ┌──────────────┐   ┌──────────┐   ┌──────────────┐
   │ cmd/atheon  │   │ core/   │   │ cmd/mcp     │
   │ (CLI scan)  ├──▶│ scanner │◀──┤ (JSON-RPC)  │
   └──────┬───────┘   └──────┬───┘   └──────────────┘
          │                  │
          ▼                  ▼
       text/JSON/SARIF    Finding{Severity,File,Line,Column,Content}
```

- **Language**: Go (RE2 regex via stdlib `regexp`, guaranteed linear time).
- **Embed**: `//go:embed patterns.bundle` so the binary is self-contained.
- **Concurrency**: parallel per-file scan via goroutines; package-level `init()` loads the bundle exactly once.

## Development Approach

**Methodology**: wave-based hardening. Each wave is one merged PR cluster, scoped by a fresh gap-analysis subagent. Waves do not have fixed sprint lengths — they end when the gap list is exhausted or risk outweighs benefit.

**Iteration cycle per wave**:
1. **Plan** (subagent or direct): enumerate gaps with risk × effort ranking.
2. **Confirm** with user via AskUserQuestion (scope, defer choices).
3. **Implement** on a `feature/wave-N-*` branch off `main`.
4. **Verify**: `go test ./...`, `gofmt -l .`, manual smoke.
5. **Open PR** via REST API (gh CLI's GraphQL is flaky on this repo).
6. **Resolve** any CodeRabbit threads before merge.
7. **Squash-merge** via REST.
8. **Cleanup**: `git branch -d`, update `MEMORY.md` with wave summary.

**Quality gates**:
- Code coverage: ≥70% on touched packages (enforced in `scripts/hooks/pre-commit`).
- `gofmt -l .` must be empty (enforced in CI Lint job).
- All commits end with `Co-Authored-By: Claude <noreply@anthropic.com>`.
- No `git stash`, `git reset --hard`, `git push --force`, `git rebase -i` — per standing repo rules.

## Hardening Waves

| Wave | PR | Theme | Status |
|------|-----|-------|--------|
| 1 | #74-76, #79-80 | Initial scaffold, rename cleanup, gap-analysis docs | Merged |
| 2 | #81 | CI/security: dependabot groups, govulncheck pin, PR template | Merged |
| 3 | #83 | Fuzz tests, detection coverage, -trimpath + SPDX SBOM | Merged |
| 4 | #84 | Pattern severity wired end-to-end (Pattern → Finding → SARIF) | Merged |
| 5 | #85 | Stats.Errors surfacing, list --category validation, hot-path benchmarks, dead-script removal | Merged |
| 6 | #86, #87, #88 | Legacy-flip log, --json --version, MCP JSON-RPC roundtrip, docs fill | Merged |
| 7 | #89 | pattern_state mutex + concurrent pattern state test harness | Merged |
| 8 | #92-98 | Detection fixtures, -race CI gate, shotgun regexes, symlink guard, SARIF completeness, MCP hardening | Merged |
| 9 | TBD | MCP protocol completion, SARIF ecosystem, community patterns, bundle integrity | In progress |

## Wave 9 — Gap Analysis (2026-06-26)

Three parallel Explore agents ran post-Wave 8: **MCP + DX**, **SARIF + Ecosystem**, **Security**. Combined findings below, ranked by risk × effort.

---

### PR #99: `feat(mcp): progress notifications, cancel handler, error sanitization, stale-bundle detection`

**Why**: MCP server is the primary interface for AI agents. Four gaps reduce its robustness as a long-lived daemon.

**Changes**:

1. **`cmd/mcp/main.go`** — Add `$/cancelRequest` handler:
   - Store active request IDs in a `sync.Map` keyed by request ID
   - `handleCall` checks `cancelMap.Load(req.ID)` at scan entry; if cancelled, return `-32802` (request was cancelled)
   - A `notifications/cancel` notification with matching `id` removes the entry from the map

2. **`cmd/mcp/main.go`** — Progress notifications during `handleUpdateBundle`:
   - Wrap `core.DownloadBundle` with a progress ticker that sends
     `{"jsonrpc":"2.0","method":"notifications/progress","params":{"token":"bundle-download","progress":N,"total":T}}`
     every 512 KiB received
   - If download completes, send final progress with `complete: true`

3. **`cmd/mcp/main.go`** — Sanitize all JSON-RPC error messages:
   - Four locations pass bare `err.Error()` to JSON-RPC responses (`handleScanFile`, `handleScanDir`, `handleUpdateBundle`, `invalidParams`)
   - A helper `safeError(err error) string` maps `os.IsNotExist` → `"file not found"`, `os.IsPermission` → `"permission denied"`, and wraps everything else as `"internal error: <category>"` (no raw paths, no raw syscall strings)
   - `handleCall`'s unknown-tool case already uses `unknown tool: " + p.Name` — safe

4. **`core/bundle.go`** — ETag-based stale-bundle detection:
   - `DownloadBundle` issues a HEAD request first to get the upstream ETag/Last-Modified
   - If the local `~/.atheon/pattern_state.json` records the same ETag and the file is recent (<24h), skip the full download
   - `update_bundle` MCP tool gains a `force: bool` parameter to bypass the cache
   - Add `bundle.etag` and `bundle.lastChecked` to the state file

**Files modified**: `cmd/mcp/main.go`, `core/bundle.go`, `core/pattern_state.go`
**New tests**: `cmd/mcp/mcp_cancel_test.go` (cancel handler), `cmd/mcp/mcp_progress_test.go` (progress notifications), `cmd/mcp/mcp_error_sanitization_test.go` (error message sanitization), `core/bundle_etag_test.go` (stale-bundle detection)

---

### PR #100: `fix(sarif): rules[].relationships + output parity + community pattern triage`

**Why**: SARIF output is functionally incomplete for non-GitHub consumers; several code-quality patterns fire on every codebase regardless of intent.

**Changes**:

1. **`cmd/atheon/main.go`** — `buildSARIFRules` add `relationships` to each rule:
   - Map pattern categories to CWE IDs. Secrets/PIE patterns → CWE-798 (Hardcoded Password), CWE-259; web-security → CWE-601 (Open Redirect), CWE-79 (XSS); code-quality patterns → CWE-483 (Dangling Temporary File) etc.
   - Emit `"relationships": [{"target": {"id": "CWE-798", "toolComponent": {"name": "CWE"}}}]` per rule
   - Only emit where a clear mapping exists; patterns without a CWE counterpart get no relationships entry

2. **`cmd/atheon/main.go`** — `printJSONFindings` parity with SARIF:
   - Add `severity` field (from `f.Severity`)
   - Add `column` field (from `f.Column`, 0 = omitted)
   - Add `fingerprint` field (same `pattern|file|line|column` key used by SARIF `partialFingerprints.atheonLoc`)
   - Add `category` field (from `core.All()` lookup by pattern name)

3. **`community/code-quality/*.yaml`** — Scope unscoped patterns:
   - `dummy-function.yaml`, `mock-stub.yaml`, `fake-data.yaml`: add word-boundary anchors and limit to test-like file paths (`_test.go`, `test_*.py`, `*_test.go`) where feasible in regex
   - `sleep-in-test.yaml`: rename to `sleep-in-test-unscoped` and add `//pytest:` or `//go:` comment context prefix — OR narrow to files named `*_test.go` / `test_*.py` using an additional context regex
   - `skip-tests.yaml`: require `skip` or `Skip` as a separate word before `test` (anchor the `skip` word boundary, remove the overly-broad `mvn.*skip.*test` branch)
   - `todo-comment.yaml`, `fixme-comment.yaml`: add severity=`info` (not `medium`) and add a note that these fire on generated files; consider narrowing to common source file extensions

4. **`cmd/atheon/main.go`** — Fix `helpUri`:
   - The `wiki/patterns#<name>` URL does not exist; remove the broken `helpUri` field from rules rather than creating dead links in the GitHub Security tab UI

**Files modified**: `cmd/atheon/main.go`, `community/code-quality/{dummy-function,mock-stub,fake-data,sleep-in-test,skip-tests,todo-comment,fixme-comment}.yaml`
**New tests**: `cmd/atheon/main_json_output_test.go` (JSON output field parity), `cmd/atheon/main_sarif_relationships_test.go` (CWE relationships present in SARIF)

---

### PR #101: `fix(core): bundle hash verification + rate-limiter hardening`

**Why**: Bundle supply chain has no integrity check; rate limiter is bypassable via connection cycling.

**Changes**:

1. **`core/bundle.go`** — SHA-256 verification for downloaded bundles:
   - Publish `checksums.txt` alongside each GitHub release: `SHA256 <bundle.gz>` computed at release time by the bundler
   - `DownloadBundle` fetches `checksums.txt` first, then verifies the bundle bytes against the expected hash before writing to disk
   - On mismatch: return `ErrBundleIntegrity` and log a warning; do NOT load the mismatched bundle
   - The embedded bundle (built at compile time) is not hash-verified at runtime — document this as a build-time integrity assumption

2. **`cmd/mcp/main.go`** — Per-connection rate limit window:
   - Instead of a single global token bucket, track requests per connection using a `sync.Map` of `*rateLimiter` instances keyed by the connection's stdin file descriptor
   - OR simpler: add a global concurrent-request cap (`maxConcurrent = 50`) using an `atomic.Int` counter — rejects new requests when the server is saturated
   - The per-connection limiter is a defense-in-depth measure; the concurrent cap prevents memory exhaustion under flood

3. **`core/runner.go`** — Binary file detection hardening:
   - The NUL-byte sniff misses UTF-16LE/UTF-16BE (common in Windows configs), uncompressed BMP, and base64-encoded blobs
   - Add an extension-based secondary heuristic: `.log`, `.cfg`, `.conf`, `.ini` files over 1 MiB are treated as likely-binary even if no NUL byte is found (they're often binary-formatted logs)
   - Add `bytes.IndexByte` check for UTF-16 surrogate pairs (`\xff\xfe` or `\xfe\xff` BOM) in the first 8 KiB as a UTF-16 detector

**Files modified**: `core/bundle.go`, `core/runner.go`, `cmd/mcp/main.go`
**New tests**: `core/bundle_hash_test.go` (hash verification), `core/binary_sniff_test.go` (UTF-16/extension heuristics), `cmd/mcp/mcp_concurrency_test.go` (concurrent request cap)

---

### PR #102: `chore: help text, Go 1.25 prep, yaml.v3 deprecation`

**Why**: Low-risk cleanup items that improve DX and prepare for the next Go version.

**Changes**:

1. **`cmd/atheon/main.go`** — Document all flags in `--help`:
   - Add `--all` (enable all patterns including disabled ones) to the flag listing
   - Add `--no-follow-symlinks` as a named flag (currently only in usage line)
   - Clarify flag ordering for `--json`/`--sarif` vs positional args

2. **Go 1.25 compatibility**: Go 1.25 is expected mid-2026. No code changes expected, but:
   - Verify `go.mod` `go 1.21` directive still works on Go 1.25 (it should)
   - Add Go 1.25 to the CI matrix (4 Go versions: 1.21, 1.22, 1.23, 1.24, 1.25)
   - Run `go vet ./...` and `golangci-lint` with the latest version to catch any new lints

3. **`gopkg.in/yaml.v3` deprecation**:
   - ~~The `yaml.v3` library is unmaintained~~ — **DONE in Wave 11 (PR #111)**. Migrated to `github.com/goccy/go-yaml`.
   - No further action needed.

**Files modified**: `cmd/atheon/main.go` (help text), `.github/workflows/ci.yml` (Go 1.25 matrix), `go.mod` (comment)

---

## Risks (Post-Wave 8)

| Risk | Impact | Mitigation |
|------|--------|------------|
| Pattern regexes accidentally match common code | High (false positives) | `TestFalsePositiveGuard` covers known clean snippets; PR review required for `community/*.yaml` |
| Bundle corruption on bad YAML | Medium (silent drop) | Bundler now logs and skips; pre-commit hook surfaces stderr |
| SARIF severity mismatch consumer expectations | Medium | CVSS-like scores (9.5/7.5/5.0/2.5) + level (error/warning/note) per spec |
| `core/` package-level state races between CLI + MCP | Medium | Documented; Wave 7 added mutex around writes |
| Coverage drop slips past CI | Low | Codecov status check |
| Bundle download has no integrity check | Medium-High | Wave 9 PR #101 adds SHA-256 verification |
| MCP error messages leak filesystem paths to agents | High | Wave 9 PR #99 sanitizes all JSON-RPC errors |
| Code-quality patterns fire on non-test files | High | Wave 9 PR #100 scopes unscoped patterns |

## Success Metrics

**Adoption** (proxy): bundle download count from GitHub releases (visible in Insights).
**Quality**:
- Pattern false-positive rate (target: <5% on `TestFalsePositiveGuard` corpus)
- Time-to-fix for a known-bad regex (target: <1 wave)
**Reliability**: zero `panic` in CI across all tested Go versions.
**Security**: zero filesystem path leaks in MCP JSON-RPC error messages.

## Timeline

No fixed timeline. Each wave lands when ready. Velocity is gated by review throughput, not calendar.

## Dependencies

**External**:
- Go 1.21+ (1.22, 1.23, 1.24, 1.25 also tested in CI matrix)
- `github.com/goccy/go-yaml` — migrated from `gopkg.in/yaml.v3` in Wave 11 (PR #111)
- GitHub Actions runners (ubuntu, windows, macos)

**Internal**:
- `core/` is the only package with embedded data; everything else imports it.
- `bundler/` is a separate `main` package that produces the embedded bundle.

## Open Questions (Post-Wave 8)

- Should `update_bundle` in the MCP server require confirmation or a `force` parameter? (Currently unconditional download.)
- Should the bundle carry a schema version (`version: 2`) so consumers can detect when the wire format changes?
