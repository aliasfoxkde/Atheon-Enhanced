# PI_INTEGRATION Architecture Implementation Tasks

**Tracking Issue:** https://github.com/aliasfoxkde/Atheon-Enhanced/issues/195
**Branch:** `feature/adapter-runtime-architecture`
**Status:** In Progress

---

## Overview

This document tracks the implementation of `docs/planning/PI_INTEGRATION.md` - restructuring Atheon-Enhanced into a single runtime with multiple adapters.

### Architecture Vision

```
                     Users
                        │
                        ▼
              AI Coding Agents
        Pi      Claude Code      Codex
        Goose     OpenHands      ...
                        │
                 Adapter Layer
          Pi Adapter | Claude Adapter | Codex Adapter
          GitForge Adapter | GitHub Actions Adapter | CLI Adapter
                        │
                Atheon Runtime
         Configuration | Scanner | Rule Engine | Pattern Engine
         File Discovery | Ignore Engine | Output Formatter
                        │
                Analysis Results
        JSON | SARIF | Markdown | Console | Exit Codes
```

---

## Phase 1 — Runtime Foundation

**Objective:** Create the reusable analysis runtime with zero Pi/GitForge/CI code.

- [x] Create `runtime/` directory structure
- [x] Create `runtime/scanner/` - file discovery and traversal
- [x] Create `runtime/rules/` - rule engine
- [x] Create `runtime/patterns/` - pattern engine
- [x] Create `runtime/parser/` - AST parsing
- [x] Create `runtime/formatter/` - output formatting
- [x] Create `runtime/config/` - configuration loading
- [x] Create `runtime/cache/` - caching mechanism
- [x] Create `runtime/diagnostics/` - diagnostics model
- [x] Verify zero Pi/GitForge/CI code in runtime
- [x] Runtime works standalone

**Status: ✅ COMPLETE** - All runtime packages implemented and tested

**Deliverables:**
```
runtime/
├── scanner/
│   └── *.go
├── rules/
│   └── *.go
├── patterns/
│   └── *.go
├── parser/
│   └── *.go
├── formatter/
│   ├── json.go
│   ├── sarif.go
│   ├── markdown.go
│   └── console.go
├── config/
│   └── *.go
├── cache/
│   └── *.go
└── diagnostics/
    └── *.go
```

---

## Phase 2 — Canonical Diagnostics Model

**Objective:** Everything returns the same diagnostics object.

- [x] Define `schemas/diagnostics.schema.json`
- [x] Define `schemas/config.schema.json`
- [x] Define `schemas/result.schema.json`
- [x] Implement diagnostics model in `runtime/diagnostics/`
- [x] All adapters consume same diagnostics object

**Status: ✅ COMPLETE**

**Diagnostics Schema:**
```json
{
  "summary": {},
  "statistics": {},
  "issues": [],
  "timing": {},
  "metadata": {}
}
```

---

## Phase 3 — Output System

**Objective:** One formatter per output type, no duplicated logic.

- [x] Console formatter (`runtime/formatter/console.go`)
- [x] Markdown formatter (`runtime/formatter/markdown.go`)
- [x] JSON formatter (`runtime/formatter/json.go`)
- [x] SARIF formatter (`runtime/formatter/sarif.go`)
- [x] Base formatter interface
- [x] Never duplicate formatting logic across adapters

**Status: ✅ COMPLETE**

---

## Phase 4 — Configuration System

**Objective:** Support multiple config file formats with deterministic merge order.

- [x] Support `atheon.json`
- [x] Support `atheon.yaml`
- [x] Support `pyproject.toml`
- [x] Support `package.json`
- [x] Document merge order
- [x] Implement config loader

**Status: ✅ COMPLETE**

**Merge Order:**
```
Defaults → Global → Project → CLI → Environment
```

---

## Phase 5 — Runtime API

**Objective:** Define one public interface for the runtime.

- [x] `Analyze(projectPath string) (*Diagnostics, error)`
- [x] `AnalyzeFiles(paths []string) (*Diagnostics, error)`
- [x] `Benchmark(project Project) (*BenchmarkResult, error)`
- [x] `ValidateConfig(configPath string) (*ValidationResult, error)`

**Status: ✅ COMPLETE**

---

## Phase 6 — Adapter Framework

**Objective:** Standardize adapter lifecycle.

- [x] Define adapter interface
- [x] Document lifecycle: Initialize → Validate → Load Config → Analyze → Format → Return → Exit
- [x] No adapter deviates from standard lifecycle

**Status: ✅ COMPLETE**

---

## Phase 7 — CLI Adapter

**Objective:** Refactor current CLI to use adapter pattern.

- [ ] Move CLI logic to `adapters/cli/`
- [ ] CLI only: parses arguments, invokes runtime, prints results
- [ ] Remove analysis logic from CLI adapter

---

## Phase 8 — Pi Adapter

**Objective:** Register one `atheon` tool for Pi.

- [ ] Create `adapters/pi/`
- [ ] Implement Pi tool registration
- [ ] Workspace, changed files, configuration inputs
- [ ] No custom analysis - uses runtime only

---

## Phase 9 — GitForge Adapter

**Objective:** Realtime validation via GitForge.

- [ ] Create `adapters/gitforge/`
- [ ] Developer → GitForge Action → Runtime → Diagnostics → AI → Fixes

---

## Phase 10 — GitHub Actions Adapter

**Objective:** Receive repository, run runtime, output SARIF, return exit code.

- [ ] Create `adapters/github-actions/`
- [ ] Input: Repository
- [ ] Output: SARIF + Exit Code

---

## Phase 11 — Pre-Commit Adapter

**Objective:** Git commit hook integration.

- [ ] Create `adapters/pre-commit/`
- [ ] Git Commit → Atheon → Diagnostics → Allow/Reject

---

## Phase 12 — Documentation

**Objective:** Document as if Atheon is a platform.

- [ ] `docs/getting-started/`
- [ ] `docs/architecture/`
- [ ] `docs/runtime/`
- [ ] `docs/diagnostics/`
- [ ] `docs/configuration/`
- [ ] `docs/adapters/`
- [ ] `docs/adapters/pi/`
- [ ] `docs/adapters/gitforge/`
- [ ] `docs/adapters/github-actions/`
- [ ] `docs/ci/`
- [ ] `docs/rules/`
- [ ] `docs/patterns/`
- [ ] `docs/performance/`
- [ ] `docs/benchmarking/`
- [ ] `docs/faq/`

---

## Phase 13 — Examples

**Objective:** Examples mirroring real workflows.

- [ ] `examples/basic-project/`
- [ ] `examples/large-monorepo/`
- [ ] `examples/python/`
- [ ] `examples/typescript/`
- [ ] `examples/mixed-language/`
- [ ] `examples/gitforge/`
- [ ] `examples/pi/`
- [ ] `examples/github-actions/`
- [ ] `examples/pre-commit/`

Each example includes:
- Sample project
- Configuration
- Expected diagnostics
- Expected formatted outputs
- Explanation of detected patterns

---

## Phase 14 — Testing

**Objective:** Separate tests by layer.

**Runtime Tests:**
- [ ] Scanner correctness
- [ ] Rule evaluation
- [ ] Pattern matching
- [ ] Configuration loading
- [ ] Diagnostic generation

**Formatter Tests:**
- [ ] JSON validity
- [ ] SARIF validity
- [ ] Markdown rendering
- [ ] Console output

**Adapter Tests:**
- [ ] CLI adapter
- [ ] Pi adapter
- [ ] GitForge adapter
- [ ] GitHub Actions adapter
- [ ] Pre-commit adapter

---

## Phase 15 — Performance

**Objective:** Measure and maintain performance targets.

**Metrics:**
- [ ] Startup time
- [ ] Configuration load time
- [ ] File discovery time
- [ ] Scan throughput (files/second)
- [ ] Pattern throughput (patterns/second)
- [ ] Memory usage
- [ ] Output generation time

**Benchmarking:**
- [ ] Set baseline benchmarks
- [ ] Regression detection
- [ ] Performance CI gate

---

## Phase 16 — Release Criteria

**Objective:** Define when a release is complete.

A release is complete only when:

- [ ] Runtime is the only location containing analysis logic
- [ ] Every adapter consumes the same runtime API
- [ ] Every adapter consumes the same diagnostics schema
- [ ] All output formats originate from the same diagnostics model
- [ ] Configuration precedence is deterministic and documented
- [ ] The diagnostics schema is versioned
- [ ] Documentation covers architecture, installation, configuration, adapter usage, and troubleshooting
- [ ] Runtime, formatter, and adapter test suites pass
- [ ] Performance benchmarks meet established targets

---

## Progress Summary

| Phase | Status | Progress |
|-------|--------|----------|
| Phase 1 - Runtime Foundation | ✅ Complete | 100% |
| Phase 2 - Diagnostics Model | ✅ Complete | 100% |
| Phase 3 - Output System | ✅ Complete | 100% |
| Phase 4 - Configuration | ✅ Complete | 100% |
| Phase 5 - Runtime API | ✅ Complete | 100% |
| Phase 6 - Adapter Framework | ✅ Complete | 100% |
| Phase 7 - CLI Adapter | 🔄 In Progress | 0% |
| Phase 8 - Pi Adapter | Not Started | 0% |
| Phase 9 - GitForge Adapter | Not Started | 0% |
| Phase 10 - GitHub Actions | Not Started | 0% |
| Phase 11 - Pre-Commit | Not Started | 0% |
| Phase 12 - Documentation | Not Started | 0% |
| Phase 13 - Examples | Not Started | 0% |
| Phase 14 - Testing | Not Started | 0% |
| Phase 15 - Performance | Not Started | 0% |
| Phase 16 - Release Criteria | Not Started | 0% |

---

## Notes

- This is a major architectural refactoring
- Current codebase will need significant restructuring
- Backward compatibility should be maintained where possible
- Incremental PRs preferred over large PRs
