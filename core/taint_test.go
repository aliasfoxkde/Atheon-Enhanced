package core

import (
	"testing"
)

func TestTaintTracker_NewTaintTracker(t *testing.T) {
	tt := NewTaintTracker()
	if tt == nil {
		t.Fatal("NewTaintTracker returned nil")
	}
	if tt.tainted == nil {
		t.Error("tainted map is nil")
	}
}

func TestTaintTracker_TrackSource(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("user_input")
	if !tt.IsTainted("user_input") {
		t.Error("expected user_input to be tainted")
	}
}

func TestTaintTracker_IsTainted(t *testing.T) {
	tt := NewTaintTracker()
	if tt.IsTainted("unknown") {
		t.Error("unknown variable should not be tainted")
	}
	tt.TrackSource("var1")
	if !tt.IsTainted("var1") {
		t.Error("var1 should be tainted after TrackSource")
	}
}

func TestTaintTracker_ClearTaint(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("var1")
	tt.ClearTaint("var1")
	if tt.IsTainted("var1") {
		t.Error("var1 should not be tainted after ClearTaint")
	}
}

func TestSeverityLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{90, "critical"},
		{50, "critical"},
		{30, "critical"},
		{29, "high"},
		{25, "high"},
		{20, "high"},
		{19, "medium"},
		{15, "medium"},
		{10, "medium"},
		{9, "low"},
		{5, "low"},
		{0, "low"},
	}
	for _, tc := range tests {
		got := SeverityLevel(tc.score)
		if got != tc.expected {
			t.Errorf("SeverityLevel(%d) = %q, want %q", tc.score, got, tc.expected)
		}
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		weight     int
		confidence float64
		min, max   int
	}{
		{40, 1.0, 40, 40},
		{20, 0.5, 10, 10},
		{40, 0.8, 32, 32},
		{100, 1.0, 100, 100}, // capped at 100
	}
	for _, tc := range tests {
		got := CalculateRiskScore(tc.weight, tc.confidence)
		if got < tc.min || got > tc.max {
			t.Errorf("CalculateRiskScore(%d, %f) = %d, want between %d and %d",
				tc.weight, tc.confidence, got, tc.min, tc.max)
		}
	}
}

func TestScanForTaintPatterns(t *testing.T) {
	content := `os.getenv("USER_INPUT") + exec()`
	findings := ScanForTaintPatterns(content)
	if len(findings) == 0 {
		t.Error("expected at least one finding for command injection")
	}
}
