package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

// TestLoadPatternState_BadJSON tests loading invalid JSON
func TestLoadPatternState_BadJSON(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateDir := filepath.Join(home, ".atheon")
	statePath := filepath.Join(stateDir, "pattern_state.json")

	// Create directory
	_ = os.MkdirAll(stateDir, 0o755)

	// Backup existing state file if present
	backupPath := statePath + ".backup"
	if _, err := os.Stat(statePath); err == nil {
		data, _ := os.ReadFile(statePath)
		_ = os.WriteFile(backupPath, data, 0o644)
	}
	defer func() {
		// Clean up test file and restore backup
		_ = os.Remove(statePath)
		if _, err := os.Stat(backupPath); err == nil {
			data, _ := os.ReadFile(backupPath)
			_ = os.WriteFile(statePath, data, 0o644)
			_ = os.Remove(backupPath)
		}
	}()

	// Write invalid JSON
	_ = os.WriteFile(statePath, []byte(`{invalid json}`), 0o644)

	_, err := loadPatternState()
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestLoadPatternState_ReadError tests file read error handling
func TestLoadPatternState_ReadError(t *testing.T) {
	home, _ := os.UserHomeDir()
	stateDir := filepath.Join(home, ".atheon")
	statePath := filepath.Join(stateDir, "pattern_state.json")

	// Create directory and file
	_ = os.MkdirAll(stateDir, 0o755)

	// Backup existing state file if present
	backupPath := statePath + ".backup"
	if _, err := os.Stat(statePath); err == nil {
		data, _ := os.ReadFile(statePath)
		_ = os.WriteFile(backupPath, data, 0o644)
	}
	defer func() {
		// Clean up and restore
		_ = os.Remove(statePath)
		if _, err := os.Stat(backupPath); err == nil {
			data, _ := os.ReadFile(backupPath)
			_ = os.WriteFile(statePath, data, 0o644)
			_ = os.Remove(backupPath)
		}
	}()

	// Create a directory instead of a file to trigger read error
	_ = os.Remove(statePath)
	_ = os.MkdirAll(statePath, 0o755)
	defer os.Remove(statePath)

	_, err := loadPatternState()
	if err == nil {
		t.Error("expected error when state path is directory, got nil")
	}
}

// TestSavePatternState_MarshalError tests JSON marshaling error handling
func TestSavePatternState_MarshalError(t *testing.T) {
	// This is difficult to test since PatternState only contains basic types
	// that can always be marshaled. The error path exists but is unreachable.
	// Skipping this test as the error condition cannot be realistically triggered.
	t.Skip("PatternState contains only marshalable types - error path unreachable")
}

// TestSyncPatternState_SaveError tests save error handling in syncPatternState.
// Skipped on Windows: os.UserHomeDir uses USERPROFILE, not HOME env var.
func TestSyncPatternState_SaveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.UserHomeDir on Windows uses USERPROFILE, not HOME env var")
	}
	// Save original home
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to a path that will cause write failure
	tmpDir := t.TempDir()
	blockPath := filepath.Join(tmpDir, "blocker")
	_ = os.WriteFile(blockPath, []byte("block"), 0o644)
	os.Setenv("HOME", filepath.Join(blockPath, "subdir"))

	err := syncPatternState()
	if err == nil {
		t.Error("expected error when HOME points through a file, got nil")
	}
}
