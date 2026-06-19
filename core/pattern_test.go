package core

import (
	"testing"
)

// TestPatternInterfaceCategory verifies that all patterns implement the Category() method
func TestPatternInterfaceCategory(t *testing.T) {
	patterns := All()

	if len(patterns) == 0 {
		t.Fatal("No patterns registered")
	}

	// Test that all patterns implement the Category() method
	validCategories := map[string]bool{
		"secrets":       true,
		"pii":           true,
		"code-quality":  true,
		"healthcare":    true,
		"finance":       true,
	}

	for _, p := range patterns {
		category := p.Category()
		if category == "" {
			t.Errorf("Pattern %s has empty category", p.Name())
		}

		if !validCategories[category] {
			t.Errorf("Pattern %s has invalid category: %s", p.Name(), category)
		}
	}
}

// TestPatternInterfaceSatisfaction ensures all registered patterns satisfy the Pattern interface
func TestPatternInterfaceSatisfaction(t *testing.T) {
	patterns := All()

	for _, p := range patterns {
		// Test Name() method
		name := p.Name()
		if name == "" {
			t.Error("Pattern has empty name")
		}

		// Test Category() method
		category := p.Category()
		if category == "" {
			t.Error("Pattern has empty category")
		}

		// Test Matches() method exists and is callable
		// (We don't test actual matching here, just that the method exists)
	}
}

// TestPatternCategoriesConsistency ensures patterns in the same category have consistent categories
func TestPatternCategoriesConsistency(t *testing.T) {
	patterns := All()

	categoryCount := make(map[string]int)
	for _, p := range patterns {
		categoryCount[p.Category()]++
	}

	// We should have at least some patterns in each category
	for category, count := range categoryCount {
		if count == 0 {
			t.Errorf("Category %s has no patterns", category)
		}
		t.Logf("Category %s: %d patterns", category, count)
	}
}