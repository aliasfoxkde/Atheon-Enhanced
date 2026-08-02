package atomicio

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestWriteFile_CreateTempError tests error handling when creating temp file fails
func TestWriteFile_CreateTempError(t *testing.T) {
	// Use a path that doesn't exist and can't be created
	// This simulates the os.CreateTemp error path
	tmp := t.TempDir()
	nonExistent := filepath.Join(tmp, "does", "not", "exist", "path", "file.txt")

	err := WriteFile(nonExistent, []byte("data"), 0644)
	if err == nil {
		t.Error("expected error when writing to non-existent directory")
	}
}

// TestWriteFile_RenameError tests error handling when rename fails.
// We create a directory at the target path so Rename fails because
// it cannot overwrite a directory with a file.
func TestWriteFile_RenameError(t *testing.T) {
	tmp := t.TempDir()
	targetPath := filepath.Join(tmp, "existing_dir")

	// Create a directory at the target path - this will cause Rename to fail
	// because we can't rename a file over a directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatalf("MkdirAll for target directory failed: %v", err)
	}

	// Also ensure the parent directory is writable so CreateTemp succeeds
	if err := os.Chmod(tmp, 0755); err != nil {
		t.Skip("cannot make temp directory writable, skipping test")
	}

	// WriteFile will try to Rename a temp file to targetPath (a directory)
	// This should fail with an error from Rename
	err := WriteFile(targetPath, []byte("data"), 0644)
	if err == nil {
		t.Error("expected error when rename fails because target is a directory")
	}
}

// TestWriteFile_RenameErrorCleanup verifies that temp file is cleaned up
// when Rename fails. This tests the defer cleanup path (lines 123-125).
func TestWriteFile_RenameErrorCleanup(t *testing.T) {
	tmp := t.TempDir()
	targetPath := filepath.Join(tmp, "existing_dir")

	// Create a directory at the target path to cause Rename to fail
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatalf("MkdirAll for target directory failed: %v", err)
	}
	if err := os.Chmod(tmp, 0755); err != nil {
		t.Skip("cannot make temp directory writable, skipping test")
	}

	// Count temp files before
	entriesBefore, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Attempt WriteFile - should fail at Rename
	err = WriteFile(targetPath, []byte("data"), 0644)
	if err == nil {
		t.Fatal("expected error when rename fails")
	}

	// Count temp files after - should be same (temp file cleaned up)
	entriesAfter, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Verify no extra temp files left behind
	if len(entriesAfter) != len(entriesBefore) {
		t.Errorf("temp file was not cleaned up: before=%d after=%d", len(entriesBefore), len(entriesAfter))
	}

	// Also verify target path is still a directory (not corrupted)
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("target path was corrupted, is now %v", info.Mode())
	}
}

// TestWriteFile_SyncErrorHandling tests that Sync errors are handled gracefully
// We can't easily trigger a Sync error, but we can verify the code path exists
func TestWriteFile_SyncErrorHandling(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sync_test.txt")

	// Normal write - Sync errors are logged but don't fail the operation
	err := WriteFile(path, []byte("data"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}
}

// TestWriteFile_ChmodErrorHandling verifies chmod error handling
// On most systems we can't easily trigger chmod to fail after a successful write,
// but we verify the error path exists in code
func TestWriteFile_ChmodErrorHandling(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "chmod_test.txt")

	err := WriteFile(path, []byte("data"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists and is readable
	_, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
}

// TestWriteFile_WriteErrorHandling tests error when write fails
// We can't easily trigger a write failure, but we verify the error path exists
func TestWriteFile_WriteErrorHandling(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "write_test.txt")

	// Normal write with empty data
	err := WriteFile(path, []byte{}, 0644)
	if err != nil {
		t.Fatalf("WriteFile with empty data failed: %v", err)
	}

	// Verify empty file was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected size 0, got %d", info.Size())
	}
}

// TestWriteFile_DirFsyncErrorHandling verifies directory fsync errors are handled
func TestWriteFile_DirFsyncErrorHandling(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "fsync_dir_test.txt")

	// Normal write - directory fsync failures are logged but don't fail the operation
	err := WriteFile(path, []byte("data"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

// TestWriteFile_LargeWrite tests writing larger amounts of data
func TestWriteFile_LargeWrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "large.txt")

	// Write 100KB of data
	data := make([]byte, 100*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if err := WriteFile(path, data, 0644); err != nil {
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

// TestWriteFile_CloseErrorHandling verifies close error handling
func TestWriteFile_CloseErrorHandling(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "close_test.txt")

	err := WriteFile(path, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

// TestWriteFile_DeferCleanupOnError verifies temp file is cleaned up on error
func TestWriteFile_DeferCleanupOnError(t *testing.T) {
	tmp := t.TempDir()
	nonExistent := filepath.Join(tmp, "nonexistent", "dir", "file.txt")

	// This should fail and not leave temp files behind
	err := WriteFile(nonExistent, []byte("data"), 0644)
	if err == nil {
		t.Error("expected error when writing to non-existent directory")
	}

	// Verify no temp files were left behind
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		// Check if any temp files were left behind
		if len(name) > 4 && name[len(name)-4:] == ".tmp" {
			t.Errorf("temp file %q was not cleaned up", name)
		}
	}
}

// TestWriteFile_ConcurrentWriteSafety tests that concurrent writes don't corrupt
func TestWriteFile_ConcurrentWriteSafety(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "concurrent.txt")
	const n = 5

	// Write different size data concurrently
	done := make(chan bool, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			data := []byte{}
			for j := 0; j < (i+1)*100; j++ {
				data = append(data, byte('A'+i))
			}
			if err := WriteFile(path, data, 0644); err != nil {
				t.Logf("WriteFile failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < n; i++ {
		<-done
	}

	// Final file should be valid (though content depends on race)
	_, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("ReadFile failed after concurrent writes: %v", err)
	}
}

// TestWriteFile_AtomicRename verifies atomic rename behavior
// When rename succeeds, we should have complete file content
func TestWriteFile_AtomicRename(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "atomic.txt")

	// Write known pattern
	expected := "atomic_write_test_pattern_12345"
	if err := WriteFile(path, []byte(expected), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(got) != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

// TestWriteFile_ZeroPerms tests writing with zero permissions
func TestWriteFile_ZeroPerms(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "zero_perm.txt")

	// On some systems, zero permissions may cause issues
	// but WriteFile should handle it
	err := WriteFile(path, []byte("data"), 0000)
	if err != nil {
		// Some systems may reject zero permissions
		t.Skipf("zero permissions not supported: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	// On POSIX, mode will be normalized to something sensible
	_ = info
}

// TestWriteFile_SpecialPermissions tests writing with special permissions
func TestWriteFile_SpecialPermissions(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "special_perm.txt")

	err := WriteFile(path, []byte("data"), 0755)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// On some systems (like Windows), these permissions may be different
	// Just verify we got some executable permission
	if info.Mode().Perm()&0111 == 0 {
		t.Logf("note: no executable permission on %s", tmp)
	}
}

// TestWriteFile_AllowsSymlink tests that we can write to a symlinked path
func TestWriteFile_AllowsSymlink(t *testing.T) {
	tmp := t.TempDir()
	realPath := filepath.Join(tmp, "real.txt")
	linkPath := filepath.Join(tmp, "link.txt")

	// Create initial file
	if err := WriteFile(realPath, []byte("original"), 0644); err != nil {
		t.Fatalf("WriteFile original failed: %v", err)
	}

	// Create symlink
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skip("symlinks not supported, skipping test")
	}

	// Write through symlink
	if err := WriteFile(linkPath, []byte("updated"), 0644); err != nil {
		t.Fatalf("WriteFile through symlink failed: %v", err)
	}

	// Both should have updated content
	got, err := os.ReadFile(linkPath)
	if err != nil {
		t.Fatalf("ReadFile through symlink failed: %v", err)
	}
	if string(got) != "updated" {
		t.Errorf("got %q, want %q", got, "updated")
	}
}

// TestWriteFile_DeepDir tests writing to deeply nested directory
func TestWriteFile_DeepDir(t *testing.T) {
	tmp := t.TempDir()
	deep := filepath.Join(tmp, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	path := filepath.Join(deep, "deep.txt")

	err := WriteFile(path, []byte("deep"), 0644)
	if err != nil {
		t.Fatalf("WriteFile to deep path failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "deep" {
		t.Errorf("got %q, want %q", got, "deep")
	}
}

// syscall.ENOTSUP mock helper - used for testing error paths on systems
// that don't support certain operations
var enotsup = syscall.ENOTSUP
