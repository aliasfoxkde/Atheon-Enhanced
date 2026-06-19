# Frequently Asked Questions (FAQ)

## 🚀 Installation & Setup

### **Q: How do I install Atheon?**
**A:** The recommended method is building from source:
```bash
# Option 1: Build from source (recommended)
git clone https://github.com/aliasfoxkde/Atheon.git
cd Atheon
go build -o atheon .
sudo mv atheon /usr/local/bin/  # or add to your PATH

# Option 2: Development version with all features
git clone https://github.com/aliasfoxkde/Atheon.git
cd Atheon
git checkout dev/full-feature
go build -o atheon .
```

### **Q: Why does `go install` fail with "does not contain package" error?**
**A:** This issue has been fixed in the latest version! The Go module path has been corrected. However, building from source is still recommended:

1. **Build from source (recommended):**
```bash
git clone https://github.com/aliasfoxkde/Atheon.git
cd Atheon
go build -o atheon .
sudo mv atheon /usr/local/bin/  # or add to your PATH
```

2. **Use the development branch:**
```bash
git clone https://github.com/aliasfoxkde/Atheon.git
cd Atheon
git checkout dev/full-feature
go build -o atheon .
```

3. **Check your Go version:**
```bash
go version  # Should be 1.21 or higher
```

### **Q: What are the system requirements?**
**A:**
- Go 1.21 or higher
- Windows, macOS, or Linux
- 50MB free disk space
- Internet connection for `go install`

---

## 🎯 Project Understanding

### **Q: What's the difference between official and enhanced Atheon?**
**A:** The official HoraDomu/Atheon is a stable production release with 57 patterns, while the enhanced aliasfoxkde/Atheon is a feature-rich testing build with 105+ patterns, performance optimizations, and experimental features. See [FEATURE_COMPARISON.md](FEATURE_COMPARISON.md) for details.

### **Q: Is this a competing project?**
**A:** No! This is an enhanced testing fork that:
- Maintains feature parity with the official project via PRs
- Tests experimental features and innovations
- Contributes improvements back upstream
- Serves as a "nightly build" for power users

### **Q: Should I use official or enhanced Atheon?**
**A:**
- **Use Official** for: Production environments, stability, conservative feature selection
- **Use Enhanced** for: Development testing, CI/CD, comprehensive scanning, performance evaluation

---

## 🔧 Usage & Configuration

### **Q: How do I scan a directory?**
**A:**
```bash
# Basic scan
atheon ./my-project

# With specific categories
atheon --categories=secrets,pii ./my-project

# With configuration profile
atheon --profile config/profiles/pipeline.json ./my-project
```

### **Q: What are configuration profiles?**
**A:** Pre-configured settings for different use cases:
- `production.json` - General use (default)
- `pipeline.json` - CI/CD optimized
- `mcp-integration.json` - MCP server settings
- `development.json` - Full feature testing

### **Q: How do I enable/disable patterns?**
**A:**
```bash
# List all patterns
atheon list

# Enable a pattern
atheon enable stripe-api-key

# Disable a pattern
atheon disable todo-comments

# List by category
atheon list --category secrets
```

### **Q: How do I scan from stdin?**
**A:**
```bash
# Scan file content
cat file.txt | atheon -

# Scan git diff
git diff | atheon -

# Scan environment variables
atheon --env
```

---

## 🔒 Security & Patterns

### **Q: How many patterns does Atheon have?**
**A:**
- **Official HoraDomu/Atheon**: 57 patterns
- **Enhanced aliasfoxkde/Atheon**: 105+ patterns

### **Q: Can I add my own patterns?**
**A:** Yes! Create a YAML file in the appropriate category:
```yaml
# community/secrets/my-service.yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

Then rebuild the bundle:
```bash
go run ./bundler
```

### **Q: How accurate are the patterns?**
**A:**
- False positive rate: <5% for enhanced version
- All patterns tested against real codebases
- Community-validated patterns
- Regular updates and improvements

### **Q: Can Atheon scan itself?**
**A:** Yes! The enhanced version includes self-scanning:
```bash
./.github/scripts/self-scan.sh
```

---

## ⚡ Performance & Benchmarks

### **Q: How fast is Atheon?**
**A:**
- **Enhanced version**: 2-3x faster than upstream
- **Large file (100MB)**: <10s scanning time
- **Large directory (10,000 files)**: <30s scanning time
- **Memory usage**: <50MB for any workload

### **Q: What makes the enhanced version faster?**
**A:**
- Streaming API for large files
- Chunked scanning (10x memory reduction)
- Parallel processing
- Optimized regex compilation

### **Q: Where can I see performance benchmarks?**
**A:**
- [Live Demo](https://atheon-benchmark.pages.dev) - Interactive benchmarking
- [Atheon-Benchmark Repository](https://github.com/aliasfoxkde/Atheon-Benchmark) - Performance tools
- CI/CD workflows include performance regression detection

---

## 🔧 Troubleshooting

### **Q: Atheon is not detecting patterns I expect**
**A:**
1. Check if the pattern is enabled: `atheon list --enabled`
2. Verify pattern syntax: `atheon list --pattern <pattern-name>`
3. Test with `--all` flag to include disabled patterns
4. Check your file is not ignored (see `.atheonignore`)

### **Q: Atheon is too slow on large files**
**A:**
1. Use configuration profile: `atheon --profile config/profiles/pipeline.json`
2. Enable streaming mode (automatic for large files)
3. Reduce number of enabled patterns
4. Use category filtering: `--categories=secrets`

### **Q: I'm getting too many false positives**
**A:**
1. Review and disable specific patterns: `atheon disable <pattern>`
2. Use more specific categories: `--categories=api-keys`
3. Create custom configuration profile
4. Report false positives to help improve patterns

### **Q: Pre-commit hooks are not working**
**A:**
1. Make hook executable: `chmod +x .git/hooks/pre-commit`
2. Test manually: `atheon ./`
3. Check Atheon is in your PATH: `which atheon`
4. Review hook script for errors

---

## 🤝 Contributing

### **Q: How do I contribute patterns?**
**A:**
1. Check pattern doesn't exist: `atheon list`
2. Create YAML file in `community/<category>/`
3. Rebuild bundle: `go run ./bundler`
4. Add test case in `core/bundle_test.go`
5. Submit PR with clear description

See [contributing.md](contributing.md) for details.

### **Q: How do I contribute Go code?**
**A:**
1. Follow Go naming conventions and idiomatic practices
2. Ensure `go fmt` and `go vet` pass
3. Add tests for new functionality
4. Justify any new dependencies
5. Submit PR with clear explanation

See [contributing.md](contributing.md) for guidelines.

### **Q: Which branch should I contribute to?**
**A:**
- **Official project**: Submit to HoraDomu/Atheon:main
- **Enhanced features**: Submit to aliasfoxkde/Atheon:main
- **Experimental features**: Submit to aliasfoxkde/Atheon:dev/full-feature

See [BRANCH_STRATEGY.md](BRANCH_STRATEGY.md) for workflow details.

---

## 📊 Live Demos & Tools

### **Q: Are there live demos available?**
**A:** Yes!
- [Atheon-GitHub-Scanner](https://atheon-scanner.pages.dev/) - Live repository scanning
- [Atheon-Benchmark](https://atheon-benchmark.pages.dev) - Performance benchmarking

### **Q: What tools are available?**
**A:**
- **Atheon** - Main pattern matching engine
- **Atheon-Benchmark** - Performance testing tools
- **Atheon-GitHub-Scanner** - Repository scanning automation
- **MCP Server** - AI assistant integration

---

## 📞 Support & Community

### **Q: Where can I get help?**
**A:**
- **Issues**: [GitHub Issues](https://github.com/aliasfoxkde/Atheon/issues)
- **Documentation**: [docs/](INDEX.md)
- **Wiki**: [GitHub Wiki](../.github/wiki/)
- **Troubleshooting**: [Troubleshooting Guide](../.github/wiki/Troubleshooting.md)

### **Q: How can I contact the maintainers?**
**A:**
- **Official project**: [dommcpro@gmail.com](mailto:dommcpro@gmail.com)
- **Enhanced fork**: [GitHub Issues](https://github.com/aliasfoxkde/Atheon/issues)

### **Q: Where can I see contributor activity?**
**A:**
- [Project Pulse](https://github.com/aliasfoxkde/Atheon/pulse) - Activity overview
- [Contributors Graph](https://github.com/aliasfoxkde/Atheon/graphs/contributors) - Contributor visualization
- [CONTRIBUTORS.md](../CONTRIBUTORS.md) - Contributor recognition

---

## 📈 Project Status

### **Q: Is this project actively maintained?**
**A:** Yes! Both projects are actively maintained:
- Regular updates and releases
- Active community contributions
- Comprehensive CI/CD pipeline
- Ongoing feature development

### **Q: What's the current status of the enhanced fork?**
**A:**
- **105+ patterns** - Comprehensive coverage
- **69.8% test coverage** - High quality assurance
- **Multi-version testing** - Go 1.21, 1.22, 1.23, 1.24
- **Active development** - See [ROADMAP.md](ROADMAP.md)

### **Q: How do I track project progress?**
**A:**
- [Roadmap](ROADMAP.md) - Development milestones
- [Project Pulse](https://github.com/aliasfoxkde/Atheon/pulse) - Activity tracking
- [CI/CD Dashboard](../.github/workflows/) - Quality metrics
- [Releases](https://github.com/aliasfoxkde/Atheon/releases) - Version history

---

**Last Updated**: 2025-01-19
**FAQ Version**: 1.0

For more information, see the [complete documentation index](INDEX.md).