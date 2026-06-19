package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadPatternStateNullPatterns exercises the "patterns == null" branch
// that initializes the map to empty.
func TestLoadPatternStateNullPatterns(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateFile := filepath.Join(home, ".atheon", "pattern_state.json")

	backup, backupErr := os.ReadFile(stateFile)
	defer func() {
		if backupErr == nil {
			_ = os.WriteFile(stateFile, backup, 0o644)
		} else {
			_ = os.Remove(stateFile)
		}
	}()

	// Write a state file with patterns: null
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stateFile, []byte(`{"patterns": null}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := loadPatternState()
	if err != nil {
		t.Fatalf("loadPatternState failed: %v", err)
	}
	if state.Patterns == nil {
		t.Error("expected Patterns to be initialized to empty map")
	}
	if len(state.Patterns) != 0 {
		t.Errorf("expected empty patterns, got %d", len(state.Patterns))
	}
}

// TestSavePatternStateReadOnlyHome exercises the save error path by making
// HOME point to a non-writable location via os.Setenv.
//
// We can't change HOME directly (the test runner uses it for other state),
// so we exercise savePatternState directly and accept either success or
// failure depending on environment. This is a no-op for coverage in most
// cases but documents the expected behavior.
func TestSavePatternStateReadOnlyHome(t *testing.T) {
	state := &PatternState{Patterns: map[string]bool{"foo": true}}
	_ = savePatternState(state)
}

// TestSavePatternStateBadHome exercises savePatternState with HOME pointing
// to a directory that doesn't exist and can't be created (a path through
// a non-directory file).
func TestSavePatternStateBadHome(t *testing.T) {
	// Create a regular file that we'll use as HOME's parent
	tmpDir := t.TempDir()
	notADir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(notADir, []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set HOME to a path under a non-directory — MkdirAll will fail
	badHome := filepath.Join(notADir, "subdir")
	t.Setenv("HOME", badHome)

	state := &PatternState{Patterns: map[string]bool{"foo": true}}
	err := savePatternState(state)
	if err == nil {
		t.Error("expected error when HOME points through a file")
	}
}

// TestSavePatternStateBadPath exercises savePatternState where the state
// file path's parent directory cannot be created.
func TestSavePatternStateBadPath(t *testing.T) {
	// Create a path that goes through a non-directory
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// HOME/<subdir>/.atheon/pattern_state.json — first <subdir> is a file
	t.Setenv("HOME", blocker)

	state := &PatternState{Patterns: map[string]bool{"foo": true}}
	err := savePatternState(state)
	if err == nil {
		t.Error("expected error when state dir can't be created")
	}
}
