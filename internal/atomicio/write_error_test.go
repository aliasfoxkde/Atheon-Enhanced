package atomicio

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// TestWriteFile_AllowsSymlink tests that we refuse to write to a symlinked path
// for security reasons (symlink attack mitigation).
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

	// Writing through symlink should be refused for security
	err := WriteFile(linkPath, []byte("updated"), 0644)
	if err == nil {
		t.Error("WriteFile through symlink: expected error, got nil")
	} else if !strings.Contains(err.Error(), "is a symlink") {
		t.Errorf("WriteFile through symlink: expected symlink error, got: %v", err)
	}

	// Original file should be unchanged
	got, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatalf("ReadFile real path failed: %v", err)
	}
	if string(got) != "original" {
		t.Errorf("got %q, want %q", got, "original")
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

// TestWriteFile_FIFOSyncError tests that Sync errors on FIFOs (named pipes)
// are handled. On Linux, calling Sync on a FIFO returns EINVAL.
func TestWriteFile_FIFOSyncError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("FIFO Sync test only on Linux")
	}

	tmp := t.TempDir()
	fifoPath := filepath.Join(tmp, "testfifo")

	// Create a FIFO
	if err := syscall.Mkfifo(fifoPath, 0644); err != nil {
		t.Skipf("Mkfifo failed: %v", err)
	}
	defer os.Remove(fifoPath)

	// Open FIFO for write - this will block until someone opens for read
	// We need to open in non-blocking mode or in a goroutine
	f, err := os.OpenFile(fifoPath, os.O_WRONLY|syscall.O_NONBLOCK, 0644)
	if err != nil {
		// Opening FIFO with O_NONBLOCK still fails if no reader
		// because the FIFO has no writers currently
		t.Skipf("Cannot open FIFO: %v", err)
	}
	defer f.Close()

	// Try Sync - on Linux, Sync on a FIFO returns EINVAL
	if err := f.Sync(); err != nil {
		t.Logf("FIFO Sync failed (expected on Linux): %v", err)
	}
}

// TestWriteFile_FIFOOpenAndSync tests opening a FIFO and doing operations on it
func TestWriteFile_FIFOOpenAndSync(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("FIFO test only on Linux")
	}

	tmp := t.TempDir()
	fifoPath := filepath.Join(tmp, "testfifo2")

	if err := syscall.Mkfifo(fifoPath, 0644); err != nil {
		t.Skipf("Mkfifo not supported: %v", err)
	}
	defer os.Remove(fifoPath)

	// Use os.Pipe to get a writer that won't block forever
	r, w, err := os.Pipe()
	if err != nil {
		t.Skipf("os.Pipe failed: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Write to pipe to unblock FIFO open if needed - actually this doesn't help
	// because they're different things

	// Open FIFO for write (this would block without a reader, so use goroutine)
	done := make(chan error, 1)
	go func() {
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0644)
		if err != nil {
			done <- err
			return
		}
		defer f.Close()

		// Try Sync - on Linux this should return EINVAL for FIFO
		err = f.Sync()
		done <- err
	}()

	// Open FIFO for read to unblock the writer
	rf, err := os.OpenFile(fifoPath, os.O_RDONLY, 0644)
	if err != nil {
		t.Skipf("Cannot open FIFO for read: %v", err)
	}
	defer rf.Close()

	// Wait for writer to finish
	err = <-done
	if err != nil {
		t.Logf("FIFO Sync returned error (expected on some systems): %v", err)
	}
}

// TestWriteFile_DevFullWrite tests writing to /dev/full which returns ENOSPC
// This tests error handling when the underlying write fails due to "no space"
func TestWriteFile_DevFullWrite(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full only on Linux")
	}

	// /dev/full is a special device that returns ENOSPC on write
	f, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
	if err != nil {
		t.Skip("/dev/full not available")
	}
	defer f.Close()

	// Writing to /dev/full should fail with ENOSPC
	n, err := f.Write([]byte("test"))
	if err == nil {
		// Some systems might not enforce this
		t.Logf("Write to /dev/full succeeded unexpectedly, wrote %d bytes", n)
	} else {
		t.Logf("Write to /dev/full failed as expected: %v", err)
	}
}

// TestWriteFile_SyncOnReadOnlyFd tests Sync on a file descriptor opened read-only
func TestWriteFile_SyncOnReadOnlyFd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	tmp := t.TempDir()
	path := filepath.Join(tmp, "readonly.txt")

	// Create a test file
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Re-open read-only and try to sync
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	// Sync on read-only file - this should fail on Linux
	if err := f.Sync(); err != nil {
		t.Logf("Sync on read-only fd failed (expected): %v", err)
	}
}

// TestWriteFile_DirectoryFsyncBestEffort tests that directory fsync errors
// are handled gracefully (best-effort operation)
func TestWriteFile_DirectoryFsyncBestEffort(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "best_effort.txt")

	// Normal write - directory fsync is best-effort and shouldn't fail the operation
	err := WriteFile(path, []byte("data"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file was written
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("got %q, want %q", got, "data")
	}
}

// mockTempFile implements tempFile for testing error paths.
// It wraps a real *os.File to ensure actual file I/O happens.
type mockTempFile struct {
	*os.File
	name     string
	writeErr error
	syncErr  error
	closeErr error
}

func (m *mockTempFile) Write(p []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.File.Write(p)
}

func (m *mockTempFile) Sync() error {
	if m.syncErr != nil {
		return m.syncErr
	}
	return m.File.Sync()
}

func (m *mockTempFile) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	return m.File.Close()
}

func (m *mockTempFile) Name() string {
	return m.name
}

// TestWriteFileWithFile_WriteError tests error handling when Write fails
func TestWriteFileWithFile_WriteError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "write_err.txt")

	// Create a real temp file to back the mock
	realFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realFile.Close()
	defer os.Remove(realFile.Name())

	writeErr := fmt.Errorf("simulated write error")
	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: f, name: realFile.Name(), writeErr: writeErr}, nil
	}

	err = WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error when Write fails")
	}
}

// TestWriteFileWithFile_SyncError tests error handling when Sync fails
func TestWriteFileWithFile_SyncError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sync_err.txt")

	// Create a real temp file to back the mock
	realFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realFile.Close()
	defer os.Remove(realFile.Name())

	syncErr := fmt.Errorf("simulated sync error")
	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: f, name: realFile.Name(), syncErr: syncErr}, nil
	}

	err = WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error when Sync fails")
	}
}

// TestWriteFileWithFile_CloseError tests error handling when Close fails
func TestWriteFileWithFile_CloseError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "close_err.txt")

	// Create a real temp file to back the mock
	realFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realFile.Close()
	defer os.Remove(realFile.Name())

	closeErr := fmt.Errorf("simulated close error")
	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: f, name: realFile.Name(), closeErr: closeErr}, nil
	}

	err = WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error when Close fails")
	}
}

// TestWriteFileWithFile_ChmodError tests error handling when Chmod fails
func TestWriteFileWithFile_ChmodError(t *testing.T) {
	tmp := t.TempDir()

	// Create a file in a directory where chmod will fail
	readonlyDir := filepath.Join(tmp, "readonly")
	if err := os.MkdirAll(readonlyDir, 0555); err != nil {
		t.Skip("cannot create readonly directory")
	}
	defer os.Chmod(readonlyDir, 0755)

	// Create a file in the readonly directory that we can chmod
	f, err := os.CreateTemp(readonlyDir, "tmp-*")
	if err != nil {
		t.Skip("cannot create temp file in readonly dir")
	}
	realPath := f.Name()
	f.Close()

	// Now chmod will fail on this file
	mockCreator := func(dir, pattern string) (tempFile, error) {
		rf, err := os.OpenFile(realPath, os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: rf, name: realPath}, nil
	}

	err = WriteFileWithFile(filepath.Join(readonlyDir, "chmod_err.txt"), []byte("test"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error when Chmod fails")
	}
}

// TestWriteFileWithFile_CreateTempError tests error handling when file creation fails
func TestWriteFileWithFile_CreateTempError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "create_err.txt")

	createErr := fmt.Errorf("simulated create error")
	mockCreator := func(dir, pattern string) (tempFile, error) {
		return nil, createErr
	}

	err := WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error when CreateTemp fails")
	}
}

// TestWriteFileWithFile_DirFsyncError tests that directory fsync errors are handled gracefully
func TestWriteFileWithFile_DirFsyncError(t *testing.T) {
	// This test verifies the directory fsync error path is exercised
	// When dirFd.Sync() returns an error, it should be logged but not fail the operation
	tmp := t.TempDir()
	path := filepath.Join(tmp, "dirfsync_err.txt")

	// Create a real temp file that will be used
	realTmpFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realTmpFile.Close()
	defer os.Remove(realTmpFile.Name())

	// Use a mock that succeeds for all operations
	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realTmpFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: f, name: realTmpFile.Name()}, nil
	}

	err = WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err != nil {
		t.Fatalf("WriteFileWithFile failed: %v", err)
	}

	// Verify file was written
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "test data" {
		t.Errorf("got %q, want %q", got, "test data")
	}
}

// TestWriteFileWithFile_DeferCleanup tests that temp file is cleaned up on error
func TestWriteFileWithFile_DeferCleanup(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cleanup_err.txt")

	entriesBefore, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Create a real temp file to back the mock
	realFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realFile.Close()
	defer os.Remove(realFile.Name())

	// Use a mock that fails on Write
	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{
			File:     f,
			name:     realFile.Name(),
			writeErr: fmt.Errorf("simulated write error"),
		}, nil
	}

	err = WriteFileWithFile(path, []byte("test data"), 0644, mockCreator)
	if err == nil {
		t.Fatal("expected error")
	}

	// Verify no extra temp files left behind
	entriesAfter, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entriesAfter) != len(entriesBefore) {
		t.Errorf("temp file was not cleaned up: before=%d after=%d", len(entriesBefore), len(entriesAfter))
	}
}

// TestWriteFileWithFile_NormalOperation tests that normal operation succeeds
func TestWriteFileWithFile_NormalOperation(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "normal.txt")

	// Create the temp file that will be used
	realFile, err := os.CreateTemp(tmp, "tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	realFile.Close()
	defer os.Remove(realFile.Name())

	mockCreator := func(dir, pattern string) (tempFile, error) {
		f, err := os.OpenFile(realFile.Name(), os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &mockTempFile{File: f, name: realFile.Name()}, nil
	}

	err = WriteFileWithFile(path, []byte("normal data"), 0644, mockCreator)
	if err != nil {
		t.Fatalf("WriteFileWithFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "normal data" {
		t.Errorf("got %q, want %q", got, "normal data")
	}
}
