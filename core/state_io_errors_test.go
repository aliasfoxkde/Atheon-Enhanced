package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestLoadPatternStateReadError exercises the read-error branch in
// loadPatternState (the path that returns nil, err for non-ENOENT errors).
// Skipped on Windows: os.Chmod for files/dirs behaves differently.
func TestLoadPatternStateReadError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod file permissions are not enforced on Windows")
	}
	tmpDir := t.TempDir()

	// Create a state file we can make unreadable
	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "pattern_state.json")
	if err := os.WriteFile(stateFile, []byte(`{"patterns":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Point HOME at tmpDir so stateFile() resolves here
	t.Setenv("HOME", tmpDir)

	// Make the file unreadable
	if err := os.Chmod(stateFile, 0o000); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(stateFile, 0o644)

	_, err := loadPatternState()
	if err == nil {
		t.Error("expected error from loadPatternState when file unreadable")
	}
}

// TestSavePatternStateWriteError exercises the os.WriteFile error branch
// in savePatternState.
// Skipped on Windows: os.Chmod directory permissions are not enforced there.
func TestSavePatternStateWriteError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod directory permissions are not enforced on Windows")
	}
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Pre-create the .atheon directory and make it read-only
	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(stateDir, 0o755)

	state := &PatternState{Patterns: map[string]bool{"foo": true}}
	err := savePatternState(state)
	if err == nil {
		t.Error("expected error from savePatternState when dir unwritable")
	}
}

// TestDownloadBundleNoChanges exercises the "No pattern changes detected"
// branch inside DownloadBundle.
func TestDownloadBundleNoChanges(t *testing.T) {
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	// Build a bundle identical to what's currently loaded
	current := make([]PatternDef, 0, len(allPatterns))
	for _, p := range allPatterns {
		current = append(current, PatternDef{
			Name:     p.name,
			Category: p.category,
			Match:    p.match,
			Enabled:  p.enabled,
		})
	}
	jb, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	body := gzipBytes(jb)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL)
	defer restore()

	if err := DownloadBundle(context.Background(), false); err != nil {
		t.Fatalf("DownloadBundle failed: %v", err)
	}
}

// TestDownloadBundleWriteFileError exercises the os.WriteFile error branch
// in DownloadBundle by making the ~/.atheon directory read-only.
// Skipped on Windows: os.Chmod is a no-op for directory permissions there.
func TestDownloadBundleWriteFileError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod directory permissions are not enforced on Windows")
	}
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
		// Restore home permissions
		tmpDir := t.TempDir()
		_ = tmpDir
	}()

	// Build a valid bundle
	defs := []PatternDef{
		{Name: "write-err-test", Category: "test", Match: `\bX\b`, Enabled: true},
	}
	body := buildTestBundleBytes(t, defs)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL)
	defer restore()

	// Use a fresh HOME with read-only .atheon
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Pre-create .atheon read-only
	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(stateDir, 0o755)

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Error("expected error from DownloadBundle when .atheon unwritable")
	}
}
