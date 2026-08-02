# Enhanced Atheon - Troubleshooting Guide

Comprehensive troubleshooting guide for common issues and their solutions.

## 🚨 Quick Fixes

### Most Common Issues

1. **Pattern count shows unexpected value**
   ```bash
   # Remove local bundle cache
   rm ~/.atheon/patterns.bundle

   # Rebuild binary
   go build -o atheon ./cmd/atheon

   # Verify
   ./atheon list | wc -l  # Should show 406 patterns
   ```

2. **Tests failing with category errors**
   ```bash
   # Update test to include new categories
   # Ensure core/pattern_test.go includes:
   # "ai-detection", "devops", "django", "nodejs", "react"
   # The -p 1 flag is MANDATORY: core has package-level state in init().
   go test ./... -p 1 -v
   ```

3. **Build fails with import errors**
   ```bash
   # Clean Go module cache
   go clean -cache -mod

   # Rebuild dependencies
   go mod tidy
   go mod download

   # Rebuild
   go build -o atheon ./cmd/atheon
   ```

## 🔧 Pattern Issues

### Patterns Not Loading

**Symptoms:**
- Pattern count doesn't match expected value
- Missing categories
- Stale pattern results

**Solutions:**

```bash
# Check local bundle cache
ls -la ~/.atheon/

# Remove stale local bundle
rm -f ~/.atheon/patterns.bundle

# Rebuild bundle
./bundler/bundler

# Rebuild binary
go build -o atheon ./cmd/atheon

# Verify patterns
./atheon list categories
# Should show: ai-detection, code-quality, devops, django, finance, healthcare, nodejs, pii, react, secrets

./atheon list | wc -l
# Should show: 406
```

### Pattern Not Matching Expected Content

**Symptoms:**
- Known pattern not detecting expected content
- False positives in results
- Regex compilation errors

**Solutions:**

```bash
# Test pattern directly
./atheon --categories=secrets --file test-file.txt

# Check if pattern is enabled
./atheon list | grep pattern-name

# Enable if disabled
./atheon enable pattern-name

# Test with --all flag
./atheon --all --categories=secrets test-file.txt
```

### Category Filtering Not Working

**Symptoms:**
- Category filter returning all patterns
- Category not recognized

**Solutions:**

```bash
# List available categories
./atheon list categories

# Test category filter
./atheon --categories=secrets,pii .

# Ensure category names are correct
# Note: case-sensitive, use hyphens not spaces
```

## 🐛 Build Issues

### Go Module Errors

**Symptoms:**
- `cannot find package` errors
- `module declares its path` errors
- Version conflicts

**Solutions:**

```bash
# Check go.mod
cat go.mod
# Should show: module github.com/aliasfoxkde/Atheon

# Update dependencies
go get -u ./...
go mod tidy

# Verify module integrity
go mod verify

# Clean and rebuild
go clean -cache
go build -o atheon ./cmd/atheon
```

### Compilation Errors

**Symptoms:**
- Syntax errors during build
- Type mismatches
- Missing imports

**Solutions:**

```bash
# Check Go version (requires 1.21+)
go version

# Run go vet for detailed errors
go vet ./...

# Check formatting
go fmt ./...

# Rebuild
go build -o atheon ./cmd/atheon
```

### Cross-platform Build Issues

**Symptoms:**
- Works on Linux but not macOS/Windows
- Platform-specific errors

**Solutions:**

```bash
# Test on current platform
go build -o atheon ./cmd/atheon

# Use GOOS/GOARCH for cross-compilation
GOOS=linux GOARCH=amd64 go build -o atheon-linux
GOOS=darwin GOARCH=amd64 go build -o atheon-macos
GOOS=windows GOARCH=amd64 go build -o atheon.exe

# Test built binary
./atheon --version
```

## 🧪 Testing Issues

### Test Coverage Below Threshold

**Symptoms:**
- Pre-commit hook fails with coverage error
- Coverage shows below 54.4%

**Solutions:**

```bash
# Check current coverage
# The -p 1 flag is MANDATORY: core has package-level state in init() that
# is not safe under parallel package execution.
go test ./... -p 1 -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total

# Run specific test packages
go test ./core -v -coverprofile=coverage.out

# Add tests for uncovered code
# Focus on core package functions

# Verify improvement
go tool cover -html=coverage.out
```

### Race Detector Failures

**Symptoms:**
- Tests fail with race condition warnings
- Inconsistent test results

**Solutions:**

```bash
# Run with race detector
go test -race ./... -v

# Identify race conditions
# Fix by adding proper synchronization

# Verify fix
go test -race ./... -v
```

### Test Timeout Issues

**Symptoms:**
- Tests timeout after 10 minutes
- Slow test execution

**Solutions:**

```bash
# Run specific test package
go test ./core -v -timeout 15m

# Identify slow tests
go test -v . 2>&1 | grep "PASS\|FAIL"

# Optimize slow tests or increase timeout
# In test files: t.Parallel() for concurrent tests
```

## 🔒 Security Issues

### Permission Denied Errors

**Symptoms:**
- Cannot access certain files
- Permission errors during scan

**Solutions:**

```bash
# Check file permissions
ls -la /path/to/file

# Run with appropriate user permissions
sudo ./atheon /protected/path

# Or exclude protected directories
echo "/protected/path" >> .atheonignore
```

### False Positives in Security Scans

**Symptoms:**
- Legitimate code flagged as security risk
- Too many warnings in output

**Solutions:**

```bash
# Check .atheonignore
cat .atheonignore

# Add false positive patterns
echo "# False positive exclusions" >> .atheonignore
echo "test/" >> .atheonignore
echo "*.test.go" >> .atheonignore

# Re-scan
./atheon . --categories=secrets
```

### Sensitive Data Exposure

**Symptoms:**
- Concerned about sensitive data in scan results
- Logs containing secrets

**Solutions:**

```bash
# Use JSON output with careful handling
./atheon . --json > findings.json

# Process findings securely
# Don't log full output, just summary

# Redact findings
./atheon . --redact
```

## 📊 Performance Issues

### Slow Scanning

**Symptoms:**
- Large codebases take long to scan
- Memory usage high

**Solutions:**

```bash
# Use category filtering
./atheon --categories=secrets,pii /large/codebase

# Use ignore patterns
echo "node_modules/" >> .atheonignore
echo "vendor/" >> .atheonignore

# Scan specific directories
./atheon src/ --categories=secrets
```

### High Memory Usage

**Symptoms:**
- Process using excessive memory
- System slows during scan

**Solutions:**

```bash
# Check memory usage
./atheon . --stats

# Use streaming for large files
# Automatically enabled for files >10MB

# Limit directory depth
find . -maxdepth 3 -type f | ./atheon - --file-list
```

## 🔄 CI/CD Issues

### GitHub Actions Failures

**Symptoms:**
- CI workflow failing
- Platform-specific failures

**Solutions:**

```bash
# Check workflow logs
gh run view --log-failed

# Test locally with same Go version
# The -p 1 flag is MANDATORY: see AUDIT/CHANGELOG notes on core init() state.
go test ./... -p 1 -v -race -coverprofile=coverage.out

# Check for Windows-specific issues
# PowerShell compatibility issues with bash scripts

# Verify all tests pass
go test ./... -p 1 -v
```

### Pre-commit Hook Failures

**Symptoms:**
- Commit blocked by pre-commit hook
- Validation errors

**Solutions:**

```bash
# Check pre-commit hook output
# Hook provides detailed error messages

# Fix common issues:
# 1. Author attribution
git config user.name "Your Name"
git config user.email "your.email@example.com"

# 2. Code formatting
go fmt ./...

# 3. Test coverage
# The -p 1 flag is MANDATORY: see core init() state note above.
go test ./... -p 1 -v -coverprofile=coverage.out

# 4. Static analysis
go vet ./...
```

## 🌐 Network Issues

### Pattern Download Failures

**Symptoms:**
- Cannot download pattern updates
- Network timeout errors

**Solutions:**

```bash
# Check connectivity
ping github.com

# Try manual update
./atheon update

# Use local patterns
# Bundle is embedded, should work without network
./atheon list | wc -l
```

### MCP Connection Issues

**Symptoms:**
- MCP server not responding
- Connection timeouts

**Solutions:**

```bash
# Check MCP server build
go build -o atheon-mcp ./cmd/mcp

# Test MCP server
./atheon-mcp &
# MCP runs on stdio, test with MCP client

# Check for process conflicts
ps aux | grep atheon-mcp
```

## 📦 Installation Issues

### Binary Not Found

**Symptoms:**
- `atheon: command not found`
- Binary not in PATH

**Solutions:**

```bash
# Build binary
go build -o atheon ./cmd/atheon

# Install to PATH
sudo mv atheon /usr/local/bin/

# Or add to PATH
export PATH=$PATH:$(pwd)

# Verify
atheon --version
```

### Go Installation Issues

**Symptoms:**
- `go: command not found`
- Wrong Go version

**Solutions:**

```bash
# Install Go (Linux)
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Set PATH
export PATH=$PATH:/usr/local/go/bin

# Verify
go version
```

## 🆘 Getting Help

### Debug Mode

```bash
# Enable debug output
export ATHEON_DEBUG=1
./atheon . --categories=secrets

# Verbose mode
./atheon . --categories=secrets --verbose
```

### Report Issues

**Include in bug reports:**
1. Atheon version: `atheon --version`
2. Go version: `go version`
3. Operating system: `uname -a`
4. Exact error message
5. Steps to reproduce
6. Expected vs actual behavior

### Useful Commands

```bash
# System information
atheon --version
go version
uname -a

# Pattern information
atheon list categories
atheon list | wc -l

# Test functionality
atheon --help
atheon --categories=secrets --file README.md

# Health check
# The -p 1 flag is MANDATORY: see core init() state note above.
go test ./... -p 1 -v
go vet ./...
```

---

**Troubleshooting Guide Version**: 1.0.0
**Last Updated**: 2026-06-19
**Maintainer**: [@aliasfoxkde](https://github.com/aliasfoxkde) — see [CONTRIBUTORS](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)