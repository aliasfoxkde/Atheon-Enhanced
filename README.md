<h1 align="center">Atheon-Enhanced</h1>

<p align="center">
  <i>Feature-rich pattern matching engine for secrets detection, AI-generated code identification, and quality enforcement</i>
</p>

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Patterns](https://img.shields.io/badge/patterns-406-blueviolet)
![Categories](https://img.shields.io/badge/categories-39-orange)
![CI](https://github.com/aliasfoxkde/Atheon-Enhanced/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/aliasfoxkde/Atheon-Enhanced/graph/badge.svg)](https://codecov.io/gh/aliasfoxkde/Atheon-Enhanced)
![Stars](https://img.shields.io/github/stars/aliasfoxkde/Atheon-Enhanced?style=social)

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

### **Enhanced aliasfoxkde/Atheon (Atheon-Enhanced)** - Feature-Rich Testing Build
- **Purpose**: Experimental "nightly build" testing the limits of pattern matching
- **Focus**: Performance optimizations, advanced features, comprehensive pattern coverage
- **Patterns**: 406 patterns across 39 categories (community-driven, comprehensive coverage)
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
- **406 patterns** across 29 categories - comprehensive coverage
- **2-3x faster** with streaming API and performance optimizations
- **10x less memory** usage with chunked file scanning
- **MCP integration** with state persistence and category filtering
- **Pattern state persistence** across sessions
- **Quality enforcement** patterns for AI/developer shortcuts
- **Configuration profiles** for different use cases
- **Comprehensive CI/CD** with multi-version testing
- **~97% test coverage** on the core scanner

### **Trade-offs to Consider**
- ✅ **More features** - Latest capabilities and experimental patterns
- ✅ **Faster updates** - Frequent enhancements and optimizations
- ⚠️ **Less stable** - Experimental features may have issues
- ⚠️ **May lag behind** - Official releases may have newer stable features
- ✅ **More testing** - Comprehensive test suite and validation

> **See detailed comparison**: [docs/reports/FEATURE_COMPARISON.md](docs/reports/FEATURE_COMPARISON.md)

</details>

### **🔗 Quick Links**

| Repository | Purpose | Status |
|------------|---------|--------|
| **[Official Project](https://github.com/HoraDomu/Atheon)** | Stable production release | ✅ Recommended for production |
| **[Enhanced Version](https://github.com/aliasfoxkde/Atheon-Enhanced)** | Feature-rich testing build | 🧪 Experimental features |
| **[Project Pulse](https://github.com/aliasfoxkde/Atheon-Enhanced/pulse)** | Activity & updates overview | 📊 Live stats |
| **[Contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)** | Project contributors | 👥 Community |
| **[Changelog](.github/CHANGELOG.md)** | Release history | 📝 Recent changes |
| **[Support](.github/SUPPORT.md)** | Help & troubleshooting | ❓ Get assistance |
| **[Docs](docs/)** | Full documentation | 📖 Browse guides |

---

## 📁 **Repository Structure**

```
Atheon-Enhanced/
├── .github/              # GitHub configuration
│   ├── workflows/        # CI/CD pipelines
│   ├── ISSUE_TEMPLATE/   # Bug/feature request templates
│   ├── SUPPORT.md        # Help & troubleshooting guide
│   ├── CHANGELOG.md      # Quick changelog reference
│   ├── CODE_OF_CONDUCT.md
│   ├── CONTRIBUTING.md
│   └── SECURITY.md
├── docs/                 # Full documentation
│   ├── architecture/     # System architecture
│   ├── development/       # Developer guides
│   ├── integrations/      # MCP & tool integrations
│   ├── reports/          # Analysis & comparison reports
│   └── *.md              # Guides and references
├── community/            # Pattern definitions
│   └── ai-detection/      # AI detection patterns
├── cmd/                  # CLI & MCP server
├── core/                 # Core scanning engine
└── config/               # Configuration profiles
```

---

## 🚀 **Quick Start**

<details>
<summary><b>📦 Installation Methods</b></summary>

### **Recommended: Build from Source**
```bash
# Clone the repository
git clone https://github.com/aliasfoxkde/Atheon-Enhanced.git
cd Atheon-Enhanced

# Build the main binary
go build -o atheon ./cmd/atheon

# Build the MCP server (optional)
go build -o atheon-mcp ./cmd/mcp

# Install to your PATH
sudo mv atheon /usr/local/bin/  # Linux/macOS
# OR add to PATH manually
export PATH=$PATH:$(pwd)

# Verify installation
atheon --version
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

### **Library Usage**
Atheon is usable as a Go library. All scanner entry points accept
`context.Context` for cancellation and return structured `core.Finding` values:

```go
import "github.com/aliasfoxkde/Atheon/core"

// Scan an in-memory string (returns []core.Finding)
findings := core.ScanString(ctx, content, "config.txt")

// Scan a file on disk (returns findings, stats, error)
findings, stats, err := core.ScanFile(ctx, "/path/to/file.go")

// Scan a directory recursively (parallel walk)
findings, stats, err := core.ScanDir(ctx, "/path/to/project")

// Scan environment variables (returns []core.Finding)
findings := core.ScanEnv(ctx)
```

Sentinel errors are exported for programmatic checks:

```go
import "errors"
import "github.com/aliasfoxkde/Atheon/core"

if errors.Is(err, core.ErrPatternNotFound) {
    // handle missing pattern
}
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

Atheon integrates cleanly as a Git pre-commit hook to block secrets and quality
issues before they reach version control.

```bash
# Create the hook file
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
# Atheon pre-commit: scan staged files for secrets and quality issues
set -e

# Collect staged files (excluding deleted)
STAGED=$(git diff --cached --name-only --diff-filter=d)
if [ -z "$STAGED" ]; then
  exit 0
fi

# Write staged content to a temp dir and scan it
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

for f in $STAGED; do
  mkdir -p "$TMP/$(dirname "$f")"
  git show ":$f" > "$TMP/$f"
done

atheon "$TMP"
EOF

chmod +x .git/hooks/pre-commit
```

> [!TIP]
> Use `--categories=secrets,pii` to limit scanning to high-severity patterns only in the hook for faster commits.

> [!NOTE]
> For team-wide enforcement, add the hook to a `scripts/` directory and document setup in your project's CONTRIBUTING guide. Tools like [pre-commit](https://pre-commit.com) can automate hook installation across the team.

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
- **[Atheon-Benchmark](https://github.com/aliasfoxkde/Atheon-Benchmark)** - Performance testing and benchmarking tools
- **[Atheon-GitHub-Scanner](https://github.com/aliasfoxkde/Atheon-GitHub-Scanner)** - GitHub repository scanning automation

### **Official Upstream**
- **[HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)** - Official stable release

### **Supporting Infrastructure**
- **[Portfolio](https://openportfolio.pages.dev)** - Project maintainer's portfolio
- **[Documentation](docs/)** - Comprehensive documentation
- **[GitHub Wiki](https://github.com/aliasfoxkde/Atheon-Enhanced/wiki)** - Community guides and tutorials

</details>

---

## 📚 **Documentation Navigation**

<details>
<summary><b>📖 Core Documentation</b></summary>

### **System Architecture & Workflow**
- **[docs/architecture/SYSTEM_ARCHITECTURE.md](docs/architecture/SYSTEM_ARCHITECTURE.md)** - Complete system architecture and workflow
- **[docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md)** - Branch structure and development workflow
- **[docs/reports/FEATURE_COMPARISON.md](docs/reports/FEATURE_COMPARISON.md)** - Upstream vs enhanced feature comparison
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
- ✅ **274 patterns** across 19 categories
- ✅ **New categories**: Accessibility, Performance, Web Development, API Integration, Security Hardening, Cloud-Native, PWA, Data Visualization
- ✅ **Enhanced coverage**: Modern web development, security best practices, performance optimization
- ✅ **AI Detection Patterns**: AI-generated code identification, template detection, safety bypasses
- ✅ **DevOps Patterns**: CI/CD configurations, Docker, Kubernetes, GitHub workflows
- ✅ **Quality Patterns**: Git --force detection, test skipping, insecure flags, placeholder code
- ✅ **Security Patterns**: Injection prevention, CORS issues, authentication, CSRF, XSS protection
- ✅ **Performance Patterns**: N+1 queries, memory leaks, caching strategies, blocking operations
- ✅ **Accessibility Patterns**: WCAG compliance, keyboard navigation, ARIA labels, semantic HTML
- ✅ **API Patterns**: REST/GraphQL best practices, rate limiting, error handling, versioning

### **Pattern Distribution**
| Category | Patterns |
|----------|----------|
| secrets | 32 |
| code-quality | 25 |
| accessibility | 19 |
| security-hardening | 14 |
| performance | 12 |
| web-development | 12 |
| web-security | 12 |
| api-integration | 9 |
| healthcare | 7 |
| ai-detection | 6 |
| cloud-native | 6 |
| devops | 6 |
| data-visualization | 5 |
| pwa | 5 |
| finance | 3 |
| pii | 3 |
| django | 1 |
| nodejs | 1 |
| react | 1 |

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
| Pattern Count | 57 | 406 |
| Categories | 5 | 29 |
| Memory Usage | Full file loading | Chunked streaming (10x reduction) |
| Performance | Baseline | 2-3x faster |
| MCP Integration | ✅ Advanced features | ✅ Advanced features + state persistence |
| Pattern Persistence | ❌ | ✅ |
| Configuration Profiles | ❌ | ✅ 4 profiles |
| Quality Enforcement | ❌ | ✅ 13 patterns |
| AI Detection | ❌ | ✅ 6 patterns |
| Self-Scanning | ❌ | ✅ |
| Streaming API | ❌ | ✅ |
| DevOps Patterns | ❌ | ✅ 6 patterns |
| Context Cancellation | ❌ | ✅ |
| Sentinel Errors | ❌ | ✅ |
| Godoc Comments | Partial | ✅ Comprehensive |
| golangci-lint | ❌ | ✅ 18 linters in CI |
| Test Coverage | — | 97%+ (core, cmd/atheon) |
| Benchmarks | ❌ | ✅ In-tree |
| Examples | ❌ | ✅ Runnable godoc examples |
| Stability | Production-ready | Testing/Experimental |
| Update Frequency | Scheduled releases | Frequent updates |
| Feature Parity | N/A (upstream) | Maintained via PRs |
| **Clone Detection** | ❌ | ✅ AST-based code clone detection |
| **Complexity Metrics** | ❌ | ✅ Cyclomatic, Cognitive, Halstead |
| **Layered Audits** | ❌ | ✅ 9-layer progressive audit system |
| **Architecture Audits** | ❌ | ✅ Engineering policy enforcement |
| **Consistency Audits** | ❌ | ✅ Boolean naming, style consistency |

> **📖 See detailed comparison**: [docs/reports/FEATURE_COMPARISON.md](docs/reports/FEATURE_COMPARISON.md)

</details>

---

## 🧪 **Testing & Validation**

<details>
<summary><b>🔬 Enhanced Test Coverage</b></summary>

### **Testing Infrastructure**
- ✅ Multi-version Go testing (1.21, 1.22, 1.23, 1.24)
- ✅ Static analysis (golangci-lint v1.64.8 with 18 linters)
- ✅ Security scanning (Atheon self-scan)
- ✅ Performance benchmarking ([BENCHMARKS.md](docs/BENCHMARKS.md))
- ✅ Integration tests for all features
- ✅ Context-cancellation tests for every scanner entry point
- ✅ Runnable godoc examples for every public API

### **Quality Metrics**
- **Test Coverage**: 97%+ across core, cmd/atheon, and bundler packages
- **CI/CD Pass Rate**: >95%
- **Pattern Validation**: All 406 patterns tested and functional
- **Pattern Coverage**: 29 categories with modern development support
- **Lint Warnings**: 0 (golangci-lint clean)
- **Clone Detection**: AST-based duplicate code detection (PMD CPD-inspired)
- **Complexity Metrics**: Cyclomatic, Cognitive, Halstead complexity analysis
- **Layered Audits**: 9-layer progressive audit system

**📖 Detailed Documentation**: See [docs/architecture/PATTERN_CATEGORIES.md](docs/architecture/PATTERN_CATEGORIES.md) for comprehensive pattern documentation

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
- **Patterns**: 274 patterns across 19 categories
- **Installation**: `go install github.com/aliasfoxkde/Atheon@latest`

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
- **[Upstream Contributors](docs/CONTRIBUTORS.md)** - Contributors to the official project
- **[Enhanced Contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)** - Contributors to this enhanced version

> **📖 See detailed system architecture**: [docs/architecture/SYSTEM_ARCHITECTURE.md](docs/architecture/SYSTEM_ARCHITECTURE.md)

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
- **Maintainer**: [@aliasfoxkde](https://github.com/aliasfoxkde) — see [CONTRIBUTORS](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)
- **Repository**: [https://github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced)
- **Portfolio**: [https://openportfolio.pages.dev](https://openportfolio.pages.dev)
- **Enhancement Purpose**: Advanced features, performance optimizations, comprehensive testing

### **Related Projects**
- **[Atheon-Benchmark](https://github.com/aliasfoxkde/Atheon-Benchmark)** - Performance testing and benchmarking tools
- **[Atheon-GitHub-Scanner](https://github.com/aliasfoxkde/Atheon-GitHub-Scanner)** - GitHub repository scanning automation

### **Development Assistance**
This enhanced version includes development support using TaskWizer technologies for systematic testing, documentation generation, and quality assurance.

</details>

<details>
<summary><b>👥 Contributors</b></summary>

Both the upstream project and this enhanced fork are built by the community. Every pattern contributed benefits all users.

- **[Upstream Contributors](docs/CONTRIBUTORS.md)** - Contributors to the official project
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
- ✅ MCP integration with advanced features

### **Use Enhanced aliasfoxkde/Atheon (Atheon-Enhanced) when you want**:
- ✅ 406 patterns across 39 categories (comprehensive coverage)
- ✅ 2-3x performance improvements
- ✅ 10x memory reduction for large files
- ✅ MCP integration with advanced features
- ✅ Pattern state persistence
- ✅ Quality enforcement patterns
- ✅ Configuration profiles for different use cases
- ✅ Comprehensive testing (97%+ coverage) and validation
- ✅ Context cancellation across all scan APIs
- ✅ Static-analysis clean (golangci-lint v1.64.8)
- ✅ Runnable godoc examples
- ✅ In-tree performance benchmarks
- ✅ Experimental feature testing

---

## 🔗 **Quick Links**

| Resource | Link | Purpose |
|----------|------|---------|
| **Official Project** | [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon) | Stable production release |
| **Enhanced Version** | [https://github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced) | Feature-rich testing build |
| **Feature Comparison** | [docs/reports/FEATURE_COMPARISON.md](docs/reports/FEATURE_COMPARISON.md) | Detailed feature comparison |
| **System Architecture** | [docs/architecture/SYSTEM_ARCHITECTURE.md](docs/architecture/SYSTEM_ARCHITECTURE.md) | Technical architecture |
| **Branch Strategy** | [docs/BRANCH_STRATEGY.md](docs/BRANCH_STRATEGY.md) | Development workflow |
| **Project Pulse** | [https://github.com/aliasfoxkde/Atheon-Enhanced/pulse](https://github.com/aliasfoxkde/Atheon-Enhanced/pulse) | Activity overview |
| **Contributors** | [https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors) | Project contributors |
| **Scanner Demo** | [https://atheon-scanner.pages.dev/](https://atheon-scanner.pages.dev/) | Live scanning demo |
| **Benchmark Demo** | [https://atheon-benchmark.pages.dev](https://atheon-benchmark.pages.dev) | Performance benchmarking |

---

## 📈 **Project Growth & Activity**

### **Repository Activity**
![GitHub Stars](https://img.shields.io/github/stars/aliasfoxkde/Atheon-Enhanced?style=for-the-badge&logo=github)
![GitHub Forks](https://img.shields.io/github/forks/aliasfoxkde/Atheon-Enhanced?style=for-the-badge&logo=github)
![GitHub Issues](https://img.shields.io/github/issues/aliasfoxkde/Atheon-Enhanced?style=for-the-badge)
![GitHub Closed PRs](https://img.shields.io/github/issues-pr-closed/aliasfoxkde/Atheon-Enhanced?style=for-the-badge)

### **Stars Over Time**
[![Star History Chart](https://api.star-history.com/svg?repos=aliasfoxkde/Atheon-Enhanced&type=Date)](https://star-history.com/#aliasfoxkde/Atheon-Enhanced&Date)

### **Recent Activity**
- ![Latest Commit](https://img.shields.io/github/last-commit/aliasfoxkde/Atheon-Enhanced?style=flat-square)
- ![Commit Activity](https://img.shields.io/github/commit-activity/y/aliasfoxkde/Atheon-Enhanced?style=flat-square)
- ![Release](https://img.shields.io/github/release/aliasfoxkde/Atheon-Enhanced?style=flat-square)

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
