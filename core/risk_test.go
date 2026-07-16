package core

import (
	"testing"
)

func TestNewRiskScore(t *testing.T) {
	findings := []Finding{
		{Pattern: "p1", Category: "secrets", Severity: "critical"},
		{Pattern: "p2", Category: "secrets", Severity: "high"},
		{Pattern: "p3", Category: "security", Severity: "medium"},
	}
	rs := NewRiskScore(findings)
	if rs.FindingCount != 3 {
		t.Errorf("expected 3 findings, got %d", rs.FindingCount)
	}
	if rs.Score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestNewRiskScore_Empty(t *testing.T) {
	rs := NewRiskScore(nil)
	if rs.Level != RiskLevelNone {
		t.Errorf("expected none, got %s", rs.Level)
	}
	if rs.Score != 0 {
		t.Error("expected 0 score for empty")
	}
}

func TestClassifyRisk(t *testing.T) {
	tests := []struct {
		score    int
		expected RiskLevel
	}{
		{100, RiskLevelCritical},
		{80, RiskLevelCritical},
		{50, RiskLevelHigh},
		{25, RiskLevelMedium},
		{1, RiskLevelLow},
		{0, RiskLevelNone},
	}
	for _, tc := range tests {
		got := ClassifyRisk(tc.score)
		if got != tc.expected {
			t.Errorf("ClassifyRisk(%d) = %s, want %s", tc.score, got, tc.expected)
		}
	}
}

func TestRiskScore_TopRisks(t *testing.T) {
	findings := []Finding{
		{Pattern: "p1", Category: "secrets", Severity: "critical"},
		{Pattern: "p2", Category: "secrets", Severity: "high"},
		{Pattern: "p3", Category: "security", Severity: "high"},
	}
	rs := NewRiskScore(findings)
	top := rs.TopRisks(2)
	if len(top) != 2 {
		t.Errorf("expected 2 top risks, got %d", len(top))
	}
	if top[0].Category != "secrets" {
		t.Errorf("expected secrets as top risk, got %s", top[0].Category)
	}
}
