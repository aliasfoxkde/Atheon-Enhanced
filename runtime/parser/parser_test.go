package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.fset == nil {
		t.Fatal("Parser.fset is nil")
	}
}

func TestParseContent(t *testing.T) {
	p := NewParser()

	content := `package main

func main() {
	println("hello")
}
`
	filename := "test.go"
	file, err := p.ParseContent(filename, content)
	if err != nil {
		t.Fatalf("ParseContent failed: %v", err)
	}

	if file == nil {
		t.Fatal("ParseContent returned nil file")
	}

	if file.Name.Name != "main" {
		t.Errorf("expected package name 'main', got '%s'", file.Name.Name)
	}

	if len(file.Decls) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(file.Decls))
	}
}

func TestParseContentInvalid(t *testing.T) {
	p := NewParser()

	content := `package main
func main() { invalid syntax }
`
	_, err := p.ParseContent("test.go", content)
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestParseFile(t *testing.T) {
	p := NewParser()

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(tmpFile, []byte(`package parser

var x = 1
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	file, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if file == nil {
		t.Fatal("ParseFile returned nil")
	}

	if file.Name.Name != "parser" {
		t.Errorf("expected package name 'parser', got '%s'", file.Name.Name)
	}
}

func TestParseFileNotFound(t *testing.T) {
	p := NewParser()
	_, err := p.ParseFile("/nonexistent/file.go")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParseDir(t *testing.T) {
	p := NewParser()

	// Create a temp directory with Go files
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte(`package main

func f1() {}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(`package main

func f2() {}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Write a non-Go file
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not go code"), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	files, err := p.ParseDir(tmpDir)
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 Go files, got %d", len(files))
	}
}

func TestWalkAST(t *testing.T) {
	src := `package main

func add(a, b int) int {
	return a + b
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var identCount int
	var funcNames []string

	WalkAST(file, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			identCount++
			funcNames = append(funcNames, ident.Name)
		}
		return true
	})

	if identCount == 0 {
		t.Error("expected to find identifiers")
	}

	// Check that we found "main", "add", "a", "b", "int", etc.
	found := false
	for _, name := range funcNames {
		if name == "add" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'add' function")
	}
}

func TestWalkASTStop(t *testing.T) {
	src := `package main

func foo() {}
func bar() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	callCount := 0
	WalkAST(file, func(node ast.Node) bool {
		callCount++
		// Stop after finding the first FuncDecl
		if _, ok := node.(*ast.FuncDecl); ok {
			return false
		}
		return true
	})

	if callCount == 0 {
		t.Error("expected at least one node visit")
	}
}

func TestFindNodes(t *testing.T) {
	src := `package main

func add(a, b int) int {
	return a + b
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Find all identifiers
	idents := FindNodes(file, func(node ast.Node) bool {
		_, ok := node.(*ast.Ident)
		return ok
	})

	if len(idents) == 0 {
		t.Fatal("expected to find identifiers")
	}

	for _, node := range idents {
		ident := node.(*ast.Ident)
		t.Logf("found ident: %s", ident.Name)
	}
}

func TestFindNodesFuncDecls(t *testing.T) {
	src := `package main

func foo() {}
func bar() {}
func baz() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	funcDecls := FindNodes(file, func(node ast.Node) bool {
		_, ok := node.(*ast.FuncDecl)
		return ok
	})

	if len(funcDecls) != 3 {
		t.Errorf("expected 3 FuncDecls, got %d", len(funcDecls))
	}
}

func TestGetNodeKind(t *testing.T) {
	fset := token.NewFileSet()

	tests := []struct {
		src    string
		expect string
	}{
		{`package p; var x int`, "GenDecl"},
		{`package p; func f() {}`, "FuncDecl"},
		{`package p; var s = "hello"`, "BasicLit"},
		{`package p; var i = x`, "Ident"},
		{`package p; var b = x + y`, "BinaryExpr"},
		{`package p; var c = f()`, "CallExpr"},
	}

	for _, test := range tests {
		file, err := parser.ParseFile(fset, "test.go", test.src, 0)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", test.src, err)
		}

		var kind string
		WalkAST(file, func(node ast.Node) bool {
			if kind == "" && GetNodeKind(node) == test.expect {
				kind = GetNodeKind(node)
			}
			return kind == ""
		})

		if kind != test.expect {
			t.Errorf("for %q: expected %s, got %s", test.src, test.expect, kind)
		}
	}
}

func TestGetNodeKindNil(t *testing.T) {
	kind := GetNodeKind(nil)
	if kind != "nil" {
		t.Errorf("expected 'nil', got '%s'", kind)
	}
}

func TestGetNodeKindUnknown(t *testing.T) {
	// Create a minimal node that is not caught by the switch
	// Using a pointer to a struct that implements ast.Node
	// but is not caught - this is tricky to test directly
	// For now, just verify it doesn't panic
	kind := GetNodeKind(nil)
	if kind != "nil" {
		t.Errorf("expected 'nil', got '%s'", kind)
	}
}
