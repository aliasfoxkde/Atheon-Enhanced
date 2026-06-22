# Installation Instructions for Atheon-Enhanced

## 🚀 Go Install Method (Recommended)

```bash
# Install latest version
go install github.com/aliasfoxkde/Atheon-Enhanced/cmd/atheon@latest

# Install MCP server (optional)
go install github.com/aliasfoxkde/Atheon-Enhanced/cmd/mcp@latest
```

> **Note:** The Go module path is `github.com/aliasfoxkde/Atheon` (no `-Enhanced` suffix).
> Use the GitHub repo URL above for cloning and releases.

## 🛠️ Build from Source

```bash
# Clone repository
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon-Enhanced

# Build binaries
go build -o atheon ./cmd/atheon
go build -o atheon-mcp ./cmd/mcp

# Install to PATH
sudo mv atheon /usr/local/bin/
sudo mv atheon-mcp /usr/local/bin/

# Install pre-commit hook (recommended for contributors)
git config core.hooksPath hooks
```

## ✅ Verify Installation

```bash
# Check version
atheon --version

# Check pattern count (should be 190+)
atheon list | wc -l

# List categories
atheon list categories
```

## 🧪 Quick Test

```bash
# Test basic functionality
atheon README.md --categories=secrets

# List all patterns
atheon list

# Scan current directory
atheon . --categories=ai-detection,secrets
```

## 📚 Documentation

- **Complete Documentation**: https://github.com/aliasfoxkde/Atheon-Enhanced/tree/main/docs
- **API Reference**: https://github.com/aliasfoxkde/Atheon-Enhanced/blob/main/docs/api/README.md
- **Development Guide**: https://github.com/aliasfoxkde/Atheon-Enhanced/blob/main/docs/development/SETUP.md
- **Troubleshooting**: https://github.com/aliasfoxkde/Atheon-Enhanced/blob/main/docs/guides/TROUBLESHOOTING.md

## 📦 Package Information

- **Module**: github.com/aliasfoxkde/Atheon
- **Go Version**: 1.21+
- **Patterns**: 190 total patterns
- **Categories**: 19 categories
- **Platforms**: Linux, macOS, Windows

## 🔧 Platform-Specific Notes

### Linux
```bash
# Install Go dependencies
sudo apt install golang-go

# Build and install
go build -o atheon ./cmd/atheon
sudo mv atheon /usr/local/bin/
```

### macOS
```bash
# Install Go using Homebrew
brew install go

# Build and install
go build -o atheon ./cmd/atheon
sudo mv atheon /usr/local/bin/
```

### Windows
```powershell
# Install Go from https://go.dev/dl/

# Build
go build -o atheon.exe ./cmd/atheon

# Add to PATH or move to a directory already in PATH
```

## ⚡ After Installation

### Remove Local Pattern Cache (if upgrading)
```bash
rm -f ~/.atheon/patterns.bundle
```

### Verify Pattern Loading
```bash
# Should show 190+ patterns
atheon list | wc -l

# Should show 19 categories
atheon list categories
```

## 🎯 Next Steps

1. **Configure Atheon** for your needs
2. **Explore patterns**: `atheon list` to see all 190 patterns
3. **Set up CI/CD integration**: see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
4. **Contribute patterns**: see [CONTRIBUTING.md](CONTRIBUTING.md)

---

**Repository**: https://github.com/aliasfoxkde/Atheon-Enhanced  
**Maintainer**: Michael Kinney (aliasfoxkde)
