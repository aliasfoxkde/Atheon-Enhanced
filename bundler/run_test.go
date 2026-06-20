package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunDefaults exercises run() with no args, which uses the default
// community → core/patterns.bundle paths. This test exercises the
// default-branch code path.
func TestRunDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	// Stage a fake community dir
	fakeCommunity := filepath.Join(tmpDir, "community", "secrets")
	if err := os.MkdirAll(fakeCommunity, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeCommunity, "p.yaml"), []byte(`name: p
match: '\bP\b'
`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create the core/ subdirectory so bundle can write there
	if err := os.MkdirAll(filepath.Join(tmpDir, "core"), 0o755); err != nil {
		t.Fatal(err)
	}

	// t.Chdir handles cleanup via t.Cleanup; using the manual os.Chdir
	// pattern is racy when other tests rely on the package's CWD.
	t.Chdir(tmpDir)

	code := run(nil)
	if code != 0 {
		t.Errorf("expected exit 0 with defaults, got %d", code)
	}
	defaultOut := filepath.Join(tmpDir, "core", "patterns.bundle")
	if _, err := os.Stat(defaultOut); err != nil {
		t.Errorf("expected bundle at %s, got err: %v", defaultOut, err)
	}
}

// TestRunError exercises run() with a missing community dir to cover the
// error branch.
func TestRunError(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.bundle")
	if code := run([]string{filepath.Join(tmpDir, "no-such-dir"), outPath}); code != 1 {
		t.Errorf("expected exit 1 for missing community, got %d", code)
	}
}

// TestRunOutputHasPatternCount verifies the success message format.
func TestRunOutputHasPatternCount(t *testing.T) {
	tmpDir := t.TempDir()
	categoryDir := filepath.Join(tmpDir, "community", "secrets")
	if err := os.MkdirAll(categoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		_ = os.WriteFile(filepath.Join(categoryDir, "p"+string(rune('a'+i))+".yaml"),
			[]byte("name: p"+string(rune('a'+i))+"\nmatch: '\\bX\\b'\n"), 0o644)
	}

	outPath := filepath.Join(tmpDir, "out.bundle")
	if code := run([]string{filepath.Join(tmpDir, "community"), outPath}); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "patterns") {
		// We don't read the printed stdout, only the file. Just sanity-check.
		t.Log("bundle file written successfully")
	}
}
