<p align="center">
  <img src="./assets/logo.svg" alt="Atheon" width="600" />
</p>

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Patterns](https://img.shields.io/badge/patterns-14-blueviolet)

> **One tool. All patterns. Any input.**

Atheon is a community-driven pattern matching engine. You define what you're looking for. You point it at anything. It finds every match and tells you exactly where  returning a clear `true` or `false` for every rule, every time.

---

## What Atheon is

Atheon is a CLI tool built around a single idea: **any pattern, any domain, any input.** It doesn't care whether you're scanning for leaked credentials, patient identifiers, financial account numbers, prohibited strings in compliance-scoped code, or anything else you can describe as a rule. If the pattern is clear  if it can return true or false  Atheon runs it.

The engine itself is deliberately minimal. It has no opinions about what matters. That knowledge lives in the patterns, and the patterns come from the community.

---

## The Mission

Atheon isn't trying to be the next big secrets scanner. It's not competing to become a giant. It's trying to be a **platform**.

Here's the idea: a developer on a team is working with sensitive data. They write a pattern for Atheon, contribute it, and it ships in the next release. Now everyone using Atheon has that pattern registered. The next team in a similar situation doesn't have to build it from scratch  it's already there.

That's Atheon: a community-driven engine where you, me, and anyone else can add patterns that every user benefits from. The goal is a library of rules that covers every domain where text contains something that matters  built not by one company, but by everyone who uses it.

**Security. Compliance. Finance. Healthcare. Legal. Operations. Gaming. Anything.**

If you can describe the rule, Atheon can run it.

---

## See it in action

[![Watch on YouTube](https://img.youtube.com/vi/4vlepIzRGqw/maxresdefault.jpg)](https://youtu.be/4vlepIzRGqw)

Skip to **9:36** for the live demo, or watch the whole thing to see what Atheon is about.

---

## The scenario that makes this real

A developer wraps up a sprint and pushes a configuration file. Buried in a comment from a debugging session three weeks ago is a production API key. The commit goes through. The pipeline passes. Two months later, someone notices unusual billing activity.

Atheon, wired into a pre-push hook:

```
$ atheon ./

openai-api-key  config/app.yaml:47
  # de****4f59

1 finding(s)
scanned 14 file(s)  22.1 KB  3ms
```

Exit code `1`. The push never happens. The key never leaves the machine.

That's it. That's the product.

---

## Install

**Homebrew (macOS / Linux):**

```
brew tap HoraDomu/atheon
brew install atheon
```

**Scoop (Windows):**

```
scoop bucket add atheon https://github.com/HoraDomu/scoop-atheon
scoop install atheon
```

**Manual:** Download the binary for your platform from [Releases](https://github.com/HoraDomu/Atheon/releases/latest). No install, no runtime, no dependencies. Drop it in your PATH and run it.

**Build from source:**

```
go build -o atheon .
```

---

## Usage

```
atheon <path>                        scan a directory or file
atheon --file <path>                 scan a single file explicitly
atheon --env                         scan all environment variables
atheon --json <path>                 output findings as JSON
atheon --categories=<c1,c2> <path>  scan specific pattern categories only
atheon --all <path>                  scan all categories (default)
atheon list                          list every loaded pattern
atheon list categories               list available categories
atheon update                        download the latest patterns bundle
atheon --help                        show help
```

Pipe support  pass `-` to read from stdin:

```
cat file.txt | atheon -
git diff | atheon -
```

Exit code `0` = clean. Exit code `1` = findings. CI-friendly by default.

---

**Category filtering**

Patterns are organized into categories. Run only what you need:

```
atheon --categories=secrets .
atheon --categories=secrets,pii .
atheon list categories
```

This keeps scans fast regardless of how many patterns are in the bundle. A pre-commit hook scanning only `secrets` costs nothing for PII patterns you don't need in that context.

---

**Cross-platform:** native binaries for Windows, macOS (Intel + Apple Silicon), and Linux. No runtime, no dependencies.

---

**Ignore rules**

Directory scans automatically respect `.gitignore`. Drop a `.atheonignore` in your project root to exclude anything not already covered  test fixtures, generated files, `.env` files:

```
# .atheonignore
test/
*.generated.go
.env
```

To suppress a single line without ignoring the whole file, add `atheon:ignore` anywhere on that line:

```
DEBUG_KEY=sk-fake-key-for-testing  # atheon:ignore
```

---

**JSON output**

Use `--json` to integrate with other tools or build your own pipeline on top:

```
atheon --json ./
```

Output is a JSON array, one object per finding:

```json
[{"pattern":"openai-api-key","file":"config/app.yaml","line":47,"match":"# debug key: sk-..."}]
```

---

**Environment scanning**

`--env` scans every variable in the current environment  useful in CI to catch secrets injected at runtime rather than stored in files:

```
atheon --env
```

---

**Pre-commit / pre-push hook**

Drop Atheon into a git hook to block bad commits before they leave the machine:

```sh
# .git/hooks/pre-push
#!/bin/sh
atheon ./
```

Or with category filtering for speed:

```sh
#!/bin/sh
atheon --categories=secrets ./
```

Wire it into whatever hook runner you already use (pre-commit, Husky, Lefthook). Atheon returns exit code `1` on any finding, which is all a hook needs to abort.

---

**MCP server**

Atheon ships a separate `atheon-mcp` binary that speaks the Model Context Protocol over stdio. Drop it into any MCP-compatible AI tool to let the assistant scan code, files, and directories for pattern matches.

### Installation

**Download releases:**
- Linux: `atheon-mcp-linux-amd64` or `atheon-mcp-linux-arm64`
- macOS: `atheon-mcp-darwin-amd64` or `atheon-mcp-darwin-arm64`
- Windows: `atheon-mcp-windows-amd64.exe`

**Homebrew:**
```bash
brew install HoraDomu/homebrew-atheon
# Installs both atheon and atheon-mcp binaries
```

**Scoop (Windows):**
```powershell
scoop bucket add HoraDomu/scoop-atheon
scoop install atheon
# Includes both atheon and atheon-mcp
```

**Build from source:**
```bash
go build -o atheon-mcp ./cmd/mcp
```

### Configuration

**Claude Code:**
```json
{
  "mcpServers": {
    "atheon": {
      "command": "/path/to/atheon-mcp"
    }
  }
}
```

**Cursor:**
```json
{
  "mcpServers": {
    "atheon": {
      "command": "atheon-mcp",
      "args": []
    }
  }
}
```

**Windsurf:**
```json
{
  "mcpServers": {
    "atheon": {
      "command": "/usr/local/bin/atheon-mcp"
    }
  }
}
```

### Available Tools

**`scan_string`** - Scan text content for patterns:
```json
{
  "name": "scan_string",
  "arguments": {
    "content": "API_KEY=sk-1234567890abcdef",
    "source": "environment",
    "categories": ["secrets"]
  }
}
```

**`scan_file`** - Scan a single file:
```json
{
  "name": "scan_file",
  "arguments": {
    "path": "/path/to/config.yaml",
    "categories": ["secrets", "pii"]
  }
}
```

**`scan_dir`** - Scan entire directories:
```json
{
  "name": "scan_dir",
  "arguments": {
    "path": "/path/to/project",
    "categories": ["secrets", "pii", "code-quality"]
  }
}
```

### Usage Examples

**Claude Code Example:**
```
User: "Can you scan the current directory for security issues?"
Assistant: [Uses scan_dir tool] "I found 3 security issues in your codebase..."
```

**Cursor Example:**
```
User: "@Atheon scan this file"
Assistant: [Uses scan_file tool] "Found 2 patterns in config.yaml..."
```

### Categories

Available categories for filtering:
- `secrets` - API keys, tokens, credentials
- `pii` - Personal information (SSN, credit cards, etc.)
- `code-quality` - Debug statements, TODOs, technical debt
- `healthcare` - Medical identifiers, PHI patterns

Omit the `categories` parameter to scan all categories.

---

## Pattern bundle

All patterns live in `community/` as plain YAML files  no Go required. The engine ships with a compiled bundle embedded in the binary. Run `atheon update` to pull the latest bundle from the release.

Adding a new pattern is one file:

```yaml
# community/secrets/my-service.yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

The folder name is the category. No engine changes, no recompile, no release gate.

---

## Contributing

Patterns are the heart of Atheon. Every pattern is one YAML file  small, fast to review, and immediately useful to every user once merged.

See [CONTRIBUTING.md](CONTRIBUTING.md) to add your own.

---

## Releases

New versions ship on the **10th and 21st of every month**. Releases are fully automated  tagging a version builds all platform binaries, generates the patterns bundle, and publishes everything to GitHub Releases, Homebrew, and Scoop automatically.

Latest release: [github.com/HoraDomu/Atheon/releases/latest](https://github.com/HoraDomu/Atheon/releases/latest)

---

## Thank you

Atheon is built by the community. Every pattern contributed ships to every user in the next release. See everyone who has helped make it here: [CONTRIBUTORS.md](CONTRIBUTORS.md)

---

## Contact

Questions, pattern requests, or anything else:

**Email:** [dommcpro@gmail.com](mailto:dommcpro@gmail.com)

---

## License

MIT with Additional Terms Copyright © 2026 Dominick Yanez

You are free to fork, clone, study, modify for personal or internal use, and contribute patterns or bug fixes back. That's encouraged.

What you may not do:
- Ship this software, or any derivative of it, as your own standalone product under a different name or brand
- Remove or obscure the author's name or copyright notice from any copy, fork, or derivative work

For permissions beyond this scope: [dommcpro@gmail.com](mailto:dommcpro@gmail.com)

See the full [LICENSE](./LICENSE) file for complete terms.
