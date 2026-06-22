# Atheon Installation Guide

## Quick Installation

### Homebrew (macOS / Linux)

```bash
brew tap aliasfoxkde/Atheon-Enhanced
brew install atheon
```

### Scoop (Windows)

```bash
scoop bucket add atheon https://github.com/aliasfoxkde/scoop-Atheon-Enhanced
scoop install atheon
```

### Manual Installation

1. Download the binary for your platform from [Releases](https://github.com/aliasfoxkde/Atheon-Enhanced/releases/latest)
2. Rename binary to `atheon` (Linux/macOS) or `atheon.exe` (Windows)
3. Move to directory in your PATH

**Linux/macOS:**
```bash
chmod +x atheon
sudo mv atheon /usr/local/bin/
```

**Windows:**
```cmd
move atheon.exe C:\Windows\System32\
```

## Build from Source

### Prerequisites

- Go 1.21 or later
- Git (for cloning repository)

### Build Steps

1. Clone repository:
```bash
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon
```

2. Build binary:
```bash
go build -o atheon .
```

3. (Optional) Install globally:
```bash
sudo cp atheon /usr/local/bin/
```

## Platform-Specific Instructions

### Linux

**Dependencies:**
```bash
# Ubuntu/Debian
sudo apt-get install golang

# Fedora
sudo dnf install golang
```

**Building:**
```bash
go build -o atheon .
```

**Installing:**
```bash
sudo cp atheon /usr/local/bin/
```

### macOS

**Dependencies:**
```bash
# Using Homebrew
brew install go

# Or install from golang.org
```

**Building:**
```bash
go build -o atheon .
```

**Installing:**
```bash
sudo cp atheon /usr/local/bin/
```

**Apple Silicon Note:**
```bash
# Set architecture if needed
export GOARCH=arm64
go build -o atheon .
```

### Windows

**Dependencies:**
- Install Go from https://golang.org/dl/

**Building:**
```cmd
go build -o atheon.exe .
```

**Installing:**
```cmd
copy atheon.exe C:\Windows\System32\
```

**PowerShell Installation:**
```powershell
# Add to PATH via environment variables
# Or copy to System32 as shown above
```

## Verify Installation

After installation, verify Atheon is working:

```bash
atheon --help
atheon list categories
```

Expected output:
```
Atheon is a CLI tool built around a single idea: any pattern, any domain, any input.
...

Usage:
  atheon <path>                        scan a directory or file
  atheon --file <path>                 scan a single file explicitly
  ...
```

## Update Patterns

Atheon ships with an embedded pattern bundle. To update to the latest patterns:

```bash
atheon update
```

This downloads the latest patterns.bundle from GitHub releases.

## MCP Server Installation

### Install from Releases

The `atheon-mcp` binary is included in all Atheon releases. It's packaged as a separate binary for MCP integration.

### Build MCP Server Only

```bash
go build -o atheon-mcp ./cmd/mcp
```

### Verify MCP Server

```bash
atheon-mcp
```

The server should start and be ready for MCP connections.

## Troubleshooting

### Build Errors

**Go version too old:**
```bash
# Check Go version
go version
# Update if needed
```

**Missing dependencies:**
```bash
go mod download
```

### Permission Issues

**Linux/macOS:**
```bash
chmod +x atheon
```

**Windows:**
- Run as Administrator if moving to System32
- Or add to user PATH instead

### Pattern Issues

**Patterns not found:**
```bash
# Update patterns bundle
atheon update

# Or rebuild from community patterns
go run ./bundler community core/patterns.bundle
```

## Development Installation

For development work, you may want to install from source and create a symlink:

```bash
# Clone repository
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon

# Build and create symlink
go build -o atheon
ln -sf $(pwd)/atheon /usr/local/bin/atheon

# For development updates
export PATH=$PATH:$(pwd):$PATH
```

## Uninstall

### Homebrew
```bash
brew uninstall atheon
brew untap aliasfoxkde/Atheon-Enhanced
```

### Scoop
```bash
scoop uninstall atheon
```

### Manual
```bash
# Linux/macOS
sudo rm /usr/local/bin/atheon

# Windows
del C:\Windows\System32\atheon.exe
```

## Next Steps

After installation:
1. Read [CONTRIBUTING.md](CONTRIBUTING.md) to add patterns
2. Try the tool: `atheon list categories`
3. Scan your code: `atheon ./`
4. Set up pre-commit hooks: [Pre-commit Hooks Guide](docs/PRE_COMMIT_HOOKS.md)