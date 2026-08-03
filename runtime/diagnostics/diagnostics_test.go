package diagnostics

import (
	"testing"
)

func TestNewDiagnostics(t *testing.T) {
	d := NewDiagnostics()
	if d == nil {
		t.Fatal("NewDiagnostics() returned nil")
	}
	if d.Issues == nil {
		t.Error("Issues slice should be initialized, got nil")
	}
	if len(d.Issues) != 0 {
		t.Errorf("Expected empty Issues slice, got length %d", len(d.Issues))
	}
}

func TestAddIssue(t *testing.T) {
	d := NewDiagnostics()

	issue := Issue{
		RuleID:   "TEST001",
		Message:  "Test issue",
		Severity: "high",
		File:     "test.go",
		Line:     10,
		Column:   5,
	}

	AddIssue(d, issue)

	if len(d.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(d.Issues))
	}
	if d.Summary.TotalFindings != 1 {
		t.Errorf("Expected TotalFindings=1, got %d", d.Summary.TotalFindings)
	}
	if d.Summary.High != 1 {
		t.Errorf("Expected High=1, got %d", d.Summary.High)
	}
}

func TestAddIssue_SeverityCounts(t *testing.T) {
	d := NewDiagnostics()

	issues := []Issue{
		{RuleID: "1", Severity: "critical"},
		{RuleID: "2", Severity: "high"},
		{RuleID: "3", Severity: "medium"},
		{RuleID: "4", Severity: "low"},
		{RuleID: "5", Severity: "info"},
		{RuleID: "6", Severity: "info"},
	}

	for _, issue := range issues {
		AddIssue(d, issue)
	}

	if d.Summary.TotalFindings != 6 {
		t.Errorf("Expected TotalFindings=6, got %d", d.Summary.TotalFindings)
	}
	if d.Summary.Critical != 1 {
		t.Errorf("Expected Critical=1, got %d", d.Summary.Critical)
	}
	if d.Summary.High != 1 {
		t.Errorf("Expected High=1, got %d", d.Summary.High)
	}
	if d.Summary.Medium != 1 {
		t.Errorf("Expected Medium=1, got %d", d.Summary.Medium)
	}
	if d.Summary.Low != 1 {
		t.Errorf("Expected Low=1, got %d", d.Summary.Low)
	}
	if d.Summary.Info != 2 {
		t.Errorf("Expected Info=2, got %d", d.Summary.Info)
	}
}

func TestAddIssue_NilDiagnostics(t *testing.T) {
	var d *Diagnostics = nil

	issue := Issue{
		RuleID:   "TEST001",
		Message:  "Test issue",
		Severity: "high",
	}

	// Should not panic, should create a new Diagnostics
	AddIssue(d, issue)
}

func TestMergeDiagnostics_BothNil(t *testing.T) {
	result := MergeDiagnostics(nil, nil)
	if result == nil {
		t.Fatal("MergeDiagnostics(nil, nil) should return new Diagnostics, got nil")
	}
	if len(result.Issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(result.Issues))
	}
}

func TestMergeDiagnostics_LeftNil(t *testing.T) {
	d2 := NewDiagnostics()
	AddIssue(d2, Issue{RuleID: "TEST001", Severity: "high"})

	result := MergeDiagnostics(nil, d2)
	if result == nil {
		t.Fatal("MergeDiagnostics(nil, d2) should return copy of d2, got nil")
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
}

func TestMergeDiagnostics_RightNil(t *testing.T) {
	d1 := NewDiagnostics()
	AddIssue(d1, Issue{RuleID: "TEST001", Severity: "high"})

	result := MergeDiagnostics(d1, nil)
	if result == nil {
		t.Fatal("MergeDiagnostics(d1, nil) should return copy of d1, got nil")
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
}

func TestMergeDiagnostics_BothEmpty(t *testing.T) {
	d1 := NewDiagnostics()
	d2 := NewDiagnostics()

	result := MergeDiagnostics(d1, d2)
	if result == nil {
		t.Fatal("MergeDiagnostics should return non-nil result")
	}
	if len(result.Issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(result.Issues))
	}
}

func TestMergeDiagnostics_BothWithIssues(t *testing.T) {
	d1 := NewDiagnostics()
	AddIssue(d1, Issue{RuleID: "D1_001", Severity: "high"})
	AddIssue(d1, Issue{RuleID: "D1_002", Severity: "critical"})

	d2 := NewDiagnostics()
	AddIssue(d2, Issue{RuleID: "D2_001", Severity: "low"})
	AddIssue(d2, Issue{RuleID: "D2_002", Severity: "medium"})

	result := MergeDiagnostics(d1, d2)

	if len(result.Issues) != 4 {
		t.Errorf("Expected 4 issues, got %d", len(result.Issues))
	}
	if result.Summary.TotalFindings != 4 {
		t.Errorf("Expected TotalFindings=4, got %d", result.Summary.TotalFindings)
	}
	if result.Summary.High != 1 {
		t.Errorf("Expected High=1, got %d", result.Summary.High)
	}
	if result.Summary.Critical != 1 {
		t.Errorf("Expected Critical=1, got %d", result.Summary.Critical)
	}
	if result.Summary.Medium != 1 {
		t.Errorf("Expected Medium=1, got %d", result.Summary.Medium)
	}
	if result.Summary.Low != 1 {
		t.Errorf("Expected Low=1, got %d", result.Summary.Low)
	}
}

func TestMergeDiagnostics_Statistics(t *testing.T) {
	d1 := NewDiagnostics()
	d1.Statistics.FilesScanned = 10
	d1.Statistics.FilesAnalyzed = 8
	d1.Statistics.PatternsMatched = 5
	d1.Statistics.Duration = 100.0

	d2 := NewDiagnostics()
	d2.Statistics.FilesScanned = 15
	d2.Statistics.FilesAnalyzed = 12
	d2.Statistics.PatternsMatched = 7
	d2.Statistics.Duration = 200.0

	result := MergeDiagnostics(d1, d2)

	if result.Statistics.FilesScanned != 25 {
		t.Errorf("Expected FilesScanned=25, got %d", result.Statistics.FilesScanned)
	}
	if result.Statistics.FilesAnalyzed != 20 {
		t.Errorf("Expected FilesAnalyzed=20, got %d", result.Statistics.FilesAnalyzed)
	}
	if result.Statistics.PatternsMatched != 12 {
		t.Errorf("Expected PatternsMatched=12, got %d", result.Statistics.PatternsMatched)
	}
	if result.Statistics.Duration != 300.0 {
		t.Errorf("Expected Duration=300.0, got %f", result.Statistics.Duration)
	}
}

func TestMergeDiagnostics_Timing(t *testing.T) {
	d1 := NewDiagnostics()
	d1.Timing.Start = 1000
	d1.Timing.End = 2000

	d2 := NewDiagnostics()
	d2.Timing.Start = 500
	d2.Timing.End = 2500

	result := MergeDiagnostics(d1, d2)

	// Start should be the earliest
	if result.Timing.Start != 500 {
		t.Errorf("Expected Start=500, got %d", result.Timing.Start)
	}
	// End should be the latest
	if result.Timing.End != 2500 {
		t.Errorf("Expected End=2500, got %d", result.Timing.End)
	}
}

func TestMergeDiagnostics_Metadata(t *testing.T) {
	d1 := NewDiagnostics()
	d1.Metadata.Version = "1.0.0"
	d1.Metadata.ConfigFile = "config1.yaml"

	d2 := NewDiagnostics()
	d2.Metadata.Version = "2.0.0"
	d2.Metadata.ConfigFile = "config2.yaml"

	result := MergeDiagnostics(d1, d2)

	// d1 values should be preferred
	if result.Metadata.Version != "1.0.0" {
		t.Errorf("Expected Version=1.0.0, got %s", result.Metadata.Version)
	}
	if result.Metadata.ConfigFile != "config1.yaml" {
		t.Errorf("Expected ConfigFile=config1.yaml, got %s", result.Metadata.ConfigFile)
	}
}

func TestMergeDiagnostics_Metadata_PartialOverlap(t *testing.T) {
	d1 := NewDiagnostics()
	d1.Metadata.Version = "1.0.0"
	// ConfigFile not set in d1

	d2 := NewDiagnostics()
	d2.Metadata.Version = ""
	d2.Metadata.ConfigFile = "config2.yaml"

	result := MergeDiagnostics(d1, d2)

	// d1 Version should be preferred, d2 ConfigFile should be used
	if result.Metadata.Version != "1.0.0" {
		t.Errorf("Expected Version=1.0.0, got %s", result.Metadata.Version)
	}
	if result.Metadata.ConfigFile != "config2.yaml" {
		t.Errorf("Expected ConfigFile=config2.yaml, got %s", result.Metadata.ConfigFile)
	}
}

func TestCopy(t *testing.T) {
	d := NewDiagnostics()
	AddIssue(d, Issue{RuleID: "TEST001", Severity: "high", Message: "Test"})
	d.Statistics.FilesScanned = 10
	d.Metadata.Version = "1.0.0"

	copy := d.copy()

	if copy == d {
		t.Error("copy should be a different pointer")
	}
	if len(copy.Issues) != len(d.Issues) {
		t.Errorf("Expected %d issues, got %d", len(d.Issues), len(copy.Issues))
	}
	if copy.Statistics.FilesScanned != d.Statistics.FilesScanned {
		t.Errorf("Expected FilesScanned=%d, got %d", d.Statistics.FilesScanned, copy.Statistics.FilesScanned)
	}
	if copy.Metadata.Version != d.Metadata.Version {
		t.Errorf("Expected Version=%s, got %s", d.Metadata.Version, copy.Metadata.Version)
	}

	// Verify it's a deep copy - modifying copy doesn't affect original
	copy.Issues[0].RuleID = "MODIFIED"
	if d.Issues[0].RuleID != "TEST001" {
		t.Error("Modifying copy affected original")
	}
}

func TestCopy_Nil(t *testing.T) {
	var d *Diagnostics = nil
	copy := d.copy()
	if copy != nil {
		t.Error("copy of nil should return nil")
	}
}
