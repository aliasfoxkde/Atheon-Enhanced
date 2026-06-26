package core

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// PatternState stores the enabled/disabled state of patterns and metadata
// about the downloaded bundle. It is persisted to ~/.atheon/pattern_state.json
// between invocations so user preferences survive across runs.
type PatternState struct {
	Patterns map[string]bool `json:"patterns"` // pattern name -> enabled state
	// BundleETag is the ETag returned by the upstream bundle server on the
	// last successful download. Used to skip redundant downloads when the
	// upstream bundle hasn't changed.
	BundleETag string `json:"bundleETag,omitempty"`
	// BundleLastChecked is the time of the last bundle download attempt
	// (successful or not). Used together with BundleETag to implement
	// stale-bundle detection: if the bundle is recent (<24h) and the ETag
	// matches, skip the download entirely.
	BundleLastChecked int64 `json:"bundleLastChecked,omitempty"` // Unix nanos, 0 = never checked
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

	// Atomic write (tempfile + rename) so a SIGKILL/power-loss mid-write
	// leaves the previous pattern_state.json intact instead of producing a
	// partial file that loadPatternState then reports as "failed to parse".
	// Cross-process safety for concurrent CLI + MCP invocations is layered
	// on top via withFileLock in flock_unix.go / flock_windows.go.
	if err := atomicWriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write pattern state: %w", err)
	}

	return nil
}

// applyPatternState applies the loaded state to allPatterns.
// Caller must hold patternMu for writing.
func applyPatternState(state *PatternState) {
	for _, p := range allPatterns {
		if enabled, exists := state.Patterns[p.name]; exists {
			p.enabled = enabled
		}
	}
}

// syncPatternState syncs the current pattern state to disk.
// Caller must hold patternMu for writing (the loadBundle/Enable/Disable
// call sites do, since they touch the same state).
func syncPatternState() error {
	return withFileLock(stateFile(), func() error {
		// Read-merge-write under the cross-process lock. Two concurrent
		// writers (CLI + MCP) each have their own in-memory allPatterns;
		// a naive last-writer-wins would clobber the other process's
		// Enable/Disable. The flock serialises us; reading the on-disk
		// state inside the lock and merging gives the second writer the
		// first writer's changes before we overwrite.
		onDisk, err := loadPatternState()
		if err != nil {
			return fmt.Errorf("sync: read on-disk state: %w", err)
		}
		// Snapshot our in-memory state — caller holds patternMu, so
		// allPatterns is stable for the rest of this closure.
		for _, p := range allPatterns {
			onDisk.Patterns[p.name] = p.enabled
		}
		return savePatternState(onDisk)
	})
}

// InitializePatternState loads the user's persisted pattern-state file
// from ~/.atheon/pattern_state.json and applies it to the in-memory
// pattern set. It is called automatically by init(); callers may invoke
// it again after programmatically modifying patterns. If the state file
// does not exist the call is a no-op. Errors are non-fatal — they are
// logged to stderr and returned for callers that want to handle them.
func InitializePatternState() error {
	state, err := loadPatternState()
	if err != nil {
		slog.Warn("failed to load pattern state", "err", err)
		return err
	}

	patternMu.Lock()
	defer patternMu.Unlock()
	applyPatternState(state)
	return nil
}
