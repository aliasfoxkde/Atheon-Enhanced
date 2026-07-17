# Claude Code MCP Configuration

Atheon-Enhanced provides an MCP server for integration with Claude Code.

## Setup Instructions

1. **Locate your Claude Code MCP configuration file:**
   - **macOS/Linux:** `~/.claude/mcp.json`
   - **Windows:** `%USERPROFILE%\.claude\mcp.json`

2. **Add the atheon server configuration:**

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

3. **Restart Claude Code** to load the new MCP server.

## Available Tools

Once configured, the following tools are available:

| Tool | Description |
|------|-------------|
| `scan_string` | Scan a string for secrets/patterns |
| `scan_file` | Scan a file for secrets/patterns |
| `scan_dir` | Scan a directory recursively |
| `scan_env` | Scan environment variables |

## Protocol

The server implements JSON-RPC 2.0 over stdin/stdout using MCP protocol version 2024-11-05.

## Verification

After setup, verify by asking Claude Code:
> "What MCP servers are available?"

You should see `atheon` listed.
