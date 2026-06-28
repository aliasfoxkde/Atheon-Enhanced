---
layout: default
title: Atheon-Enhanced
---

# Atheon-Enhanced

Pattern-matching engine for secrets, PII, and code quality. 274 patterns, SARIF output, MCP server for AI assistant integration.

[![CI](https://github.com/aliasfoxkde/Atheon-Enhanced/actions/workflows/ci.yml/badge.svg)](https://github.com/aliasfoxkde/Atheon-Enhanced/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/aliasfoxkde/Atheon-Enhanced/graph/badge.svg)](https://codecov.io/gh/aliasfoxkde/Atheon-Enhanced)

## Quick Start

```bash
go install github.com/aliasfoxkde/Atheon-Enhanced/cmd/atheon@latest
atheon .                        # scan current directory
atheon --json .                 # JSON output
atheon --sarif . > results.sarif  # SARIF for GitHub Security tab
```

## Documentation

| | |
|---|---|
| [API Reference](api/README.md) | Go library API — ScanFile, ScanDir, ScanString, ScanEnv |
| [Pattern Format](PATTERN_FORMAT.md) | YAML schema for writing patterns |
| [Development Guide](development.md) | Branch strategy, local setup, release process |
| [Self-Scan Integration](self-scan.md) | Running Atheon against your own CI pipeline |
| [MCP Server Setup](integrations/mcp.md) | Use Atheon from Claude Desktop, VS Code, Cursor |
| [GitHub Agent Integration](integrations/github-agents.md) | GitHub Models API and Agentic Workflows |
| [Owner Checklist](OWNER_CHECKLIST.md) | Setup tasks, recommendations, maintenance |

## Source

[github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced)
