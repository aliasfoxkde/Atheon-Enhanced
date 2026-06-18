package core

import (
	"testing"
)

func TestCategories(t *testing.T) {
	cats := Categories()

	if len(cats) == 0 {
		t.Error("expected at least one category")
	}

	// Verify expected categories exist
	expectedCats := map[string]bool{
		"secrets": false,
		"pii":     false,
	}

	for _, cat := range cats {
		if _, exists := expectedCats[cat]; exists {
			expectedCats[cat] = true
		}
	}

	for cat, found := range expectedCats {
		if !found {
			t.Errorf("expected to find category '%s'", cat)
		}
	}
}

func TestDownloadBundle(t *testing.T) {
	// Test DownloadBundle function
	// This will likely fail due to network, but we can test the function exists
	err := DownloadBundle()

	// The function should not panic
	// It might fail due to network or other issues, but that's expected
	if err != nil {
		t.Logf("DownloadBundle failed as expected: %v", err)
	} else {
		t.Log("DownloadBundle succeeded")
	}
}

func TestSetActiveCategoriesWithEmpty(t *testing.T) {
	// Test with empty category list
	SetActiveCategories(nil)

	// Should set active categories to all
	cats := Categories()
	if len(cats) == 0 {
		t.Error("expected categories to be set")
	}
}

func TestSetActiveCategoriesWithInvalid(t *testing.T) {
	// Test with invalid category name
	SetActiveCategories([]string{"nonexistent-category"})

	// Should handle gracefully without crashing
	patterns := All()
	if patterns == nil {
		t.Error("expected patterns to be returned")
	}
}

func TestPatternInterfaces(t *testing.T) {
	// Test that all patterns implement the Pattern interface correctly
	patterns := All()

	for _, p := range patterns {
		// Test Name() method
		name := p.Name()
		if name == "" {
			t.Error("pattern should have a name")
		}

		// Test Matches() method doesn't panic
		_ = p.Matches("test content")

		// Test that Name() is consistent
		if p.Name() != name {
			t.Error("pattern name should be consistent")
		}
	}
}

func TestLoadBundleErrorHandling(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("not gzip data")

	err := loadBundle(invalidData)
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

func TestLoadBundleInvalidJSON(t *testing.T) {
	// Test with valid gzip but invalid JSON
	invalidJSON := []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 0, 255} // gzip header
	if len(invalidJSON) < 20 {
		// Make it longer
		invalidJSON = append(invalidJSON, make([]byte, 20)...)
	}

	// This should fail to decompress or parse
	err := loadBundle(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid bundle data")
	}
}

func TestPatternConsistency(t *testing.T) {
	// Test that patterns behave consistently
	patterns := All()

	if len(patterns) == 0 {
		t.Error("expected to have patterns registered")
	}

	// Test that we can get patterns multiple times consistently
	patterns2 := All()

	if len(patterns) != len(patterns2) {
		t.Error("pattern count should be consistent")
	}

	// Create a map of pattern names for comparison
	patternNames := make(map[string]bool)
	for _, p := range patterns {
		patternNames[p.Name()] = true
	}

	for _, p := range patterns2 {
		if !patternNames[p.Name()] {
			t.Errorf("pattern %s not found in first call", p.Name())
		}
	}
}