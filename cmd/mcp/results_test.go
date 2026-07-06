package main

import (
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// TestPatternsResult tests the patternsResult helper function
func TestPatternsResult(t *testing.T) {
	// Create a fake pattern for testing
	fp := &fakePatternForTest{
		name:     "test-pattern",
		category: "test",
		severity: "medium",
		enabled:  true,
	}

	// Test with patterns
	result := patternsResult([]core.Pattern{fp}, "")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	content := result["content"].([]map[string]interface{})
	if len(content) == 0 {
		t.Fatal("expected content array")
	}
	text := content[0]["text"].(string)
	if text == "" {
		t.Fatal("expected non-empty text")
	}

	// Test with category filter that matches
	result = patternsResult([]core.Pattern{fp}, "test")
	content = result["content"].([]map[string]interface{})
	text = content[0]["text"].(string)
	if text == "" {
		t.Fatal("expected non-empty text")
	}

	// Test with category filter that doesn't match
	result = patternsResult([]core.Pattern{fp}, "nonexistent")
	content = result["content"].([]map[string]interface{})
	text = content[0]["text"].(string)
	if text == "" {
		t.Fatal("expected non-empty text")
	}
}

// TestCategoriesResult tests the categoriesResult helper function
func TestCategoriesResult(t *testing.T) {
	// Test with categories
	result := categoriesResult([]string{"secrets", "pii", "code-quality"})
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	content := result["content"].([]map[string]interface{})
	if len(content) == 0 {
		t.Fatal("expected content array")
	}
	text := content[0]["text"].(string)
	if text == "" {
		t.Fatal("expected non-empty text")
	}

	// Test with empty categories
	result = categoriesResult([]string{})
	content = result["content"].([]map[string]interface{})
	text = content[0]["text"].(string)
	if text == "" {
		t.Fatal("expected non-empty text")
	}
}

// fakePatternForTest implements core.Pattern for testing
type fakePatternForTest struct {
	name     string
	category string
	severity string
	enabled  bool
}

func (f *fakePatternForTest) Name() string             { return f.name }
func (f *fakePatternForTest) Category() string         { return f.category }
func (f *fakePatternForTest) Enabled() bool           { return f.enabled }
func (f *fakePatternForTest) Severity() string        { return f.severity }
func (f *fakePatternForTest) Confidence() string      { return "medium" }
func (f *fakePatternForTest) Matches(line string) bool { return false }
func (f *fakePatternForTest) Description() string     { return "" }
func (f *fakePatternForTest) Reference() string       { return "" }
func (f *fakePatternForTest) Tags() []string         { return nil }
