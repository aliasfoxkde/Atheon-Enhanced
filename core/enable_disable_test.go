package core

import (
	"os"
	"testing"
)

// TestEnablePatternNotFound exercises the not-found branch of EnablePattern.
func TestEnablePatternNotFound(t *testing.T) {
	if EnablePattern("definitely-not-a-real-pattern-xyz-12345") {
		t.Error("EnablePattern should return false for unknown pattern")
	}
}

// TestDisablePatternNotFound exercises the not-found branch of DisablePattern.
func TestDisablePatternNotFound(t *testing.T) {
	if DisablePattern("definitely-not-a-real-pattern-xyz-12345") {
		t.Error("DisablePattern should return false for unknown pattern")
	}
}

// TestEnablePatternSyncError exercises the syncPatternState error branch by
// pointing HOME to a read-only directory temporarily.
//
// We can't easily change HOME for the duration of one call without affecting
// other tests, so instead we make the .atheon directory read-only and restore
// it afterward. This forces syncPatternState → savePatternState → WriteFile
// to fail, hitting the warning branch.
func TestEnablePatternSyncError(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateDir := home + "/.atheon"

	// Make the directory read-only so writes fail
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Skipf("cannot chmod state dir: %v", err)
	}
	defer func() {
		_ = os.Chmod(stateDir, 0o755) // restore so other tests work
	}()

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	// This call should still succeed but print a warning to stderr.
	// We don't capture stderr; we just verify the function returns true
	// (the pattern was found and enabled despite the sync failure).
	originalState := false
	for _, p := range allPatterns {
		if p.name == name {
			originalState = p.enabled.Load()
			break
		}
	}

	if !EnablePattern(name) {
		t.Error("EnablePattern should return true even when sync fails")
	}

	// Restore for subsequent tests
	if !originalState {
		DisablePattern(name)
	}
}

// TestDisablePatternSyncError exercises the syncPatternState error branch
// of DisablePattern.
func TestDisablePatternSyncError(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateDir := home + "/.atheon"

	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Skipf("cannot chmod state dir: %v", err)
	}
	defer func() {
		_ = os.Chmod(stateDir, 0o755)
	}()

	patterns := All()
	if len(patterns) == 0 {
		t.Skip("no patterns")
	}
	name := patterns[0].Name()

	originalState := false
	for _, p := range allPatterns {
		if p.name == name {
			originalState = p.enabled.Load()
			break
		}
	}

	if !DisablePattern(name) {
		t.Error("DisablePattern should return true even when sync fails")
	}

	// Restore
	if originalState {
		EnablePattern(name)
	}
}
