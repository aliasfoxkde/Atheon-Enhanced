I actually think this is the right long-term decision.

The key realization is that **Pi should not be the integration point**. Instead, **Atheon should expose a stable runtime/API, and Pi, GitForge, CI/CD, GitHub Actions, pre-commit hooks, IDEs, and future tools should all integrate with that same runtime.**

That avoids creating a "Pi version" and a "GitForge version" of Atheon. There is only **one Atheon runtime**, with multiple adapters.

I would standardize on the following architecture.

---

# Atheon-Enhanced Architecture

```
                     Users

                        │
                        ▼

              AI Coding Agents
        Pi      Claude Code      Codex
        Goose     OpenHands      ...

                        │

                 Adapter Layer

          Pi Adapter
          Claude Adapter
          Codex Adapter
          GitForge Adapter
          GitHub Actions Adapter
          CLI Adapter

                        │

                Atheon Runtime

         Configuration
         Scanner
         Rule Engine
         Pattern Engine
         File Discovery
         Ignore Engine
         Output Formatter

                        │

                Analysis Results

        JSON
        SARIF
        Markdown
        Console
        Exit Codes

                        │

               Consumer Systems

       AI Agent
       GitForge
       CI/CD
       GitHub Actions
       Humans
```

The critical design principle is:

> **There is exactly one implementation of analysis. Everything else is an adapter.**

---

# Project Structure

I'd organize the repository around that architecture rather than around integrations.

```text
atheon-enhanced/

├── runtime/
│   ├── scanner/
│   ├── rules/
│   ├── patterns/
│   ├── parser/
│   ├── formatter/
│   ├── config/
│   ├── cache/
│   └── diagnostics/
│
├── adapters/
│   ├── cli/
│   ├── pi/
│   ├── gitforge/
│   ├── github-actions/
│   ├── pre-commit/
│   └── api/
│
├── schemas/
│   ├── diagnostics.schema.json
│   ├── config.schema.json
│   └── result.schema.json
│
├── docs/
├── examples/
├── tests/
└── benchmarks/
```

Everything depends on the runtime.

Nothing depends on Pi.

---

# Guiding Principles

These become architectural requirements.

## Single Runtime

Only one implementation of scanning exists.

No duplicated logic.

---

## Stateless Execution

Each analysis run is independent.

Adapters never maintain scan state.

---

## Structured Output First

The runtime never generates "pretty" output first.

Internally:

```
Analysis

↓

Diagnostics

↓

Output Formatter

↓

Markdown
JSON
SARIF
Console
```

Everything derives from diagnostics.

---

## Adapter Isolation

Every adapter should be under ~500 lines.

Its only responsibilities are:

* receive request
* validate input
* invoke runtime
* return results

No analysis logic.

---

## Runtime Independence

The runtime cannot know:

* Pi exists
* GitForge exists
* GitHub Actions exist

It only knows:

```
Analyze(project)

↓

Return diagnostics
```

---

# Phase 1 — Runtime Foundation

Objective:

Create the reusable analysis runtime.

Deliverables:

```
runtime/

scanner/

rule engine

pattern engine

diagnostics

configuration

formatter
```

Acceptance criteria

* zero Pi code
* zero GitForge code
* zero CI code

Only analysis.

---

# Phase 2 — Canonical Diagnostics Model

Everything returns the same object.

Example

```json
{
  "summary": {},
  "statistics": {},
  "issues": [],
  "timing": {},
  "metadata": {}
}
```

Every adapter consumes this.

No exceptions.

---

# Phase 3 — Output System

One formatter per output type.

```
Formatter

↓

Console

↓

Markdown

↓

JSON

↓

SARIF
```

Never duplicate formatting logic.

---

# Phase 4 — Configuration System

Support

```
atheon.json

atheon.yaml

pyproject.toml

package.json
```

Merge order

```
Defaults

↓

Global

↓

Project

↓

CLI

↓

Environment
```

Single source of truth.

---

# Phase 5 — Runtime API

Define one public interface.

```
analyze()

analyzeFiles()

analyzeProject()

benchmark()

validateConfig()
```

Everything else is internal.

---

# Phase 6 — Adapter Framework

Every adapter follows the same lifecycle.

```
Initialize

↓

Validate

↓

Load Config

↓

Analyze

↓

Format

↓

Return

↓

Exit
```

No adapter deviates.

---

# Phase 7 — CLI Adapter

Current CLI becomes an adapter.

Responsibilities

* parse arguments
* invoke runtime
* print results

Nothing else.

---

# Phase 8 — Pi Adapter

Responsibilities

Register one tool.

```
atheon
```

Inputs

* workspace
* changed files
* configuration
* optional verbosity

Workflow

```
Pi

↓

Atheon Tool

↓

Runtime

↓

Diagnostics

↓

Pi
```

No custom analysis.

No custom rules.

The AI decides when to call the tool naturally. A separate `/atheon` command can exist as an optional manual/debug entry point, but it should invoke the same tool path rather than introduce parallel logic.

---

# Phase 9 — GitForge Adapter

Purpose

Realtime validation.

Workflow

```
Developer

↓

GitForge Action

↓

Atheon Runtime

↓

Diagnostics

↓

AI

↓

Fixes

↓

Run Again
```

Same runtime.

---

# Phase 10 — GitHub Actions Adapter

Responsibilities

Receive

```
Repository

↓

Run Runtime

↓

SARIF

↓

Exit Code
```

Nothing more.

---

# Phase 11 — Pre-Commit Adapter

Workflow

```
Git Commit

↓

Atheon

↓

Diagnostics

↓

Allow

or

Reject
```

Uses same runtime.

---

# Phase 12 — Documentation

I'd write documentation as if Atheon is a platform, not a CLI utility.

```
docs/

getting-started/

architecture/

runtime/

diagnostics/

configuration/

adapters/

pi/

gitforge/

github-actions/

ci/

rules/

patterns/

performance/

benchmarking/

faq/
```

---

# Pi Documentation

```
Overview

Installation

How Pi Calls Atheon

Configuration

Examples

Troubleshooting

FAQ
```

Important section

```
What Pi Does

What Atheon Does
```

Explain separation of responsibility.

---

# GitForge Documentation

```
Architecture

Realtime Feedback

Pipeline Integration

Configuration

Examples

Performance

Troubleshooting
```

---

# Runtime Documentation

Include

* execution pipeline
* diagnostic lifecycle
* formatter pipeline
* adapter model
* configuration hierarchy
* schema definitions
* extension points

---

# Phase 13 — Examples

Examples should mirror real workflows.

```
examples/

basic-project/

large-monorepo/

python/

typescript/

mixed-language/

gitforge/

pi/

github-actions/

pre-commit/
```

Each example should include:

* sample project
* configuration
* expected diagnostics
* expected formatted outputs
* explanation of the detected patterns

---

# Phase 14 — Testing

Separate tests by layer.

## Runtime

* scanner correctness
* rule evaluation
* pattern matching
* configuration loading
* diagnostic generation

## Formatter

* JSON validity
* SARIF validity
* Markdown rendering
* console output

## Adapter

* CLI
* Pi
* GitForge
* GitHub Actions
* pre-commit

Adapters should be tested primarily for request/response translation, not analysis behavior.

---

# Phase 15 — Performance

Since speed is a core value of Atheon, treat performance as a feature with measurable targets.

Measure:

* startup time
* configuration load time
* file discovery time
* scan throughput (files/second)
* pattern throughput (patterns/second)
* memory usage
* output generation time

Maintain regression benchmarks so new features cannot silently degrade performance.

---

# Phase 16 — Release Criteria

A release is complete only when:

* The runtime is the only location containing analysis logic.
* Every adapter consumes the same runtime API.
* Every adapter consumes the same diagnostics schema.
* All output formats originate from the same diagnostics model.
* Configuration precedence is deterministic and documented.
* The diagnostics schema is versioned.
* Documentation covers architecture, installation, configuration, adapter usage, and troubleshooting.
* Runtime, formatter, and adapter test suites pass.
* Performance benchmarks meet established targets.

---

## Final Architecture Summary

The architecture can be reduced to four permanent layers:

```text
                Clients
──────────────────────────────────────
Pi
GitForge
CLI
GitHub Actions
Pre-commit
Future Integrations

                │

──────────────────────────────────────
Adapter Layer

Thin translators only

                │

──────────────────────────────────────
Atheon Runtime

Scanner
Rule Engine
Pattern Engine
Configuration
Diagnostics
Formatter

                │

──────────────────────────────────────
Outputs

Diagnostics Model
JSON
SARIF
Markdown
Console
Exit Codes
```

This gives you a stable foundation that scales cleanly. As new integrations emerge—whether another AI coding agent, an IDE plugin, or a different CI platform—you add only a new adapter. The runtime, diagnostic model, configuration system, and output pipeline remain unchanged, ensuring a single source of truth for Atheon's analysis behavior across every environment.
