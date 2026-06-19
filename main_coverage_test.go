package main

import (
	"strings"
	"testing"

	"atheon/core"
)

func TestParseCategories(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expected  []string
		remaining int
	}{
		{
			name:      "no categories",
			args:      []string{"scan", "."},
			expected:  nil,
			remaining: 2,
		},
		{
			name:      "single category",
			args:      []string{"--categories=secrets", "scan", "."},
			expected:  []string{"secrets"},
			remaining: 2,
		},
		{
			name:      "multiple categories",
			args:      []string{"--categories=secrets,pii", "scan", "."},
			expected:  []string{"secrets", "pii"},
			remaining: 2,
		},
		{
			name:      "all flag",
			args:      []string{"--all", "scan", "."},
			expected:  nil,
			remaining: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cats, remaining := parseCategories(tt.args)

			// Check categories
			if tt.expected == nil {
				if cats != nil {
					t.Errorf("expected nil categories, got %v", cats)
				}
			} else {
				if len(cats) != len(tt.expected) {
					t.Errorf("expected %d categories, got %d", len(tt.expected), len(cats))
				}
				for i, cat := range tt.expected {
					if cats[i] != cat {
						t.Errorf("expected category %d to be %s, got %s", i, cat, cats[i])
					}
				}
			}

			// Check remaining args
			if len(remaining) != tt.remaining {
				t.Errorf("expected %d remaining args, got %d", tt.remaining, len(remaining))
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	// Test that printHelp doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printHelp panicked: %v", r)
		}
	}()

	printHelp()
}

func TestCmdList(t *testing.T) {
	// Test cmdList doesn't panic with various args
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("cmdList panicked: %v", r)
		}
	}()

	tests := [][]string{
		{},
		{"categories"},
		{"--category=secrets"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("cmdList with args %v panicked: %v", args, r)
				}
			}()
			cmdList(args)
		})
	}
}

func TestRedact(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "short",
			expected: "***",
		},
		{
			input:    "a much longer string that should be redacted",
			expected: "a mu****cted",
		},
		{
			input:    "",
			expected: "***",
		},
		{
			input:    "exactlytwentychars!!",
			expected: "exac****rs!!",
		},
		{
			input:    "12345678",
			expected: "***",
		},
		{
			input:    "123456789",
			expected: "1234****6789",
		},
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

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1.5, "1.5 MB"},
		{1024 * 1024 * 1024, "1024.0 MB"},
		{1024 * 1024 * 1024 * 1.5, "1536.0 MB"},
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

func TestPrintFindings(t *testing.T) {
	// Test that printFindings doesn't panic
	findings := []core.Finding{
		{
			Pattern: "test-pattern",
			File:    "test.txt",
			Line:    1,
			Content: "test content",
		},
	}

	stats := &core.Stats{
		Files:     1,
		Bytes:     1024,
		ElapsedMs: 100,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printFindings panicked: %v", r)
		}
	}()

	printFindings(findings, stats, false)
}

func TestPrintJSONFindings(t *testing.T) {
	// Test JSON output format
	findings := []core.Finding{
		{
			Pattern: "test-pattern",
			File:    "test.txt",
			Line:    1,
			Content: "test content",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printJSONFindings panicked: %v", r)
		}
	}()

	printJSONFindings(findings)
}

func TestPrintFindingsWithNilStats(t *testing.T) {
	// Test handling of nil stats
	findings := []core.Finding{
		{
			Pattern: "test-pattern",
			File:    "test.txt",
			Line:    1,
			Content: "test content",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printFindings with nil stats panicked: %v", r)
		}
	}()

	printFindings(findings, nil, false)
}

func TestPrintFindingsWithEmptyFindings(t *testing.T) {
	// Test handling of empty findings
	findings := []core.Finding{}
	stats := &core.Stats{
		Files:     0,
		Bytes:     0,
		ElapsedMs: 0,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printFindings with empty findings panicked: %v", r)
		}
	}()

	printFindings(findings, stats, false)
}

func TestParseCategoriesEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func([]string) bool
	}{
		{
			name: "empty category value",
			args: []string{"--categories=", "scan", "."},
			check: func(cats []string) bool {
				return len(cats) == 0 // empty strings are trimmed
			},
		},
		{
			name: "category with trailing comma",
			args: []string{"--categories=secrets,", "scan", "."},
			check: func(cats []string) bool {
				return len(cats) == 1 && cats[0] == "secrets"
			},
		},
		{
			name: "multiple commas",
			args: []string{"--categories=secrets,,pii", "scan", "."},
			check: func(cats []string) bool {
				return len(cats) == 2 && cats[0] == "secrets" && cats[1] == "pii"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cats, _ := parseCategories(tt.args)

			if !tt.check(cats) {
				t.Errorf("parseCategories(%v) = %v, expected different result", tt.args, cats)
			}
		})
	}
}

func TestMainIntegration(t *testing.T) {
	// Test main function components
	// Can't test main() directly because it calls os.Exit()

	// Test 1: Help functionality
	t.Run("help", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printHelp panicked: %v", r)
			}
		}()
		printHelp()
	})

	// Test 2: List functionality
	t.Run("list", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("cmdList panicked: %v", r)
			}
		}()
		cmdList([]string{})
	})
}
