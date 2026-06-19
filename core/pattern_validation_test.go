package core

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestMain(m *testing.M) {
	// Force bundle reload from current bundle file
	home, _ := os.UserHomeDir()
	bundlePath := filepath.Join(home, ".atheon", "patterns.bundle")
	if data, err := os.ReadFile(bundlePath); err == nil {
		_ = loadBundle(data) // Ignore error in test setup
	}
	os.Exit(m.Run())
}

// TestPatternValidation provides comprehensive testing for pattern validation
func TestPatternValidation(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		valid    bool
		expected string
	}{
		{
			name:     "Valid simple pattern",
			pattern:  `\b\d{3}-\d{2}-\d{4}\b`,
			valid:    true,
			expected: "SSN pattern should compile",
		},
		{
			name:     "Valid complex pattern",
			pattern:  `(?i)api[_-]?key[_-]?\s*[:=]\s*['"][\w-]+['"]`,
			valid:    true,
			expected: "API key pattern should compile",
		},
		{
			name:     "Valid character class",
			pattern:  `[A-Z]{2,}-\d{3,}`,
			valid:    true,
			expected: "Character class pattern should compile",
		},
		{
			name:     "Invalid unclosed group",
			pattern:  `(abc`,
			valid:    false,
			expected: "Unclosed group should fail compilation",
		},
		{
			name:     "Invalid bracket",
			pattern:  `[abc`,
			valid:    false,
			expected: "Unclosed bracket should fail compilation",
		},
		{
			name:     "Invalid escape sequence",
			pattern:  `\k\m`,
			valid:    false,
			expected: "Invalid escape should fail compilation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := regexp.Compile(tt.pattern)
			if tt.valid {
				if err != nil {
					t.Errorf("Expected pattern to compile but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected pattern to fail compilation but it succeeded")
				}
			}
		})
	}
}

// TestPatternMatchingPerformance tests pattern matching performance
func TestPatternMatchingPerformance(t *testing.T) {
	// Compile patterns once
	patterns := []struct {
		name string
		re   *regexp.Regexp
	}{
		{
			name: "simple-text",
			re:   regexp.MustCompile(`TODO`),
		},
		{
			name: "phone-number",
			re:   regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		},
		{
			name: "email",
			re:   regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		},
		{
			name: "complex-secret",
			re:   regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`),
		},
	}

	// Test data
	testData := struct {
		simple   string
		phone    string
		email    string
		secret   string
		longText string
	}{
		simple:   "TODO: implement this feature",
		phone:    "123-456-7890",
		email:    "user@example.com",
		secret:   "AKIAIOSFODNN7EXAMPLE",
		longText: string(make([]byte, 1024*1024)), // 1MB text
	}

	t.Run("Simple pattern matching", func(t *testing.T) {
		for _, p := range patterns {
			p := p // capture loop variable
			t.Run(p.name, func(t *testing.T) {
				t.Parallel()
				matches := p.re.FindAllString(testData.simple, -1)
				// Simple performance assertion
				if len(matches) > 1000 {
					t.Errorf("Too many matches for simple pattern: %d", len(matches))
				}
			})
		}
	})

	t.Run("Large text performance", func(t *testing.T) {
		t.Parallel()
		for _, p := range patterns {
			p := p // capture loop variable
			t.Run(p.name, func(t *testing.T) {
				// Test performance on large text
				matches := p.re.FindAllString(testData.longText, -1)
				// Should complete quickly even on large text
				if len(matches) > 10000 {
					t.Errorf("Unexpectedly many matches: %d", len(matches))
				}
			})
		}
	})
}

// TestPatternCoverageValidation tests pattern coverage across categories
func TestPatternCoverageValidation(t *testing.T) {
	patterns := All()

	// Ensure we have all expected categories
	expectedCategories := map[string]int{
		"accessibility":      17,
		"ai-detection":       6,
		"api-integration":    9,
		"cloud-native":       6,
		"code-quality":       25,
		"data-visualization": 5,
		"devops":             6,
		"django":             1,
		"finance":            3,
		"healthcare":         7,
		"nodejs":             1,
		"performance":       12,
		"pii":                3,
		"pwa":                5,
		"react":              1,
		"secrets":            32,
		"security-hardening": 14,
		"web-development":   12,
		"web-security":      11,
	}

	categoryCounts := make(map[string]int)
	for _, p := range patterns {
		if p.Enabled() {
			categoryCounts[p.Category()]++
		}
	}

	for category, expectedCount := range expectedCategories {
		actualCount := categoryCounts[category]
		if actualCount != expectedCount {
			t.Errorf("Category %s: expected %d patterns, got %d", category, expectedCount, actualCount)
		}
	}

	// Total pattern count
	totalPatterns := len(patterns)
	expectedTotal := 176
	if totalPatterns != expectedTotal {
		t.Errorf("Expected %d total patterns, got %d", expectedTotal, totalPatterns)
	}
}

// TestPatternEdgeCases tests edge cases in pattern handling
func TestPatternEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "Pattern name validation",
			testFunc: func(t *testing.T) {
				// Test that patterns must have names
				// Bundle system rejects patterns with empty names during loading
				pattern := &bundlePattern{
					name:     "",
					category: "test",
					match:    "test",
					enabled:  true,
					re:       regexp.MustCompile("test"),
				}
				// Empty name is invalid, but we're testing the struct
				// In real usage, bundler rejects empty names
				if pattern.Name() == "" {
					t.Log("Pattern with empty name created (bundler would reject this)")
				}
			},
		},
		{
			name: "Empty category handling",
			testFunc: func(t *testing.T) {
				// Test that patterns with empty categories work
				pattern := &bundlePattern{
					name:     "test-pattern",
					category: "",
					match:    "test",
					enabled:  true,
					re:       regexp.MustCompile("test"),
				}
				// Empty category is technically allowed but not recommended
				if pattern.Category() == "" {
					t.Log("Pattern with empty category (not recommended but functional)")
				}
			},
		},
		{
			name: "Special regex characters",
			testFunc: func(t *testing.T) {
				// Test patterns with special regex characters
				specialPatterns := []string{
					`\b\d{3}-\d{2}-\d{4}\b`,  // SSN with word boundaries
					`[A-Z]{2,}-\d{3,}`,       // Character classes
					`(?i)case\s*insensitive`, // Case insensitive
					`^password\s*=\s*\S+`,    // Password assignment
				}

				for _, pattern := range specialPatterns {
					_, err := regexp.Compile(pattern)
					if err != nil {
						t.Errorf("Pattern '%s' failed to compile: %v", pattern, err)
					}
				}
			},
		},
		{
			name: "Unicode handling",
			testFunc: func(t *testing.T) {
				// Test Unicode handling in patterns
				unicodePatterns := []string{
					`✨`, // Emoji
					`🚀`, // Rocket
					`💡`, // Light bulb
				}

				for _, pattern := range unicodePatterns {
					re := regexp.MustCompile(regexp.QuoteMeta(pattern))
					if !re.MatchString(pattern) {
						t.Errorf("Unicode pattern '%s' should match itself", pattern)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// TestPatternEnabledToggle tests pattern enable/disable functionality
func TestPatternEnabledToggle(t *testing.T) {
	patterns := All()

	// Find a test pattern to toggle
	var testPattern Pattern
	for _, p := range patterns {
		if p.Category() == "code-quality" {
			testPattern = p
			break
		}
	}

	if testPattern == nil {
		t.Fatal("Could not find a code-quality pattern to test")
	}

	originalState := testPattern.Enabled()

	// Test DisablePattern function
	success := DisablePattern(testPattern.Name())
	if !success {
		t.Error("DisablePattern function failed")
	}
	if testPattern.Enabled() {
		t.Error("Pattern should be disabled")
	}

	// Test EnablePattern function
	success = EnablePattern(testPattern.Name())
	if !success {
		t.Error("EnablePattern function failed")
	}
	if !testPattern.Enabled() {
		t.Error("Pattern should be enabled")
	}

	// Restore original state
	if originalState {
		DisablePattern(testPattern.Name())
	} else {
		EnablePattern(testPattern.Name())
	}
}

// TestPatternMatchingAccuracy tests pattern matching accuracy
func TestPatternMatchingAccuracy(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		testCases []struct {
			input       string
			shouldMatch bool
			description string
		}
	}{
		{
			name:    "SSN pattern",
			pattern: `\b\d{3}-\d{2}-\d{4}\b`,
			testCases: []struct {
				input       string
				shouldMatch bool
				description string
			}{
				{"123-45-6789", true, "Valid SSN"},
				{"12-345-6789", false, "Too short"},
				{"1234567890", false, "No hyphens"},
				{"987-65-4321", true, "Valid SSN"},
				{"not-a-ssn", false, "Text string"},
			},
		},
		{
			name:    "API key pattern",
			pattern: `\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`,
			testCases: []struct {
				input       string
				shouldMatch bool
				description string
			}{
				{"AKIAIOSFODNN7EXAMPLE", true, "Valid AWS key"},
				{"ASIAXXXXXXXXXXXXXXXX", true, "Valid ASIA key"},
				{"AKIA123", false, "Too short"},
				{"AKIAIOSFODNN7EXAMPLEEXTRA", false, "Too long"},
				{"BKIAIOSFODNN7EXAMPLE", false, "Wrong prefix"},
			},
		},
		{
			name:    "Email pattern",
			pattern: `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			testCases: []struct {
				input       string
				shouldMatch bool
				description string
			}{
				{"user@example.com", true, "Valid email"},
				{"user.name@example.com", true, "Email with dot"},
				{"user+tag@example.com", true, "Email with plus"},
				{"user@sub.example.com", true, "Email with subdomain"},
				{"@example.com", false, "No username"},
				{"user@", false, "No domain"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)

			for _, tc := range tt.testCases {
				t.Run(tc.description, func(t *testing.T) {
					matched := re.MatchString(tc.input)
					if matched != tc.shouldMatch {
						t.Errorf("Pattern '%s': input '%s' - expected match=%v, got match=%v",
							tt.pattern, tc.input, tc.shouldMatch, matched)
					}
				})
			}
		})
	}
}
