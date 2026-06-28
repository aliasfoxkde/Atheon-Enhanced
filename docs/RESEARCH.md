# AST Pattern Research

## Problem Statement

Current Atheon-Enhanced uses regex patterns only. Regex cannot detect:
- SQL injection across string concatenation
- Command injection via string building
- Unsafe deserialization
- Context-aware security issues

## Solution: Go AST Analysis

Using Go's standard `go/ast` and `go/packages` for:
1. No external dependencies
2. Native Go code understanding
3. Precise line/column reporting

## Implementation Approach

### Phase 1: Built-in AST Patterns (using standard library only)

Built-in patterns using `go/ast`:
- `unsafe-deserialization` - encoding.BinaryUnmarshaler with user input
- `sql-injection` - String concat in database queries
- `command-injection` - exec.Command with string building
- `path-traversal` - os.Open with user input in path
- `hardcoded-credentials` - Credentials assigned without env var
- `error-not-handled` - Unchecked error returns

### Phase 2: AST Pattern File Format

Extend pattern YAML to support AST-based patterns:

```yaml
name: go-sql-injection
type: ast  # new field
language: go
pattern: |
  call(Query|Exec|Execute) with Concat(arg, userInput)
severity: high
```

### Phase 3: Integration

- Add `atheon scan --ast` flag
- AST patterns run alongside regex patterns
- Unified Finding output

## Constraints

- Must work with Go 1.21+
- No external C dependencies (no tree-sitter)
- Performance: AST parsing is slower than regex
- Option to disable AST scanning for speed

## Technical Notes

- Use `go/ast.Inspect` for tree traversal
- Use `go/packages` for package-level analysis
- AST patterns complement regex, don't replace

## See Also

- `core/ast_patterns.go` - Implementation
- `docs/ROADMAP.md` - Future plans
