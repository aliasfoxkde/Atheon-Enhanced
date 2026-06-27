package core

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// atomicWriteFile writes data to path via tempfile-then-rename. POSIX rename
// is atomic within a filesystem, so a process crash, power loss, or SIGKILL
// in the middle of the write either leaves the previous file intact (if
// the rename hadn't run yet) or replaces it cleanly (if it had). Without
// this, `os.WriteFile` does truncate-then-write, and a crash in the gap
// leaves a zero-byte or partial file that downstream loads see as corrupt.
//
// The tmp file is placed in the same directory as path so the rename is
// on the same filesystem (a cross-filesystem rename is not atomic on most
// kernels and degrades to copy+delete).
//
// On Windows the rename is also atomic if the destination already exists,
// since Go's os.Rename on Windows uses MoveFileEx with MOVEFILE_REPLACE_EXISTING.
// New behavior, no caller needs to change.
func atomicWriteFile(path string, data []byte, perm os.FileMode) (retErr error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("atomic write: create temp: %w", err)
	}
	tmpName := tmp.Name()
	// Ensure the tmp file is removed on any failure path. On success the
	// rename consumes the inode and the Remove is a harmless no-op.
	defer func() {
		if retErr != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("atomic write: write data: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("atomic write: fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("atomic write: close: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("atomic write: chmod: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic write: rename: %w", err)
	}
	// Directory fsync: POSIX rename is atomic within a filesystem, but the
	// new directory entry isn't durable until the parent dir's metadata is
	// flushed. Without this, a power loss after Rename but before the dir
	// sync can leave the rename uncommitted — the file is gone, the .tmp
	// is gone, and the previous content isn't visible. Best-effort: skip
	// on platforms (Windows) where opening a directory is restricted.
	//
	// dir is filepath.Dir(path) — the parent of the file we're writing.
	// Callers pass user-home paths (e.g. ~/.atheon/patterns.bundle), so
	// the directory is the user's own home subdirectory, not a request-
	// controlled value. gosec G304 (file inclusion) does not apply.
	// #nosec G304
	if dirFd, err := os.Open(dir); err == nil {
		if err := dirFd.Sync(); err != nil {
			slog.Warn("directory fsync failed", "dir", dir, "err", err)
		}
		_ = dirFd.Close()
	}
	return nil
}
