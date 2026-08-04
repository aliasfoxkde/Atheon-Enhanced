package formatter

import (
	"encoding/json"
	"fmt"
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

	runs, ok := result["runs"].([]any)
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
		{"unknown", "none", "5.0"},
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

func TestConsoleFormatterAllSeverities(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"critical", "CRITICAL"},
		{"high", "HIGH"},
		{"medium", "MEDIUM"},
		{"low", "LOW"},
		{"info", "INFO"},
		{"unknown", "UNKNOWN"},
	}

	for _, tt := range tests {
		f := NewConsoleFormatter()
		diag := diagnostics.NewDiagnostics()
		diagnostics.AddIssue(diag, diagnostics.Issue{
			RuleID:   "test-rule",
			Message:  "Test message",
			Severity: tt.severity,
			File:     "test.go",
			Line:     1,
		})

		output, err := f.Format(diag)
		if err != nil {
			t.Fatalf("Format failed for severity %s: %v", tt.severity, err)
		}
		if !strings.Contains(string(output), tt.want) {
			t.Errorf("Output should contain %s for severity %s", tt.want, tt.severity)
		}
	}
}

func TestConsoleFormatterEmptyIssues(t *testing.T) {
	f := NewConsoleFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 0

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "no findings") {
		t.Error("Output should contain 'no findings' for empty issues")
	}
}

func TestConsoleFormatterIssueWithoutLine(t *testing.T) {
	f := NewConsoleFormatter()
	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "rule-no-line",
		Message:  "Issue without line number",
		Severity: "high",
		File:     "test.go",
		Line:     0,
		Column:   0,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "test.go") {
		t.Error("Output should contain file without line number")
	}
}

func TestConsoleFormatterIssueWithFix(t *testing.T) {
	f := NewConsoleFormatter()
	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "rule-with-fix",
		Message:  "Issue with fix",
		Severity: "high",
		File:     "test.go",
		Line:     10,
		Fix:      "Replace with proper code",
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "fix: Replace with proper code") {
		t.Error("Output should contain fix suggestion")
	}
}

func TestConsoleFormatterStatistics(t *testing.T) {
	f := NewConsoleFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Statistics.FilesScanned = 42
	diag.Statistics.Duration = 123.456

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "scanned 42 file(s)") {
		t.Error("Output should contain files scanned")
	}
	if !strings.Contains(string(output), "duration: 123.46ms") {
		t.Error("Output should contain duration")
	}
}

func TestConsoleFormatterSeverityBreakdown(t *testing.T) {
	f := NewConsoleFormatter()
	diag := diagnostics.NewDiagnostics()

	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r1", Message: "m", Severity: "critical", File: "f.go"})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r2", Message: "m", Severity: "high", File: "f.go"})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r3", Message: "m", Severity: "medium", File: "f.go"})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r4", Message: "m", Severity: "low", File: "f.go"})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r5", Message: "m", Severity: "info", File: "f.go"})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(output), "severity breakdown:") {
		t.Error("Output should contain severity breakdown header")
	}
	if !strings.Contains(string(output), "critical: 1") {
		t.Error("Output should contain critical count")
	}
	if !strings.Contains(string(output), "high: 1") {
		t.Error("Output should contain high count")
	}
	if !strings.Contains(string(output), "medium: 1") {
		t.Error("Output should contain medium count")
	}
	if !strings.Contains(string(output), "low: 1") {
		t.Error("Output should contain low count")
	}
	if !strings.Contains(string(output), "info: 1") {
		t.Error("Output should contain info count")
	}
}

func TestJSONFormatterEmptyIssues(t *testing.T) {
	f := NewJSONFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Statistics.FilesScanned = 10

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	issues := result["issues"].([]any)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestJSONFormatterAllSeverities(t *testing.T) {
	f := NewJSONFormatter()

	for _, sev := range []string{"critical", "high", "medium", "low", "info", "unknown"} {
		diag := diagnostics.NewDiagnostics()
		diagnostics.AddIssue(diag, diagnostics.Issue{
			RuleID:   "rule-" + sev,
			Message:  "Test message",
			Severity: sev,
			File:     "test.go",
			Line:     1,
		})

		output, err := f.Format(diag)
		if err != nil {
			t.Fatalf("Format failed for severity %s: %v", sev, err)
		}

		var result map[string]any
		if err := json.Unmarshal(output, &result); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}
	}
}

func TestSARIFFormatterSetVersion(t *testing.T) {
	f := NewSARIFFormatter()
	f.SetVersion("1.2.3")

	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "test-rule",
		Message:  "Test finding",
		Severity: "high",
		File:     "test.go",
		Line:     1,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	if driver["version"] != "1.2.3" {
		t.Errorf("version = %v, want 1.2.3", driver["version"])
	}
}

func TestSARIFFormatterMultipleIssuesSameRule(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()

	// Add multiple issues with the same rule ID - should deduplicate
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "duplicate-rule",
		Message:  "First issue",
		Severity: "high",
		File:     "file1.go",
		Line:     10,
	})
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "duplicate-rule",
		Message:  "Second issue",
		Severity: "high",
		File:     "file2.go",
		Line:     20,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	rules := driver["rules"].([]any)

	// Should only have 1 rule (deduplicated)
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule (deduplicated), got %d", len(rules))
	}
}

func TestSARIFFormatterIssueWithoutColumn(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "no-column-rule",
		Message:  "Issue without column",
		Severity: "medium",
		File:     "test.go",
		Line:     5,
		Column:   0,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result0 := results[0].(map[string]any)
	locations := result0["locations"].([]any)
	loc0 := locations[0].(map[string]any)
	physLoc := loc0["physicalLocation"].(map[string]any)
	region := physLoc["region"].(map[string]any)

	if _, hasColumn := region["startColumn"]; hasColumn {
		t.Error("Region should not have startColumn when column is 0")
	}
}

func TestSARIFFormatterIssueWithFix(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "fix-rule",
		Message:  "Issue with fix",
		Severity: "low",
		File:     "test.go",
		Line:     1,
		Fix:      "Apply this fix",
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result0 := results[0].(map[string]any)

	props := result0["properties"].(map[string]any)
	fix := props["fix"].(map[string]any)
	if fix["description"] != "Apply this fix" {
		t.Errorf("fix description = %v, want 'Apply this fix'", fix["description"])
	}
}

func TestSARIFFormatterAllSeverities(t *testing.T) {
	for _, sev := range []string{"critical", "high", "medium", "low", "info", "unknown"} {
		f := NewSARIFFormatter()
		diag := diagnostics.NewDiagnostics()
		diagnostics.AddIssue(diag, diagnostics.Issue{
			RuleID:   "test-rule-" + sev,
			Message:  "Test message",
			Severity: sev,
			File:     "test.go",
			Line:     1,
		})

		output, err := f.Format(diag)
		if err != nil {
			t.Fatalf("Format failed for severity %s: %v", sev, err)
		}

		var result map[string]any
		if err := json.Unmarshal(output, &result); err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}
	}
}

func TestMarkdownFormatterVersion(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Metadata.Version = "1.2.3"

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "**Version:** 1.2.3") {
		t.Error("Output should contain version")
	}
}

func TestMarkdownFormatterTiming(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Timing.Start = 1000
	diag.Timing.End = 2000

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "## Timing") {
		t.Error("Output should contain Timing section")
	}
	if !strings.Contains(string(output), "Start | 1000ms") {
		t.Error("Output should contain start time")
	}
	if !strings.Contains(string(output), "End | 2000ms") {
		t.Error("Output should contain end time")
	}
}

func TestMarkdownFormatterAllSeverityBreakdown(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()

	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r1", Message: "m", Severity: "critical", File: "f.go", Line: 1})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r2", Message: "m", Severity: "high", File: "f.go", Line: 2})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r3", Message: "m", Severity: "medium", File: "f.go", Line: 3})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r4", Message: "m", Severity: "low", File: "f.go", Line: 4})
	diagnostics.AddIssue(diag, diagnostics.Issue{RuleID: "r5", Message: "m", Severity: "info", File: "f.go", Line: 5})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(output), "## Severity Breakdown") {
		t.Error("Output should contain Severity Breakdown section")
	}
	if !strings.Contains(string(output), "| Critical | 1 |") {
		t.Error("Output should contain Critical row")
	}
	if !strings.Contains(string(output), "| High | 1 |") {
		t.Error("Output should contain High row")
	}
	if !strings.Contains(string(output), "| Medium | 1 |") {
		t.Error("Output should contain Medium row")
	}
	if !strings.Contains(string(output), "| Low | 1 |") {
		t.Error("Output should contain Low row")
	}
	if !strings.Contains(string(output), "| Info | 1 |") {
		t.Error("Output should contain Info row")
	}
}

func TestMarkdownFormatterEmptyIssues(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 0

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "No findings to display") {
		t.Error("Output should indicate no findings")
	}
}

func TestMarkdownFormatterLongMessage(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 1
	longMsg := "This is a very long message that exceeds sixty characters and should be truncated in the output"
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "long-msg-rule",
		Message:  longMsg,
		Severity: "medium",
		File:     "test.go",
		Line:     1,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	// Long messages should be truncated with "..."
	if !strings.Contains(string(output), "...") {
		t.Error("Output should contain truncated message indicator")
	}
}

func TestMarkdownFormatterSpecialCharacters(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 1
	// Message with markdown special characters
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "pipe|rule",
		Message:  "Message with | pipe and\nnewline",
		Severity: "high",
		File:     "file|with?special.go",
		Line:     1,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	// Should not panic and should contain escaped content
	if !strings.Contains(string(output), "\\|") {
		t.Error("Output should escape pipe characters")
	}
}

func TestMarkdownFormatterEmptySeverity(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 1
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "no-severity-rule",
		Message:  "Issue without severity",
		Severity: "",
		File:     "test.go",
		Line:     1,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "| unknown |") {
		t.Error("Output should show 'unknown' for empty severity")
	}
}

func TestMarkdownFormatterIssueWithoutLine(t *testing.T) {
	f := NewMarkdownFormatter()
	diag := diagnostics.NewDiagnostics()
	diag.Summary.TotalFindings = 1
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "no-line-rule",
		Message:  "Issue without line",
		Severity: "high",
		File:     "test.go",
		Line:     0,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.Contains(string(output), "| test.go | 0 |") {
		t.Error("Output should show file with line 0")
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"pipe|character", "pipe\\|character"},
		{"newline\nhere", "newline here"},
		{"carriage\rreturn", "carriagereturn"},
		{"mixed|pipe\nand\rchar", "mixed\\|pipe andchar"},
	}

	for _, tt := range tests {
		got := escapeMarkdown(tt.input)
		if got != tt.expected {
			t.Errorf("escapeMarkdown(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatSeverityAllCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "CRITICAL"},
		{"high", "HIGH"},
		{"medium", "MEDIUM"},
		{"low", "LOW"},
		{"info", "INFO"},
		{"unknown", "UNKNOWN"},
		{"", ""},
	}

	for _, tt := range tests {
		got := formatSeverity(tt.input)
		if got != tt.expected {
			t.Errorf("formatSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSARIFFormatterMultipleUniqueRules(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()

	// Add 3 issues with different rule IDs to ensure sorting is exercised
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "zzz-rule",
		Message:  "Third issue",
		Severity: "low",
		File:     "file3.go",
		Line:     30,
	})
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "aaa-rule",
		Message:  "First issue",
		Severity: "high",
		File:     "file1.go",
		Line:     10,
	})
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "mmm-rule",
		Message:  "Second issue",
		Severity: "medium",
		File:     "file2.go",
		Line:     20,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	rules := driver["rules"].([]any)

	// Should have 3 unique rules, sorted alphabetically
	if len(rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(rules))
	}

	// Verify alphabetical sorting
	for i, r := range rules {
		rule := r.(map[string]any)
		var expectedID string
		switch i {
		case 0:
			expectedID = "aaa-rule"
		case 1:
			expectedID = "mmm-rule"
		case 2:
			expectedID = "zzz-rule"
		}
		if rule["id"] != expectedID {
			t.Errorf("rules[%d].id = %v, want %s", i, rule["id"], expectedID)
		}
	}
}

func TestSARIFFormatterEmptyRuleID(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()
	diagnostics.AddIssue(diag, diagnostics.Issue{
		RuleID:   "",
		Message:  "Issue with empty rule ID",
		Severity: "high",
		File:     "test.go",
		Line:     1,
	})

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	rules := driver["rules"].([]any)

	// Should have 1 rule with empty ID
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestSARIFFormatterManyIssues(t *testing.T) {
	f := NewSARIFFormatter()
	diag := diagnostics.NewDiagnostics()

	// Add many issues to ensure all code paths are exercised
	for i := 0; i < 20; i++ {
		severities := []string{"critical", "high", "medium", "low", "info"}
		diagnostics.AddIssue(diag, diagnostics.Issue{
			RuleID:   fmt.Sprintf("rule-%d", i),
			Message:  fmt.Sprintf("Message %d", i),
			Severity: severities[i%len(severities)],
			File:     fmt.Sprintf("file%d.go", i),
			Line:     i * 10,
			Column:   i,
		})
	}

	output, err := f.Format(diag)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	runs := result["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	rules := driver["rules"].([]any)

	// Should have 20 unique rules
	if len(rules) != 20 {
		t.Errorf("Expected 20 rules, got %d", len(rules))
	}
}

func TestBuildRulesEmptyIssues(t *testing.T) {
	// Test buildRules directly with empty slice
	rules := buildRules([]diagnostics.Issue{})
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules for empty issues, got %d", len(rules))
	}
}

func TestBuildResultsEmptyIssues(t *testing.T) {
	// Test buildResults directly with empty slice
	results := buildResults([]diagnostics.Issue{})
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty issues, got %d", len(results))
	}
}

func TestBuildResultsWithColumn(t *testing.T) {
	// Test buildResults with column > 0
	results := buildResults([]diagnostics.Issue{
		{
			RuleID:   "col-rule",
			Message:  "Issue with column",
			Severity: "high",
			File:     "test.go",
			Line:     10,
			Column:   5,
		},
	})

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	locations := result["locations"].([]map[string]any)
	loc0 := locations[0]
	physLoc := loc0["physicalLocation"].(map[string]any)
	region := physLoc["region"].(map[string]any)

	if region["startLine"] != 10 {
		t.Errorf("startLine = %v, want 10", region["startLine"])
	}
	if region["startColumn"] != 5 {
		t.Errorf("startColumn = %v, want 5", region["startColumn"])
	}
}
