package core

import (
	"os"
	"testing"
)

func TestAtomicWriteFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "atomic-test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmppath := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(tmppath)

	content := []byte("test content")
	if err := atomicWriteFile(tmppath, content, 0o644); err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(tmppath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "test content" {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestAtomicWriteFile_ReadOnlyDir(t *testing.T) {
	err := atomicWriteFile("/readonly/test.txt", []byte("test"), 0o644)
	if err == nil {
		t.Error("expected error writing to readonly dir")
	}
}
