# Troubleshooting Guide

## Installation Issues

### "command not found" after installation
```bash
# Verify Go bin is in PATH
echo $PATH | grep "$HOME/go/bin"

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

### Build failures
```bash
# Ensure Go 1.21+ is installed
go version

# Update Go modules
go mod tidy
go mod download
```

### Permission denied on binary
```bash
# Make binary executable
chmod +x atheon
chmod +x atheon-mcp
```

## Scanning Issues

### "No findings" but expected results

**Check pattern status:**
```bash
# List enabled patterns
atheon list --enabled

# List all patterns including disabled
atheon list

# Try with all patterns
atheon --all ./
```

**Verify pattern syntax:**
```bash
# Check specific pattern
atheon list --pattern stripe-api-key
```

### Slow scanning on large projects

**Use category filtering:**
```bash
# Scan only specific categories
atheon --categories=secrets ./large-project

# Use pipeline profile
atheon --profile config/profiles/pipeline.json ./large-project
```

**Memory issues:**
```bash
# Enhanced version uses streaming automatically
# Check available memory
free -h  # Linux
vm_stat  # macOS
```

### False positives

**Add ignore rules:**
```bash
# Create .atheonignore file
cat > .atheonignore << EOF
test/
*.generated.go
.env
EOF
```

**Inline ignore:**
```bash
# Add atheon:ignore to specific line
DEBUG_KEY=sk-fake-key-for-testing  # atheon:ignore
```

## Pattern Issues

### Pattern not matching

**Test regex pattern:**
```bash
# Use online regex tester to verify pattern
# Ensure pattern uses word boundaries: \b
```

**Pattern syntax errors:**
```bash
# Validate pattern file
cat community/secrets/my-pattern.yaml

# Test with simple example
echo "test_content" | atheon -
```

### Pattern not loading

**Check pattern file:**
```bash
# Verify YAML syntax
cat community/secrets/my-pattern.yaml

# Rebuild bundle
go run ./bundler
```

## CI/CD Issues

### Hook not executing

**Verify hook permissions:**
```bash
# Make executable
chmod +x .git/hooks/pre-commit

# Test manually
.git/hooks/pre-commit
```

**Check git configuration:**
```bash
# Verify hook path
git config core.hooksPath

# Test hook
git run hook pre-commit
```

### Pipeline failures

**Debug JSON output:**
```bash
# Test JSON generation
atheon --json ./test-project > test-output.json
cat test-output.json
```

**Exit code issues:**
```bash
# Check exit code explicitly
atheon ./test-project
echo "Exit code: $?"
```

## MCP Integration Issues

### MCP server not starting

**Build verification:**
```bash
# Rebuild MCP server
go build -o atheon-mcp ./cmd/mcp

# Test directly
./atheon-mcp
```

**Configuration issues:**
```bash
# Test MCP tools
echo '{"name": "scan_string", "arguments": {"content": "test"}}' | ./atheon-mcp
```

### AI assistant not using Atheon

**Configuration verification:**
```json
{
  "mcpServers": {
    "atheon": {
      "command": "/absolute/path/to/atheon-mcp"
    }
  }
}
```

**Path issues:**
- Use absolute paths for MCP command
- Verify binary is executable
- Check AI assistant MCP configuration

## Performance Issues

### High CPU usage

**Category filtering:**
```bash
# Use fewer categories
atheon --categories=secrets ./project

# Profile mode (slower but thorough)
--profile config/profiles/development.json
```

### High memory usage

**Enhanced version optimizations:**
```bash
# Streaming is automatic in enhanced version
# For large files, memory usage is 10x lower than upstream

# Check process memory
ps aux | grep atheon
```

## Configuration Issues

### Profile not loading

**Verify profile syntax:**
```bash
# Check JSON syntax
cat config/profiles/production.json | jq empty

# Test profile explicitly
atheon --profile config/profiles/production.json --list
```

**Default configuration:**
```bash
# Check embedded configuration
atheon --help

# Test without profile
atheon ./test-project
```

### Pattern state not persisting

**Check state file:**
```bash
# Verify state file exists
cat ~/.atheon/pattern_state.json

# Check permissions
ls -la ~/.atheon/
```

**Reset pattern state:**
```bash
# Enable all patterns
atheon enable-all

# Reset to defaults
rm ~/.atheon/pattern_state.json
```

## Development Issues

### Test failures

**Run specific tests:**
```bash
# Run all tests
go test ./...

# Run specific package
go test ./core

# Run with coverage
go test -cover ./...
```

### Build issues

**Clean build:**
```bash
# Clean build artifacts
go clean -cache
go mod tidy

# Rebuild
go build -o atheon .
```

## Getting Help

### Documentation
- [Getting Started](https://github.com/aliasfoxkde/Atheon/wiki/Getting-Started)
- [Configuration Profiles](https://github.com/aliasfoxkde/Atheon/wiki/Configuration-Profiles)
- [Pattern Development](https://github.com/aliasfoxkde/Atheon/wiki/Pattern-Development)

### Community Support
- **Official Project**: [https://github.com/HoraDomu/Atheon/issues](https://github.com/HoraDomu/Atheon/issues)
- **Enhanced Version**: [https://github.com/aliasfoxkde/Atheon/issues](https://github.com/aliasfoxkde/Atheon/issues)

### Debug Mode
```bash
# Enable debug output
atheon --profile config/profiles/development.json ./project

# Verbose mode
atheon ./project --verbose
```

## Common Error Messages

### "bundle load failed"
- Delete local bundle: `rm ~/.atheon/patterns.bundle`
- Run `atheon update`
- Check disk space

### "pattern validation failed"
- Check pattern file syntax
- Verify regex is valid
- Test pattern with simple examples

### "permission denied"
- Check file permissions
- Verify executable permissions
- Run with appropriate user privileges

## Performance Tuning

### For large codebases
```bash
# Use category filtering
atheon --categories=secrets,pii ./large-project

# Use pipeline profile
atheon --profile config/profiles/pipeline.json ./large-project

# Monitor progress
atheon --stats ./large-project
```

### For CI/CD pipelines
```bash
# Fast security scan
atheon --categories=secrets --json . > findings.json

# Fail fast on findings
atheon --categories=secrets --fail-on-error .
```

### For development
```bash
# Comprehensive testing
atheon --profile config/profiles/development.json --all ./project

# Enable debug mode
atheon --profile config/profiles/development.json --stats ./project
```