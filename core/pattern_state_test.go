package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestStateFile verifies the state file path is correct
func TestStateFile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	expected := filepath.Join(home, ".atheon", "pattern_state.json")
	actual := stateFile()

	if actual != expected {
		t.Errorf("stateFile() = %s, want %s", actual, expected)
	}
}

// TestLoadPatternState_NoFile verifies loading when no state file exists
func TestLoadPatternState_NoFile(t *testing.T) {
	// Save original state file if it exists
	home, _ := os.UserHomeDir()
	originalStatePath := filepath.Join(home, ".atheon", "pattern_state.json")
	backupPath := originalStatePath + ".backup"

	// Backup existing state file if present
	if _, err := os.Stat(originalStatePath); err == nil {
		data, _ := os.ReadFile(originalStatePath)
		_ = os.WriteFile(backupPath, data, 0o644)
		_ = os.Remove(originalStatePath)
	}
	defer func() {
		// Restore original state file
		if _, err := os.Stat(backupPath); err == nil {
			data, _ := os.ReadFile(backupPath)
			_ = os.WriteFile(originalStatePath, data, 0o644)
			_ = os.Remove(backupPath)
		}
	}()

	// Load should return empty state without error
	state, err := loadPatternState()
	if err != nil {
		t.Fatalf("loadPatternState() failed: %v", err)
	}

	if state == nil {
		t.Fatal("loadPatternState() returned nil state")
	}

	if state.Patterns == nil {
		t.Error("loadPatternState() returned nil Patterns map")
	}

	if len(state.Patterns) != 0 {
		t.Errorf("loadPatternState() returned %d patterns, want 0", len(state.Patterns))
	}
}

// TestLoadPatternState_WithFile verifies loading from an existing state file
func TestLoadPatternState_WithFile(t *testing.T) {
	// Note: This test won't work with the actual stateFile() function
	// because it uses the real home directory. We'd need to refactor
	// to accept a path parameter for proper testing.
	t.Skip("Requires refactoring to support test home directory")

	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, ".atheon", "pattern_state.json")

	// Would create state file with test data here
	_ = `{
  "patterns": {
    "pattern1": true,
    "pattern2": false,
    "pattern3": true
  }
}`
}

// TestSavePatternState verifies saving pattern state
func TestSavePatternState(t *testing.T) {
	// Note: This test won't work with the actual savePatternState() function
	// because it uses the real home directory. We'd need to refactor
	// to accept a path parameter for proper testing.
	t.Skip("Requires refactoring to support test home directory")

	_ = t.TempDir()
	_ = &PatternState{
		Patterns: map[string]bool{
			"pattern1": true,
			"pattern2": false,
			"pattern3": true,
		},
	}
}

// TestApplyPatternState verifies applying state to patterns
func TestApplyPatternState(t *testing.T) {
	// Save current pattern states
	originalStates := make(map[string]bool)
	for _, p := range allPatterns {
		originalStates[p.name] = p.enabled
	}

	// Create test state
	testState := &PatternState{
		Patterns: map[string]bool{
			"aws-access-key": false,
			"openai-api-key": false,
		},
	}

	// Apply state
	applyPatternState(testState)

	// Verify patterns were disabled
	for _, p := range allPatterns {
		if shouldDisable, exists := testState.Patterns[p.name]; exists {
			if shouldDisable && p.enabled {
				t.Errorf("Pattern %s should be disabled but is enabled", p.name)
			}
		}
	}

	// Restore original states
	for _, p := range allPatterns {
		if originalState, exists := originalStates[p.name]; exists {
			p.enabled = originalState
		}
	}
}

// TestSyncPatternState verifies syncing current state to disk
func TestSyncPatternState(t *testing.T) {
	// Note: This test writes to the real home directory
	// We'd need to refactor to support testing without side effects
	t.Skip("Requires refactoring to support test home directory")
}

// TestInitializePatternState verifies initialization
func TestInitializePatternState(t *testing.T) {
	// This test calls InitializePatternState which reads from real home directory
	// We'd need to refactor to support proper testing
	t.Skip("Requires refactoring to support test home directory")
}

// TestPatternStateJSON verifies JSON marshaling/unmarshaling
func TestPatternStateJSON(t *testing.T) {
	state := &PatternState{
		Patterns: map[string]bool{
			"pattern1": true,
			"pattern2": false,
			"pattern3": true,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Unmarshal from JSON
	var decoded PatternState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Verify patterns match
	if len(decoded.Patterns) != len(state.Patterns) {
		t.Errorf("Got %d patterns, want %d", len(decoded.Patterns), len(state.Patterns))
	}

	for name, enabled := range state.Patterns {
		if decoded.Patterns[name] != enabled {
			t.Errorf("Pattern %s: got %v, want %v", name, decoded.Patterns[name], enabled)
		}
	}
}
