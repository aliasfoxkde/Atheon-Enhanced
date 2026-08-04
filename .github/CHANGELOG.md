# Changelog

All notable changes to **Atheon-Enhanced** are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

For detailed recent changes, see [docs/CHANGELOG_RECENT.md](docs/CHANGELOG_RECENT.md).

## [1.3.18-enhanced] - 2026-07-17

### Fixed
- golangci-lint: replace `len(s)==0` with `s==""` per gocritic emptyStringTest rule
- CI: Self-scan jq filters updated for new JSON output format
- Release: golangci-lint install path fixed in release workflow validation
- Release: GoReleaser archive syntax fixed for v2.6+
- Go version upgraded to 1.25 to address 21 Go standard library vulnerabilities

### Merged
- #153, #154, #144, #142, #141, #138

## [1.3.14-enhanced] - 2026-07-17

### Fixed
- GoReleaser syntax for v2.6+

### Added
- 19 new security patterns, AST enhancements, MCP security patterns

## [1.3.13-enhanced] - 2026-07-17

### Added
- Framework patterns for Vue, Angular, Rails, Flask, Laravel, Express, Spring

## [1.3.12-enhanced] - 2026-07-17

### Added
- AST patterns, dead code patterns, test fixtures

## [1.3.8-enhanced] - 2026-06-29

### Added
- Taint tracking, risk scoring, baseline suppression, YARA scanning
- Risk scoring in JSON and SARIF output
- `--baseline` flag for suppressing known findings
- 384 security patterns across 28 categories

### Changed
- Coverage: 78.5%

## [1.3.1-enhanced] - 2026-06-18

### Added
- Comprehensive security scanning patterns
- Accessibility, secrets, web security patterns
- Community pattern contributions

## [1.0.0] - 2026-06-17

### Added
- Initial release
- Core scanning engine
- MCP server integration
- GitHub Actions CI/CD
