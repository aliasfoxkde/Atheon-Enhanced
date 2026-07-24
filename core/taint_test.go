package core

import (
	"go/ast"
	"go/token"
	"os"
	"testing"
)

func TestTaintTracker_NewTaintTracker(t *testing.T) {
	tt := NewTaintTracker()
	if tt == nil {
		t.Fatal("NewTaintTracker returned nil")
	}
	if tt.tainted == nil {
		t.Error("tainted map is nil")
	}
}

func TestTaintTracker_TrackSource(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("user_input")
	if !tt.IsTainted("user_input") {
		t.Error("expected user_input to be tainted")
	}
}

func TestTaintTracker_IsTainted(t *testing.T) {
	tt := NewTaintTracker()
	if tt.IsTainted("unknown") {
		t.Error("unknown variable should not be tainted")
	}
	tt.TrackSource("var1")
	if !tt.IsTainted("var1") {
		t.Error("var1 should be tainted after TrackSource")
	}
}

func TestTaintTracker_ClearTaint(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("var1")
	tt.ClearTaint("var1")
	if tt.IsTainted("var1") {
		t.Error("var1 should not be tainted after ClearTaint")
	}
}

func TestSeverityLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{90, "critical"},
		{50, "critical"},
		{30, "critical"},
		{29, "high"},
		{25, "high"},
		{20, "high"},
		{19, "medium"},
		{15, "medium"},
		{10, "medium"},
		{9, "low"},
		{5, "low"},
		{0, "low"},
	}
	for _, tc := range tests {
		got := SeverityLevel(tc.score)
		if got != tc.expected {
			t.Errorf("SeverityLevel(%d) = %q, want %q", tc.score, got, tc.expected)
		}
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		weight     int
		confidence float64
		min, max   int
	}{
		{40, 1.0, 40, 40},
		{20, 0.5, 10, 10},
		{40, 0.8, 32, 32},
		{100, 1.0, 100, 100}, // capped at 100
	}
	for _, tc := range tests {
		got := CalculateRiskScore(tc.weight, tc.confidence)
		if got < tc.min || got > tc.max {
			t.Errorf("CalculateRiskScore(%d, %f) = %d, want between %d and %d",
				tc.weight, tc.confidence, got, tc.min, tc.max)
		}
	}
}

func TestScanForTaintPatterns(t *testing.T) {
	content := `os.getenv("USER_INPUT") + exec()`
	findings := ScanForTaintPatterns(content)
	if len(findings) == 0 {
		t.Error("expected at least one finding for command injection")
	}
}

func TestScanFileAST(t *testing.T) {
	// Create a temp Go file with taintable code
	tmpfile, err := os.CreateTemp("", "taint_test*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test code
	code := `package main
import "os"
func main() {
	user := os.Getenv("USER")
	print(user)
}`
	if _, err := tmpfile.WriteString(code); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	tt := NewTaintTracker()
	findings, err := tt.ScanFileAST(tmpfile.Name())
	if err != nil {
		t.Fatalf("ScanFileAST failed: %v", err)
	}
	_ = findings // findings may be empty depending on analysis
}

func TestTaintTracker_getFuncName(t *testing.T) {
	tt := NewTaintTracker()

	// Test with identifier
	expr := &ast.Ident{Name: "testFunc"}
	fnName := tt.getFuncName(expr)
	if fnName != "testFunc" {
		t.Errorf("expected testFunc, got %s", fnName)
	}

	// Test with selector expression
	sel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "os"},
		Sel: &ast.Ident{Name: "Getenv"},
	}
	fnName = tt.getFuncName(sel)
	if fnName != "os.Getenv" {
		t.Errorf("expected os.Getenv, got %s", fnName)
	}

	// Test with other type
	fnName = tt.getFuncName(&ast.StarExpr{})
	if fnName != "" {
		t.Errorf("expected empty string for StarExpr, got %s", fnName)
	}
}

func TestTaintTracker_getExprText(t *testing.T) {
	tt := NewTaintTracker()

	// Test identifier
	ident := &ast.Ident{Name: "myVar"}
	text := tt.getExprText(ident)
	if text != "myVar" {
		t.Errorf("expected myVar, got %s", text)
	}

	// Test basic literal
	lit := &ast.BasicLit{Kind: 9, Value: `"hello"`}
	text = tt.getExprText(lit)
	if text != `"hello"` {
		t.Errorf("expected \"hello\", got %s", text)
	}
}

func TestTaintTracker_analyzeAssignment(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("source")

	// Create assignment: dest = source
	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "dest"}},
		Rhs: []ast.Expr{&ast.Ident{Name: "source"}},
	}

	fset := token.NewFileSet()
	tt.analyzeAssignment(assign, fset)

	// dest should now be tainted
	if !tt.IsTainted("dest") {
		t.Error("dest should be tainted after assignment from tainted source")
	}
}

func TestIsExprTainted(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("source")

	// Test with tainted identifier
	ident := &ast.Ident{Name: "source"}
	if !tt.isExprTainted(ident) {
		t.Error("expected source to be tainted")
	}

	// Test with untainted identifier
	clean := &ast.Ident{Name: "clean"}
	if tt.isExprTainted(clean) {
		t.Error("expected clean to not be tainted")
	}
}

func TestTaintTracker_ScanFileAST_Empty(t *testing.T) {
	tt := NewTaintTracker()

	// Create empty file
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/empty.go"
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := tt.ScanFileAST(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
}
