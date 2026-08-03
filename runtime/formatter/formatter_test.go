package formatter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

func TestConsoleFormatter(t *testing.T) {
	f := NewConsoleFormatter()

	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "test-rule",
		Message:  "Test finding message",
		Severity: "high",
		File:     "test.go",
		Line:     10,
		Column:   5,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if f.ContentType() != "text/plain" {
		t.Errorf("ContentType() = %s, want text/plain", f.ContentType())
	}

	if f.FileExtension() != ".txt" {
		t.Errorf("FileExtension() = %s, want .txt", f.FileExtension())
	}

	if !strings.Contains(string(output), "HIGH") {
		t.Errorf("Output should contain HIGH severity badge")
	}
	if !strings.Contains(string(output), "test-rule") {
		t.Errorf("Output should contain rule ID")
	}
	if !strings.Contains(string(output), "test.go:10") {
		t.Errorf("Output should contain file location")
	}
}

func TestConsoleFormatterNil(t *testing.T) {
	f := NewConsoleFormatter()

	output, err := f.Format(nil)
	if err != nil {
		t.Fatalf("Format failed for nil diagnostics: %v", err)
	}
	if !strings.Contains(string(output), "no findings") {
		t.Errorf("Output should contain 'no findings' for nil input")
	}
}

func TestJSONFormatter(t *testing.T) {
	f := NewJSONFormatter()

	diag := diagnostics.NewDiagnostics()
	diag.Statistics.FilesScanned = 5
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "json-test-rule",
		Message:  "JSON test finding",
		Severity: "medium",
		File:     "example.go",
		Line:     20,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if f.ContentType() != "application/json" {
		t.Errorf("ContentType() = %s, want application/json", f.ContentType())
	}

	if f.FileExtension() != ".json" {
		t.Errorf("FileExtension() = %s, want .json", f.FileExtension())
	}

	// Verify valid JSON
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	issues, ok := result["issues"].([]any)
	if !ok {
		t.Fatalf("issues field is not an array")
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
}

func TestJSONFormatterNil(t *testing.T) {
	f := NewJSONFormatter()

	output, err := f.Format(nil)
	if err != nil {
		t.Fatalf("Format failed for nil diagnostics: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	issues := result["issues"].([]any)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for nil input, got %d", len(issues))
	}
}

func TestSARIFFormatter(t *testing.T) {
	f := NewSARIFFormatter()

	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "sarif-test-rule",
		Message:  "SARIF test finding",
		Severity: "critical",
		File:     "app.go",
		Line:     42,
		Column:   10,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if f.ContentType() != "application/sarif+json" {
		t.Errorf("ContentType() = %s, want application/sarif+json", f.ContentType())
	}

	if f.FileExtension() != ".sarif" {
		t.Errorf("FileExtension() = %s, want .sarif", f.FileExtension())
	}

	// Verify valid JSON and SARIF structure
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if result["$schema"] == nil {
		t.Error("SARIF should have $schema field")
	}
	if result["version"] != "2.1.0" {
		t.Errorf("version = %v, want 2.1.0", result["version"])
	}

	runs, ok := result["runs"]. ([]any)
	if !ok || len(runs) == 0 {
		t.Fatal("SARIF should have at least one run")
	}
}

func TestSARIFFormatterNil(t *testing.T) {
	f := NewSARIFFormatter()

	output, err := f.Format(nil)
	if err != nil {
		t.Fatalf("Format failed for nil diagnostics: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for nil input, got %d", len(results))
	}
}

func TestMarkdownFormatter(t *testing.T) {
	f := NewMarkdownFormatter()

	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 1
	diag.Statistics.FilesScanned = 5
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "md-test-rule",
		Message:  "Markdown test finding with long message that should be truncated",
		Severity: "low",
		File:     "readme.md",
		Line:     5,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if f.ContentType() != "text/markdown" {
		t.Errorf("ContentType() = %s, want text/markdown", f.ContentType())
	}

	if f.FileExtension() != ".md" {
		t.Errorf("FileExtension() = %s, want .md", f.FileExtension())
	}

	content := string(output)
	if !strings.Contains(content, "# Diagnostics Report") {
		t.Error("Output should contain header")
	}
	if !strings.Contains(content, "## Findings") {
		t.Error("Output should contain Findings section")
	}
	if !strings.Contains(content, "| md-test-rule |") {
		t.Error("Output should contain rule ID in table")
	}
}

func TestMarkdownFormatterNil(t *testing.T) {
	f := NewMarkdownFormatter()

	output, err := f.Format(nil)
	if err != nil {
		t.Fatalf("Format failed for nil diagnostics: %v", err)
	}
	if !strings.Contains(string(output), "No findings") {
		t.Error("Output should indicate no findings for nil input")
	}
}

func TestMarkdownFormatterNoFindings(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "No findings to display") {
		t.Error("Output should indicate no findings")
	}
}

func TestSeverityMappings(t *testing.T) {
	tests := []struct {
		severity  string
		wantLevel string
		wantScore string
	}{
		{"critical", "error", "9.5"},
		{"high", "error", "7.5"},
		{"medium", "warning", "5.0"},
		{"low", "note", "2.5"},
		{"info", "none", "5.0"},
	}

	for _, tt := range tests {
		if got := sarifLevel(tt.severity); got != tt.wantLevel {
			t.Errorf("sarifLevel(%s) = %s, want %s", tt.severity, got, tt.wantLevel)
		}
		if got := sarifSeverityScore(tt.severity); got != tt.wantScore {
			t.Errorf("sarifSeverityScore(%s) = %s, want %s", tt.severity, got, tt.wantScore)
		}
	}
}
