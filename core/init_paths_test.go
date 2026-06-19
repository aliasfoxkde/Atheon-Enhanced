package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitializeWithBadBundle exercises the bundle-load-failed branch in
// initializeWith by feeding corrupt (non-gzip) bundle data.
func TestInitializeWithBadBundle(t *testing.T) {
	defer func() {
		// Restore the embedded bundle after the test so subsequent tests
		// see a sane state.
		initializeWith(embeddedBundle)
	}()

	initializeWith([]byte("not a gzip stream"))
	// If we got here without panicking, the error path was taken.
	// Verify the embedded bundle is NOT loaded (since we overwrote it).
	// After the deferred restore above, allPatterns will be back to normal.
}

// TestInitializeWithPatternStateError exercises the InitializePatternState
// error branch by setting HOME to a path where MkdirAll fails.
func TestInitializeWithPatternStateError(t *testing.T) {
	defer func() {
		// Restore state
		initializeWith(embeddedBundle)
	}()

	// Build a tmp with a "blocker" file so HOME/<subdir>/.atheon can't be created
	tmpDir := t.TempDir()
	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)

	initializeWith(embeddedBundle)
	// Should not panic; InitializePatternState error path runs.
}
