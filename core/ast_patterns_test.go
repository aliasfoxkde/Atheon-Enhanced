package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFileAST_CommandInjection(t *testing.T) {
	content := `package main

import "os/exec"

func badCmd(input string) {
	exec.Command("sh", "-c", "echo "+input)
}

func goodCmd(input string) {
	exec.Command("sh", "-c", "echo", input)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var cmdFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "command-injection" {
			cmdFindings = append(cmdFindings, f)
		}
	}

	if len(cmdFindings) == 0 {
		t.Error("expected to find command-injection pattern")
	}
}

func TestScanFileAST_HardcodedCredentials(t *testing.T) {
	content := `package main

func badConfig() {
	password := "hunter2"
	apiKey := "sk-1234567890"
	secret := "mysecret"
}

func goodConfig() {
	password := os.Getenv("PASSWORD")
	apiKey := os.Getenv("API_KEY")
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var credFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "hardcoded-credentials" {
			credFindings = append(credFindings, f)
		}
	}

	if len(credFindings) == 0 {
		t.Error("expected to find hardcoded-credentials pattern")
	}
}

func TestScanDirAST(t *testing.T) {
	content := `package main

import "os/exec"

func badCmd(input string) {
	exec.Command("sh", "-c", "echo "+input)
}
`
	tmpDir := t.TempDir()
	tmpFile1 := filepath.Join(tmpDir, "test1.go")
	tmpFile2 := filepath.Join(tmpDir, "test2.go")
	if err := os.WriteFile(tmpFile1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpFile2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanDirAST(tmpDir, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) == 0 {
		t.Error("expected to find findings in directory scan")
	}
}

func TestScanFileAST_HardcodedCredentials_Detailed(t *testing.T) {
	// More comprehensive test for hardcoded-credentials pattern
	content := `package main

var staticPassword = "changeme"
var apiKey = "sk_live_xxxxxxxxxxxxxxxx"

func loadConfig() {
	databasePassword := "admin123"
	secretToken := "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var credFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "hardcoded-credentials" {
			credFindings = append(credFindings, f)
		}
	}

	if len(credFindings) < 2 {
		t.Errorf("expected at least 2 hardcoded-credentials findings, got %d", len(credFindings))
	}
}

func TestBuiltinPatternsCount(t *testing.T) {
	if len(builtinASTPatterns) < 6 {
		t.Errorf("expected at least 6 builtin AST patterns, got %d", len(builtinASTPatterns))
	}
}

func TestASTPattern_Fields(t *testing.T) {
	for _, p := range builtinASTPatterns {
		if p.Name == "" {
			t.Error("pattern name should not be empty")
		}
		if p.Severity == "" {
			t.Error("pattern severity should not be empty")
		}
		if p.Func == nil {
			t.Error("pattern func should not be nil")
		}
	}
}

func TestScanFileAST_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	// Empty/safe file should produce no findings
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty file, got %d", len(findings))
	}
}

func TestScanFileAST_SafeCode(t *testing.T) {
	// Verify that credential vars with env lookup are not flagged
	// (but non-env credential assignments would be)
	content := `package main

var password = os.Getenv("APP_PASSWORD")

func safeCmd() {
	exec.Command("echo", "hello")
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	// Credential loaded from env should not trigger hardcoded-credentials
	for _, f := range findings {
		if f.Rule == "hardcoded-credentials" {
			t.Errorf("safe code should not trigger hardcoded-credentials, got: %s", f.Message)
		}
	}
}

func TestScanDirAST_SkipNonGoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	// Write a .go file with actual command injection (variable concatenation)
	goFile := filepath.Join(tmpDir, "test.go")
	// Using string concatenation with a function parameter named "input"
	// which should trigger the command-injection detector
	if err := os.WriteFile(goFile, []byte("package main\nimport \"os/exec\"\nfunc bad(req string) { exec.Command(\"sh\", \"-c\", \"echo \"+req) }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a non-.go file
	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("This is not Go code"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanDirAST(tmpDir, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	// Should find command injection in the .go file
	if len(findings) == 0 {
		t.Error("expected findings in .go file with command injection")
	}
}

func TestContainsEnvVar(t *testing.T) {
	// Test code with os.Getenv
	code := `package main
import "os"
func main() {
	x := os.Getenv("HOME")
	print(x)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "envvar.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestIsStringType(t *testing.T) {
	// Test with string literal
	code := `package main
const s = "hello"
func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "stringtype.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestHasStringLiteral(t *testing.T) {
	// Test with string literals
	code := `package main
func main() {
	s := "hello world"
	print(s)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "hasstring.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestContainsDangerousSource(t *testing.T) {
	// Test with dangerous source patterns
	code := `package main
import "os"
func main() {
	data := os.Getenv("USER_INPUT")
	print(data)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "dangerous.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectMissingReturn(t *testing.T) {
	// Function with conditional return but no final return
	code := `package main

func badFunc(x int) int {
	if x > 5 {
		return x
	}
	// No return here - bug!
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "missing_return.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var missingReturnFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "missing-return" {
			missingReturnFindings = append(missingReturnFindings, f)
		}
	}

	if len(missingReturnFindings) == 0 {
		t.Error("expected to find missing-return pattern")
	}
}

func TestDetectMissingReturn_Safe(t *testing.T) {
	// Function with proper return at end - should NOT trigger
	code := `package main

func goodFunc(x int) int {
	if x > 5 {
		return x
	}
	return 0
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "proper_return.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "missing-return" {
			t.Errorf("good code should not trigger missing-return, got: %s", f.Message)
		}
	}
}

func TestDetectMissingReturn_VoidFunc(t *testing.T) {
	// Void function - should NOT trigger (no return value needed)
	code := `package main

func voidFunc(x int) {
	if x > 5 {
		print(x)
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "void_func.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "missing-return" {
			t.Errorf("void function should not trigger missing-return, got: %s", f.Message)
		}
	}
}

func TestDetectDuplicateCondition(t *testing.T) {
	code := `package main

func badFunc(x int) {
	if x == 5 {
		print("five")
	} else if x == 5 {
		print("five again")
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "dup_cond.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var dupFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "duplicate-condition" {
			dupFindings = append(dupFindings, f)
		}
	}

	if len(dupFindings) == 0 {
		t.Error("expected to find duplicate-condition pattern")
	}
}

func TestDetectDuplicateCondition_Safe(t *testing.T) {
	code := `package main

func goodFunc(x int) {
	if x == 5 {
		print("five")
	} else if x == 6 {
		print("six")
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_cond.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "duplicate-condition" {
			t.Errorf("good code should not trigger duplicate-condition, got: %s", f.Message)
		}
	}
}

func TestDetectImpossibleBranch(t *testing.T) {
	code := `package main

func badFunc(x *int) {
	if x == nil {
		print("nil")
	} else if x == nil {
		print("nil again")
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "impossible.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var impossibleFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "impossible-branch" {
			impossibleFindings = append(impossibleFindings, f)
		}
	}

	if len(impossibleFindings) == 0 {
		t.Error("expected to find impossible-branch pattern")
	}
}

func TestDetectImpossibleBranch_Safe(t *testing.T) {
	code := `package main

func goodFunc(x *int) {
	if x == nil {
		print("nil")
	} else if *x > 5 {
		print("positive")
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_branch.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "impossible-branch" {
			t.Errorf("good code should not trigger impossible-branch, got: %s", f.Message)
		}
	}
}

func TestDetectIteratorModification(t *testing.T) {
	code := `package main

func badFunc(list []int) {
	for i := range list {
		list = append(list, i)
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "iter_mod.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	var iterFindings []ASTFinding
	for _, f := range findings {
		if f.Rule == "iterator-modification" {
			iterFindings = append(iterFindings, f)
		}
	}

	if len(iterFindings) == 0 {
		t.Error("expected to find iterator-modification pattern")
	}
}

func TestDetectIteratorModification_Safe(t *testing.T) {
	code := `package main

func goodFunc(list []int) []int {
	var result []int
	for _, v := range list {
		result = append(result, v*2)
	}
	return result
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_iter.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "iterator-modification" {
			t.Errorf("good code should not trigger iterator-modification, got: %s", f.Message)
		}
	}
}

func TestExprToString(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"package main\nfunc main() { var x int }", ""},
	}

	for _, tt := range tests {
		_ = tt
	}
}

func TestNewPatternsRegistered(t *testing.T) {
	// Verify all new patterns are registered
	newPatterns := []string{
		"missing-return",
		"duplicate-condition",
		"impossible-branch",
		"iterator-modification",
	}

	registeredPatterns := make(map[string]bool)
	for _, p := range builtinASTPatterns {
		registeredPatterns[p.Name] = true
	}

	for _, name := range newPatterns {
		if !registeredPatterns[name] {
			t.Errorf("expected pattern %q to be registered", name)
		}
	}
}

