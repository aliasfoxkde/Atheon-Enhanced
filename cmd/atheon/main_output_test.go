package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// TestPrintFindingsFull tests printFindings with real data
func TestPrintFindingsFull(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "test-pattern", File: "test.txt", Line: 1, Content: "test content"},
	}
	stats := &core.Stats{
		Files:     1,
		Bytes:     1024,
		ElapsedMs: 100,
	}

	// Just verify the function doesn't panic
	// Output testing would require proper synchronization
	printFindings(findings, stats, false, false)
}

// TestScanErrorsPresent covers the exit-code-bump helper. A scan that
// silently dropped files (permission denied, unreadable) must register
// as a partial failure even when no findings are reported.
func TestScanErrorsPresent(t *testing.T) {
	cases := []struct {
		name string
		stat *core.Stats
		want bool
	}{
		{"nil stats", nil, false},
		{"empty stats", &core.Stats{}, false},
		{"with errors", &core.Stats{Errors: []error{errors.New("perm denied")}}, true},
		{"multiple errors", &core.Stats{Errors: []error{errors.New("a"), errors.New("b")}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scanErrorsPresent(c.stat); got != c.want {
				t.Errorf("scanErrorsPresent(%+v) = %v, want %v", c.stat, got, c.want)
			}
		})
	}
}

// TestPrintFindingsSurfacesErrors verifies that printFindings writes
// per-file read errors to stderr when present, so silent data loss
// during a scan is observable.
func TestPrintFindingsSurfacesErrors(t *testing.T) {
	findings := []core.Finding{}
	stats := &core.Stats{
		Files:  2,
		Bytes:  100,
		Errors: []error{&os.PathError{Op: "open", Path: "/root/secret.env", Err: os.ErrPermission}},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printFindings(findings, stats, false, false)
	w.Close()
	os.Stderr = oldStderr

	buf, _ := io.ReadAll(r)
	output := string(buf)
	if !strings.Contains(output, "permission denied") {
		t.Errorf("expected stderr to contain sanitized error, got: %q", output)
	}
	if !strings.Contains(output, "1 file(s) could not be read") {
		t.Errorf("expected stderr to surface error count, got: %q", output)
	}
}

// TestPrintJSONFindingsFull tests printJSONFindings
func TestPrintJSONFindingsFull(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "test-pattern", File: "test.txt", Line: 1, Content: "test content"},
	}

	// Just verify the function doesn't panic
	printJSONFindings(findings, nil)
}

// TestCmdListFull tests cmdList with various arguments
func TestCmdListFull(t *testing.T) {
	tests := [][]string{
		{},
		{"categories"},
		{"--enabled"},
		{"--disabled"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			// Just verify the function doesn't panic
			cmdList(args)
		})
	}
}

// TestCmdListUnknownCategory verifies that --category=<bogus> is rejected
// with a clear error and a non-zero exit code, instead of silently
// filtering to zero matches.
func TestCmdListUnknownCategory(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := cmdList([]string{"--category=does-not-exist"})
	w.Close()
	os.Stderr = oldStderr

	buf, _ := io.ReadAll(r)
	if code == 0 {
		t.Errorf("expected non-zero exit code for unknown category, got 0")
	}
	if !strings.Contains(string(buf), "unknown category") {
		t.Errorf("expected stderr to mention 'unknown category', got: %q", buf)
	}
	if !strings.Contains(string(buf), "secrets") {
		t.Errorf("expected stderr to list known categories, got: %q", buf)
	}
}

// TestCmdListKnownCategory verifies that --category=<real> succeeds.
func TestCmdListKnownCategory(t *testing.T) {
	cats := core.Categories()
	if len(cats) == 0 {
		t.Skip("no categories available")
	}
	if code := cmdList([]string{"--category=" + cats[0]}); code != 0 {
		t.Errorf("expected 0 exit code for known category %q, got %d", cats[0], code)
	}
}

// TestPrintHelpFull tests printHelp function
func TestPrintHelpFull(t *testing.T) {
	// Just verify the function doesn't panic
	printHelp()
}

// TestRedactFull tests redact function
func TestRedactFull(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "***"},
		{"a much longer string that should be redacted", "a mu****cted"},
		{"", "***"},
		{"12345678", "***"},
		{"123456789", "1234****6789"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := redact(tt.input)
			if result != tt.expected {
				t.Errorf("redact(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatBytesFull tests formatBytes function
func TestFormatBytesFull(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1536 * 1024, "1.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}
