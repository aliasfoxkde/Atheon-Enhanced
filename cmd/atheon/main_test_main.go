package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain sets up test isolation by removing any stale ~/.atheon/patterns.bundle
// before tests run. This prevents DownloadBundle tests in other packages
// (run in parallel via `go test ./...`) from leaving behind a corrupted bundle
// that would cause subsequent tests in this package to fail.
func TestMain(m *testing.M) {
	// Clean up any stale bundle left by parallel tests in other packages.
	if home, err := os.UserHomeDir(); err == nil {
		bundlePath := filepath.Join(home, ".atheon", "patterns.bundle")
		_ = os.Remove(bundlePath)
	}
	os.Exit(m.Run())
}
