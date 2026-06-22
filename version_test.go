package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestVersionFlag tests the --version flag via subprocess to avoid os.Exit side effects.
func TestVersionFlag(t *testing.T) {
	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Logf("--version flag error: %v", err)
	}

	if !strings.Contains(string(out), "atheon") {
		t.Errorf("Version output should contain 'atheon', got: %s", out)
	}

	if !strings.Contains(string(out), "dev") && !strings.Contains(string(out), "v") {
		t.Errorf("Version output should contain version number, got: %s", out)
	}
}

// TestDevVersion tests that dev version works correctly
func TestDevVersion(t *testing.T) {
	if version != "dev" {
		t.Logf("Note: version is '%s' (expected 'dev' for development builds)", version)
	}

	// Version should not be empty
	if version == "" {
		t.Error("Version should not be empty")
	}
}
