# Future Roadmap — 2026-2027

**Date**: 2026-06-27
**Status**: Draft — For review and planning

---

## Executive Summary

Atheon-Enhanced has completed 12 waves of hardening. The codebase is stable, CI/CD is robust, and the pattern matching engine is battle-tested. This document identifies the next horizon of opportunities across five dimensions: **Intelligence**, **Ecosystem**, **Performance**, **Security**, and **Developer Experience**.

---

## 1. Intelligence Layer

### 1.1 AI-Assisted Pattern Generation

**Current State**: Patterns are YAML-defined manually.
**Future**: Use LLM to generate patterns from natural language descriptions.

```
feat: "detect JWT tokens with weak secret"
→ generates re2 regex + severity + CWE mapping
```

**Approach**:
- Add `ateon gen-pattern "description"` CLI command
- Use GitHub Models API (already in use for community-pattern-review.yml)
- Pattern wizard generates YAML, human reviews before merge

**Priority**: HIGH — reduces barrier to contribution

### 1.2 ML-Based False Positive Detection

**Current State**: Manual exclusion patterns.
**Future**: Train on historical accept/reject decisions.

**Approach**:
- Collect anonymous pattern-match decisions (opt-in telemetry)
- Train lightweight classifier on match context
- Score threshold for auto-dismiss low-confidence matches

**Priority**: MEDIUM — requires privacy review

### 1.3 Semantic Pattern Matching

**Current State**: Pure regex matching.
**Future**: AST-aware matching for code patterns.

**Approach**:
- Add optional tree-sitter based matching for Go/Python/JS
- Detect: insecure deserialization, SQL injection, XSS contexts
- Superset of regex — doesn't replace existing patterns

**Priority**: MEDIUM — significant complexity

---

## 2. Ecosystem Expansion

### 2.1 Plugin Architecture

**Current State**: Monolithic pattern engine.
**Future**: Plugin system for custom evaluators.

```go
type Plugin interface {
    Name() string
    Evaluate(ctx *Context) (*Finding, error)
}
```

**Use Cases**:
- Custom secret scanning (HashiCorp Vault, AWS secrets)
- License compliance checking
- SBOM vulnerability correlation

**Priority**: HIGH — enables enterprise adoption

### 2.2 Language Server Protocol (LSP) Integration

**Current State**: CLI only.
**Future**: IDE plugins for real-time scanning.

**Targets**:
- VS Code extension
- JetBrains plugin
- Neovim LSP client

**Benefit**: Developers see issues before committing

**Priority**: MEDIUM — significant development effort

### 2.3 Package Registry Integration

**Current State**: Local bundle only.
**Future**: Publish patterns to registry.

```
atheon patterns publish --registry github.com/aliasfoxkde/patterns
atheon patterns install @secretlint/rules-aws
```

**Priority**: LOW — existing community/*.yaml works well

---

## 3. Performance

### 3.1 Incremental Scanning

**Current State**: Full rescan on every run.
**Future**: Delta scanning with file hash tracking.

**Implementation**:
- Store file hashes in `.atheon/cache.db`
- Only rescan changed files
- Invalidate on pattern bundle update

**Priority**: HIGH — significant speedup for large repos

### 3.2 Parallel Directory Walking

**Current State**: Sequential walk.
**Future**: Worker pool with bounded concurrency.

**Target**: 10x speedup on multi-core machines

**Priority**: MEDIUM

### 3.3 WASM Compilation

**Current State**: Native Go binary.
**Future**: Compile to WASM for browser/edge execution.

**Use Cases**:
- Cloudflare Workers scanning
- Browser extension
- Cross-platform without Go runtime

**Priority**: LOW — complex, limited immediate use case

---

## 4. Security Hardening

### 4.1 Signed Pattern Bundles

**Current State**: SHA-256 hash verification.
**Future**: Sigstore/kms-signed bundles.

**Benefit**: Verify pattern bundle integrity beyond hash

**Priority**: MEDIUM

### 4.2 YARA Integration

**Current State**: RE2 regex only.
**Future**: YARA-L support for malware detection patterns.

**Priority**: LOW — specialized use case

### 4.3 Supply Chain Security

**Future**:
- SLSA provenance for releases (partially done via goreleaser)
- Sigstore cosign signing for pattern bundles
- SBOM generation with syft (blocked on CI)

**Priority**: MEDIUM

---

## 5. Developer Experience

### 5.1 Interactive TUI

**Current State**: JSON/CLI output.
**Future**: Terminal UI for scanning.

```
┌─────────────────────────────────────┐
│ 🔍 Scanning: /project              │
│ ████████████░░░░░░░ 67% (234/348) │
│                                     │
│ ⚠️  HIGH: hardcoded-aws-key        │
│    src/config/aws.go:42            │
│                                     │
│ ⚠️  MEDIUM: sql-injection          │
│    src/handlers/user.go:128         │
└─────────────────────────────────────┘
```

**Priority**: MEDIUM — low effort, high impact

### 5.2 GitHub App Deployment

**Current State**: Manual `atheon scan` runs.
**Future**: GitHub App for automatic PR scanning.

**Benefits**:
- No local setup required
- Per-repo configuration
- Dashboard for tracking issues over time

**Priority**: HIGH — major adoption driver

### 5.3 VS Code Extension

**Current State**: CLI only.
**Future**: Inline diagnostics, hover explanations.

**Priority**: MEDIUM

### 5.4 Commit Template (Pending)

Add conventional commit template:

```bash
git config commit.template .github/commit-template.txt
```

**File**: `.github/commit-template.txt`
```
# <type>(<scope>): <short description>
#
# Types: feat, fix, docs, test, refactor, chore, ci, build, perf, release
# Scope: (optional) module name
#
# Body: explain WHAT and WHY (not how)
#
# Footer: BREAKING CHANGE or Closes #<issue>
```

**Priority**: LOW — cosmetic improvement

---

## 6. Testing & Quality

### 6.1 Property-Based Testing

**Current**: Example-based tests.
**Future**: Hypothesis/property-based tests.

```go
// Property: regex that matches a secret always produces a finding
forall(secret in validSecrets) { assert(detect(secret).finding) }
```

**Priority**: MEDIUM

### 6.2 Chaos Engineering

**Future**:
- Inject malformed bundles
- Test corruption handling
- Verify graceful degradation

**Priority**: LOW

---

## 7. Wave 13 Proposed Plan

### PR Candidates (Priority Order)

1. **Incremental scanning** — file hash caching for delta scans
2. **AI pattern generation** — `ateon gen-pattern` command
3. **Interactive TUI** — terminal UI for scan results
4. **GitHub App skeleton** — repository structure for GitHub App

### Not Recommended for Near-Term

- WASM compilation (complex, limited use)
- ML-based FP detection (privacy concerns)
- YARA integration (specialized)

---

## 8. Technical Debt

### High Priority

| Item | Description | Effort |
|------|-------------|--------|
| MCP `isError`/`structuredContent` fields | Per-tool MCP response enhancement | MEDIUM |
| Bundle format schema v2 | Versioned bundle format with backwards compat | MEDIUM |
| `update_bundle --force` confirmation | Safety confirm for force updates | LOW |

### Medium Priority

| Item | Description | Effort |
|------|-------------|--------|
| Refactor pattern engine to plugin interface | Internal refactor, enables future plugins | HIGH |
| Extract shared `internal/errors` package | Already partially done in PR #119 | LOW |
| Transition to `go.mod` proxy blocking | Ensure reproducible builds | MEDIUM |

### Low Priority

| Item | Description |
|------|-------------|
| Commit template | Cosmetic improvement |
| Additional language support in tree-sitter | High effort |

---

## 9. Community Growth

### 9.1 Pattern Library Curation

**Future**:
- Weekly "Pattern Spotlight" in release notes
- Pattern of the month voting
- Community pattern showcase page

### 9.2 Integration Partners

**Target Organizations**:
- Snyk (pattern sharing)
- GitLab (SAST integration)
- JetBrains (IDE plugin)

### 9.3 Documentation Overhaul

**Current**: Scattered across README, AGENTS.md, docs/
**Future**: Unified docs at `atheon.dev` or GitHub Pages

---

## 10. Success Metrics

| Metric | Current | Target (12 months) |
|--------|---------|---------------------|
| Community patterns | ~280 | 500+ |
| GitHub stars | ~50 | 500+ |
| Downloads (monthly) | ~1K | 10K+ |
| Contributors | 1 | 5+ |
| GitHub App installations | 0 | 50+ |
| IDE extensions | 0 | 2+ |

---

## Appendix: Competitor Analysis

| Tool | Strengths | Weaknesses | Opportunity |
|------|-----------|------------|-------------|
| Semgrep | ML-powered, large ecosystem | Complex rule syntax | Better DX, simpler rules |
| CodeQL | Deep semantic analysis | Query language barrier | Accessibility |
| Bandit | Python-specific | Limited scope | Multi-language |
| TruffleHog | Secrets focus | No broader patterns | Pattern breadth |

**Differentiation**: Focus on **simplicity** + **comprehensiveness** + **developer experience** over semantic depth.

---

*This document is a living plan. Update quarterly.*
