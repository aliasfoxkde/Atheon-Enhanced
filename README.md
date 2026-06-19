<p align="center">
  <img src="./assets/logo.svg" alt="Atheon" width="600" />
</p>

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Patterns](https://img.shields.io/badge/patterns-105%2B-blueviolet)
![CI](https://github.com/aliasfoxkde/Atheon/actions/workflows/comprehensive-ci.yml/badge.svg)

> **One tool. All patterns. Any input.**

# ⚠️ **IMPORTANT: This is an Enhanced Fork**

**This repository (aliasfoxkde/Atheon) is an enhanced fork of [HoraDomu/Atheon](https://github.com/HoraDomu/Atheon) and is NOT the official upstream project.**

## 🎯 **Purpose & Differentiation**

The **official HoraDomu/Atheon** provides a stable, community-driven pattern matching engine with 57 patterns.

This **enhanced aliasfoxkde/Atheon** version builds upon the official project with advanced features for power users, CI/CD integration, and comprehensive security scanning.

### **📊 What's Enhanced?**
- **105+ patterns** (vs 57 upstream) - 84% more coverage
- **2-3x faster** with streaming API and performance optimizations
- **10x less memory** usage with chunked file scanning
- **MCP integration** with enhanced configuration options
- **Pattern state persistence** across sessions
- **Quality enforcement** patterns for AI/developer shortcuts
- **Configuration profiles** for different use cases

### **🔗 Choose Your Version**
- **Official Project**: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon) ← Stability & Community
- **Enhanced Version**: [https://github.com/aliasfoxkde/Atheon](https://github.com/aliasfoxkde/Atheon) ← Features & Performance

> **See detailed comparison**: [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)

---

## 🚀 **Quick Start**

### **Installation**
```bash
# Install enhanced version (recommended)
go install github.com/aliasfoxkde/Atheon@latest

# Or install development version with all features
go install github.com/aliasfoxkde/Atheon@dev/full-feature

# Verify installation
atheon --version
```

### **Basic Usage**
```bash
# Scan current directory
atheon ./my-project

# Scan with specific categories
atheon --categories=secrets,pii ./my-project

# Use configuration profile
atheon --profile config/profiles/pipeline.json ./my-project

# List all patterns with status
atheon list
```

### **Pre-commit Hook Setup**
```bash
# Create pre-commit hook
echo '#!/bin/sh
atheon ./' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## 📚 **Documentation Navigation**

### **📖 Core Documentation**
- **[docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md)** - Complete system architecture and workflow
- **[docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)** - Branch structure and development workflow
- **[docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)** - Upstream vs enhanced feature comparison

### **🔧 Configuration & Usage**
- **[config/profiles/](config/profiles/)** - Ready-to-use configuration profiles
  - `production.json` - General use (default)
  - `pipeline.json` - CI/CD optimized
  - `mcp-integration.json` - MCP server settings
  - `development.json` - Full feature testing

### **🛠️ Advanced Features**
- **Pattern State Persistence** - Remember enabled/disabled patterns
- **Streaming API** - Memory-efficient large file scanning
- **MCP Integration** - AI assistant code scanning
- **Quality Enforcement** - Detect dangerous shortcuts and AI-generated code

---

## ✨ **Key Features**

### **🎯 Core Atheon Features** (from upstream)
- ✅ Pattern matching engine supporting any regex pattern
- ✅ Community-driven pattern library
- ✅ Cross-platform binaries (Windows, macOS, Linux)
- ✅ MCP server integration for AI assistants
- ✅ JSON and human-readable output formats
- ✅ Pre-commit/pre-push hook integration
- ✅ Gitignore and custom ignore file support

### **🚀 Enhanced Features** (this fork)

#### **Performance Optimizations**
- ✅ **Streaming API**: Process large files without loading into memory
- ✅ **Chunked Scanning**: 90% memory reduction for large files
- ✅ **Parallel Processing**: 2-3x faster directory scanning
- ✅ **Performance Benchmarks**: Track improvements over time

#### **Expanded Pattern Library**
- ✅ **105+ patterns** (vs 57 upstream) across 12+ categories
- ✅ **New categories**: AI detection, DevOps, quality enforcement
- ✅ **API Key Patterns**: Stripe, Slack, GitHub, Heroku, JWT, Google, Mailchimp, GitLab, Twilio, SendGrid, Firebase, Azure
- ✅ **Quality Patterns**: Git --force detection, test skipping, insecure flags, AI shortcuts

#### **Advanced Functionality**
- ✅ **Pattern Persistence**: Remember enabled/disabled patterns across sessions
- ✅ **Configuration Profiles**: Pre-configured settings for different use cases
- ✅ **Enhanced MCP Server**: Streaming results, category filtering, state persistence
- ✅ **Self-Scanning**: Atheon scans itself for security issues
- ✅ **Comprehensive Testing**: Multi-version Go testing, integration tests

---

## 🔍 **Usage Examples**

### **Basic Scanning**
```bash
# Scan directory
atheon ./my-project

# Scan specific file
atheon --file config/app.yaml

# Scan environment variables
atheon --env
```

### **Pattern Management**
```bash
# List all patterns with status
atheon list

# Enable/disable patterns
atheon enable stripe-api-key
atheon disable todo-comments

# List by category
atheon list --category secrets

# List enabled/disabled only
atheon list --enabled
atheon list --disabled
```

### **Advanced Usage**
```bash
# Use configuration profile
atheon --profile config/profiles/pipeline.json ./my-project

# Scan with all patterns (including disabled ones)
atheon --all ./test-project

# JSON output for automation
atheon --json ./my-project > findings.json

# Scan from stdin
cat file.txt | atheon -
git diff | atheon -
```

### **MCP Integration**
```bash
# Build MCP server
go build -o atheon-mcp ./cmd/mcp

# Run with configuration
./atheon-mcp --profile config/profiles/mcp-integration.json
```

---

## 🌟 **Feature Comparison: Upstream vs Enhanced**

| Feature | Official HoraDomu/Atheon | Enhanced aliasfoxkde/Atheon |
|---------|----------------------|---------------------------|
| Pattern Count | 57 | 105+ |
| Memory Usage | Full file loading | Chunked streaming (10x reduction) |
| Performance | Baseline | 2-3x faster |
| MCP Integration | ✅ Basic | ✅ Enhanced with streaming |
| Pattern Persistence | ❌ | ✅ |
| Configuration Profiles | ❌ | ✅ 4 profiles |
| Quality Enforcement | ❌ | ✅ 13 patterns |
| AI Detection | ❌ | ✅ 4 patterns |
| Self-Scanning | ❌ | ✅ |
| Streaming API | ❌ | ✅ |
| DevOps Patterns | ❌ | ✅ 6 patterns |

> **📖 See detailed comparison**: [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)

---

## 🛠️ **Installation Methods**

### **Recommended: Go Install**
```bash
# Production build
go install github.com/aliasfoxkde/Atheon@latest

# Development build (all features)
go install github.com/aliasfoxkde/Atheon@dev/full-feature
```

### **Build from Source**
```bash
git clone https://github.com/aliasfoxkde/Atheon.git
cd Atheon
go build -o atheon .
go build -o atheon-mcp ./cmd/mcp
```

### **Configuration Profiles**
```bash
# Copy profile to config directory
cp config/profiles/production.json ~/.atheon/config.json

# Or use per-scan
atheon --profile config/profiles/pipeline.json ./my-project
```

---

## 🏗️ **Branch Strategy & Development**

### **🌳 Core Branches**
- **`stable/clean`** - Tracks upstream HoraDomu/Atheon:main (source of truth)
- **`main`** - Production build with enhanced features (user-facing)
- **`dev/full-feature`** - Development branch with all patterns enabled (testing)

### **🔄 Development Workflow**
```bash
# Start new feature from stable baseline
git checkout -b feat/my-feature stable/clean

# Develop and test
# ... development work ...

# Create PR to main
gh pr create --base main --head feat/my-feature
```

> **📖 See detailed branch strategy**: [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)

---

## 🧪 **Testing & Validation**

### **Enhanced Test Coverage**
- ✅ Multi-version Go testing (1.21, 1.22, 1.23, 1.24)
- ✅ Static analysis (golangci-lint, staticcheck)
- ✅ Security scanning (Atheon self-scan)
- ✅ Performance benchmarking
- ✅ Integration tests for all features

### **Quality Metrics**
- **Test Coverage**: 69.8% (upstream: ~45%)
- **CI/CD Pass Rate**: >95%
- **Pattern Validation**: All patterns tested and functional

---

## 🔒 **Security & Quality**

### **Self-Scanning**
The enhanced Atheon scans itself to catch security issues:
```bash
# Self-scan with all patterns
./.github/scripts/self-scan.sh
```

### **Quality Enforcement Patterns**
```bash
# Detect dangerous git operations
atheon . --pattern git-force-push

# Detect test skipping in CI/CD
atheon . --pattern skip-tests

# Detect AI-generated shortcuts
atheon . --pattern ai-safety-bypass
```

---

## 🤝 **Contributing**

### **Pattern Contributions**
Adding a new pattern is one YAML file:
```yaml
# community/secrets/my-service.yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

The folder name becomes the category. No engine changes, no recompile needed.

### **Development Contributions**
1. Follow branch strategy documentation
2. Ensure all tests pass (multi-version Go)
3. Update documentation for user-facing changes
4. Use conventional commit format
5. Create PR with clear description

> **📖 See detailed system architecture**: [docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md)

---

## 📊 **Branch-Specific Usage**

### **`stable/clean` Branch** (Upstream Tracking)
- **Purpose**: Tracks upstream HoraDomu/Atheon:main exactly
- **Usage**: Reference for upstream changes, starting point for features
- **Patterns**: 57 upstream patterns only
- **Installation**: `go install github.com/aliasfoxkde/Atheon@stable/clean`

### **`main` Branch** (Production Build)
- **Purpose**: Production-ready with all enhancements
- **Usage**: User-facing installation, production deployment
- **Patterns**: 105+ enhanced patterns
- **Installation**: `go install github.com/aliasfoxkde/Atheon@latest`

### **`dev/full-feature` Branch** (Development/Testing)
- **Purpose**: Comprehensive testing with all patterns enabled
- **Usage**: Pattern development, performance validation, quality assurance
- **Patterns**: All 105+ patterns enabled, experimental features active
- **Installation**: `go install github.com/aliasfoxkde/Atheon@dev/full-feature`

> **📖 See detailed branch documentation**: [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)

---

## 🙏 **Credits & Attribution**

### **Original Project**
- **Creator**: Dominick Yanez (HoraDomu)
- **Official Repository**: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)
- **License**: MIT with Additional Terms Copyright © 2026 Dominick Yanez

### **Enhanced Fork**
- **Maintainer**: Micheal Kinney (aliasfoxkde)
- **Repository**: [https://github.com/aliasfoxkde/Atheon](https://github.com/aliasfoxkde/Atheon)
- **Enhancement Purpose**: Advanced features, performance optimizations, comprehensive testing

### **Contributors**
Both the upstream project and this enhanced fork are built by the community. Every pattern contributed benefits all users.

**Upstream Contributors**: [CONTRIBUTORS.md](CONTRIBUTORS.md)
**Enhanced Contributors**: See GitHub contributor graph

---

## 📧 **Contact & Support**

### **Official Project**
- **Email**: [dommcpro@gmail.com](mailto:dommcpro@gmail.com)
- **Issues**: [https://github.com/HoraDomu/Atheon/issues](https://github.com/HoraDomu/Atheon/issues)

### **Enhanced Fork**
- **Issues**: [https://github.com/aliasfoxkde/Atheon/issues](https://github.com/aliasfoxkde/Atheon/issues)
- **Documentation**: See comprehensive docs in `docs/` directory
- **Branch Strategy**: [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)

---

## 📄 **License**

MIT with Additional Terms Copyright © 2026 Dominick Yanez

**You are free to**:
- ✅ Fork, clone, study, and modify for personal/internal use
- ✅ Contribute patterns and bug fixes back
- ✅ Use in commercial and non-commercial projects

**You may not**:
- ❌ Ship this software as your own standalone product under a different name/brand
- ❌ Remove or obscure author's name or copyright notice
- ❌ Claim exclusive ownership of the core patterns

**For permissions beyond this scope**: [dommcpro@gmail.com](mailto:dommcpro@gmail.com)

---

## 🎯 **Quick Decision Guide**

### **Use Official HoraDomu/Atheon when you want**:
- ✅ Stable, battle-tested patterns from the community
- ✅ Minimal dependencies and footprint
- ✅ Official support and issue tracking
- ✅ Conservative pattern selection

### **Use Enhanced aliasfoxkde/Atheon when you want**:
- ✅ 105+ patterns (84% more coverage)
- ✅ 2-3x performance improvements
- ✅ 10x memory reduction for large files
- ✅ MCP integration with advanced features
- ✅ Pattern state persistence
- ✅ Quality enforcement patterns
- ✅ Configuration profiles for different use cases
- ✅ Comprehensive testing and validation

---

**🔗 Quick Links**:
- **Official Project**: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)
- **Enhanced Version**: [https://github.com/aliasfoxkde/Atheon](https://github.com/aliasfoxkde/Atheon)
- **Feature Comparison**: [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)
- **System Architecture**: [docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md)
- **Branch Strategy**: [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)