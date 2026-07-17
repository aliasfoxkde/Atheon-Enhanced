# Roadmap

Atheon-Enhanced's trajectory: a larger, more authoritative pattern library across more domains. The engine is stable. What grows from here are the patterns.

---

## Current state (July 2026)

- **400+ patterns** across 36 categories
- **AI detection** with patterns for prompt injection, MCP security, AI-generated content
- CLI with category filtering, enable/disable with persistent state, JSON/SARIF output, stdin piping
- Streaming API for memory-efficient large file scanning
- MCP server (`atheon-mcp`) for Claude Code, Cursor, and Windsurf integration
- Git hook support (pre-commit, pre-push)
- CI/CD integration with native binaries for Windows, macOS, and Linux
- Self-scanning validation in CI
- Pattern state persistence across sessions
- Entropy caching for 100x speedup on repeated calculations
- Risk scoring (0-100 scale) and baseline comparison
- **81%+ test coverage**

---

## Completed (2026)

- ✅ Entropy caching implementation
- ✅ Risk scoring system
- ✅ Baseline comparison feature
- ✅ Self-scanning validation in CI
- ✅ Rust security patterns (7 patterns)
- ✅ Go security patterns (3 patterns)
- ✅ Kubernetes security patterns (expanded to 8 patterns)
- ✅ Cloud-native security patterns (AWS, EKS, ECS, Lambda)
- ✅ Terraform patterns (AWS, Azure, GCP)
- ✅ Supply chain patterns (typosquatting, malicious packages)
- ✅ Container security patterns (Dockerfile, Docker Compose)
- ✅ GraphQL security patterns
- ✅ CloudFormation security patterns
- ✅ ARM templates security patterns

---

## Near term

**Pattern expansion**

- Expand `frameworks` category (more Go patterns)
- More `compliance` patterns (SOC2, ISO 27001)
- Context-aware matching — anchor patterns to specific file types

**Tooling improvements**

- Smart pre-commit hooks (scan only changed files)
- Pattern validation tool (`atheon check <file>`)
- Pattern overlap detection
- Confidence metadata for patterns

---

## Medium term

**Pattern quality**

- Pattern deprecation workflow
- False positive rate tracking
- Pattern confidence scoring

**Enhanced integrations**

- MCP server improvements
- IDE integrations
- Dashboard plugins

---

## Contributing

Every pattern on the roadmap is open for contribution. See [.github/CONTRIBUTING.md](../.github/CONTRIBUTING.md) to add yours.
