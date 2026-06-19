package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCompileIgnoreFileMissing exercises the error branch when the file
// doesn't exist.
func TestCompileIgnoreFileMissing(t *testing.T) {
	_, err := compileIgnoreFile("/nonexistent/path/.atheonignore")
	if err == nil {
		t.Error("expected error for missing ignore file")
	}
}

// TestCompileIgnoreFileInvalidRegex exercises the invalid-regex branch.
// When a pattern can't be compiled, it's skipped (continue), so the test
// just verifies no panic and the function returns.
func TestCompileIgnoreFileInvalidRegex(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), ".atheonignore")
	content := `# comment
*.go
[unterminated
valid-pattern
`
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := compileIgnoreFile(tmp)
	if err != nil {
		t.Fatalf("compileIgnoreFile failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil matcher")
	}
	// Verify valid pattern matches
	if !m.matchesPath("test.go") {
		t.Error("expected *.go to match test.go")
	}
}

// TestCompileIgnoreFileSlashOnly exercises the slash-only pattern that
// triggers ignorePatternToRegexp's "empty pattern" error inside
// compileIgnoreFile. The pattern is skipped (continue) without panic.
func TestCompileIgnoreFileSlashOnly(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), ".atheonignore")
	content := `/
valid-pattern
`
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := compileIgnoreFile(tmp)
	if err != nil {
		t.Fatalf("compileIgnoreFile failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil matcher")
	}
	// The "/" pattern is skipped silently; valid-pattern remains
	if !m.matchesPath("valid-pattern") {
		t.Error("expected valid-pattern to be matched")
	}
}

// TestLoadPatternStateNoFile exercises the no-file branch returning empty state.
func TestLoadPatternStateNoFile(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateFile := filepath.Join(home, ".atheon", "pattern_state.json")

	// Save and remove
	backup, _ := os.ReadFile(stateFile)
	defer func() {
		if backup != nil {
			_ = os.WriteFile(stateFile, backup, 0o644)
		}
	}()
	_ = os.Remove(stateFile)

	state, err := loadPatternState()
	if err != nil {
		t.Fatalf("loadPatternState failed: %v", err)
	}
	if state == nil || state.Patterns == nil {
		t.Error("expected non-nil state with empty patterns map")
	}
	if len(state.Patterns) != 0 {
		t.Errorf("expected empty patterns, got %d", len(state.Patterns))
	}
}

// TestScanStringWithNewlines exercises ScanString with multi-line content.
func TestScanStringWithNewlines(t *testing.T) {
	findings := ScanString("line1\nAKIAIOSFODNN7EXAMPLE\nline3", "test")
	if len(findings) == 0 {
		t.Error("expected findings from multi-line scan string")
	}
	if len(findings) > 0 && findings[0].Pattern == "" {
		t.Error("expected non-empty pattern in finding")
	}
}

// TestIsIgnoredExercised verifies the isIgnored function with several patterns.
func TestIsIgnoredExercised(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".atheonignore"), []byte("*.log\n!important.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	patterns := loadIgnorePatternsMatcher(tmp)
	if patterns == nil {
		t.Skip("no patterns loaded")
	}

	if !isIgnored("test.log", patterns) {
		t.Error("expected test.log to be ignored")
	}
	// Negation rule should un-ignore important.log
	if isIgnored("important.log", patterns) {
		t.Error("expected important.log NOT to be ignored (negation rule)")
	}
}
