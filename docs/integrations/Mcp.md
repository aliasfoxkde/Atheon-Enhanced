# MCP Server Integration

Atheon ships an MCP (Model Context Protocol) server at `cmd/mcp`. It lets AI assistants — Claude Desktop, VS Code Copilot, Cursor, Windsurf — call the Atheon pattern engine as a tool during conversations.

## Build

The MCP binary is separate from the main CLI:

```bash
# Linux / macOS
go build -o atheon-mcp ./cmd/mcp

# Windows
go build -o atheon-mcp.exe ./cmd/mcp
```

The binary communicates over **stdio** (stdin/stdout), which is the standard transport for local MCP servers. No network port is needed.

## What the MCP Server Exposes

| Tool name | Description |
|-----------|-------------|
| `scan_file` | Scan a single file, returns findings array |
| `scan_dir` | Scan a directory recursively |
| `scan_string` | Scan arbitrary text (useful for scanning clipboard content or inline code) |
| `scan_env` | Scan current environment variables |
| `list_patterns` | List all patterns with enabled/disabled status |
| `enable_pattern` | Enable a pattern by name |
| `disable_pattern` | Disable a pattern by name |

## Claude Desktop

Add to `~/.claude/claude_desktop_config.json` (create if it doesn't exist):

```json
{
  "mcpServers": {
    "atheon": {
      "command": "/path/to/atheon-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

On Windows, use the full path with forward slashes or escaped backslashes:

```json
{
  "mcpServers": {
    "atheon": {
      "command": "C:/Repos/Atheon-Enhanced/atheon-mcp.exe"
    }
  }
}
```

Restart Claude Desktop after saving. You will see an "atheon" tool listed in the tools panel.

## VS Code (GitHub Copilot)

Add to `.vscode/settings.json` in any workspace, or to your user `settings.json` (`Ctrl+Shift+P → Open User Settings JSON`):

```json
{
  "github.copilot.chat.mcp.servers": {
    "atheon": {
      "type": "stdio",
      "command": "/path/to/atheon-mcp"
    }
  }
}
```

Requires GitHub Copilot Chat extension. After saving, the MCP server appears in the Copilot tools list. You can then ask Copilot: *"scan this directory for secrets using Atheon"*.

## Cursor

1. Open Cursor Settings (`Ctrl+,`)
2. Navigate to **Tools & MCP**
3. Click **Add MCP Server**
4. Enter:
   - Name: `atheon`
   - Type: `stdio`
   - Command: `/path/to/atheon-mcp`
5. Save — Cursor auto-reloads the config

## Windsurf / Codeium

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "atheon": {
      "command": "/path/to/atheon-mcp",
      "args": []
    }
  }
}
```

## Distributing to Users

MCP servers have no central registry. To help users configure Atheon's MCP server, document the binary path and config snippet in your project README. Users point the config at whichever binary they built or downloaded.

For a GitHub Releases approach: add `atheon-mcp` as a release artifact (see goreleaser setup in `docs/OWNER_CHECKLIST.md`), then users download and configure once.

## Logging

The MCP server logs to stderr (not stdout, which is reserved for the JSON-RPC protocol). When running under an AI client, stderr is captured and shown in the client's MCP diagnostics panel.

To see raw stdio traffic during development:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./atheon-mcp
```
