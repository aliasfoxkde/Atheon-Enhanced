# MCP Server Setup

Atheon-Enhanced includes an MCP (Model Context Protocol) server that can be integrated with Claude Code for enhanced security scanning capabilities.

## Configuration

Add the following to your `~/.claude/mcp.json` file:

```json
{
  "mcpServers": {
    "atheon": {
      "command": "atheon-mcp",
      "args": [],
      "cwd": "/path/to/Atheon-Enhanced",
      "env": {}
    }
  }
}
```

## Available Tools

The MCP server provides the following tools:

- `scan_string` - Scan a string for secrets/patterns
- `scan_file` - Scan a file for secrets/patterns
- `scan_dir` - Scan a directory for secrets/patterns
- `scan_env` - Scan environment variables for secrets

## Protocol

The server uses JSON-RPC 2.0 over stdin/stdout. It implements the MCP 2024-11-05 protocol specification.

## Example Usage

After configuring, you can use Claude Code to scan codebases for security issues using natural language commands like:

- "Scan this directory for API keys"
- "Check environment variables for secrets"
- "Find all hardcoded passwords in the project"
