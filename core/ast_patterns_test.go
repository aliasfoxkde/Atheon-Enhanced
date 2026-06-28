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
