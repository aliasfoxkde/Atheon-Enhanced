package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Test clone detection helper functions

func TestExprTokens_Ident(t *testing.T) {
	code := `package main

func main() {
	x := 1
	y := 2
	print(x, y)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "ident.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestExprTokens_BasicLit(t *testing.T) {
	code := `package main

func main() {
	x := 42
	y := 3.14
	z := "hello"
	_ = x
	_ = y
	_ = z
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "lit.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestExprTokens_BinaryExpr(t *testing.T) {
	code := `package main

func main() {
	x := 1 + 2
	y := 3 * 4
	z := x + y
	_ = z
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "binary.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestExprTokens_UnaryExpr(t *testing.T) {
	code := `package main

func main() {
	x := 5
	y := -x
	z := !true
	_ = y
	_ = z
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unary.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestExprTokens_CallExpr(t *testing.T) {
	code := `package main

func foo() int { return 1 }
func bar() int { return 2 }

func main() {
	x := foo()
	y := bar()
	_ = x
	_ = y
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "call.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}
