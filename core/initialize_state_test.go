package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitializePatternStateMalformed exercises InitializePatternState
// when the state file is malformed JSON. The function should print a
// warning and return an error.
func TestInitializePatternStateMalformed(t *testing.T) {
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

	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stateFile, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := InitializePatternState()
	if err == nil {
		t.Error("expected error from InitializePatternState with malformed JSON")
	}
}

// TestInitializePatternStateNoFile exercises InitializePatternState when
// no state file exists. Should succeed with empty state.
func TestInitializePatternStateNoFile(t *testing.T) {
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

	_ = os.Remove(stateFile)

	if err := InitializePatternState(); err != nil {
		t.Errorf("expected no error with no state file, got: %v", err)
	}
}
