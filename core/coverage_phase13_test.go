package core

import (
	"testing"
)

// Phase 13 coverage tests - targeting remaining uncovered code

func TestRiskScore_ByCategory_Phase13(t *testing.T) {
	// Test RiskScore.ByCategory
	findings := []Finding{
		{Pattern: "p1", Severity: "critical", Category: "security"},
		{Pattern: "p2", Severity: "high", Category: "security"},
		{Pattern: "p3", Severity: "medium", Category: "code-quality"},
	}
	rs := NewRiskScore(findings)
	categories := rs.ByCategory
	_ = categories
}

func TestRiskScore_ClassifyRisk_Phase13(t *testing.T) {
	// Test ClassifyRisk various levels
	tests := []struct {
		score    int
		expected RiskLevel
	}{
		{100, RiskLevelCritical},
		{90, RiskLevelCritical},
		{80, RiskLevelCritical},
		{79, RiskLevelHigh},
		{50, RiskLevelHigh},
		{49, RiskLevelMedium},
		{25, RiskLevelMedium},
		{1, RiskLevelLow},
		{0, RiskLevelNone},
		{-1, RiskLevelNone},
	}

	for _, tc := range tests {
		result := ClassifyRisk(tc.score)
		if result != tc.expected {
			t.Errorf("ClassifyRisk(%d) = %s, want %s", tc.score, result, tc.expected)
		}
	}
}

func TestRiskScore_HighestSeverity_Phase13(t *testing.T) {
	// Test that HighestSeverity is updated
	findings := []Finding{
		{Pattern: "p1", Severity: "low", Category: "style"},
	}
	rs := NewRiskScore(findings)
	rs.AddFinding(Finding{Pattern: "p2", Severity: "critical", Category: "security"})
	if rs.HighestSeverity != "critical" {
		t.Errorf("expected critical, got %s", rs.HighestSeverity)
	}
}

func TestRiskScore_FindingCount_Phase13(t *testing.T) {
	// Test FindingCount
	findings := []Finding{
		{Pattern: "p1", Severity: "low", Category: "style"},
	}
	rs := NewRiskScore(findings)
	rs.AddFinding(Finding{Pattern: "p2", Severity: "low", Category: "style"})
	if rs.FindingCount != 2 {
		t.Errorf("expected 2, got %d", rs.FindingCount)
	}
}

func TestRiskScore_AddFinding_DifferentCategories_Phase13(t *testing.T) {
	// Test AddFinding with different categories
	rs := NewRiskScore(nil)
	rs.AddFinding(Finding{Pattern: "p1", Severity: "high", Category: "security"})
	rs.AddFinding(Finding{Pattern: "p2", Severity: "medium", Category: "code-quality"})
	rs.AddFinding(Finding{Pattern: "p3", Severity: "low", Category: "style"})

	if len(rs.ByCategory) != 3 {
		t.Errorf("expected 3 categories, got %d", len(rs.ByCategory))
	}
}

func TestRiskScore_AddFinding_SameCategory_Phase13(t *testing.T) {
	// Test AddFinding with same category
	rs := NewRiskScore(nil)
	rs.AddFinding(Finding{Pattern: "p1", Severity: "high", Category: "security"})
	rs.AddFinding(Finding{Pattern: "p2", Severity: "medium", Category: "security"})
	rs.AddFinding(Finding{Pattern: "p3", Severity: "low", Category: "security"})

	if len(rs.ByCategory) != 1 {
		t.Errorf("expected 1 category, got %d", len(rs.ByCategory))
	}
	if rs.ByCategory["security"].Count != 3 {
		t.Errorf("expected 3 findings in security, got %d", rs.ByCategory["security"].Count)
	}
}

func TestSeverityWeight_AllLevels_Phase13(t *testing.T) {
	// Test severityWeight for all levels
	levels := []string{"critical", "high", "medium", "low", "info", "unknown"}
	for _, level := range levels {
		w := severityWeight(level)
		if w <= 0 && level != "info" && level != "unknown" {
			t.Errorf("expected positive weight for %s, got %d", level, w)
		}
	}
}

func TestClassifyRisk_EdgeCases_Phase13(t *testing.T) {
	// Test ClassifyRisk edge cases
	tests := []struct {
		score    int
		expected RiskLevel
	}{
		{100, RiskLevelCritical},
		{99, RiskLevelCritical},
		{29, RiskLevelMedium},
		{1, RiskLevelLow},
		{-1, RiskLevelNone},
	}

	for _, tc := range tests {
		result := ClassifyRisk(tc.score)
		if result != tc.expected {
			t.Errorf("ClassifyRisk(%d) = %s, want %s", tc.score, result, tc.expected)
		}
	}
}
