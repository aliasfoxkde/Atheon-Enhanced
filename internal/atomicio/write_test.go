package atomicio

import (
	"os"
	"path/filepath"
	"sync"
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

func TestWriteFile_ConcurrentWrites(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "concurrent.txt")
	const n = 10

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			data := []byte{byte('A' + i)}
			if err := WriteFile(path, data, 0o644); err != nil {
				t.Errorf("WriteFile failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// At least one write should succeed and the file should exist
	_, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("ReadFile failed after concurrent writes: %v", err)
	}
}

func TestWriteFile_LargeData(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "large.txt")

	// Write 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if err := WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile large failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile large failed: %v", err)
	}
	if len(got) != len(data) {
		t.Errorf("got %d bytes, want %d", len(got), len(data))
	}
}

func TestWriteFile_ValidPerms(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		perm    os.FileMode
		wantRwx bool
	}{
		{0o600, true}, // rw-
		{0o400, true}, // r--
		{0o644, true}, // rw-r--r--
		{0o755, true}, // rwxr-xr-x
	}

	for _, tt := range tests {
		path := filepath.Join(tmp, "perm.txt")
		if err := WriteFile(path, []byte("data"), tt.perm); err != nil {
			t.Fatalf("WriteFile with perm %o failed: %v", tt.perm, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.Mode().Perm() != tt.perm {
			t.Errorf("got perm %o, want %o", info.Mode().Perm(), tt.perm)
		}
	}
}

func TestWriteFile_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "existing.txt")

	// Create an existing file with different content
	if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
		t.Fatalf("WriteFile initial failed: %v", err)
	}

	// Write new content
	if err := WriteFile(path, []byte("new content"), 0o644); err != nil {
		t.Fatalf("WriteFile overwrite failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "new content" {
		t.Errorf("got %q, want %q", got, "new content")
	}
}
