# Improvement Plan - 2026-07-17

## Current State
- **Coverage:** 81.1% overall
- **Patterns:** 406 enabled patterns across 36 categories
- **Tests:** All passing
- **Release:** v1.3.18-enhanced live with 7 artifacts

## Identified Gaps & Improvements

### 1. Coverage Gaps

| Package | Current | Target | Gap |
|---------|---------|--------|-----|
| internal/atomicio | 64.0% | 80% | Error paths require mocking os package |
| core | 77.8% | 80% | AST pattern tests, entropy edge cases |

**Action:** Coverage at 81.1% is healthy. The atomicio gap requires architectural changes (mocking) that add complexity without proportional benefit.

### 2. Pattern Gaps

| Category | Current | Potential |
|----------|---------|-----------|
| Rust | 7 patterns | Add more Rust-specific patterns ( lifetimes, borrowing issues) |
| Go | 3 patterns | Add Go-specific security patterns |
| Kubernetes | 11 patterns | Expand to include more K8s security scenarios |
| Terraform | 8 patterns | Expand AWS/Azure/GCP security configurations |

**Action:** Rust patterns are new and comprehensive. Consider Go-specific patterns.

### 3. Documentation Gaps

| Item | Status |
|------|--------|
| README.md | 28KB, comprehensive |
| docs/ARCHITECTURE.md | Needs update for recent changes |
| docs/ROADMAP.md | Should reflect completed work |
| API docs | Missing |

**Action:** Update ARCHITECTURE.md and ROADMAP.md

### 4. Performance

| Area | Status |
|------|--------|
| Entropy caching | ✅ Done (100x faster) |
| Bundle loading | Benchmarked |
| Large directory scans | Benchmarked |

**Action:** Current performance is adequate.

### 5. CI/CD Gaps

| Area | Status |
|------|--------|
| Go version | ✅ 1.25 |
| golangci-lint | ✅ v2.5 |
| CodeQL | ✅ Active |
| govulncheck | ✅ Active |
| Auto-merge | ✅ Working |

**Action:** CI/CD is healthy.

## Recommended Next Steps (Priority Order)

### High Priority
1. ~~**Update docs/ARCHITECTURE.md**~~ - ✅ Done
2. ~~**Update docs/ROADMAP.md**~~ - ✅ Done
3. ~~**Add Go security patterns**~~ - ✅ Done (3 patterns added)

### Medium Priority
4. ~~**Add Kubernetes patterns**~~ - ✅ Done (11 patterns now)
5. ~~**Add Terraform patterns**~~ - ✅ Done (8 patterns now)
6. **Performance: Bundle caching** - Cache compiled patterns between invocations

### Low Priority
7. **Homebrew/Scoop publishing** - Requires dedicated tap/bucket repos
8. **SBOM generation** - Requires syft installation in CI

## Implementation Plan

1. Update ARCHITECTURE.md and ROADMAP.md
2. Add 5-10 Go security patterns
3. Expand Kubernetes patterns
4. Expand Terraform patterns
5. Commit/push and create PR

## Success Criteria
- Coverage: 81%+ (maintain)
- Patterns: 400+ enabled
- All tests passing
- All CI green
