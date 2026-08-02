package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestExprTokens_Direct tests exprTokens directly with parsed AST expressions
func TestExprTokens_Direct(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantOpts []string // substring options we expect in tokens
	}{
		{
			name: "ident_expression",
			code: `package main
func main() { x := 1; _ = x }`,
			wantOpts: []string{"VAR"},
		},
		{
			name: "basic_lit_int",
			code: `package main
func main() { x := 42; _ = x }`,
			wantOpts: []string{"LIT_INT"},
		},
		{
			name: "basic_lit_float",
			code: `package main
func main() { x := 3.14; _ = x }`,
			wantOpts: []string{"LIT_FLOAT"},
		},
		{
			name: "basic_lit_char",
			code: `package main
func main() { x := 'a'; _ = x }`,
			wantOpts: []string{"LIT_CHAR"},
		},
		{
			name: "basic_lit_string",
			code: `package main
func main() { x := "hello"; _ = x }`,
			wantOpts: []string{"LIT_STRING"},
		},
		{
			name: "binary_expr",
			code: `package main
func main() { x := 1 + 2; _ = x }`,
			wantOpts: []string{"EXPR_BIN", "+"},
		},
		{
			name: "unary_expr_negate",
			code: `package main
func main() { x := -5; _ = x }`,
			wantOpts: []string{"EXPR_UN", "-"},
		},
		{
			name: "unary_expr_not",
			code: `package main
func main() { x := !true; _ = x }`,
			wantOpts: []string{"EXPR_UN", "!"},
		},
		{
			name: "call_expr_no_args",
			code: `package main
func foo() {}
func main() { foo() }`,
			wantOpts: []string{"CALL", "VAR"},
		},
		{
			name: "call_expr_with_args",
			code: `package main
func foo(int) {}
func main() { foo(1) }`,
			wantOpts: []string{"CALL", "LIT_INT"},
		},
		{
			name: "selector_expr",
			code: `package main
type T struct{}
func (t T) Method() {}
func main() { var t T; t.Method() }`,
			wantOpts: []string{"SEL", "VAR"},
		},
		{
			name: "index_expr",
			code: `package main
func main() { arr := []int{1,2,3}; _ = arr[0] }`,
			wantOpts: []string{"EXPR_INDEX", "LIT_INT"},
		},
		{
			name: "slice_expr",
			code: `package main
func main() { arr := []int{1,2,3}; _ = arr[:] }`,
			wantOpts: []string{"EXPR_SLICE"},
		},
		{
			name: "composite_lit",
			code: `package main
type Point struct{ X, Y int }
func main() { p := Point{X: 1, Y: 2}; _ = p }`,
			wantOpts: []string{"EXPR_COMPOSITE", "LIT_INT"},
		},
		{
			name: "func_literal",
			code: `package main
func main() { fn := func() {}; _ = fn }`,
			wantOpts: []string{"FUNC_LIT"},
		},
		{
			name: "paren_expr",
			code: `package main
func main() { x := (1 + 2); _ = x }`,
			wantOpts: []string{"EXPR_PAREN", "EXPR_BIN"},
		},
		{
			name: "type_assert_expr",
			code: `package main
func main() { var i interface{} = 1; _ = i.(int) }`,
			wantOpts: []string{"EXPR_TYPEASSERT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			detector := NewCloneDetector(DefaultCloneDetectionConfig())

			var tokens []string
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					for _, stmt := range fn.Body.List {
						tokens = append(tokens, detector.stmtTokens(stmt)...)
					}
				}
			}

			for _, want := range tt.wantOpts {
				if !containsToken(tokens, want) {
					t.Errorf("exprTokens missing expected token containing %q, got %v", want, tokens)
				}
			}
		})
	}
}

func containsToken(tokens []string, substr string) bool {
	for _, tok := range tokens {
		if len(tok) >= len(substr) && (tok == substr || len(substr) < len(tok)) {
			if substringMatch(tok, substr) {
				return true
			}
		}
	}
	return false
}

func substringMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestExprTokens_BinaryExpr_Subtraction tests binary expr with subtraction
func TestExprTokens_BinaryExpr_Subtraction(t *testing.T) {
	code := `package main
func main() { x := 1 - 2; _ = x }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "EXPR_BIN") {
		t.Errorf("expected EXPR_BIN, got %v", tokens)
	}
	if !sliceContains(tokens, "-") {
		t.Errorf("expected -, got %v", tokens)
	}
}

// TestExprTokens_BinaryExpr_Division tests binary expr with division
func TestExprTokens_BinaryExpr_Division(t *testing.T) {
	code := `package main
func main() { x := 10 / 2; _ = x }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "/") {
		t.Errorf("expected /, got %v", tokens)
	}
}

// TestExprTokens_UnaryExpr_Address tests unary expr with address operator
func TestExprTokens_UnaryExpr_Address(t *testing.T) {
	code := `package main
func main() { x := 5; y := &x; _ = y }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "&") {
		t.Errorf("expected &, got %v", tokens)
	}
}

// TestExprTokens_CallExpr_MultipleArgs tests call expr with multiple args
func TestExprTokens_CallExpr_MultipleArgs(t *testing.T) {
	code := `package main
func add(a, b int) int { return a + b }
func main() { x := add(1, 2); _ = x }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "CALL") {
		t.Errorf("expected CALL, got %v", tokens)
	}
}

// TestExprTokens_CompositeLit_Slice tests composite literal for slice
func TestExprTokens_CompositeLit_Slice(t *testing.T) {
	code := `package main
func main() { x := []int{1, 2, 3}; _ = x }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "EXPR_COMPOSITE") {
		t.Errorf("expected EXPR_COMPOSITE, got %v", tokens)
	}
}

// TestExprTokens_CompositeLit_Map tests composite literal for map
func TestExprTokens_CompositeLit_Map(t *testing.T) {
	code := `package main
func main() { x := map[string]int{"a": 1}; _ = x }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "EXPR_COMPOSITE") {
		t.Errorf("expected EXPR_COMPOSITE, got %v", tokens)
	}
}

// TestExprTokens_SelectorExpr_Chain tests chained selector expressions
func TestExprTokens_SelectorExpr_Chain(t *testing.T) {
	code := `package main
type A struct{ B struct{ C int } }
func main() { var a A; _ = a.B.C }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "SEL") {
		t.Errorf("expected SEL, got %v", tokens)
	}
}

// TestExprTokens_SliceExpr_WithBounds tests slice with bounds
func TestExprTokens_SliceExpr_WithBounds(t *testing.T) {
	code := `package main
func main() { arr := []int{1,2,3,4,5}; _ = arr[1:3] }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "EXPR_SLICE") {
		t.Errorf("expected EXPR_SLICE, got %v", tokens)
	}
}

// TestExprTokens_TypeAssertExpr_OkPattern tests type assertion with ok pattern
func TestExprTokens_TypeAssertExpr_OkPattern(t *testing.T) {
	code := `package main
func main() { var i interface{} = 1; if v, ok := i.(int); ok { _ = v } }`
	tokens := getExprTokens(t, code)
	if !sliceContains(tokens, "EXPR_TYPEASSERT") {
		t.Errorf("expected EXPR_TYPEASSERT, got %v", tokens)
	}
}

func getExprTokens(t *testing.T, code string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	detector := NewCloneDetector(DefaultCloneDetectionConfig())

	var tokens []string
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			for _, stmt := range fn.Body.List {
				tokens = append(tokens, detector.stmtTokens(stmt)...)
			}
		}
	}
	return tokens
}

func sliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
