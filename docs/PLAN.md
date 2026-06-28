# AST Pattern Implementation Plan

## Overview

Add AST-based pattern analysis using Go's standard `go/ast` library to detect security issues that regex cannot find.

## Phases

### Phase 1: Core AST Scanner
- [x] Create `core/ast_patterns.go`
- [x] Define `ASTFinding` and `ASTPattern` types
- [x] Implement `ScanFileAST()` and `ScanDirAST()`
- [ ] Add pattern registration

### Phase 2: Built-in Patterns
- [ ] `unsafe-deserialization` - encoding.Unmarshal with user input
- [ ] `sql-injection` - String concat in Query/Exec
- [ ] `command-injection` - exec.Command with concat
- [ ] `path-traversal` - os.Open with user input
- [ ] `hardcoded-credentials` - Direct assignment of secrets
- [ ] `error-not-handled` - Unchecked error returns

### Phase 3: CLI Integration
- [ ] Add `--ast` flag to `atheon scan`
- [ ] Add `--ast-only` flag to skip regex
- [ ] Add `--categories=ast-security` for AST patterns

### Phase 4: Pattern Format Extension
- [ ] Extend YAML format with `type: ast`
- [ ] Add `language: go` field
- [ ] Document AST pattern syntax

## Files to Create/Modify

| File | Change |
|------|--------|
| `core/ast_patterns.go` | NEW - AST pattern engine |
| `cmd/atheon/main.go` | Add --ast flags |
| `community/ast-security/` | NEW - AST pattern YAML files |
| `docs/guides/AST_PATTERNS.md` | NEW - User guide |

## Testing

- [ ] Add `core/ast_patterns_test.go`
- [ ] Test each built-in pattern
- [ ] Integration test with CLI

## Rollout

1. CLI flag `--ast` enables AST scanning
2. AST patterns in separate category
3. Performance: warn if scanning large codebases
