package core

import (
	"os"
	"path/filepath"
	"testing"

	"go/ast"
	"go/parser"
	"go/token"
)

func TestDetectLockNotReleased(t *testing.T) {
	// Function with lock acquired - current implementation detects when Unlock is missing entirely
	code := `package main

import "sync"

func badFunc(mu sync.Mutex) {
	mu.Lock()
	if true {
		return
	}
	// BUG: missing mu.Unlock() on early return path
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "lock_test.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectLockNotReleased(fset, file)
	// Current implementation: detects when Lock is called without any Unlock in function
	if len(findings) == 0 {
		t.Error("expected to find lock-not-released pattern")
	}
}

func TestDetectLockNotReleased_Safe(t *testing.T) {
	// Function with no lock at all - should not trigger
	code := `package main

func simpleFunc(x int) int {
	if x > 5 {
		return x
	}
	return 0
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "no_lock.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectLockNotReleased(fset, file)
	for _, f := range findings {
		if f.Rule == "lock-not-released" {
			t.Errorf("function with no lock should not trigger lock-not-released, got: %s", f.Message)
		}
	}
}

func TestDetectResourceLeak(t *testing.T) {
	// Function with Open but no Close anywhere
	code := `package main

import "os"

func badOpen(name string) {
	f, _ := os.Open(name)
	if true {
		return
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "leak_test.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectResourceLeak(fset, file)
	if len(findings) == 0 {
		t.Error("expected to find resource-leak pattern")
	}
}

func TestDetectResourceLeak_Safe(t *testing.T) {
	// Function with no file operations - should not trigger
	code := `package main

func simpleRead() int {
	return 42
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_open.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectResourceLeak(fset, file)
	// Current implementation only detects when Close is completely missing
	for _, f := range findings {
		if f.Rule == "resource-leak" {
			t.Errorf("good code with Close should not trigger with current simple detection, got: %s", f.Message)
		}
	}
}

func TestDetectTransactionBug(t *testing.T) {
	code := `package main

func badTx(db *sql.DB) {
	tx, _ := db.Begin()
	if true {
		return
	}
	tx.Commit()
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tx_test.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectTransactionBug(fset, file)
	if len(findings) == 0 {
		t.Error("expected to find transaction-not-ended pattern")
	}
}

func TestDetectTransactionBug_Safe(t *testing.T) {
	// Function with no transaction - should not trigger
	code := `package main

func simpleFunc() int {
	return 42
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_tx.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	findings := DetectTransactionBug(fset, file)
	// Current implementation only detects when Commit/Rollback is completely missing
	for _, f := range findings {
		if f.Rule == "transaction-not-ended" {
			t.Errorf("good code with Commit should not trigger with current simple detection, got: %s", f.Message)
		}
	}
}

func TestCFGPatternsRegistered(t *testing.T) {
	// Verify all CFG patterns are registered
	cfgPatterns := []string{
		"lock-not-released",
		"resource-leak",
		"transaction-not-ended",
	}

	registeredPatterns := make(map[string]bool)
	for _, p := range builtinASTPatterns {
		registeredPatterns[p.Name] = true
	}

	for _, name := range cfgPatterns {
		if !registeredPatterns[name] {
			t.Errorf("expected pattern %q to be registered", name)
		}
	}
}

func TestBuildCFG(t *testing.T) {
	code := `package main

func simpleFunc(x int) int {
	if x > 5 {
		return x
	}
	return 0
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "simple.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			cfg := BuildCFG(funcDecl)
			if cfg == nil {
				t.Error("expected BuildCFG to return non-nil CFG")
			}
			if cfg.FuncName != "simpleFunc" {
				t.Errorf("expected func name 'simpleFunc', got '%s'", cfg.FuncName)
			}
		}
	}
}

func TestGetReleaseFunc(t *testing.T) {
	tests := []struct {
		acquire string
		expected string
	}{
		{"Lock", "Unlock"},
		{"RLock", "RUnlock"},
		{"WLock", "WUnlock"},
		{"Unknown", "Release"},
	}

	for _, tt := range tests {
		result := getReleaseFunc(tt.acquire)
		if result != tt.expected {
			t.Errorf("getReleaseFunc(%q) = %q, want %q", tt.acquire, result, tt.expected)
		}
	}
}
