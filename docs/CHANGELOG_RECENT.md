# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.18-enhanced] - 2026-07-17

### Fixed
- golangci-lint: replace `len(s)==0` with `s==""` per gocritic emptyStringTest rule
- CI: Self-scan jq filters updated for new JSON output format
- Release: golangci-lint install path fixed in release workflow validation
- Release: GoReleaser archive syntax fixed for v2.6+

### Changed
- Go version upgraded to 1.25 to address 21 Go standard library vulnerabilities
- Build: Go 1.25 now tested in CI matrix alongside 1.22-1.24

### Merged
- #153: fix: update Self-Scan jq filters for new JSON output format
- #154: fix: replace len(s)==0 with s=="" per golangci-lint
- #144: build(deps): bump actions-minor-and-patch group with 5 updates
- #142: feat: implement code quality roadmap (entropy, confidence, pattern fixes)
- #141: feat(ci): add job dependency gating and ARM64 cross-compilation testing
- #138: build(deps): bump github.com/goccy/go-yaml from 1.11.0 to 1.19.2

## [1.3.14-enhanced] - 2026-07-17

### Fixed
- GoReleaser syntax for v2.6+ (format singular, format_overrides)

### Added
- 19 new security patterns across 5 categories
- AST pattern enhancements with prompt injection and MCP security patterns

## [1.3.13-enhanced] - 2026-07-17

### Added
- Pattern categories for Vue, Angular, Rails, Flask, Laravel, Express, Spring

## [1.3.12-enhanced] - 2026-07-17

### Added
- AST patterns, dead code patterns, and test fixture improvements

## [1.3.8-enhanced] - 2026-06-29

### Added
- Comprehensive enhancement audit and documentation
- Taint tracking, risk scoring, baseline suppression, and YARA scanning
- Risk scoring integrated into JSON and SARIF output
- `--baseline` flag to suppress known findings

### Changed
- Coverage improved to 78.5%
