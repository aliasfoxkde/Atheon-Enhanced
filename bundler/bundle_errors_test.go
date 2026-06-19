package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBundleReadFileError exercises the os.ReadFile error branch by
// creating a directory with a non-readable file inside.
func TestBundleReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community", "secrets")
	if err := os.MkdirAll(communityDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file then chmod it to be unreadable
	badPath := filepath.Join(communityDir, "unreadable.yaml")
	if err := os.WriteFile(badPath, []byte(`name: x
match: '\bX\b'
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(badPath, 0o000); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(badPath, 0o644)

	outPath := filepath.Join(tmpDir, "out.bundle")
	_, err := bundle(filepath.Join(tmpDir, "community"), outPath)
	if err == nil {
		t.Error("expected error for unreadable file")
	}
}

// TestBundleWriteFileError exercises the os.WriteFile error branch by
// pointing outPath to an unwritable directory.
func TestBundleWriteFileError(t *testing.T) {
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community", "secrets")
	if err := os.MkdirAll(communityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(communityDir, "p.yaml"), []byte(`name: p
match: '\bP\b'
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Make the output dir read-only
	outDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(outDir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(outDir, 0o755)

	outPath := filepath.Join(outDir, "out.bundle")
	_, err := bundle(filepath.Join(tmpDir, "community"), outPath)
	if err == nil {
		t.Error("expected error for unwritable output path")
	}
}
