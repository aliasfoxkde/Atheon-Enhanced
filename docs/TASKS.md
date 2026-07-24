# Tasks

Current and upcoming development tasks.

## In Progress

- Research and implement SkillSpector-inspired enhancements
- Unicode deception detection patterns

### Wave 18 (2026-07-24) — Coverage Phase 2
- [x] Add comprehensive phase 2 coverage tests
- [x] Core coverage: 84.9% → 85.3%
- [x] Add tests for typeToString, exprToString, callExprToString
- [x] Add tests for extractConstraints, slicesEqual, countInterfaceDepth
- [x] Add tests for hasConditionalReturnPath, hasUnconditionalReturnAtEnd
- [x] Add tests for detectPoorlyNamedIdentifier, modifiesVariable
- [x] PR #177 merged

### Wave 19 (2026-07-24) — Coverage Phase 3-9
- [x] Add phase 3 coverage tests (PR #179 merged)
- [x] Add phase 4 coverage tests
- [x] Add phase 5 coverage tests (PR #181 merged)
- [x] Add phase 6 coverage tests (PR #182 merged)
- [x] Add phase 7 coverage tests (PR #183 merged)
- [x] Add phase 8 coverage tests (PR #184 merged)
- [x] Add phase 9 coverage tests (PR #185 merged)
- [x] Core coverage: 85.3% → 87.6%

## Backlog

## Completed

### Wave 14 (2026-06-28) — SkillSpector Research
- [x] Research NVIDIA SkillSpector architecture
- [x] Add AST7: dynamic getattr detection
- [x] Add AST8: dangerous execution chain detection
- [x] Add AST9: reflective getattr sink detection
- [x] Add prompt injection patterns (ignore, override, jailbreak)
- [x] Add hidden instruction patterns (HTML/markdown comments)
- [x] Add MCP security patterns (hidden instructions, exfil, malicious defaults)
- [x] Add AI skill security patterns (credential exfil, remote eval, destructive actions)
- [x] Total patterns: 366 across 23 categories

### Wave 13 (2026-06-27)
- [x] Add anti-cheating AI detection patterns (21 patterns)
- [x] Add magic number detection patterns
- [x] Add AI harness anti-pattern patterns
- [x] Expand frameworks patterns (Angular, Express, Flask, Laravel, Rails, Spring, Vue)
- [x] Fix auto-merge workflow (timeout exit 1 instead of 0)
- [x] Add .github/SUPPORT.md
- [x] Add .github/CHANGELOG.md
- [x] Add .editorconfig
- [x] Remove redundant documentation files
- [x] Rename lowercase markdown files to UPPERCASE
- [x] Generate README.md for each category folder
- [x] Update community/README.md with accurate pattern counts
- [x] Dead code quality patterns (empty-if-block, empty-else-block, empty-for-loop, etc.)

### Wave 12 (2026-06-26)
- [x] SDLC audit fixes
- [x] Dependabot groups
- [x] PR template improvements
- [x] Release environment variables
- [x] GO_VERSION in goreleaser

## Backlog

- Unicode deception detection (RTL overrides, homoglyphs, zero-width chars)
- OSV.dev CVE lookups
- Full YARA library integration (currently simplified version in core/yara_scanner.go)

### Recently Completed (Wave 17 - 2026-07-24)
- [x] Bundle decode tests (decodeJSONStrict, decodeBundleDefs, trimSpace)
- [x] normalizeConfidence tests with valid confidence levels
- [x] Clone detection tests (variety of statements, custom config)
- [x] Audit layer tests (type checking, security, code smells)
- [x] Expression and call expression tests (exprToString, callExprToString)
- [x] Test coverage: 81.1% → 84.7%
- [x] Added comprehensive coverage_test.go with AST pattern tests
- [x] Python-specific functions (containsDangerousSource, hasStringLiteral, isStringType) remain at 0% - these only execute on Python code scans, not Go tests

### Recently Completed (Wave 16 - 2026-07-24)
- [x] Null dereference detection pattern (null-dereference)
- [x] Dead assignment detection pattern (dead-assignment)
- [x] CFG-based bug detection: lock-not-released, resource-leak, transaction-not-ended
- [x] Circular import detection (import graph analysis)
- [x] Test coverage: 79.9% → 80.3%

## Recently Completed (Wave 15 - 2026-07-16)
- [x] Taint tracking analysis (source → sink flow) - core/taint.go
- [x] Taint command injection pattern - community/security-hardening/taint-command-injection.yaml
- [x] Risk scoring system (0-100) - core/risk.go, integrated into JSON/SARIF output
- [x] Baseline suppression for re-scans - core/suppression.go with --baseline flag
- [x] YARA scanner (simplified) - core/yara_scanner.go
- [x] Test coverage: 66.2% → 77.0%
- [x] Fix Findigs typo → Findings in suppression.go
- [x] Add tests for atomic_file, ignore, yara_scanner modules

## See Also

- [PROGRESS.md](PROGRESS.md) - Completed work
- [ROADMAP.md](ROADMAP.md) - Future plans
- [docs/planning/IMPLEMENTATION_PLAN.md](planning/IMPLEMENTATION_PLAN.md) - Detailed implementation roadmap
