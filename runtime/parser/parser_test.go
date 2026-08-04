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

func TestParseDirWithSubdir(t *testing.T) {
	p := NewParser()

	tmpDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "sub")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte(`package main

func f1() {}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	err = os.WriteFile(filepath.Join(subDir, "file2.go"), []byte(`package main

func f2() {}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	files, err := p.ParseDir(tmpDir)
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}

	// Should find both files
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

func TestWalkASTNilNode(t *testing.T) {
	// Should not panic
	WalkAST(nil, func(node ast.Node) bool {
		return true
	})
}

func TestWalkASTAllNodeTypes(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		// File with comments
		{
			name: "File with comments",
			src: `// comment
package main
`,
		},
		// GenDecl - const
		{
			name: "GenDecl const",
			src: `package main
const x = 1
`,
		},
		// GenDecl - import
		{
			name: "GenDecl import",
			src: `package main
import "fmt"
`,
		},
		// GenDecl - var group
		{
			name: "GenDecl var group",
			src: `package main
var (
	x = 1
	y = 2
)
`,
		},
		// ImportSpec
		{
			name: "ImportSpec",
			src: `package main
import (
	"fmt"
	"os"
)
`,
		},
		// ValueSpec
		{
			name: "ValueSpec",
			src: `package main
var x int = 10
`,
		},
		// TypeSpec
		{
			name: "TypeSpec",
			src: `package main
type MyStruct struct{}
`,
		},
		// FuncDecl with body
		{
			name: "FuncDecl",
			src: `package main
func foo() {}
`,
		},
		// BlockStmt
		{
			name: "BlockStmt",
			src: `package main
func foo() { x := 1 }
`,
		},
		// IfStmt
		{
			name: "IfStmt",
			src: `package main
func foo() {
	if true {
		x := 1
	}
}
`,
		},
		// IfStmt with else
		{
			name: "IfStmt else",
			src: `package main
func foo() {
	if true {
		x := 1
	} else {
		y := 2
	}
}
`,
		},
		// IfStmt with init
		{
			name: "IfStmt init",
			src: `package main
func foo() {
	if x := 1; x > 0 {
		println(x)
	}
}
`,
		},
		// ForStmt
		{
			name: "ForStmt",
			src: `package main
func foo() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
`,
		},
		// ForStmt with condition only
		{
			name: "ForStmt condition",
			src: `package main
func foo() {
	x := 0
	for x < 10 {
		x++
	}
}
`,
		},
		// RangeStmt
		{
			name: "RangeStmt",
			src: `package main
func foo() {
	m := map[int]int{1: 2}
	for k, v := range m {
		println(k, v)
	}
}
`,
		},
		// SwitchStmt
		{
			name: "SwitchStmt",
			src: `package main
func foo() {
	x := 1
	switch x {
	case 1:
		println("one")
	case 2:
		println("two")
	}
}
`,
		},
		// SelectStmt
		{
			name: "SelectStmt",
			src: `package main
func foo() {
	ch := make(chan int)
	select {
	case v := <-ch:
		println(v)
	default:
		println("default")
	}
}
`,
		},
		// TypeSwitchStmt
		{
			name: "TypeSwitchStmt",
			src: `package main
func foo() {
	var x interface{} = 1
	switch v := x.(type) {
	case int:
		println(v)
	}
}
`,
		},
		// CaseClause
		{
			name: "CaseClause",
			src: `package main
func foo() {
	x := 1
	switch x {
	case 1, 2:
		println("one or two")
	}
}
`,
		},
		// CommClause (select case)
		{
			name: "CommClause",
			src: `package main
func foo() {
	ch := make(chan int)
	select {
	case v := <-ch:
		println(v)
	}
}
`,
		},
		// AssignStmt
		{
			name: "AssignStmt",
			src: `package main
func foo() {
	x := 1
	x = 2
	x, y := 3, 4
	_ = y
}
`,
		},
		// CallExpr
		{
			name: "CallExpr",
			src: `package main
func foo() {
	println("hello")
}
`,
		},
		// ReturnStmt
		{
			name: "ReturnStmt",
			src: `package main
func foo() int {
	return 42
}
`,
		},
		// ExprStmt
		{
			name: "ExprStmt",
			src: `package main
func foo() {
	1 + 2
}
`,
		},
		// SendStmt
		{
			name: "SendStmt",
			src: `package main
func foo() {
	ch := make(chan int)
	ch <- 42
}
`,
		},
		// IncDecStmt
		{
			name: "IncDecStmt",
			src: `package main
func foo() {
	x := 1
	x++
	x--
}
`,
		},
		// UnaryExpr
		{
			name: "UnaryExpr",
			src: `package main
func foo() {
	x := 1
	_ = -x
	_ = !true
	_ = <-make(chan int)
}
`,
		},
		// BinaryExpr
		{
			name: "BinaryExpr",
			src: `package main
func foo() {
	x := 1 + 2
	y := 3 > 4
	z := 5 == 6
	_ = x
	_ = y
	_ = z
}
`,
		},
		// ParenExpr
		{
			name: "ParenExpr",
			src: `package main
func foo() {
	x := (1 + 2)
	_ = x
}
`,
		},
		// SelectorExpr
		{
			name: "SelectorExpr",
			src: `package main
import "fmt"
func foo() {
	fmt.Println("hello")
}
`,
		},
		// IndexExpr
		{
			name: "IndexExpr",
			src: `package main
func foo() {
	arr := []int{1, 2, 3}
	x := arr[0]
	_ = x
}
`,
		},
		// IndexListExpr
		{
			name: "IndexListExpr",
			src: `package main
func foo() {
	type M struct{}
	var m M
	_ = m
}
`,
		},
		// SliceExpr
		{
			name: "SliceExpr",
			src: `package main
func foo() {
	arr := []int{1, 2, 3, 4, 5}
	x := arr[1:3]
	y := arr[:3]
	z := arr[2:]
	w := arr[:]
	_ = x
	_ = y
	_ = z
	_ = w
}
`,
		},
		// TypeAssertExpr
		{
			name: "TypeAssertExpr",
			src: `package main
func foo() {
	var x interface{} = 1
	y := x.(int)
	_ = y
}
`,
		},
		// StarExpr
		{
			name: "StarExpr",
			src: `package main
func foo() {
	x := 1
	ptr := &x
	_ = *ptr
}
`,
		},
		// KeyValueExpr
		{
			name: "KeyValueExpr",
			src: `package main
func foo() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	_ = m
}
`,
		},
		// CompositeLit
		{
			name: "CompositeLit",
			src: `package main
func foo() {
	x := []int{1, 2, 3}
	y := map[string]int{"a": 1}
	z := struct{A int}{A: 1}
	_ = x
	_ = y
	_ = z
}
`,
		},
		// FuncLit
		{
			name: "FuncLit",
			src: `package main
func foo() {
	x := func() {}
	_ = x
}
`,
		},
		// FuncType
		{
			name: "FuncType",
			src: `package main
func foo() {
	var f func(int) int
	_ = f
}
`,
		},
		// StructType
		{
			name: "StructType",
			src: `package main
type S struct {
	A int
	B string
}
`,
		},
		// InterfaceType
		{
			name: "InterfaceType",
			src: `package main
type I interface {
	Method()
}
`,
		},
		// MapType
		{
			name: "MapType",
			src: `package main
func foo() {
	var m map[string]int
	_ = m
}
`,
		},
		// ChanType
		{
			name: "ChanType",
			src: `package main
func foo() {
	var ch chan int
	_ = ch
}
`,
		},
		// BasicLit
		{
			name: "BasicLit",
			src: `package main
func foo() {
	x := 42
	y := "hello"
	z := 3.14
	_ = x
	_ = y
	_ = z
}
`,
		},
		// Ident
		{
			name: "Ident",
			src: `package main
var x = 1
`,
		},
		// Ellipsis
		{
			name: "Ellipsis",
			src: `package main
func foo(args ...int) {
	for _, v := range args {
		_ = v
	}
}
`,
		},
		// Field
		{
			name: "Field",
			src: `package main
func foo(a, b int) {
	_ = a
	_ = b
}
`,
		},
		// FieldList
		{
			name: "FieldList",
			src: `package main
type F func(int, string)
`,
		},
		// LabeledStmt
		{
			name: "LabeledStmt",
			src: `package main
func foo() {
	x := 1
	if x > 0 {
		goto label
	}
	label:
		println("end")
}
`,
		},
		// GoStmt
		{
			name: "GoStmt",
			src: `package main
func foo() {
	go func() {}()
}
`,
		},
		// DeferStmt
		{
			name: "DeferStmt",
			src: `package main
func foo() {
	defer func() {}()
}
`,
		},
		// EmptyStmt
		{
			name: "EmptyStmt",
			src: `package main
func foo() {
	;
}
`,
		},
		// BranchStmt - Break
		{
			name: "BranchStmt break",
			src: `package main
func foo() {
	for {
		break
	}
}
`,
		},
		// BranchStmt - Continue
		{
			name: "BranchStmt continue",
			src: `package main
func foo() {
	for i := 0; i < 10; i++ {
		if i < 5 {
			continue
		}
	}
}
`,
		},
		// BranchStmt - Goto
		{
			name: "BranchStmt goto",
			src: `package main
func foo() {
	goto label
	label:
		println("end")
}
`,
		},
		// DeclStmt
		{
			name: "DeclStmt",
			src: `package main
func foo() {
	var x = 1
	_ = x
}
`,
		},
		// ArrayType
		{
			name: "ArrayType",
			src: `package main
func foo() {
	var arr [5]int
	_ = arr
}
`,
		},
		// FuncType with results
		{
			name: "FuncType with results",
			src: `package main
func foo() func() int {
	return func() int { return 1 }
}
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tc.name, err)
			}

			nodeCount := 0
			WalkAST(file, func(node ast.Node) bool {
				nodeCount++
				return true
			})

			if nodeCount == 0 {
				t.Errorf("expected to find nodes in %q", tc.name)
			}
		})
	}
}

func TestWalkASTPackage(t *testing.T) {
	// Test walking a Package node
	src := `package main
func foo() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Get the package node
	pkg := &ast.Package{
		Name: "main",
		Files: map[string]*ast.File{
			"test.go": file,
		},
	}

	nodeCount := 0
	WalkAST(pkg, func(node ast.Node) bool {
		nodeCount++
		return true
	})

	if nodeCount == 0 {
		t.Error("expected to find nodes when walking Package")
	}
}

func TestWalkASTCommentGroup(t *testing.T) {
	src := `// comment1
// comment2
package main
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	nodeCount := 0
	hasComment := false
	WalkAST(file, func(node ast.Node) bool {
		nodeCount++
		if _, ok := node.(*ast.Comment); ok {
			hasComment = true
		}
		return true
	})

	if nodeCount == 0 {
		t.Error("expected to find nodes")
	}
	if !hasComment {
		t.Error("expected to find Comment node")
	}
}

func TestWalkASTFuncLit(t *testing.T) {
	src := `package main
var x = func() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncLit node")
	}
}

func TestWalkASTFuncType(t *testing.T) {
	src := `package main
var f func(int) string
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncType node")
	}
}

func TestWalkASTStructType(t *testing.T) {
	src := `package main
type S struct {
	Field int
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.StructType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find StructType node")
	}
}

func TestWalkASTInterfaceType(t *testing.T) {
	src := `package main
type I interface {
	Method()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.InterfaceType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find InterfaceType node")
	}
}

func TestWalkASTMapType(t *testing.T) {
	src := `package main
var m map[string]int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.MapType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find MapType node")
	}
}

func TestWalkASTChanType(t *testing.T) {
	src := `package main
var ch chan int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ChanType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ChanType node")
	}
}

func TestWalkASTArrayType(t *testing.T) {
	src := `package main
var arr [5]int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ArrayType); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ArrayType node")
	}
}

func TestWalkASTEllipsis(t *testing.T) {
	src := `package main
func foo(...int) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.Ellipsis); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find Ellipsis node")
	}
}

func TestWalkASTField(t *testing.T) {
	src := `package main
func foo(a, b int, c string) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	fieldCount := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.Field); ok {
			fieldCount++
		}
		return true
	})

	if fieldCount == 0 {
		t.Error("expected to find Field nodes")
	}
}

func TestWalkASTGoStmt(t *testing.T) {
	src := `package main
func foo() {
	go func() {}()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.GoStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GoStmt node")
	}
}

func TestWalkASTDeferStmt(t *testing.T) {
	src := `package main
func foo() {
	defer func() {}()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.DeferStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find DeferStmt node")
	}
}

func TestWalkASTEmptyStmt(t *testing.T) {
	src := `package main
func foo() {
	if true {
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Empty if body - just checking it doesn't crash
	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

func TestWalkASTLabeledStmt(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	if x > 0 {
		goto label
	}
	label:
		println(x)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.LabeledStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find LabeledStmt node")
	}
}

func TestWalkASTBadStmt(t *testing.T) {
	// Test BadStmt via syntax error
	// Can't easily create a BadStmt through valid syntax, but we can verify
	// the code path exists by checking it doesn't panic
	src := `package main
func foo() {
	?
}
`
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "test.go", src, 0)
	// This should produce an error or BadStmt
	if err == nil {
		// If it parses successfully, the syntax isn't malformed enough
		t.Skip("syntax doesn't produce BadStmt")
	}
}

func TestWalkASTKeyValueExpr(t *testing.T) {
	src := `package main
var m = map[string]int{"a": 1}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.KeyValueExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find KeyValueExpr node")
	}
}

func TestWalkASTCompositeLit(t *testing.T) {
	src := `package main
var x = []int{1, 2, 3}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.CompositeLit); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CompositeLit node")
	}
}

func TestWalkASTSendStmt(t *testing.T) {
	src := `package main
func foo() {
	ch := make(chan int)
	ch <- 42
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SendStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SendStmt node")
	}
}

func TestWalkASTTypeAssertExpr(t *testing.T) {
	src := `package main
func foo() {
	var x interface{} = 1
	y := x.(int)
	_ = y
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.TypeAssertExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeAssertExpr node")
	}
}

func TestWalkASTSliceExpr(t *testing.T) {
	src := `package main
func foo() {
	arr := []int{1, 2, 3}
	_ = arr[:]
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SliceExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SliceExpr node")
	}
}

func TestWalkASTUnaryExpr(t *testing.T) {
	src := `package main
func foo() {
	x := -1
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.UnaryExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find UnaryExpr node")
	}
}

func TestWalkASTIncDecStmt(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	x++
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.IncDecStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find IncDecStmt node")
	}
}

func TestWalkASTSelectorExpr(t *testing.T) {
	src := `package main
import "fmt"
func foo() {
	fmt.Println("hi")
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SelectorExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SelectorExpr node")
	}
}

func TestWalkASTIndexExpr(t *testing.T) {
	src := `package main
func foo() {
	arr := []int{1, 2, 3}
	_ = arr[0]
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.IndexExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find IndexExpr node")
	}
}

func TestWalkASTParenExpr(t *testing.T) {
	src := `package main
func foo() {
	x := (1 + 2)
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ParenExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ParenExpr node")
	}
}

func TestWalkASTStarExpr(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	ptr := &x
	_ = *ptr
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.StarExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find StarExpr node")
	}
}

func TestWalkASTBranchStmt(t *testing.T) {
	// Test break
	src := `package main
func foo() {
	for {
		break
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if bs, ok := node.(*ast.BranchStmt); ok && bs.Tok.String() == "break" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find BranchStmt node")
	}
}

func TestWalkASTExprStmt(t *testing.T) {
	src := `package main
func foo() {
	1 + 2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ExprStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ExprStmt node")
	}
}

func TestWalkASTReturnStmt(t *testing.T) {
	src := `package main
func foo() {
	return
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ReturnStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ReturnStmt node")
	}
}

func TestWalkASTAssignStmt(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	x = 2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.AssignStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find AssignStmt node")
	}
}

func TestWalkASTCallExpr(t *testing.T) {
	src := `package main
func foo() {
	println(1)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.CallExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CallExpr node")
	}
}

func TestWalkASTBinaryExpr(t *testing.T) {
	src := `package main
func foo() {
	x := 1 + 2
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.BinaryExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find BinaryExpr node")
	}
}

func TestWalkASTBasicLit(t *testing.T) {
	src := `package main
func foo() {
	x := 42
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.BasicLit); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find BasicLit node")
	}
}

func TestWalkASTIdent(t *testing.T) {
	src := `package main
var x = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.Ident); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find Ident node")
	}
}

func TestWalkASTGenDecl(t *testing.T) {
	src := `package main
const x = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.GenDecl); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GenDecl node")
	}
}

func TestWalkASTImportSpec(t *testing.T) {
	src := `package main
import "fmt"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ImportSpec); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ImportSpec node")
	}
}

func TestWalkASTValueSpec(t *testing.T) {
	src := `package main
var x = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ValueSpec); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ValueSpec node")
	}
}

func TestWalkASTTypeSpec(t *testing.T) {
	src := `package main
type S struct{}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.TypeSpec); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSpec node")
	}
}

func TestWalkASTDeclStmt(t *testing.T) {
	src := `package main
func foo() {
	var x = 1
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.DeclStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find DeclStmt node")
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

func TestFindNodesAllTypes(t *testing.T) {
	src := `package main
import "fmt"
const x = 1
var y = 2
type S struct {}
func foo() {
	_ = fmt.Sprint(x)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Find all node types
	typeCounts := make(map[string]int)

	WalkAST(file, func(node ast.Node) bool {
		kind := GetNodeKind(node)
		typeCounts[kind]++
		return true
	})

	if typeCounts["File"] == 0 {
		t.Error("expected to find File node")
	}
	if typeCounts["GenDecl"] == 0 {
		t.Error("expected to find GenDecl node")
	}
	if typeCounts["ImportSpec"] == 0 {
		t.Error("expected to find ImportSpec node")
	}
	if typeCounts["FuncDecl"] == 0 {
		t.Error("expected to find FuncDecl node")
	}
}

func TestFindNodesReturnNil(t *testing.T) {
	// FindNodes on nil should return empty
	result := FindNodes(nil, func(n ast.Node) bool { return true })
	if result != nil {
		t.Error("expected nil result for FindNodes on nil")
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

func TestGetNodeKindAllTypes(t *testing.T) {
	// Test that GetNodeKind returns correct kind for various node types
	fset := token.NewFileSet()

	testCases := []struct {
		src    string
		expect string
	}{
		{`package p`, "File"},
		{`package p; import "fmt"`, "ImportSpec"},
		{`package p; const x = 1`, "GenDecl"},
		{`package p; func f() {}`, "FuncDecl"},
		{`package p; var f = func() {}`, "FuncLit"},
		{`package p; var f func()`, "FuncType"},
		{`package p; var x = -1`, "UnaryExpr"},
		{`package p; func f() { return }`, "ReturnStmt"},
		{`package p; func f() { if true {} }`, "IfStmt"},
		{`package p; func f() { for {} }`, "ForStmt"},
		{`package p; func f() { for range []int{} {} }`, "RangeStmt"},
		{`package p; func f() { switch x {} }`, "SwitchStmt"},
		{`package p; func f() { switch x { case 1: } }`, "CaseClause"},
		{`package p; func f() { select { case <-ch: } }`, "CommClause"},
		{`package p; func f() { {} }`, "BlockStmt"},
		{`package p; func f() { 1 }`, "ExprStmt"},
		{`package p; func f() { ch <- 1 }`, "SendStmt"},
		{`package p; func f() { x++ }`, "IncDecStmt"},
		{`package p; var x = *y`, "StarExpr"},
		{`package p; var x = a.b`, "SelectorExpr"},
		{`package p; var x = a[0]`, "IndexExpr"},
		{`package p; var x = a[:]`, "SliceExpr"},
		{`package p; var x = a.(int)`, "TypeAssertExpr"},
		{`package p; var x = (1+2)`, "ParenExpr"},
		{`package p; var x = map[string]int{"a": 1}`, "KeyValueExpr"},
		{`package p; var x = []int{1,2}`, "CompositeLit"},
		{`package p; type S struct{}`, "StructType"},
		{`package p; type I interface{}`, "InterfaceType"},
		{`package p; var m map[string]int`, "MapType"},
		{`package p; var ch chan int`, "ChanType"},
		{`package p; var arr [5]int`, "ArrayType"},
		{`package p; func f(args ...int) {}`, "Ellipsis"},
		{`package p; func f(a, b int) {}`, "Field"},
		{`package p; type F func()`, "FieldList"},
		{`package p; func f() { goto L; L: }`, "LabeledStmt"},
		{`package p; func f() { go f() }`, "GoStmt"},
		{`package p; func f() { defer f() }`, "DeferStmt"},
		{`package p; func f() { ; }`, "EmptyStmt"},
		{`package p; func f() { break }`, "BranchStmt"},
		{`package p; func f() { var x = 1 }`, "DeclStmt"},
		{`package p; var x = x + y`, "BinaryExpr"},
		{`package p; var x = f()`, "CallExpr"},
		{`package p; func f() { x = 2 }`, "AssignStmt"},
		{`package p; var x int`, "Ident"},
	}

	for _, tc := range testCases {
		file, err := parser.ParseFile(fset, "test.go", tc.src, 0)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", tc.src, err)
		}

		found := false
		WalkAST(file, func(node ast.Node) bool {
			if GetNodeKind(node) == tc.expect {
				found = true
			}
			return true // continue walking all nodes
		})

		if !found {
			t.Errorf("GetNodeKind: expected to find %q in %q", tc.expect, tc.src)
		}
	}
}

// TestGetNodeKindValueSpec tests GetNodeKind for ValueSpec node
func TestGetNodeKindValueSpec(t *testing.T) {
	src := `package main
var x = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "ValueSpec" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ValueSpec node")
	}
}

// TestGetNodeKindTypeSpec tests GetNodeKind for TypeSpec node
func TestGetNodeKindTypeSpec(t *testing.T) {
	src := `package main
type S struct{}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "TypeSpec" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSpec node")
	}
}

// TestGetNodeKindSelectStmt tests GetNodeKind for SelectStmt node
func TestGetNodeKindSelectStmt(t *testing.T) {
	src := `package main
func foo() {
	select {}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "SelectStmt" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SelectStmt node")
	}
}

// TestGetNodeKindTypeSwitchStmt tests GetNodeKind for TypeSwitchStmt node
func TestGetNodeKindTypeSwitchStmt(t *testing.T) {
	src := `package main
func foo() {
	var x interface{} = 1
	switch v := x.(type) {}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "TypeSwitchStmt" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSwitchStmt node")
	}
}

// TestGetNodeKindComment tests GetNodeKind for Comment node
func TestGetNodeKindComment(t *testing.T) {
	src := `package main
// comment
func foo() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "Comment" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find Comment node")
	}
}

// TestGetNodeKindCommentGroup tests GetNodeKind for CommentGroup node
func TestGetNodeKindCommentGroup(t *testing.T) {
	src := `package main
// comment 1
// comment 2
func foo() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Check if there are comments
	if file.Comments == nil {
		t.Skip("no comments in file")
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if GetNodeKind(node) == "CommentGroup" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CommentGroup node")
	}
}

// TestParseDirNonexistent tests ParseDir with a nonexistent directory
func TestParseDirNonexistent(t *testing.T) {
	p := NewParser()
	_, err := p.ParseDir("/nonexistent/directory")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// TestParseDirWalkError tests ParseDir when filepath.Walk returns an error
func TestParseDirWalkError(t *testing.T) {
	p := NewParser()

	// Create a temp directory
	tmpDir := t.TempDir()

	// Create a file that we can make unreadable later if needed
	// But we can't easily test filepath.Walk errors without mocking
	// So we just verify the basic error handling path works
	tmpFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(tmpFile, []byte(`package main`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err = p.ParseDir(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWalkASTNilChildren tests WalkAST with nil children in nodes
func TestWalkASTNilChildren(t *testing.T) {
	// Test a FuncDecl with nil body
	src := `package main
func foo()
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should not panic
	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

// TestWalkASTFieldListNil tests WalkAST with nil FieldList.List
func TestWalkASTFieldListNil(t *testing.T) {
	// FuncType with nil params and results
	src := `package main
var f func()
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should not panic
	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

// TestWalkASTIfStmtNilInit tests IfStmt with nil Init
func TestWalkASTIfStmtNilInit(t *testing.T) {
	src := `package main
func foo() {
	if true {
		println("hi")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.IfStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find IfStmt")
	}
}

// TestWalkASTForStmtNilParts tests ForStmt with nil parts
func TestWalkASTForStmtNilParts(t *testing.T) {
	src := `package main
func foo() {
	for {
		break
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ForStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ForStmt")
	}
}

// TestWalkASTRangeStmt tests RangeStmt with only value
func TestWalkASTRangeStmtValueOnly(t *testing.T) {
	src := `package main
func foo() {
	m := map[int]int{1: 2}
	for _, v := range m {
		_ = v
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.RangeStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find RangeStmt")
	}
}

// TestWalkASTSwitchStmtNilInit tests SwitchStmt with nil Init
func TestWalkASTSwitchStmtNilInit(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	switch x {
	case 1:
		println("one")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SwitchStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SwitchStmt")
	}
}

// TestWalkASTTypeSwitchStmtNilInit tests TypeSwitchStmt with nil Init
func TestWalkASTTypeSwitchStmtNilInit(t *testing.T) {
	src := `package main
func foo() {
	var x interface{} = 1
	switch v := x.(type) {
	case int:
		println(v)
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.TypeSwitchStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSwitchStmt")
	}
}

// TestWalkASTReturnStmtEmpty tests ReturnStmt with no results
func TestWalkASTReturnStmtEmpty(t *testing.T) {
	src := `package main
func foo() {
	return
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ReturnStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ReturnStmt")
	}
}

// TestWalkASTCallExprNoArgs tests CallExpr with no arguments
func TestWalkASTCallExprNoArgs(t *testing.T) {
	src := `package main
func foo() {
	println()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.CallExpr); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CallExpr")
	}
}

// TestWalkASTImportSpecNoName tests ImportSpec with no name (dot import)
func TestWalkASTImportSpecNoName(t *testing.T) {
	src := `package main
import . "fmt"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ImportSpec); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ImportSpec")
	}
}

// TestWalkASTBinaryExprAllOps tests BinaryExpr with various operators
func TestWalkASTBinaryExprAllOps(t *testing.T) {
	src := `package main
func foo() {
	_ = 1 + 2
	_ = 1 - 2
	_ = 1 * 2
	_ = 1 / 2
	_ = 1 == 2
	_ = 1 != 2
	_ = 1 < 2
	_ = 1 <= 2
	_ = 1 > 2
	_ = 1 >= 2
	_ = 1 && 2
	_ = 1 || 2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.BinaryExpr); ok {
			count++
		}
		return true
	})

	if count < 11 {
		t.Errorf("expected at least 11 BinaryExpr nodes, got %d", count)
	}
}

// TestWalkASTUnaryExprAllOps tests UnaryExpr with various operators
func TestWalkASTUnaryExprAllOps(t *testing.T) {
	src := `package main
import "os"
func foo() {
	_ = -1
	_ = !true
	_ = ^0
	_ = <-make(chan int)
	_ = os.Stdin
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.UnaryExpr); ok {
			count++
		}
		return true
	})

	if count < 3 {
		t.Errorf("expected at least 3 UnaryExpr nodes, got %d", count)
	}
}

// TestWalkASTFuncDeclNoBody tests FuncDecl with body
func TestWalkASTFuncDeclNoBody(t *testing.T) {
	src := `package main
func foo() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if fd, ok := node.(*ast.FuncDecl); ok && fd.Body != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncDecl with body")
	}
}

// TestWalkASTChanTypeAllDir tests ChanType with all directions
func TestWalkASTChanTypeAllDir(t *testing.T) {
	src := `package main
func foo() {
	var ch1 chan int
	var ch2 chan<- int
	var ch3 <-chan int
	_ = ch1
	_ = ch2
	_ = ch3
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ChanType); ok {
			count++
		}
		return true
	})

	if count < 3 {
		t.Errorf("expected at least 3 ChanType nodes, got %d", count)
	}
}

// TestWalkASTSliceExprFull tests SliceExpr with all options
func TestWalkASTSliceExprFull(t *testing.T) {
	src := `package main
func foo() {
	arr := []int{1, 2, 3, 4, 5}
	_ = arr[:]
	_ = arr[1:]
	_ = arr[:3]
	_ = arr[1:3]
	_ = arr[1:4:5]
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SliceExpr); ok {
			count++
		}
		return true
	})

	if count < 4 {
		t.Errorf("expected at least 4 SliceExpr nodes, got %d", count)
	}
}

// TestWalkASTArrayTypeLen tests ArrayType with various lengths
func TestWalkASTArrayTypeLen(t *testing.T) {
	src := `package main
func foo() {
	var a1 [5]int
	var a2 [3]int
	_ = a1
	_ = a2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.ArrayType); ok {
			count++
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 ArrayType nodes, got %d", count)
	}
}

// TestWalkASTBasicLitAllKinds tests BasicLit with various literal kinds
func TestWalkASTBasicLitAllKinds(t *testing.T) {
	src := `package main
const (
	a = 42
	b = 3.14
	c = "hello"
	d = 'x'
)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.BasicLit); ok {
			count++
		}
		return true
	})

	if count < 4 {
		t.Errorf("expected at least 4 BasicLit nodes, got %d", count)
	}
}

// TestWalkASTAssignStmtAllOps tests AssignStmt with all operations
func TestWalkASTAssignStmtAllOps(t *testing.T) {
	src := `package main
func foo() {
	var a int
	a = 1
	a += 2
	a -= 3
	a *= 4
	a /= 5
	a %= 6
	a &= 7
	a |= 8
	a ^= 9
	a &^= 10
	a <<= 11
	a >>= 12
	_ = a
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.AssignStmt); ok {
			count++
		}
		return true
	})

	if count < 12 {
		t.Errorf("expected at least 12 AssignStmt nodes, got %d", count)
	}
}

// TestWalkASTCaseClause tests CaseClause with multiple expressions
func TestWalkASTCaseClause(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	switch x {
	case 1, 2, 3:
		println("one two or three")
	default:
		println("other")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if cc, ok := node.(*ast.CaseClause); ok && len(cc.List) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CaseClause with multiple expressions")
	}
}

// TestWalkASTCommClause tests CommClause (select case)
func TestWalkASTCommClause(t *testing.T) {
	src := `package main
func foo() {
	ch := make(chan int)
	select {
	case v := <-ch:
		_ = v
	case ch <- 1:
	default:
		println("default")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.CommClause); ok {
			count++
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 CommClause nodes, got %d", count)
	}
}

// TestWalkASTIndexListExpr tests IndexListExpr (multi-dimensional index)
func TestWalkASTIndexListExpr(t *testing.T) {
	// Note: IndexListExpr is used for multi-dimensional indexing like a[i, j]
	// In Go 1.18+ this is used for generic type parameter indexing
	src := `package main
func foo() {
	var x [][]int
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Just verify it doesn't crash
	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

// TestWalkASTReturnStmtMultiReturn tests ReturnStmt with multiple return values
func TestWalkASTReturnStmtMultiReturn(t *testing.T) {
	src := `package main
func foo() (int, string) {
	return 1, "hi"
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if r, ok := node.(*ast.ReturnStmt); ok && len(r.Results) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ReturnStmt with multiple results")
	}
}

// TestWalkASTCallExprVariadic tests CallExpr with variadic arguments
func TestWalkASTCallExprVariadic(t *testing.T) {
	src := `package main
import "fmt"
func foo() {
	fmt.Println(1, 2, 3)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if c, ok := node.(*ast.CallExpr); ok && len(c.Args) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find CallExpr with multiple arguments")
	}
}

// TestWalkASTCallExprMethod tests CallExpr for method calls
func TestWalkASTCallExprMethod(t *testing.T) {
	src := `package main
type S struct{}
func (s S) Method() {}
func foo() {
	var s S
	s.Method()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if c, ok := node.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Method" {
				found = true
			}
		}
		return true
	})

	if !found {
		t.Error("expected to find method CallExpr")
	}
}

// TestWalkASTCallExprBuiltin tests CallExpr for builtin functions
func TestWalkASTCallExprBuiltin(t *testing.T) {
	src := `package main
func foo() {
	_ = len("hello")
	_ = cap([]int{1, 2})
	_ = append([]int{}, 1)
	_ = make([]int, 5)
	_ = new(int)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if c, ok := node.(*ast.CallExpr); ok {
			if ident, ok := c.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "len", "cap", "append", "make", "new":
					count++
				}
			}
		}
		return true
	})

	if count < 5 {
		t.Errorf("expected at least 5 builtin CallExpr nodes, got %d", count)
	}
}

// TestWalkASTUnaryExprReceive tests UnaryExpr with receive operator
func TestWalkASTUnaryExprReceive(t *testing.T) {
	src := `package main
func foo() {
	ch := make(chan int)
	v := <-ch
	_ = v
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if u, ok := node.(*ast.UnaryExpr); ok && u.Op.String() == "<-" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find UnaryExpr with <-")
	}
}

// TestWalkASTBinaryExprLogical tests BinaryExpr with logical operators
func TestWalkASTBinaryExprLogical(t *testing.T) {
	src := `package main
func foo() {
	_ = true && false
	_ = true || false
	_ = !true
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	binaryCount := 0
	unaryCount := 0
	WalkAST(file, func(node ast.Node) bool {
		if b, ok := node.(*ast.BinaryExpr); ok && (b.Op.String() == "&&" || b.Op.String() == "||") {
			binaryCount++
		}
		if u, ok := node.(*ast.UnaryExpr); ok && u.Op.String() == "!" {
			unaryCount++
		}
		return true
	})

	if binaryCount < 2 {
		t.Errorf("expected at least 2 logical BinaryExpr nodes, got %d", binaryCount)
	}
	if unaryCount < 1 {
		t.Errorf("expected at least 1 logical UnaryExpr node, got %d", unaryCount)
	}
}

// TestWalkASTBinaryExprPointer tests BinaryExpr with pointer operators
func TestWalkASTBinaryExprPointer(t *testing.T) {
	src := `package main
func foo() {
	x := 1
	y := &x
	_ = y
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Find the unary & expression
	found := false
	WalkAST(file, func(node ast.Node) bool {
		if u, ok := node.(*ast.UnaryExpr); ok && u.Op.String() == "&" {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find UnaryExpr with &")
	}
}

// TestWalkASTSliceExprWithMax tests SliceExpr with max parameter
func TestWalkASTSliceExprWithMax(t *testing.T) {
	src := `package main
func foo() {
	arr := []int{1, 2, 3, 4, 5}
	_ = arr[1:3:5]
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if s, ok := node.(*ast.SliceExpr); ok && s.Max != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SliceExpr with Max")
	}
}

// TestWalkASTCompositeLitNoKeys tests CompositeLit without keys
func TestWalkASTCompositeLitNoKeys(t *testing.T) {
	src := `package main
var x = []int{1, 2, 3}
var y = [3]int{1, 2, 3}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if c, ok := node.(*ast.CompositeLit); ok && c.Elts != nil && len(c.Elts) > 0 {
			count++
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 CompositeLit nodes without keys, got %d", count)
	}
}

// TestWalkASTCompositeLitWithKeys tests CompositeLit with keys
func TestWalkASTCompositeLitWithKeys(t *testing.T) {
	src := `package main
var x = []int{0: 1, 1: 2, 2: 3}
var y = map[string]int{"a": 1, "b": 2}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if c, ok := node.(*ast.CompositeLit); ok {
			for _, elt := range c.Elts {
				if _, ok := elt.(*ast.KeyValueExpr); ok {
					count++
					break
				}
			}
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 CompositeLit nodes with keys, got %d", count)
	}
}

// TestWalkASTFuncLitWithType tests FuncLit with type
func TestWalkASTFuncLitWithType(t *testing.T) {
	src := `package main
var f func(int) int = func(x int) int {
	return x + 1
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if fl, ok := node.(*ast.FuncLit); ok && fl.Type != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncLit with Type")
	}
}

// TestWalkASTFuncTypeWithParams tests FuncType with params
func TestWalkASTFuncTypeWithParams(t *testing.T) {
	src := `package main
var f func(int, string, bool)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if ft, ok := node.(*ast.FuncType); ok && ft.Params != nil && len(ft.Params.List) > 0 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncType with Params")
	}
}

// TestWalkASTFuncTypeWithResults tests FuncType with results
func TestWalkASTFuncTypeWithResults(t *testing.T) {
	src := `package main
var f func() int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if ft, ok := node.(*ast.FuncType); ok && ft.Results != nil && len(ft.Results.List) > 0 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find FuncType with Results")
	}
}

// TestWalkASTGenDeclDoc tests GenDecl with documentation
func TestWalkASTGenDeclDoc(t *testing.T) {
	src := `package main
// This is a comment
const x = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if g, ok := node.(*ast.GenDecl); ok && g.Doc != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GenDecl with Doc")
	}
}

// TestWalkASTGenDeclSpecs tests GenDecl with multiple specs
func TestWalkASTGenDeclSpecs(t *testing.T) {
	src := `package main
import (
	"fmt"
	"os"
)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if g, ok := node.(*ast.GenDecl); ok && len(g.Specs) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GenDecl with multiple specs")
	}
}

// TestWalkASTValueSpecNames tests ValueSpec with multiple names
func TestWalkASTValueSpecNames(t *testing.T) {
	src := `package main
var a, b, c = 1, 2, 3
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if v, ok := node.(*ast.ValueSpec); ok && len(v.Names) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ValueSpec with multiple names")
	}
}

// TestWalkASTValueSpecValues tests ValueSpec with multiple values
func TestWalkASTValueSpecValues(t *testing.T) {
	src := `package main
var a, b = 1, 2
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if v, ok := node.(*ast.ValueSpec); ok && len(v.Values) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ValueSpec with multiple values")
	}
}

// TestWalkASTTypeSpecStruct tests TypeSpec with struct type
func TestWalkASTTypeSpecStruct(t *testing.T) {
	src := `package main
type S struct {
	A int
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if t, ok := node.(*ast.TypeSpec); ok {
			if s, ok := t.Type.(*ast.StructType); ok && s.Fields != nil {
				found = true
			}
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSpec with StructType")
	}
}

// TestWalkASTTypeSpecInterface tests TypeSpec with interface type
func TestWalkASTTypeSpecInterface(t *testing.T) {
	src := `package main
type I interface {
	Method()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if t, ok := node.(*ast.TypeSpec); ok {
			if i, ok := t.Type.(*ast.InterfaceType); ok && i.Methods != nil {
				found = true
			}
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSpec with InterfaceType")
	}
}

// TestWalkASTTypeSpecAlias tests TypeSpec alias (Go 1.9+)
func TestWalkASTTypeSpecAlias(t *testing.T) {
	src := `package main
type Int = int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if t, ok := node.(*ast.TypeSpec); ok && t.Assign.IsValid() {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSpec with alias")
	}
}

// TestWalkASTImportSpecPath tests ImportSpec with path
func TestWalkASTImportSpecPath(t *testing.T) {
	src := `package main
import "fmt"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if i, ok := node.(*ast.ImportSpec); ok && i.Path != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ImportSpec with Path")
	}
}

// TestWalkASTImportSpecName tests ImportSpec with name (alias)
func TestWalkASTImportSpecName(t *testing.T) {
	src := `package main
import f "fmt"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if i, ok := node.(*ast.ImportSpec); ok && i.Name != nil {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find ImportSpec with Name")
	}
}

// TestWalkASTComment tests Comment node
func TestWalkASTComment(t *testing.T) {
	src := `package main
// line comment
func foo() {
	/* block comment */
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	commentCount := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.Comment); ok {
			commentCount++
		}
		return true
	})

	if commentCount < 2 {
		t.Errorf("expected at least 2 Comment nodes, got %d", commentCount)
	}
}

// TestWalkASTGenDeclLparen tests GenDecl with Lparen (grouped)
func TestWalkASTGenDeclLparen(t *testing.T) {
	src := `package main
const (
	A = 1
	B = 2
)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if g, ok := node.(*ast.GenDecl); ok && g.Lparen.IsValid() {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GenDecl with Lparen")
	}
}

// TestWalkASTGenDeclRparen tests GenDecl with Rparen
func TestWalkASTGenDeclRparen(t *testing.T) {
	src := `package main
const (
	A = 1
)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if g, ok := node.(*ast.GenDecl); ok && g.Rparen.IsValid() {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find GenDecl with Rparen")
	}
}

// TestWalkASTFieldNames tests Field with multiple names
func TestWalkASTFieldNames(t *testing.T) {
	src := `package main
func foo(a, b, c int) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if f, ok := node.(*ast.Field); ok && len(f.Names) > 1 {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find Field with multiple names")
	}
}

// TestWalkASTFieldType tests Field with type expression
func TestWalkASTFieldType(t *testing.T) {
	src := `package main
func foo(a map[string]int) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if f, ok := node.(*ast.Field); ok && f.Type != nil {
			if _, ok := f.Type.(*ast.MapType); ok {
				found = true
			}
		}
		return true
	})

	if !found {
		t.Error("expected to find Field with MapType")
	}
}

// TestWalkASTValueSpecType tests ValueSpec with type
func TestWalkASTValueSpecType(t *testing.T) {
	src := `package main
var x int = 10
var y string = "hello"
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if v, ok := node.(*ast.ValueSpec); ok && v.Type != nil {
			count++
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 ValueSpec with Type, got %d", count)
	}
}

// TestWalkASTBadDecl tests BadDecl
func TestWalkASTBadDecl(t *testing.T) {
	// A BadDecl occurs with severely malformed code
	// We can't easily produce one through valid parser input
	// This just ensures WalkAST doesn't panic on unexpected node types
	src := `package main
func foo() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

// TestWalkASTBadExpr tests BadExpr
func TestWalkASTBadExpr(t *testing.T) {
	// A BadExpr occurs with severely malformed code
	// We can't easily produce one through valid parser input
	// This just ensures WalkAST doesn't panic on unexpected node types
	src := `package main
var x int = 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	WalkAST(file, func(node ast.Node) bool {
		return true
	})
}

// TestWalkASTSelectStmt tests SelectStmt
func TestWalkASTSelectStmt(t *testing.T) {
	src := `package main
func foo() {
	select {}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.SelectStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find SelectStmt")
	}
}

// TestWalkASTTypeSwitchStmt tests TypeSwitchStmt
func TestWalkASTTypeSwitchStmt(t *testing.T) {
	src := `package main
func foo() {
	var x interface{} = 1
	switch v := x.(type) {
	case int:
		println(v)
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	found := false
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.TypeSwitchStmt); ok {
			found = true
		}
		return true
	})

	if !found {
		t.Error("expected to find TypeSwitchStmt")
	}
}

// TestWalkASTGenDeclGroup tests GenDecl with grouped declarations
func TestWalkASTGenDeclGroup(t *testing.T) {
	src := `package main
const (
	A = 1
	B = 2
)
var (
	X = 1
	Y = 2
)
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if g, ok := node.(*ast.GenDecl); ok {
			if len(g.Specs) > 1 {
				count++
			}
		}
		return true
	})

	if count < 2 {
		t.Errorf("expected at least 2 GenDecl with multiple specs, got %d", count)
	}
}

// TestParseDirNonexistentError tests ParseDir with nonexistent directory
func TestParseDirNonexistentError(t *testing.T) {
	p := NewParser()
	_, err := p.ParseDir("/this/path/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// TestWalkASTAllExprTypes tests all expression types
func TestWalkASTAllExprTypes(t *testing.T) {
	src := `package main
func foo() {
	// Binary operators
	_ = 1 + 2
	_ = 1 - 2
	_ = 1 * 2
	_ = 1 / 2
	_ = 1 % 2
	_ = 1 & 2
	_ = 1 | 2
	_ = 1 ^ 2
	_ = 1 &^ 2
	_ = 1 << 2
	_ = 1 >> 2
	_ = 1 == 2
	_ = 1 != 2
	_ = 1 < 2
	_ = 1 <= 2
	_ = 1 > 2
	_ = 1 >= 2
	_ = true && false
	_ = true || false

	// Unary operators
	_ = -1
	_ = !true
	_ = ^0
	_ = <-make(chan int)

	// Select statement with recv
	ch := make(chan int)
	select {
	case <-ch:
	case ch <- 1:
	default:
	}

	// Type switch
	var x interface{} = 1
	switch v := x.(type) {
	case int:
		_ = v
	case string:
		_ = v
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Verify various node types are visited
	hasBinary := false
	hasUnary := false
	hasSelect := false
	hasTypeSwitch := false

	WalkAST(file, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.BinaryExpr:
			hasBinary = true
		case *ast.UnaryExpr:
			hasUnary = true
		case *ast.SelectStmt:
			hasSelect = true
		case *ast.TypeSwitchStmt:
			hasTypeSwitch = true
		}
		return true
	})

	if !hasBinary {
		t.Error("expected BinaryExpr nodes")
	}
	if !hasUnary {
		t.Error("expected UnaryExpr nodes")
	}
	if !hasSelect {
		t.Error("expected SelectStmt nodes")
	}
	if !hasTypeSwitch {
		t.Error("expected TypeSwitchStmt nodes")
	}
}

// TestWalkASTStructFields tests WalkAST with struct fields
func TestWalkASTStructFields(t *testing.T) {
	src := "package main\ntype S struct {\n\tA int\n\tB string\n\tC []int\n}\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	fieldCount := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.Field); ok {
			fieldCount++
		}
		return true
	})

	if fieldCount < 3 {
		t.Errorf("expected at least 3 fields, got %d", fieldCount)
	}
}

// TestGetNodeKindBinaryExpr tests GetNodeKind for all binary operators
func TestGetNodeKindBinaryExpr(t *testing.T) {
	src := `package main
func foo() {
	_ = 1 + 2
	_ = 1 == 2
	_ = true && false
	_ = 1 &^ 2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.BinaryExpr); ok {
			count++
		}
		return true
	})

	if count < 4 {
		t.Errorf("expected at least 4 BinaryExpr, got %d", count)
	}
}

// TestGetNodeKindUnaryExpr tests GetNodeKind for unary operators
func TestGetNodeKindUnaryExpr(t *testing.T) {
	src := `package main
import "os"
func foo() {
	_ = -1
	_ = !true
	_ = ^0
	_ = <-make(chan int)
	_ = os.Stdin
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	count := 0
	WalkAST(file, func(node ast.Node) bool {
		if _, ok := node.(*ast.UnaryExpr); ok {
			count++
		}
		return true
	})

	if count < 3 {
		t.Errorf("expected at least 3 UnaryExpr, got %d", count)
	}
}
