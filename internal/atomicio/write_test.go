package atomicio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")

	data := []byte("hello world")
	if err := WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestWriteFile_CreatesParentDir(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "subdir", "nested")
	path := filepath.Join(subdir, "test.txt")

	// Create parent directory since os.CreateTemp needs it
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	data := []byte("nested")
	if err := WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "nested" {
		t.Errorf("got %q, want %q", got, "nested")
	}
}

func TestWriteFile_Overwrites(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "overwrite.txt")

	if err := WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("WriteFile first failed: %v", err)
	}
	if err := WriteFile(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("WriteFile second failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}

func TestWriteFile_EmptyData(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.txt")

	if err := WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile empty failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d bytes, want 0", len(got))
	}
}

func TestWriteFile_Permission(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "perms.txt")

	if err := WriteFile(path, []byte("perms"), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("got perm %o, want %o", info.Mode().Perm(), 0o600)
	}
}

func TestWriteFile_ReadOnlyDir(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "readonly")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	// Make the subdir read-only so we can't create files in it
	if err := os.Chmod(subdir, 0o444); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}
	defer os.Chmod(subdir, 0o755)

	path := filepath.Join(subdir, "test.txt")
	err := WriteFile(path, []byte("data"), 0o644)
	if err == nil {
		t.Error("expected error when writing to read-only directory")
	}
}
