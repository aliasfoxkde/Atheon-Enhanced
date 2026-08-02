package core

import (
	"os"
	"testing"
)

// TestInitializePatternStateOK exercises InitializePatternState by setting
// a known state file and reloading.
func TestInitializePatternStateOK(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateDir := home + "/.atheon"
	stateFile := stateDir + "/pattern_state.json"

	// Save existing state if any
	backup, backupErr := os.ReadFile(stateFile)
	defer func() {
		if backupErr == nil {
			_ = os.WriteFile(stateFile, backup, 0o644)
		} else {
			_ = os.Remove(stateFile)
		}
	}()

	// Write a state that disables the first pattern and enables a non-existent one
	all := All()
	if len(all) == 0 {
		t.Skip("No patterns available")
	}
	targetName := all[0].Name()

	state := &PatternState{
		Patterns: map[string]bool{
			targetName:        false,
			"phantom-pattern": true,
		},
	}
	if err := savePatternState(state); err != nil {
		t.Fatalf("savePatternState failed: %v", err)
	}

	if err := InitializePatternState(); err != nil {
		t.Fatalf("InitializePatternState failed: %v", err)
	}

	// Verify target is now disabled
	for _, p := range allPatterns {
		if p.name == targetName && p.enabled.Load() {
			t.Errorf("Pattern %s should be disabled after InitializePatternState", targetName)
		}
	}

	// Restore the pattern's enabled state so other tests aren't affected
	EnablePattern(targetName)
}

// TestSyncPatternStateRun calls syncPatternState end-to-end after toggling
// patterns, verifying the state file reflects the current pattern states.
func TestSyncPatternStateRun(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Skip("No patterns available")
	}
	name := all[0].Name()

	// Save original state
	originalEnabled := false
	for _, p := range allPatterns {
		if p.name == name {
			originalEnabled = p.enabled.Load()
			break
		}
	}

	// Toggle and sync
	DisablePattern(name)
	if err := syncPatternState(); err != nil {
		t.Fatalf("syncPatternState failed: %v", err)
	}

	// Read back and verify
	loaded, err := loadPatternState()
	if err != nil {
		t.Fatalf("loadPatternState failed: %v", err)
	}
	if got, ok := loaded.Patterns[name]; !ok || got {
		t.Errorf("Expected %s disabled in state file, got %v (exists=%v)", name, got, ok)
	}

	// Restore
	if originalEnabled {
		EnablePattern(name)
	} else {
		DisablePattern(name)
	}
}

// TestApplyPatternStateAllDisabled verifies applyPatternState disables every
// pattern that has an entry in the state map.
func TestApplyPatternStateAllDisabled(t *testing.T) {
	if len(allPatterns) < 2 {
		t.Skip("Need at least 2 patterns")
	}
	name1 := allPatterns[0].name
	name2 := allPatterns[1].name

	state := &PatternState{
		Patterns: map[string]bool{
			name1: false,
			name2: false,
		},
	}
	applyPatternState(state)

	for _, p := range allPatterns {
		if p.name == name1 || p.name == name2 {
			if p.enabled.Load() {
				t.Errorf("Pattern %s should be disabled", p.name)
			}
		}
	}

	// Restore for subsequent tests
	EnableAllPatterns()
}
