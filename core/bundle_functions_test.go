package core

import (
	"testing"
)

// TestCategories tests the Categories() function
func TestCategories(t *testing.T) {
	cats := Categories()

	if len(cats) == 0 {
		t.Error("Categories() returned empty list")
	}

	// Verify expected categories exist
	expectedCats := map[string]bool{
		"secrets":      false,
		"pii":          false,
		"code-quality": false,
		"healthcare":   false,
		"finance":      false,
	}

	for _, cat := range cats {
		if _, exists := expectedCats[cat]; exists {
			expectedCats[cat] = true
		}
	}

	for cat, found := range expectedCats {
		if !found {
			t.Errorf("Expected category '%s' not found in Categories() output", cat)
		}
	}
}

// TestEnablePattern tests the EnablePattern() function
func TestEnablePattern(t *testing.T) {
	// Find a bundle pattern to test
	var testPattern *bundlePattern
	for _, p := range allPatterns {
		testPattern = p
		break
	}

	if testPattern == nil {
		t.Fatal("No pattern found to test")
	}

	// Disable it first
	testPattern.enabled = false

	// Test EnablePattern
	result := EnablePattern(testPattern.name)
	if !result {
		t.Errorf("EnablePattern(%q) returned false, expected true", testPattern.name)
	}

	if !testPattern.enabled {
		t.Errorf("Pattern %q was not enabled after EnablePattern call", testPattern.name)
	}
}

// TestDisablePattern tests the DisablePattern() function
func TestDisablePattern(t *testing.T) {
	// Find a pattern to test
	var testPattern *bundlePattern
	for _, p := range allPatterns {
		testPattern = p
		break
	}

	if testPattern == nil {
		t.Fatal("No pattern found to test")
	}

	// Enable it first
	testPattern.enabled = true

	// Test DisablePattern
	result := DisablePattern(testPattern.name)
	if !result {
		t.Errorf("DisablePattern(%q) returned false, expected true", testPattern.name)
	}

	if testPattern.enabled {
		t.Errorf("Pattern %q was not disabled after DisablePattern call", testPattern.name)
	}
}

// TestSetPatternEnabled tests the SetPatternEnabled() function
func TestSetPatternEnabled(t *testing.T) {
	var testPattern *bundlePattern
	for _, p := range allPatterns {
		testPattern = p
		break
	}

	if testPattern == nil {
		t.Fatal("No pattern found to test")
	}

	// Test setting to false
	result := SetPatternEnabled(testPattern.name, false)
	if !result {
		t.Errorf("SetPatternEnabled(%q, false) returned false, expected true", testPattern.name)
	}
	if testPattern.enabled {
		t.Errorf("Pattern %q was not disabled after SetPatternEnabled call", testPattern.name)
	}

	// Test setting to true
	result = SetPatternEnabled(testPattern.name, true)
	if !result {
		t.Errorf("SetPatternEnabled(%q, true) returned false, expected true", testPattern.name)
	}
	if !testPattern.enabled {
		t.Errorf("Pattern %q was not enabled after SetPatternEnabled call", testPattern.name)
	}
}

// TestListDisabledPatterns tests the ListDisabledPatterns() function
func TestListDisabledPatterns(t *testing.T) {
	// Disable a known pattern first
	var testPattern *bundlePattern
	for _, p := range allPatterns {
		testPattern = p
		break
	}

	if testPattern != nil {
		originalState := testPattern.enabled
		testPattern.enabled = false
		defer func() { testPattern.enabled = originalState }()
	}

	disabled := ListDisabledPatterns()

	// Should include our disabled pattern
	if testPattern != nil {
		found := false
		for _, name := range disabled {
			if name == testPattern.name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListDisabledPatterns() did not include disabled pattern %q", testPattern.name)
		}
	}
}

// TestListEnabledPatterns tests the ListEnabledPatterns() function
func TestListEnabledPatterns(t *testing.T) {
	// Ensure at least one pattern is enabled
	var testPattern *bundlePattern
	for _, p := range allPatterns {
		if p.enabled {
			testPattern = p
			break
		}
	}

	enabled := ListEnabledPatterns()

	if len(enabled) == 0 {
		t.Error("ListEnabledPatterns() returned empty list")
	}

	// Should include our enabled pattern
	if testPattern != nil {
		found := false
		for _, name := range enabled {
			if name == testPattern.name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListEnabledPatterns() did not include enabled pattern %q", testPattern.name)
		}
	}
}

// TestEnableAllPatterns tests the EnableAllPatterns() function
func TestEnableAllPatterns(t *testing.T) {
	// Disable some patterns first
	for _, p := range allPatterns {
		p.enabled = false
	}

	// Call EnableAllPatterns
	EnableAllPatterns()

	// Verify all patterns are now enabled
	for _, p := range allPatterns {
		if !p.enabled {
			t.Errorf("Pattern %q was not enabled after EnableAllPatterns call", p.name)
		}
	}
}

// TestRebuildActiveScanners tests the rebuildActiveScanners() function
func TestRebuildActiveScanners(t *testing.T) {
	// Save original category filter
	originalFilter := activeCategoryFilter
	defer func() { activeCategoryFilter = originalFilter }()

	// Change category filter
	SetActiveCategories([]string{"secrets"})

	// Call rebuildActiveScanners
	rebuildActiveScanners()

	// Verify activeScanners was rebuilt
	if len(activeScanners) == 0 {
		t.Error("rebuildActiveScanners() resulted in no active scanners")
	}
}

// TestContains tests the contains() helper function
func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"test"}, "test", true},
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
		}
	}
}
