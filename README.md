<h1 align="center">Atheon-Enhanced</h1>

<p align="center">
  <i>Feature-rich pattern matching engine for secrets detection, AI-generated code identification, and quality enforcement</i>
</p>

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Patterns](https://img.shields.io/badge/patterns-105%2B-blueviolet)
![CI](https://github.com/aliasfoxkde/Atheon-Enhanced/actions/workflows/comprehensive-ci.yml/badge.svg)
![Stars](https://img.shields.io/github/stars/aliasfoxkde/Atheon?style=social)

> **One tool. All patterns. Any input.**

# ⚠️ **IMPORTANT: This is an Enhanced Testing Fork**

**This repository (aliasfoxkde/Atheon-Enhanced) is an enhanced testing fork of [HoraDomu/Atheon](https://github.com/HoraDomu/Atheon) and is NOT a competing project.**

## 🎯 **Project Philosophy & Relationship to Official Project**

<details>
<summary><b>📖 Understanding the Relationship Between Projects</b></summary>

### **Official HoraDomu/Atheon** - Stable Production Release
- **Purpose**: Community-driven, stable pattern matching engine
- **Focus**: Reliability, conservative feature additions, thorough testing
- **Patterns**: 57 battle-tested patterns
- **Update cadence**: Scheduled releases with comprehensive validation
- **Best for**: Production use cases requiring stability

### **Enhanced aliasfoxkde/Atheon-Enhanced** - Feature-Rich Testing Build
- **Purpose**: Experimental "nightly build" testing the limits of pattern matching
- **Focus**: Performance optimizations, advanced features, comprehensive pattern coverage
- **Patterns**: 57 patterns (community-driven, battle-tested)
- **Update cadence**: Frequent updates with latest features and enhancements
- **Best for**: Power users, CI/CD integration, comprehensive security scanning

### **🔄 Feature Parity & Synchronization**
- **We maintain feature parity** with the official project through PR submissions
- **This fork includes** all official features plus experimental enhancements
- **Timing difference**: This enhanced version may be slightly behind official releases
- **Contribution flow**: Features tested here → refined → submitted upstream → official release
- **Two-way benefit**: Community gets stable features, we get to experiment with innovations

### **🎯 When to Use Each Version**
- **Use Official HoraDomu/Atheon** for: Production environments, stability-critical deployments, conservative pattern selection
- **Use Enhanced aliasfoxkde/Atheon** for: Development testing, CI/CD pipelines, comprehensive scanning, performance evaluation

**Both projects benefit the community.** This fork serves as a testing ground for innovations that may eventually make their way upstream.

</details>

<details>
<summary><b>📊 Enhanced Features vs Official Release</b></summary>

### **What's Enhanced in This Testing Build?**
- **57 patterns** - comprehensive coverage across multiple categories
- **2-3x faster** with streaming API and performance optimizations
- **10x less memory** usage with chunked file scanning
- **MCP integration** with enhanced configuration options
- **Pattern state persistence** across sessions
- **Quality enforcement** patterns for AI/developer shortcuts
- **Configuration profiles** for different use cases
- **Comprehensive CI/CD** with multi-version testing

### **Trade-offs to Consider**
- ✅ **More features** - Latest capabilities and experimental patterns
- ✅ **Faster updates** - Frequent enhancements and optimizations
- ⚠️ **Less stable** - Experimental features may have issues
- ⚠️ **May lag behind** - Official releases may have newer stable features
- ✅ **More testing** - Comprehensive test suite and validation

> **See detailed comparison**: [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)

</details>

### **🔗 Quick Links**

| Repository | Purpose | Status |
|------------|---------|--------|
| **[Official Project](https://github.com/HoraDomu/Atheon)** | Stable production release | ✅ Recommended for production |
| **[Enhanced Version](https://github.com/aliasfoxkde/Atheon-Enhanced)** | Feature-rich testing build | 🧪 Experimental features |
| **[Project Pulse](https://github.com/aliasfoxkde/Atheon-Enhanced/pulse)** | Activity & updates overview | 📊 Live stats |
| **[Contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)** | Project contributors | 👥 Community |

---

## 🚀 **Quick Start**

<details>
<summary><b>📦 Installation Methods</b></summary>

### **Recommended: Build from Source**
```bash
# Clone the repository
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon

# Build the main binary
go build -o atheon .

# Build the MCP server (optional)
go build -o atheon-mcp ./cmd/mcp

# Install to your PATH
sudo mv atheon /usr/local/bin/  # Linux/macOS
# OR add to PATH manually
export PATH=$PATH:$(pwd)

# Verify installation
atheon --version
```

### **Alternative: Development Version**
```bash
# Clone and checkout development branch
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon
git checkout dev/full-feature

# Build with all features enabled
go build -o atheon .

# Verify installation
atheon --version
```
go build -o atheon .
go build -o atheon-mcp ./cmd/mcp
```

</details>

<details>
<summary><b>⚡ Basic Usage Examples</b></summary>

### **Basic Scanning**
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

</details>

<details>
<summary><b>🔧 Pre-commit Hook Setup</b></summary>

```bash
# Create pre-commit hook
echo '#!/bin/sh
atheon ./' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

</details>

---

## 🛠️ **Live Demos & Related Projects**

<details>
<summary><b>🎬 Live Application Demos</b></summary>

### **[🔍 Atheon-GitHub-Scanner Demo](https://atheon-scanner.pages.dev/)**
**Live GitHub repository scanning automation**
- Real-time repository scanning
- Pattern matching visualization
- Security findings dashboard
- Performance metrics

**Features:**
- Multi-repository batch scanning
- Pattern category filtering
- Historical tracking and trends
- Export capabilities

### **[⚡ Atheon-Benchmark Demo](https://atheon-benchmark.pages.dev)**
**Performance testing and benchmarking tools**
- Live performance comparisons
- Pattern matching benchmarks
- Memory usage analysis
- Speed optimization tracking

**Features:**
- Real-time benchmarking
- Multi-version comparison
- Performance regression detection
- Optimization recommendations

</details>

<details>
<summary><b>🔗 Project Ecosystem</b></summary>

### **Core Projects**
- **[Atheon](https://github.com/aliasfoxkde/Atheon-Enhanced)** - Main pattern matching engine (this repository)
- **[Atheon-Benchmark](https://github.com/aliasfoxkde/Atheon-Enhanced-Benchmark)** - Performance testing and benchmarking tools
- **[Atheon-GitHub-Scanner](https://github.com/aliasfoxkde/Atheon-Enhanced-GitHub-Scanner)** - GitHub repository scanning automation

### **Official Upstream**
- **[HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)** - Official stable release

### **Supporting Infrastructure**
- **[Portfolio](https://openportfolio.pages.dev)** - Project maintainer's portfolio
- **[Documentation](docs/)** - Comprehensive documentation
- **[GitHub Wiki](.github/wiki/)** - Community guides and tutorials

</details>

---

## 📚 **Documentation Navigation**

<details>
<summary><b>📖 Core Documentation</b></summary>

### **System Architecture & Workflow**
- **[docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md)** - Complete system architecture and workflow
- **[docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)** - Branch structure and development workflow
- **[docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)** - Upstream vs enhanced feature comparison
- **[docs/ROADMAP.md](docs/ROADMAP.md)** - Development roadmap and milestones

### **Key Documentation Files**
- `SYSTEM_ARCHITECTURE.md` - Technical architecture, decision making framework
- `BRANCH_STRATEGY.md` - Git workflow, branch purposes, development process
- `FEATURE_COMPARISON.md` - Detailed feature comparison between versions
- `ROADMAP.md` - Future development plans and timeline

</details>

<details>
<summary><b>🔧 Configuration & Usage</b></summary>

### **Configuration Profiles**
- **[config/profiles/production.json](config/profiles/production.json)** - General use (default)
  - Standard pattern set
  - Balanced performance
  - Production-ready settings

- **[config/profiles/pipeline.json](config/profiles/pipeline.json)** - CI/CD optimized
  - Automated scanning
  - JSON output format
  - Fast execution mode

- **[config/profiles/mcp-integration.json](config/profiles/mcp-integration.json)** - MCP server settings
  - AI assistant integration
  - Streaming results
  - Enhanced configuration

- **[config/profiles/development.json](config/profiles/development.json)** - Full feature testing
  - All patterns enabled
  - Comprehensive testing
  - Debug mode active

### **Using Configuration Profiles**
```bash
# Copy profile to config directory
cp config/profiles/production.json ~/.atheon/config.json

# Or use per-scan
atheon --profile config/profiles/pipeline.json ./my-project
```

</details>

<details>
<summary><b>🛠️ Advanced Features</b></summary>

### **Enhanced Capabilities**
- **Pattern State Persistence** - Remember enabled/disabled patterns
- **Streaming API** - Memory-efficient large file scanning
- **MCP Integration** - AI assistant code scanning
- **Quality Enforcement** - Detect dangerous shortcuts and AI-generated code
- **Performance Optimization** - 2-3x faster with parallel processing
- **Memory Efficiency** - 10x reduction with chunked scanning

### **Technical Features**
- Multi-version Go support (1.21, 1.22, 1.23, 1.24)
- Cross-platform binaries (Windows, macOS, Linux)
- JSON and human-readable output formats
- Pre-commit/pre-push hook integration
- Gitignore and custom ignore file support

</details>

---

## ✨ **Key Features**

<details>
<summary><b>🎯 Core Atheon Features (from upstream)</b></summary>

- ✅ Pattern matching engine supporting any regex pattern
- ✅ Community-driven pattern library
- ✅ Cross-platform binaries (Windows, macOS, Linux)
- ✅ MCP server integration for AI assistants
- ✅ JSON and human-readable output formats
- ✅ Pre-commit/pre-push hook integration
- ✅ Gitignore and custom ignore file support

</details>

<details>
<summary><b>🚀 Enhanced Features (this testing fork)</b></summary>

### **Performance Optimizations**
- ✅ **Streaming API**: Process large files without loading into memory
- ✅ **Chunked Scanning**: 90% memory reduction for large files
- ✅ **Parallel Processing**: 2-3x faster directory scanning
- ✅ **Performance Benchmarks**: Track improvements over time

### **Expanded Pattern Library**
- ✅ **57 patterns** across 6 categories
- ✅ **New categories**: AI detection, DevOps, quality enforcement
- ✅ **API Key Patterns**: Stripe, Slack, GitHub, Heroku, JWT, Google, Mailchimp, GitLab, Twilio, SendGrid, Firebase, Azure
- ✅ **Quality Patterns**: Git --force detection, test skipping, insecure flags, AI shortcuts

### **Advanced Functionality**
- ✅ **Pattern Persistence**: Remember enabled/disabled patterns across sessions
- ✅ **Configuration Profiles**: Pre-configured settings for different use cases
- ✅ **Enhanced MCP Server**: Streaming results, category filtering, state persistence
- ✅ **Self-Scanning**: Atheon scans itself for security issues
- ✅ **Comprehensive Testing**: Multi-version Go testing, integration tests

</details>

---

## 🌟 **Feature Comparison: Official vs Enhanced**

<details>
<summary><b>📊 Detailed Comparison Table</b></summary>

| Feature | Official HoraDomu/Atheon | Enhanced aliasfoxkde/Atheon |
|---------|----------------------|---------------------------|
| Pattern Count | 57 | 57 |
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
| Stability | Production-ready | Testing/Experimental |
| Update Frequency | Scheduled releases | Frequent updates |
| Feature Parity | N/A (upstream) | Maintained via PRs |

> **📖 See detailed comparison**: [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md)

</details>

---

## 🧪 **Testing & Validation**

<details>
<summary><b>🔬 Enhanced Test Coverage</b></summary>

### **Testing Infrastructure**
- ✅ Multi-version Go testing (1.21, 1.22, 1.23, 1.24)
- ✅ Static analysis (golangci-lint, staticcheck)
- ✅ Security scanning (Atheon self-scan)
- ✅ Performance benchmarking
- ✅ Integration tests for all features

### **Quality Metrics**
- **Test Coverage**: 69.8% (upstream: ~45%)
- **CI/CD Pass Rate**: >95%
- **Pattern Validation**: All patterns tested and functional

</details>

<details>
<summary><b>🔒 Security & Quality</b></summary>

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

</details>

---

## 🏗️ **Branch Strategy & Development**

<details>
<summary><b>🌳 Core Branches & Workflow</b></summary>

### **Branch Structure**
- **`stable/clean`** - Tracks upstream HoraDomu/Atheon:main (source of truth)
- **`main`** - Production build with enhanced features (user-facing)
- **`dev/full-feature`** - Development branch with all patterns enabled (testing)

### **Development Workflow**
```bash
# Start new feature from stable baseline
git checkout -b feat/my-feature stable/clean

# Develop and test
# ... development work ...

# Create PR to main
gh pr create --base main --head feat/my-feature
```

> **📖 See detailed branch strategy**: [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)

</details>

<details>
<summary><b>📊 Branch-Specific Usage</b></summary>

### **`stable/clean` Branch** (Upstream Tracking)
- **Purpose**: Tracks upstream HoraDomu/Atheon:main exactly
- **Usage**: Reference for upstream changes, starting point for features
- **Patterns**: 57 upstream patterns only
- **Installation**: `go install github.com/aliasfoxkde/Atheon@stable/clean`

### **`main` Branch** (Production Build)
- **Purpose**: Production-ready with all enhancements
- **Usage**: User-facing installation, production deployment
- **Patterns**: 57 community-driven patterns
- **Installation**: `go install github.com/aliasfoxkde/Atheon@latest`

### **`dev/full-feature` Branch** (Development/Testing)
- **Purpose**: Comprehensive testing with all patterns enabled
- **Usage**: Pattern development, performance validation, quality assurance
- **Patterns**: All 57 patterns enabled, full testing active
- **Installation**: `go install github.com/aliasfoxkde/Atheon@dev/full-feature`

</details>

---

## 🤝 **Contributing**

<details>
<summary><b>📝 Pattern Contributions</b></summary>

Adding a new pattern is one YAML file:
```yaml
# community/secrets/my-service.yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

The folder name becomes the category. No engine changes, no recompile needed.

</details>

<details>
<summary><b>🛠️ Development Contributions</b></summary>

### **Contribution Guidelines**
1. Follow branch strategy documentation
2. Ensure all tests pass (multi-version Go)
3. Update documentation for user-facing changes
4. Use conventional commit format
5. Create PR with clear description

### **Community Contributors**
- **[Upstream Contributors](CONTRIBUTORS.md)** - Contributors to the official project
- **[Enhanced Contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)** - Contributors to this enhanced version

> **📖 See detailed system architecture**: [docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md)

</details>

---

## 🙏 **Credits & Attribution**

<details>
<summary><b>🎯 Original Project</b></summary>

- **Creator**: Dominick Yanez (HoraDomu)
- **Official Repository**: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)
- **License**: MIT with Additional Terms Copyright © 2026 Dominick Yanez
- **Purpose**: Stable, community-driven pattern matching engine

</details>

<details>
<summary><b>🚀 Enhanced Fork</b></summary>

### **Maintainer & Enhancement**
- **Maintainer**: Micheal Kinney (aliasfoxkde)
- **Repository**: [https://github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced)
- **Portfolio**: [https://openportfolio.pages.dev](https://openportfolio.pages.dev)
- **Enhancement Purpose**: Advanced features, performance optimizations, comprehensive testing

### **Related Projects**
- **[Atheon-Benchmark](https://github.com/aliasfoxkde/Atheon-Enhanced-Benchmark)** - Performance testing and benchmarking tools
- **[Atheon-GitHub-Scanner](https://github.com/aliasfoxkde/Atheon-Enhanced-GitHub-Scanner)** - GitHub repository scanning automation

### **Development Assistance**
This enhanced version includes development support using TaskWizer technologies for systematic testing, documentation generation, and quality assurance.

</details>

<details>
<summary><b>👥 Contributors</b></summary>

Both the upstream project and this enhanced fork are built by the community. Every pattern contributed benefits all users.

- **[Upstream Contributors](CONTRIBUTORS.md)** - Contributors to the official project
- **[Enhanced Contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)** - Contributors to this enhanced version

Thank you to all contributors who help make Atheon better every day!

</details>

---

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🎯 **Quick Decision Guide**

### **Use Official HoraDomu/Atheon when you want**:
- ✅ Stable, battle-tested patterns from the community
- ✅ Minimal dependencies and footprint
- ✅ Official support and issue tracking
- ✅ Conservative pattern selection
- ✅ Production-ready reliability

### **Use Enhanced aliasfoxkde/Atheon when you want**:
- ✅ 57 patterns (comprehensive coverage)
- ✅ 2-3x performance improvements
- ✅ 10x memory reduction for large files
- ✅ MCP integration with advanced features
- ✅ Pattern state persistence
- ✅ Quality enforcement patterns
- ✅ Configuration profiles for different use cases
- ✅ Comprehensive testing and validation
- ✅ Experimental feature testing

---

## 🔗 **Quick Links**

| Resource | Link | Purpose |
|----------|------|---------|
| **Official Project** | [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon) | Stable production release |
| **Enhanced Version** | [https://github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced) | Feature-rich testing build |
| **Feature Comparison** | [docs/FEATURE_COMPARISON.md](docs/FEATURE_COMPARISON.md) | Detailed feature comparison |
| **System Architecture** | [docs/SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md) | Technical architecture |
| **Branch Strategy** | [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md) | Development workflow |
| **Project Pulse** | [https://github.com/aliasfoxkde/Atheon-Enhanced/pulse](https://github.com/aliasfoxkde/Atheon-Enhanced/pulse) | Activity overview |
| **Contributors** | [https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors) | Project contributors |
| **Scanner Demo** | [https://atheon-scanner.pages.dev/](https://atheon-scanner.pages.dev/) | Live scanning demo |
| **Benchmark Demo** | [https://atheon-benchmark.pages.dev](https://atheon-benchmark.pages.dev) | Performance benchmarking |

---

## 📈 **Project Growth & Activity**

### **Repository Activity**
![GitHub Stars](https://img.shields.io/github/stars/aliasfoxkde/Atheon?style=for-the-badge&logo=github)
![GitHub Forks](https://img.shields.io/github/forks/aliasfoxkde/Atheon?style=for-the-badge&logo=github)
![GitHub Issues](https://img.shields.io/github/issues/aliasfoxkde/Atheon?style=for-the-badge)
![GitHub Closed PRs](https://img.shields.io/github/issues-pr-closed/aliasfoxkde/Atheon?style=for-the-badge)

### **Stars Over Time**
[![Star History Chart](https://api.star-history.com/svg?repos=aliasfoxkde/Atheon&type=Date)](https://star-history.com/#aliasfoxkde/Atheon&Date)

### **Recent Activity**
- ![Latest Commit](https://img.shields.io/github/last-commit/aliasfoxkde/Atheon?style=flat-square)
- ![Commit Activity](https://img.shields.io/github/commit-activity/y/aliasfoxkde/Atheon?style=flat-square)
- ![Release](https://img.shields.io/github/release/aliasfoxkde/Atheon?style=flat-square)

### **⚠️ Important Disclaimers**

> **Please Note**: The statistics and badges shown above reflect activity on this enhanced testing fork only. These metrics do not represent the official HoraDomu/Atheon project.
>
> **For official project statistics**, please visit: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)
>
> **Fork Attribution**: This repository maintains feature parity with upstream through PR submissions. Stars, forks, and other metrics shown here are specific to this enhanced version and do not imply ownership or endorsement of the original project.

---

<p align="center">
  <b>Both projects serve the community. This fork tests boundaries while maintaining upstream compatibility.</b>
</p>

<p align="center">
  <i>For production stability, use the official project. For cutting-edge features, use this enhanced version.</i>
</p>
