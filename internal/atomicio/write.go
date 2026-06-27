// Package atomicio provides atomic file writing utilities.
package atomicio

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// WriteFile writes data to path via tempfile-then-rename. POSIX rename
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
//
// The directory fsync ensures directory entry durability after rename.
// Without this, a power loss after Rename but before the dir sync can
// leave the rename uncommitted. Best-effort on platforms (Windows) where
// opening a directory is restricted.
//
// Callers pass user-home paths (e.g. ~/.atheon/patterns.bundle), so
// the directory is the user's own home subdirectory, not a user-controlled
// value. gosec G304 (file inclusion) does not apply.
func WriteFile(path string, data []byte, perm os.FileMode) (retErr error) {
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
	// Directory fsync for durability.
	// #nosec G304 -- dir is derived from the caller's path argument which
	// is not user-controlled for our use cases.
	if dirFd, err := os.Open(dir); err == nil {
		if err := dirFd.Sync(); err != nil {
			slog.Warn("directory fsync failed", "dir", dir, "err", err)
		}
		_ = dirFd.Close()
	}
	return nil
}
