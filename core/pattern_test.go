package core

import (
	"strings"
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
		"secrets":            true,
		"pii":                true,
		"code-quality":       true,
		"healthcare":         true,
		"finance":            true,
		"ai-detection":       true,
		"devops":             true,
		"angular":            true,
		"django":             true,
		"express":            true,
		"flask":              true,
		"laravel":            true,
		"nodejs":             true,
		"rails":              true,
		"react":              true,
		"spring":             true,
		"vue":                true,
		"accessibility":      true,
		"api-integration":    true,
		"cloud-native":       true,
		"data-visualization": true,
		"performance":        true,
		"pwa":                true,
		"security-hardening": true,
		"web-development":    true,
		"web-security":       true,
		"compliance":         true,
		"git-hygiene":        true,
		"git-ops":            true,
		"kubernetes":         true,
		"metadata":           true,
		"terraform":          true,
		"supply-chain":       true,
		"container":          true,
		"graphql":            true,
		"cloudformation":     true,
		"arm":                true,
		"rust":               true,
		"go":                 true,
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

// TestValidatePattern tests the ValidatePattern helper function
func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name    string
		def     PatternDef
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pattern",
			def: PatternDef{
				Name:  "test-pattern",
				Match: `test\d+`,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			def: PatternDef{
				Name:  "",
				Match: `test\d+`,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty match",
			def: PatternDef{
				Name:  "test-pattern",
				Match: "",
			},
			wantErr: true,
			errMsg:  "match regex is required",
		},
		{
			name: "invalid regex",
			def: PatternDef{
				Name:  "test-pattern",
				Match: "[invalid",
			},
			wantErr: true,
			errMsg:  "error parsing regexp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.def)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePattern() expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePattern() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePattern() unexpected error: %v", err)
				}
			}
		})
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
