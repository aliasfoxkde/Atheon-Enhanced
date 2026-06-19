package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestVersionFlag tests the --version flag
func TestVersionFlag(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Test --version flag
	os.Args = []string{"atheon", "--version"}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a buffered copy of stdout
	out := make(chan string)
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			t.Logf("Failed to capture stdout: %v", err)
		}
		out <- buf.String()
	}()

	// Run main (which should exit with --version)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected to exit, so panic is OK
				t.Logf("Recovered panic (expected): %v", r)
			}
		}()
		main()
	}()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Get output
	output := <-out

	// Check that output contains "atheon" and a version
	if !strings.Contains(output, "atheon") {
		t.Errorf("Version output should contain 'atheon', got: %s", output)
	}

	if !strings.Contains(output, "dev") && !strings.Contains(output, "v") {
		t.Errorf("Version output should contain version number, got: %s", output)
	}

	t.Logf("Version output: %s", strings.TrimSpace(output))
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
