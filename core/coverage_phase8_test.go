package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// Phase 8 coverage tests - targeting remaining low coverage functions

func TestRiskScore_AddFinding_Severity_Phase8(t *testing.T) {
	// Test AddFinding with different severity levels
	findings := []Finding{
		{Pattern: "test-rule", Severity: "critical", Category: "security"},
		{Pattern: "test-rule-2", Severity: "high", Category: "security"},
		{Pattern: "test-rule-3", Severity: "medium", Category: "security"},
		{Pattern: "test-rule-4", Severity: "low", Category: "security"},
		{Pattern: "test-rule-5", Severity: "info", Category: "security"},
	}
	rs := NewRiskScore(findings)
	rs.AddFinding(Finding{Pattern: "new-critical", Severity: "critical", Category: "security"})
	rs.AddFinding(Finding{Pattern: "new-high", Severity: "high", Category: "security"})
	rs.AddFinding(Finding{Pattern: "new-medium", Severity: "medium", Category: "security"})
	rs.AddFinding(Finding{Pattern: "new-low", Severity: "low", Category: "security"})
	rs.AddFinding(Finding{Pattern: "new-info", Severity: "info", Category: "security"})
	_ = rs
}

func TestRiskScore_Summary_Phase8(t *testing.T) {
	findings := []Finding{
		{Pattern: "test", Severity: "high", Category: "security"},
	}
	rs := NewRiskScore(findings)
	summary := rs.Summary()
	_ = summary
}

func TestBaselineMatcher_Baseline_Phase8(t *testing.T) {
	// Test NewBaselineMatcher with various inputs
	tmpDir := t.TempDir()

	// Empty baseline
	baselineFile := filepath.Join(tmpDir, "baseline.yaml")
	os.WriteFile(baselineFile, []byte(`version: "1.0"
findings: []`), 0644)

	matcher, err := NewBaselineMatcher(baselineFile)
	if err != nil {
		t.Fatal(err)
	}
	_ = matcher

	// Baseline with entries
	baselineContent := `version: "1.0"
findings:
  - pattern_id: "test-rule"
    file: "test.go"
    line: 10
    suppression_type: "exact"
  - pattern_id: "other-rule"
    file: "other.go"
    line: 20
    suppression_type: "file"
`
	os.WriteFile(baselineFile, []byte(baselineContent), 0644)

	matcher2, err := NewBaselineMatcher(baselineFile)
	if err != nil {
		t.Fatal(err)
	}
	_ = matcher2
}

func TestIsSuppressed_Baseline_Phase8(t *testing.T) {
	// Test IsSuppressed with baseline
	tmpDir := t.TempDir()

	baselineFile := filepath.Join(tmpDir, "baseline.yaml")
	baselineContent := `version: "1.0"
findings:
  - pattern_id: "suppressed-rule"
    file: test.go
    line: 10
    suppression_type: exact
`
	os.WriteFile(baselineFile, []byte(baselineContent), 0644)

	matcher, err := NewBaselineMatcher(baselineFile)
	if err != nil {
		t.Fatal(err)
	}

	f1 := Finding{Pattern: "suppressed-rule", File: "test.go", Line: 10}
	if !matcher.IsSuppressed(f1) {
		t.Error("expected finding to be suppressed")
	}

	f2 := Finding{Pattern: "suppressed-rule", File: "test.go", Line: 20}
	if matcher.IsSuppressed(f2) {
		t.Error("expected finding NOT to be suppressed")
	}
}

func TestLoadDefaultBaseline_Phase8(t *testing.T) {
	// Test LoadDefaultBaseline
	// It tries common locations - this just exercises the function
	_, err := LoadDefaultBaseline()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateBaselineFile_Phase8(t *testing.T) {
	// Test CreateBaselineFile
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, "baseline.yaml")

	findings := []Finding{
		{Pattern: "rule1", File: "f1.go", Line: 10, Severity: "high"},
		{Pattern: "rule2", File: "f2.go", Line: 20, Severity: "medium"},
	}

	err := CreateBaselineFile(findings, baselineFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was created
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("baseline file should not be empty")
	}
}

func TestTaintTracker_AnalyzeCall_Phase8(t *testing.T) {
	// Test analyzeCall with various call patterns
	code := `package main

func safeFunc(input string) string { return input }
func dangerousFunc(input string) { eval(input) }

func test() {
	dangerousFunc("user input")
	safeFunc("safe input")
}

func main() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}

	tracker := NewTaintTracker()
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					_ = tracker.analyzeCall(call, fset)
				}
				return true
			})
		}
	}
}

func TestTaintTracker_IsTainted_Phase8(t *testing.T) {
	// Test IsTainted with various names
	tracker := NewTaintTracker()
	_ = tracker.IsTainted("unknown")
	_ = tracker.IsTainted("")
}

func TestTaintTracker_GetExprText_Phase8(t *testing.T) {
	// Test getExprText with various expressions
	code := `package main

func test() {
	x := 1
	y := "hello"
	z := x + 1
	_ = x; _ = y; _ = z
}

func main() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}

	tracker := NewTaintTracker()
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				if binExpr, ok := n.(*ast.BinaryExpr); ok {
					_ = tracker.getExprText(binExpr)
				}
				return true
			})
		}
	}
}

func TestCalculateRiskScore_Phase8(t *testing.T) {
	// Test CalculateRiskScore with various inputs
	score := CalculateRiskScore(10, 0.8)
	_ = score

	score = CalculateRiskScore(0, 0.5)
	_ = score

	score = CalculateRiskScore(5, 1.0)
	_ = score
}

func TestYARAScanFile_Phase8(t *testing.T) {
	// Test YARA ScanFile
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("sensitive data: password123"), 0644)

	scanner := NewYARAScanner("")
	findings, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestYARAScanFile_NoMatch_Phase8(t *testing.T) {
	// Test YARA ScanFile with no matches
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("just some regular text"), 0644)

	scanner := NewYARAScanner("")
	findings, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestYARAScanDir_Phase8(t *testing.T) {
	// Test YARA ScanDir
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)

	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("secret: abc123"), 0644)
	os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("password: def456"), 0644)

	scanner := NewYARAScanner("")
	findings, err := scanner.ScanDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestYARAScanDir_Empty_Phase8(t *testing.T) {
	// Test YARA ScanDir with empty directory
	tmpDir := t.TempDir()

	scanner := NewYARAScanner("")
	findings, err := scanner.ScanDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}
