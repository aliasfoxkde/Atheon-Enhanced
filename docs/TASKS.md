# Tasks

Current and upcoming development tasks.

## In Progress

- Research and implement SkillSpector-inspired enhancements
- Unicode deception detection patterns
- Taint tracking analysis

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
- Taint tracking analysis (source → sink flow)
- Risk scoring system (0-100)
- Baseline suppression for re-scans
- YARA rule integration
- OSV.dev CVE lookups

## See Also

- [PROGRESS.md](PROGRESS.md) - Completed work
- [ROADMAP.md](ROADMAP.md) - Future plans
- [docs/planning/IMPLEMENTATION_PLAN.md](planning/IMPLEMENTATION_PLAN.md) - Detailed implementation roadmap
