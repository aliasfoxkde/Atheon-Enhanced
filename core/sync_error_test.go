package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEnablePatternSyncErrorEnv exercises the syncPatternState error branch
// by setting HOME to a path through a non-directory file. This makes
// savePatternState → MkdirAll fail.
func TestEnablePatternSyncErrorEnv(t *testing.T) {
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	// Save original state
	originalEnabled := false
	for _, p := range allPatterns {
		if p.name == name {
			originalEnabled = p.enabled.Load()
			break
		}
	}

	// This should still return true even though sync fails
	if !EnablePattern(name) {
		t.Error("EnablePattern should return true even when sync fails")
	}

	// Restore
	if !originalEnabled {
		DisablePattern(name)
	}
}

// TestDisablePatternSyncErrorEnv exercises DisablePattern's sync error path.
func TestDisablePatternSyncErrorEnv(t *testing.T) {
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	originalEnabled := false
	for _, p := range allPatterns {
		if p.name == name {
			originalEnabled = p.enabled.Load()
			break
		}
	}

	if !DisablePattern(name) {
		t.Error("DisablePattern should return true even when sync fails")
	}

	// Restore
	if originalEnabled {
		EnablePattern(name)
	}
}

// setBlockedHome points HOME (Unix) or USERPROFILE (Windows) at a regular
// file so that os.UserHomeDir() returns a file path, causing
// savePatternState's os.MkdirAll to fail with "not a directory".
func setBlockedHome(t *testing.T, blockerFile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", blockerFile)
	} else {
		t.Setenv("HOME", blockerFile)
	}
}

// TestSavePatternStateMkdirErrorCrossplatform exercises the MkdirAll error
// branch in savePatternState on both Unix (via HOME) and Windows (via
// USERPROFILE). Pointing the home dir at a regular file makes the
// os.MkdirAll(.atheon) call fail with "not a directory".
func TestSavePatternStateMkdirErrorCrossplatform(t *testing.T) {
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setBlockedHome(t, blocker)

	state := &PatternState{Patterns: map[string]bool{"foo": true}}
	err := savePatternState(state)
	if err == nil {
		t.Error("expected error from savePatternState when home dir is a file")
	}
}

// TestEnablePatternSyncErrorCrossplatform is the cross-platform version of
// TestEnablePatternSyncErrorEnv using USERPROFILE on Windows.
func TestEnablePatternSyncErrorCrossplatform(t *testing.T) {
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setBlockedHome(t, blocker)

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	originalEnabled := false
	for _, p := range allPatterns {
		if p.name == name {
			originalEnabled = p.enabled.Load()
			break
		}
	}

	if !EnablePattern(name) {
		t.Error("EnablePattern should return true even when sync fails")
	}

	if !originalEnabled {
		DisablePattern(name)
	}
}

// TestDisablePatternSyncErrorCrossplatform is the cross-platform version of
// TestDisablePatternSyncErrorEnv using USERPROFILE on Windows.
func TestDisablePatternSyncErrorCrossplatform(t *testing.T) {
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setBlockedHome(t, blocker)

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	originalEnabled := false
	for _, p := range allPatterns {
		if p.name == name {
			originalEnabled = p.enabled.Load()
			break
		}
	}

	if !DisablePattern(name) {
		t.Error("DisablePattern should return true even when sync fails")
	}

	if originalEnabled {
		EnablePattern(name)
	}
}
