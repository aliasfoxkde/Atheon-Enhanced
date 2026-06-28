# Roadmap

Atheon-Enhanced's trajectory: a larger, more authoritative pattern library across more domains. The engine is stable. What grows from here are the patterns.

---

## Current state (June 2026)

- **335 patterns** across 23 categories
- **AI detection** with 21 patterns (anti-cheating, harness integrity)
- CLI with category filtering, enable/disable with persistent state, JSON/SARIF output, stdin piping
- Streaming API for memory-efficient large file scanning
- MCP server (`atheon-mcp`) for Claude, Cursor, and Windsurf integration
- Git hook support (pre-commit, pre-push)
- CI/CD integration with native binaries for Windows, macOS, and Linux
- Self-scanning validation in CI
- Pattern state persistence across sessions

---

## Near term

**Pattern expansion**

- Expand `frameworks` category (Django, React, Vue, Angular patterns)
- Expand `kubernetes` and `terraform` categories
- More `compliance` patterns (SOC2, ISO 27001)

**Tooling improvements**

- Smart pre-commit hooks (scan only changed files)
- Pattern validation tool (`atheon check <file>`)
- Pattern overlap detection
- Confidence metadata for patterns

---

## Medium term

**Pattern quality**

- Context-aware matching — anchor patterns to specific file types
- Pattern deprecation workflow
- False positive rate tracking

**Enhanced integrations**

- MCP server improvements
- IDE integrations
- Dashboard plugins

---

## Contributing

Every pattern on the roadmap is open for contribution. See [.github/CONTRIBUTING.md](../.github/CONTRIBUTING.md) to add yours.
