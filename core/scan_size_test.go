// Scaffolding for the file-size cap behaviour that Wave 8 PR #95 introduces
// in ScanDir. PR #95 adds a readFileCapped(path, maxBytes) helper and wires
// it into the worker goroutines so a single 10GB file can't exhaust process
// memory. The test below exercises the contract once the helper lands —
// until then it documents the expected behaviour with t.Skip so the file
// compiles and the test is discoverable via `go test -list`.
//
// What's being asserted when PR #95 ships:
//   - files < maxBytes: full content returned
//   - files == maxBytes: full content returned (boundary inclusive)
//   - files >  maxBytes: ErrFileTooLarge returned, no content read
//   - files == 0: empty content returned, no error (zero is a valid size)
//   - files unreadable (perm denied): read error surfaced, not silently dropped
//
// Keep this file's helpers (writeFileWithSize) — PR #95's tests will reuse
// them.

package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ErrFileTooLarge is the sentinel error PR #95 returns when a file exceeds
// the configured cap. Declared here so tests can assert on it before the
// production helper exists.
var ErrFileTooLarge = errors.New("core: file exceeds configured max bytes")

// writeFileWithSize creates a file of exactly `size` bytes at the given
// path. Uses Truncate so it works for size==0 (Seek(size-1) would fail for
// size==0 because there's no negative offset to seek to in a fresh file).
// Sparse-friendly: no actual data is written for non-zero sizes.
func writeFileWithSize(t *testing.T, path string, size int64) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		t.Fatalf("truncate %s to %d: %v", path, size, err)
	}
}

// TestScanSizeCap_Scaffold is a placeholder until PR #95 lands the helper.
// When the helper exists, replace this with concrete assertions:
//   - under-cap file: returned in full
//   - over-cap file: ErrFileTooLarge
//   - zero-byte file: returned empty, no error
func TestScanSizeCap_Scaffold(t *testing.T) {
	t.Skip("readFileCapped helper ships in Wave 8 PR #95 — see plan:fix/wave8-runner-safety")
	dir := t.TempDir()
	small := filepath.Join(dir, "small.txt")
	writeFileWithSize(t, small, 1024)
	_ = small
}
