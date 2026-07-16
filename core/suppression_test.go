package core

import (
	"os"
	"testing"
)

func TestNewBaselineMatcher(t *testing.T) {
	bm, err := NewBaselineMatcher("/nonexistent/path")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if bm == nil {
		t.Error("expected non-nil matcher")
	}
}

func TestBaselineMatcher_IsSuppressed(t *testing.T) {
	// Create temp baseline file
	content := `version: "1.0"
findings:
  - pattern_id: "test-pattern"
    file: test.go
    line: 10
`
	tmpfile, err := os.CreateTemp("", "baseline-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	bm, err := NewBaselineMatcher(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be suppressed
	f1 := Finding{Pattern: "test-pattern", File: "test.go", Line: 10}
	if !bm.IsSuppressed(f1) {
		t.Error("expected finding to be suppressed")
	}

	// Should NOT be suppressed
	f2 := Finding{Pattern: "other-pattern", File: "test.go", Line: 10}
	if bm.IsSuppressed(f2) {
		t.Error("expected finding to NOT be suppressed")
	}
}

func TestBaselineMatcher_FilterFindings(t *testing.T) {
	content := `version: "1.0"
findings:
  - pattern_id: "suppressed-1"
    file: test.go
    line: 5
`
	tmpfile, err := os.CreateTemp("", "baseline-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	bm, err := NewBaselineMatcher(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := []Finding{
		{Pattern: "suppressed-1", File: "test.go", Line: 5},
		{Pattern: "active-1", File: "test.go", Line: 10},
	}

	filtered := bm.FilterFindings(findings)
	if len(filtered) != 1 {
		t.Errorf("expected 1 filtered finding, got %d", len(filtered))
	}
	if filtered[0].Pattern != "active-1" {
		t.Error("expected active-1 to remain")
	}
}

func TestBaselineStats(t *testing.T) {
	bm, _ := NewBaselineMatcher("/nonexistent")
	bm.totalCount = 10
	bm.suppressedCount = 3

	stats := bm.Stats()
	if stats.Total != 10 {
		t.Errorf("expected Total=10, got %d", stats.Total)
	}
	if stats.Suppressed != 3 {
		t.Errorf("expected Suppressed=3, got %d", stats.Suppressed)
	}
	if stats.Active != 7 {
		t.Errorf("expected Active=7, got %d", stats.Active)
	}
}
