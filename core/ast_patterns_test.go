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
