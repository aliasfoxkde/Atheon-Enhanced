# Enhancement Plan for Atheon-Enhanced

## Overview
This document outlines a comprehensive enhancement plan for the Atheon-Enhanced project, addressing architecture improvements, pattern expansion, performance optimization, and documentation enhancement.

## Completed Enhancements

### Patterns Added (August 2026)
- `wcag-aaa-contrast.yaml` - WCAG AAA color contrast pattern
- `wcag-aaa-text-spacing.yaml` - WCAG AAA text spacing pattern
- `wcag-aaa-focus-visible.yaml` - WCAG AAA focus visible pattern

### Documentation Created
- `docs/planning/ENHANCEMENT_PLAN.md` - This comprehensive enhancement plan

### Test Coverage Improvements
- Coverage improved from 93.4% to 93.5%
- New tests for clone detection, audit layers, and AST patterns

## Current State Assessment

### Strengths
- **406 patterns** across 29 categories
- Multi-format output: JSON, SARIF, Markdown, plain text
- MCP server for AI assistant integration
- Pi CLI adapter for coding agent workflows
- Self-scanning CI pipeline
- Clone detection and code quality auditing

### Areas for Improvement
- Test coverage: 93.5% (target: 99%)
- Some patterns use simple regex (not AST-aware)
- No incremental/scanned-file caching
- No network scanning capability
- No daemon mode for background service
- Documentation could be more comprehensive

---

## Phase 1: Quick Wins & Documentation

### 1.1 Documentation Enhancement
- [ ] Improve README.md with clearer getting-started
- [ ] Add architecture diagrams to docs/ARCHITECTURE.md
- [ ] Create comprehensive API documentation in docs/api/
- [ ] Add troubleshooting section with common issues

### 1.2 Pattern Quality
- [ ] Audit existing patterns for false positives
- [ ] Add WCAG 2.1 AAA compliance patterns
- [ ] Add more semantic web patterns (ARIA, HTML5)

### 1.3 Branch Cleanup
- [ ] Delete merged/unused branches
- [ ] Create branch strategy documentation

---

## Phase 2: AST Pattern Enhancement

### 2.1 AST Pattern Improvements
Current patterns are regex-based. Enhance with AST-aware patterns:

#### Python Patterns (ast-grep inspired)
```
# Detect exec with user input
exec(input_string)

# Detect eval with user input
eval(user_data)

# Detect dynamic attribute access
getattr(obj, user_controlled_attr)

# Detect SQL injection via string formatting
"SELECT * FROM %s" % user_input
```

#### JavaScript/TypeScript Patterns
```
# Detect innerHTML assignment with user input
element.innerHTML = userInput

# Detect eval with user input
eval(userInput)

# Detect document.write with user input
document.write(userInput)
```

### 2.2 Implement AST Pattern Registry
- Create `ASTPattern` type that uses Go's ast.Inspect
- Add `builtinASTPatterns` expansion
- Support multiple language parsers

---

## Phase 3: Performance Optimization

### 3.1 Incremental Scanning
- [ ] Add file hash caching for skip-if-unchanged
- [ ] Track last scan timestamp per file
- [ ] Only re-scan modified files

### 3.2 Parallel Processing
- [ ] Add worker pool for multi-file scanning
- [ ] Use sync/errgroup for parallel AST parsing
- [ ] Consider SIMD for regex matching (RE2 does this internally)

### 3.3 Memory Optimization
- [ ] Stream large files instead of loading entirely
- [ ] Reuse byte buffers in hot paths
- [ ] Profile with `pprof` to identify bottlenecks

---

## Phase 4: New Features

### 4.1 Network Scanner
- [ ] Add `--network-scan` flag
- [ ] Detect exposed API keys in network traffic patterns
- [ ] Scan for open security headers

### 4.2 Daemon Mode (atheon-svc)
- [ ] Create lightweight daemon
- [ ] Add Unix socket or TCP endpoint
- [ ] Support hot-reload of patterns
- [ ] Add status/health endpoint

### 4.3 Enhanced CI/CD Integration
- [ ] Add GitHub Actions cache integration
- [ ] Support GitLab CI variables
- [ ] Add pre-commit hook optimization

---

## Phase 5: Code Quality

### 5.1 Test Coverage (Target: 99%)
Blocking items for 99% coverage:
- Refactor `main()` functions to use error returns instead of `os.Exit()`
- Convert `SetBundleDownloadURL` panic() calls to error returns
- Add HTTP mocking interfaces for network tests
- Build AST fixture libraries for statement type coverage

### 5.2 Code Refactoring
- [ ] Extract interfaces for testing
- [ ] Add error wrapping conventions
- [ ] Improve documentation comments

---

## Phase 6: Upstream Integration

### 6.1 Sync with HoraDomu/Atheon
- [ ] Add upstream as remote (if accessible)
- [ ] Review changes for cherry-picking
- [ ] Run upstream tests on our patterns

### 6.2 Contribution Back
- [ ] Package improvements as upstream PRs
- [ ] Share AST pattern approach
- [ ] Document custom extensions

---

## Implementation Notes

### SIMD and Performance
Go's `regexp` package (RE2) already uses prefix matching and is quite efficient. For most use cases:
- RE2 is single-threaded but deterministic
- For true SIMD, consider `github.com/rivo/uniseg` for Unicode
- Focus on algorithmic improvements (caching, incremental) before SIMD

### ast-grep Integration
ast-grep uses tree-sitter for parsing. Consider:
- Adding tree-sitter as a dependency
- Creating pattern translations from ast-grep rules
- Using ast-grep's multi-language support

### Rust Rewrite
A full Rust rewrite would provide:
- Memory safety guarantees
- Better performance for compute-heavy tasks
- Richer type system

However, this is a significant undertaking. Consider:
- Partial rewrite of hot paths
- Using Rust for new standalone tools
- Keeping Go for CLI/main application

---

## Timeline

| Phase | Effort | Impact |
|-------|--------|--------|
| Phase 1: Quick Wins | Low | Medium |
| Phase 2: AST Patterns | High | High |
| Phase 3: Performance | Medium | Medium |
| Phase 4: New Features | High | High |
| Phase 5: Code Quality | Medium | Medium |
| Phase 6: Upstream | Low | Low |

---

## References

- [ast-grep](https://github.com/ast-grep/ast-grep) - AST-based code search
- [tree-sitter](https://tree-sitter.github.io/tree-sitter/) - Multi-language parser
- [Go AST Package](https://pkg.go.dev/go/ast)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
