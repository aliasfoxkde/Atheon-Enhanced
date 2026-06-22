package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PatternState stores the enabled/disabled state of patterns
type PatternState struct {
	Patterns map[string]bool `json:"patterns"` // pattern name -> enabled state
}

// stateFile returns the path to the pattern state file
func stateFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".atheon", "pattern_state.json")
}

// loadPatternState loads the pattern state from disk
func loadPatternState() (*PatternState, error) {
	path := stateFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No state file exists, return empty state
			return &PatternState{Patterns: make(map[string]bool)}, nil
		}
		return nil, err
	}

	var state PatternState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse pattern state: %w", err)
	}

	if state.Patterns == nil {
		state.Patterns = make(map[string]bool)
	}

	return &state, nil
}

// savePatternState saves the pattern state to disk
func savePatternState(state *PatternState) error {
	path := stateFile()
	dir := filepath.Dir(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pattern state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write pattern state: %w", err)
	}

	return nil
}

// applyPatternState applies the loaded state to allPatterns
func applyPatternState(state *PatternState) {
	for _, p := range allPatterns {
		if enabled, exists := state.Patterns[p.name]; exists {
			p.enabled = enabled
		}
	}
}

// syncPatternState syncs the current pattern state to disk
func syncPatternState() error {
	state := &PatternState{Patterns: make(map[string]bool)}

	// Collect current state from allPatterns
	for _, p := range allPatterns {
		state.Patterns[p.name] = p.enabled
	}

	return savePatternState(state)
}

// InitializePatternState loads and applies pattern state on startup
func InitializePatternState() error {
	state, err := loadPatternState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load pattern state: %v\n", err)
		return err
	}

	applyPatternState(state)
	return nil
}
