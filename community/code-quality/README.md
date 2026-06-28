# code-quality

Patterns in this category:

| Pattern | Description | Severity |
|---------|-------------|----------|
| `auto-confirm` | Pattern: auto confirm. Pattern in category . This is a low-severity | low |
| `bare-exception-block` | Pattern: bare exception block. Pattern in category . This is | low |
| `build-force` | Pattern: build force. Pattern in category . This is a low-severity | low |
| `commented-out-code` | Pattern: commented out code block. Pattern in category . This | low |
| `commented-password` | Pattern: commented password. Pattern in category code-quality. Detects passwords or secrets in code comments. | medium |
| `commented-secret` | Pattern: Detect commented-out secrets and keys | critical |
| `console-log-production` | Pattern: Detect console.log/debug/warn left in production code | low |
| `console-log` | Detected debug or logging statement: console log. Debug code should | low |
| `cruft` | Pattern: cruft. Pattern in category . This is a low-severity | low |
| `debug-breakpoint` | Detected debug or logging statement: debug breakpoint. Debug code | low |
| `debug-statement` | Detected debug or logging statement: debug statement. Debug code | low |
| `deprecated-api` | Pattern: Detect usage of deprecated APIs | medium |
| `deprecated-function` | Pattern: deprecated function. Pattern in category . This is | low |
| `direct-sql-query` | Detected potential SQL injection risk: direct sql query. User input | low |
| `dummy-function` | Pattern: dummy function. Pattern in category . This is a low-severity | low |
| `empty-catch-block` | Pattern: empty catch block. Pattern in category . This is a | low |
| `empty-code-block` | Pattern: empty code block. Pattern in category code-quality. Detects empty if/for/while blocks that may indicate unfinished code. | low |
| `eval-usage` | Pattern: Detect eval() and similar code execution functions | critical |
| `fake-data` | Pattern: fake data. Pattern in category . This is a low-severity | low |
| `fixme-comment` | Detected TODO/FIXME comment: fixme comment. Pending tasks should | low |
| `fmt-println-prod` | Pattern: fmt println prod. Pattern in category . This is a low-severity | low |
| `git-clean-force` | Pattern: git clean force. Pattern in category . This is a low-severity | low |
| `git-force-push` | Pattern: git force push. Pattern in category . This is a low-severity | low |
| `git-hard-reset` | Pattern: git hard reset. Pattern in category . This is a low-severity | low |
| `global-variable` | Pattern: global variable. Pattern in category . This is a low-severity | low |
| `hardcoded-ip` | Pattern: Detect hardcoded IP addresses | high |
| `hardcoded-localhost` | Detected hardcoded value: hardcoded localhost. Hardcoded values | low |
| `hardcoded-url` | Detected hardcoded value: hardcoded url. Hardcoded values reduce | low |
| `ignored-error-underscore` | Pattern: Detect error values assigned to underscore (suppressed errors in Go, Python, Rust) | medium |
| `infinite-loop` | Pattern: Detect obvious infinite loops | high |
| `innerhtml-usage` | Pattern: Detect innerHTML assignments (XSS risk) | high |
| `insecure-flag` | Pattern: insecure flag. Pattern in category . This is a low-severity | low |
| `insecure-randomness` | Pattern: Detect Math.random() or similar for security-sensitive randomness | high |
| `magic-number` | Detects common magic numbers (86400 seconds/day, 1024 bytes/KB, 65535 max unsigned 16-bit, etc.) that should be replaced with named constants for readability and maintainability. | low |
| `missing-await` | Pattern: Detect async calls without await (missing await in async functions) | medium |
| `mock-stub` | Pattern: mock stub. Pattern in category . This is a low-severity | low |
| `package-manager-force` | Pattern: package manager force. Pattern in category . This is | low |
| `panic-in-handler` | Pattern: panic in handler. Pattern in category . This is a low-severity | low |
| `placeholder-code` | Pattern: placeholder code. Pattern in category . This is a low-severity | low |
| `print-debug-leftover` | Pattern: print debug leftover. Pattern in category code-quality. Detects debug print statements left in production code. | low |
| `print-statement` | Pattern: Detect print statements that may be debug leftovers | low |
| `race-condition` | Pattern: Detect potential race conditions (shared variable modified without synchronization) | high |
| `skip-hooks` | Detected potential CI/security bypass: skip hooks. Bypass mechanisms | low |
| `skip-tests` | Detected potential CI/security bypass: skip tests. Bypass mechanisms | low |
| `sleep-in-test` | Pattern: sleep in test. Pattern in category . This is a low-severity | low |
| `sql-concatenation` | Pattern: Detect SQL built via string concatenation (SQL injection risk) | critical |
| `temporary-code` | Pattern: temporary code. Pattern in category . This is a low-severity | low |
| `todo-comment` | Detected TODO/FIXME comment: todo comment. Pending tasks should | low |
| `todo-fixme` | Pattern: Detect TODO/FIXME/HACK comments that need attention | low |
| `todo-stub` | Detected TODO/FIXME comment: todo stub. Pending tasks should be | low |
| `unreachable-code` | Pattern: unreachable code. Pattern in category . This is a low-severity | low |
| `unused-import-comment` | Pattern: unused import comment. Pattern in category . This is | low |
| `unused-variable` | Pattern: Detect unused variable declarations | low |

---
*Auto-generated from pattern YAML files*
