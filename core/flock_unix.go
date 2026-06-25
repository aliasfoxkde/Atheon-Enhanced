//go:build !windows

package core

import (
	"fmt"
	"os"
	"syscall"
)

// withFileLock takes an exclusive flock on path for the duration of fn and
// releases it on return. Used to serialize concurrent saves of the user
// pattern-state file across `atheon enable` (CLI) and `atheon-mcp` agent
// invocations on the same host — without it, two writers can interleave
// reads and writes and lose each other's intent.
//
// On POSIX, syscall.Flock is a real OS-level lock that survives fork/exec
// and works across processes on the same host. Combined with the atomic
// rename in atomicWriteFile, it gives us read-modify-write atomicity for
// the pattern_state.json file.
func withFileLock(path string, fn func() error) error {
	// path is the user-home lockfile (~/.atheon/patterns.bundle.lock or
	// ~/.atheon/pattern_state.json.lock). It is computed inside this
	// package from os.UserHomeDir() — not a request-controlled value —
	// so G304 (potential file inclusion) does not apply.
	// #nosec G304
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		// ENOENT is fine — first save creates the file; we'll try again.
		if !os.IsNotExist(err) {
			return fmt.Errorf("flock open: %w", err)
		}
		// #nosec G304 -- see comment above
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return fmt.Errorf("flock create: %w", err)
		}
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock acquire: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	return fn()
}
