# Branch Strategy Documentation

## 🎯 Overview

The aliasfoxkde/Atheon repository uses a systematic branch strategy designed to maintain high-quality releases while staying synchronized with upstream HoraDomu/Atheon.

## 📋 Core Branches

### **`stable/clean`** (Upstream Tracking Branch)
**Purpose**: Source of truth that tracks upstream HoraDomu/Atheon:main

**Characteristics**:
- ✅ Clean upstream code without modifications
- ✅ Updated daily via automated sync
- ✅ Reference point for upstream changes
- ✅ Protected branch (maintainer-only access)

**Usage**:
```bash
# Sync with upstream
./.github/scripts/sync-stable-clean.sh

# Manual sync
git checkout stable/clean
git fetch upstream
git merge upstream/main
git push origin stable/clean
```

**When to use**:
- Starting new feature development
- Checking upstream changes
- Resolving merge conflicts
- Testing upstream compatibility

### **`main`** (Production Build Branch)
**Purpose**: Production-ready build with all validated enhancements

**Characteristics**:
- ✅ All validated PRs merged
- ✅ Enhanced features (105+ patterns vs 57 upstream)
- ✅ Performance optimizations
- ✅ MCP integration
- ✅ Pattern state persistence
- ✅ Comprehensive testing

**Installation**:
```bash
go install github.com/aliasfoxkde/Atheon@latest
```

**When to use**:
- User-facing installation
- Production deployment
- Feature integration testing
- Release tagging

### **`dev/full-feature`** (Development Branch)
**Purpose**: Comprehensive testing with ALL patterns enabled

**Characteristics**:
- ✅ All 105+ patterns enabled
- ✅ Experimental features active
- ✅ Comprehensive test suite
- ✅ Debug mode enabled
- ✅ Performance benchmarking
- ✅ Self-scanning validation

**Installation**:
```bash
go install github.com/aliasfoxkde/Atheon@dev/full-feature
```

**When to use**:
- Comprehensive testing
- Pattern development
- Performance validation
- Feature experimentation
- Quality assurance

## 🔄 Feature Branch Strategy

### **Branch Naming Conventions**

```
feat/feature-name          # New features
perf/performance-name     # Performance improvements
patterns/category-name     # Pattern library expansion
docs/documentation-name   # Documentation updates
infra/infrastructure-name # CI/CD, tooling, infrastructure
fix/bug-fix-name          # Bug fixes
test/test-improvement     # Test enhancements
```

### **Feature Development Lifecycle**

```
1. Branch Creation
   git checkout -b feat/feature-name stable/clean

2. Development & Testing
   - Implement feature
   - Test locally (go test ./...)
   - Run with config/profiles/development.json
   - Validate patterns (atheon list --all)

3. Create PR
   gh pr create --base main --head feat/feature-name

4. CI/CD Validation
   - Multi-version Go testing
   - Static analysis
   - Security scanning
   - Performance benchmarks

5. Code Review
   - Peer review
   - Documentation review
   - Architecture assessment

6. Merge to main
   - Auto-merge when checks pass
   - Update changelog
   - Version bump
```

## 🔧 Branch-Specific Configurations

### **stable/clean Configuration**
- **Profile**: Default upstream settings
- **Patterns**: Upstream 57 patterns only
- **Testing**: Standard upstream test suite
- **Performance**: Baseline upstream performance

### **main Configuration**
- **Profile**: `config/profiles/production.json`
- **Patterns**: 105+ enhanced patterns
- **Testing**: Extended test suite + integration tests
- **Performance**: Optimized with streaming API
- **Features**: MCP integration, pattern persistence

### **dev/full-feature Configuration**
- **Profile**: `config/profiles/development.json`
- **Patterns**: All 105+ patterns enabled
- **Testing**: Comprehensive test suite + self-scanning
- **Performance**: Comprehensive mode with profiling
- **Features**: All experimental features enabled

## 🚨 Branch Protection Rules

### **Protected Branches**
- `stable/clean`: Maintainer-only push
- `main`: PR required, CI checks required
- `dev/full-feature`: PR required, CI checks required

### **Required Checks**
- ✅ All tests pass (multi-version Go)
- ✅ Lint checks pass (golangci-lint, staticcheck)
- ✅ Security scan passes (Atheon self-scan)
- ✅ Performance benchmarks pass
- ✅ Code coverage threshold met (≥45%)

## 📊 Branch Comparison

| Feature | stable/clean | main | dev/full-feature |
|---------|-------------|------|------------------|
| Upstream Sync | ✅ Complete | ⚠️ Periodic | ❌ None |
| Pattern Count | 57 | 105+ | 105+ |
| Enhanced Features | ❌ None | ✅ Yes | ✅ Yes |
| MCP Integration | ❌ None | ✅ Yes | ✅ Yes |
| Performance Opt | ❌ None | ✅ Yes | ✅ Yes |
| Pattern Persistence | ❌ None | ✅ Yes | ✅ Yes |
| Experimental Features | ❌ None | ❌ No | ✅ Yes |
| Debug Mode | ❌ No | ❌ No | ✅ Yes |
| Self-Scanning | ❌ No | ⚠️ Optional | ✅ Yes |
| User Installation | ❌ No | ✅ Yes | ⚠️ Testing |

## 🔄 Update & Sync Workflow

### **Daily Automated Sync**
```bash
# GitHub Action runs daily
./.github/scripts/sync-stable-clean.sh
```

### **Weekly Integration**
```bash
# Sync main with stable/clean
git checkout main
git merge stable/clean
# Test integration
git push origin main
```

### **Feature Branch Updates**
```bash
# Update feature branch with latest stable/clean
git checkout feat/my-feature
git merge stable/clean
# Resolve conflicts
# Continue development
```

## 🎯 Decision Tree

```
Starting Development?
│
├─ Need clean baseline? → stable/clean
├─ Adding new feature? → feat/feature-name → main
├─ Testing integration? → dev/full-feature
└─ Production deployment? → main
```

## 📝 Branch Maintenance

### **Weekly Tasks**
- Review and merge completed PRs to main
- Update feature branches with stable/clean
- Clean up stale branches
- Review branch protection settings

### **Monthly Tasks**
- Comprehensive branch audit
- Update documentation
- Review CI/CD performance
- Assess upstream changes impact

### **Quarterly Tasks**
- Branch strategy review
- Architecture assessment
- Workflow optimization
- Documentation overhaul

---

## 🚀 Quick Start Guide

### **For Users**
```bash
# Install production version
go install github.com/aliasfoxkde/Atheon@latest

# Use with default settings
atheon scan ./my-project

# Use with profile
atheon scan ./my-project --profile pipeline
```

### **For Developers**
```bash
# Start new feature
git checkout -b feat/feature-name stable/clean

# Run with development profile
cp config/profiles/development.json ~/.atheon/config.json
atheon scan ./test-project --all

# Create PR
gh pr create --base main --head feat/feature-name
```

### **For Maintainers**
```bash
# Sync with upstream
./.github/scripts/sync-stable-clean.sh

# Review PRs
gh pr list --state open

# Merge to main
gh pr merge <pr-number> --merge

# Update documentation
make docs-update
```

---

**This branch strategy ensures systematic development while maintaining high quality and upstream compatibility.**