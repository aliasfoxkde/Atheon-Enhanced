# Getting Started with Atheon

## Quick Installation

```bash
# Install enhanced version
go install github.com/aliasfoxkde/Atheon@latest

# Verify installation
atheon --version
```

## Your First Scan

```bash
# Scan current directory
atheon .

# Scan specific file
atheon --file config/app.yaml

# Scan with categories
atheon --categories=secrets,pii .
```

## Understanding the Output

```
openai-api-key  config/app.yaml:47
  # debug key: sk-****4f59

stripe-api-key  .env:3
  sk_test_****MzN3

2 finding(s)
scanned 23 file(s)  41.3 KB  4ms
```

- **Pattern name**: What was detected
- **File location**: Where it was found
- **Match content**: The actual matching text (redacted)
- **Summary**: Total findings and scan statistics

## Common Use Cases

### Pre-commit Hook
```bash
# Create hook
echo '#!/bin/sh
atheon ./' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### CI/CD Integration
```bash
# Add to pipeline
atheon --categories=secrets --json . > security-findings.json

# Exit code 1 if findings found
if [ $? -ne 0 ]; then
    echo "Security issues detected!"
    exit 1
fi
```

### MCP Server Setup
```bash
# Build MCP server
go build -o atheon-mcp ./cmd/mcp

# Add to Claude Code config
{
  "mcpServers": {
    "atheon": {
      "command": "/path/to/atheon-mcp"
    }
  }
}
```

## Configuration Profiles

Atheon comes with pre-configured profiles:

- **production.json** - General use (default)
- **pipeline.json** - CI/CD optimized
- **mcp-integration.json** - MCP server settings
- **development.json** - Full feature testing

```bash
# Use profile
atheon --profile config/profiles/pipeline.json ./my-project
```

## Pattern Management

```bash
# List all patterns
atheon list

# Enable/disable patterns
atheon enable stripe-api-key
atheon disable todo-comments

# Update pattern bundle
atheon update
```

## Troubleshooting

### "No findings" but expected results
- Check if patterns are enabled: `atheon list --enabled`
- Try `--all` flag to include disabled patterns
- Verify pattern syntax with `atheon list`

### Performance issues on large files
- Use streaming API (automatic in enhanced version)
- Consider category filtering: `--categories=secrets`
- Check memory usage with statistics

### Pattern not working
- Test pattern regex: use regex tester
- Check if pattern is enabled: `atheon list --pattern <name>`
- Verify pattern syntax: `atheon list --disabled`

## Next Steps

- Explore [Configuration Profiles](https://github.com/aliasfoxkde/Atheon/wiki/Configuration-Profiles)
- Learn about [Pattern Development](https://github.com/aliasfoxkde/Atheon/wiki/Pattern-Development)
- Set up [CI/CD Integration](https://github.com/aliasfoxkde/Atheon/wiki/CI-CD-Integration)
- Configure [MCP Integration](https://github.com/aliasfoxkde/Atheon/wiki/MCP-Integration)