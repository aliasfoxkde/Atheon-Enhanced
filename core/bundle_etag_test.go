package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestShouldSkipDownloadFresh verifies that a bundle checked within 24 hours
// with a matching ETag is skipped.
func TestShouldSkipDownloadFresh(t *testing.T) {
	// Set up a pattern state with a recent ETag.
	home := t.TempDir()
	state := &PatternState{
		Patterns:          map[string]bool{},
		BundleETag:        `"W/"etag-123""`,
		BundleLastChecked: time.Now().Add(-1 * time.Hour).UnixNano(),
	}
	statePath := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(statePath, "pattern_state.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Swap the state file path for this test.
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	skip, etag := shouldSkipDownload()
	if !skip {
		t.Error("expected skip=true for fresh recent ETag")
	}
	if etag == "" {
		t.Error("expected non-empty etag")
	}
}

// TestShouldSkipDownloadStale verifies that a bundle checked more than
// 24 hours ago is NOT skipped even with a stored ETag.
func TestShouldSkipDownloadStale(t *testing.T) {
	home := t.TempDir()
	state := &PatternState{
		Patterns:          map[string]bool{},
		BundleETag:        `"W/"etag-456""`,
		BundleLastChecked: time.Now().Add(-25 * time.Hour).UnixNano(),
	}
	statePath := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(statePath, "pattern_state.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	skip, _ := shouldSkipDownload()
	if skip {
		t.Error("expected skip=false for stale (>24h) ETag")
	}
}

// TestShouldSkipDownloadNoState verifies that missing ETag returns skip=false.
func TestShouldSkipDownloadNoState(t *testing.T) {
	home := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	skip, _ := shouldSkipDownload()
	if skip {
		t.Error("expected skip=false when no state file exists")
	}
}

// TestShouldSkipDownloadEmptyETag verifies that an empty stored ETag
// returns skip=false (forces a fresh fetch).
func TestShouldSkipDownloadEmptyETag(t *testing.T) {
	home := t.TempDir()
	state := &PatternState{
		Patterns:          map[string]bool{},
		BundleETag:        "",
		BundleLastChecked: time.Now().Add(-1 * time.Hour).UnixNano(),
	}
	statePath := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(statePath, "pattern_state.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	skip, _ := shouldSkipDownload()
	if skip {
		t.Error("expected skip=false for empty ETag")
	}
}

// TestRecordBundleETag verifies that recordBundleETag persists the
// ETag and timestamp to the pattern state file. We call savePatternState
// directly to avoid withFileLock acquiring the state file as the lock
// handle (which truncates it on open), since atomicWriteFile's rename-
// over-open-file behavior varies across platforms.
func TestRecordBundleETag(t *testing.T) {
	home := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	athomeDir := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(athomeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a valid empty state file so loadPatternState can read+merge.
	emptyState := &PatternState{Patterns: map[string]bool{}}
	initialData, err := json.Marshal(emptyState)
	if err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(athomeDir, "pattern_state.json")
	if err := os.WriteFile(statePath, initialData, 0o600); err != nil {
		t.Fatal(err)
	}

	// Now call recordBundleETag which acquires the lock and updates.
	const testETag = `"W/"test-etag-789""`
	if err := recordBundleETag(testETag); err != nil {
		t.Fatalf("recordBundleETag failed: %v", err)
	}

	state, err := loadPatternState()
	if err != nil {
		t.Fatalf("loadPatternState failed: %v", err)
	}
	if state.BundleETag != testETag {
		t.Errorf("expected etag %q, got %q", testETag, state.BundleETag)
	}
	if state.BundleLastChecked == 0 {
		t.Error("expected non-zero BundleLastChecked")
	}
}

// TestLoadBundleETag verifies that loadBundleETag returns the stored
// ETag and last-checked time.
func TestLoadBundleETag(t *testing.T) {
	home := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", home)
	t.Cleanup(func() { t.Setenv("HOME", origHome) })

	// Persist a state with known ETag.
	before := time.Now().Add(-1 * time.Minute).UnixNano()
	state := &PatternState{
		Patterns:          map[string]bool{},
		BundleETag:        `"W/"load-etag-test""`,
		BundleLastChecked: before,
	}
	statePath := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(statePath, "pattern_state.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	etag, lastChecked, err := loadBundleETag()
	if err != nil {
		t.Fatalf("loadBundleETag failed: %v", err)
	}
	if etag != `"W/"load-etag-test""` {
		t.Errorf("expected etag %q, got %q", `"W/"load-etag-test""`, etag)
	}
	if lastChecked.IsZero() {
		t.Error("expected non-zero lastChecked time")
	}
}
