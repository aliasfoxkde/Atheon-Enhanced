package core

import (
	"os"
	"path/filepath"
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
			originalEnabled = p.enabled
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
			originalEnabled = p.enabled
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
