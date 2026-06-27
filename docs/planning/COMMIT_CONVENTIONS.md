# Commit Conventions — Atheon Enhanced

## Format

Atheon Enhanced follows the [Conventional Commits](https://www.conventionalcommits.org/) specification.

```
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

### Type Prefixes

| Type | Use For | Example |
|------|---------|---------|
| `feat` | New feature | `feat(mcp): add cancel request handler` |
| `fix` | Bug fix | `fix(core): guard against nil bundle panic` |
| `docs` | Documentation only | `docs: update CONTRIBUTING.md` |
| `test` | Adding or updating tests | `test: add MCP panic recovery tests` |
| `refactor` | Code change with no behavior change | `refactor(bundle): extract hash computation` |
| `chore` | Build, tooling, dependency updates | `chore: bump golangci-lint to v2.1.0` |
| `ci` | CI/CD configuration changes | `ci: add Go 1.25 to test matrix` |
| `build` | Build system or dependency changes | `build(deps): migrate yaml.v3 → goccy/go-yaml` |
| `perf` | Performance improvements | `perf(core): cache compiled regex patterns` |
| `release` | Release-related actions | `release: tag v0.7.0` |
| `revert` | Reverting a previous commit | `revert: revert commit abc1234` |

### Scope

Scope is the package or module affected (use folder names):

- `cmd/atheon` → `cmd/atheon`
- `cmd/mcp` → `mcp`
- `core` → `core`
- `bundler` → `bundler`
- `community` → `community`
- `.github/workflows` → `ci`, `security`, `release`
- `docs` → `docs`

### Breaking Changes

Add `!` before the colon:
```
feat!]: remove deprecated --legacy flag
```

Or add `BREAKING CHANGE:` in the footer:
```
BREAKING CHANGE: ScanDir signature changed to take ctx, root, opts
```

### Rules

1. **Title line**: max 72 characters, imperative mood, lowercase after colon
2. **Body**: wrap at 72 characters, explain *what* and *why*
3. **Footer**: reference issues/PRs: `Closes #123`, `Fixes aliasfoxkde/Atheon-Enhanced#456`
4. **Co-Authored-By**: use for AI-generated commits:
   ```
   Co-Authored-By: Claude <noreply@anthropic.com>
   ```

### Envelope

**Enforced by**: `make commitlint` (local) and `commitlint` step in CI (when added)

**NOT enforced by**: git hooks (currently), existing CI (gaps documented in SDLC_AUDIT)

### Examples

```
feat(mcp): add concurrent request cap of 50

Prevents memory exhaustion under flood. Each incoming request increments
mcpInflight atomic counter; if it exceeds cap, returns -32001 error.

Closes #97
```

```
fix(test): use cmd.Output() instead of CombinedOutput() to avoid pipe race

CombinedOutput() races with subprocess exit on Go 1.22 in CI. Using
Output() captures only stdout, avoiding the stderr pipe timing issue.

Fixes #107
Co-Authored-By: Claude <noreply@anthropic.com>
```

```
chore: update Go test matrix to 1.21-1.26

Go 1.26 is the latest stable. Update CI matrix and verify backward
compatibility down to Go 1.21 minimum supported version.
```

```
docs: add pre-commit hook documentation

Document how to install and configure the pre-commit hook for
local validation before pushing.
```

## PR Title Convention

PR titles follow the same format and become the squash-merge commit subject.

**Format**: `type(scope): description`

Examples:
- `fix(test): use cmd.Output() instead of CombinedOutput() to avoid pipe race`
- `docs: fix CHANGELOG - version Wave 9/10 content`
- `build(deps): migrate gopkg.in/yaml.v3 → github.com/goccy/go-yaml`
