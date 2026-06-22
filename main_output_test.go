package main

import (
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
	printFindings(findings, stats, false)
}

// TestPrintJSONFindingsFull tests printJSONFindings
func TestPrintJSONFindingsFull(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "test-pattern", File: "test.txt", Line: 1, Content: "test content"},
	}

	// Just verify the function doesn't panic
	printJSONFindings(findings)
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
