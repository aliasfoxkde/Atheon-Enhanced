package pi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// ============================================================
// Adapter Interface Tests
// ============================================================

func TestAdapterName(t *testing.T) {
	adapter := New(DefaultConfig())
	if got := adapter.Name(); got != "pi" {
		t.Errorf("Name() = %q, want %q", got, "pi")
	}
}

func TestAdapterVersion(t *testing.T) {
	adapter := New(DefaultConfig())
	if got := adapter.Version(); got != "1.0.0" {
		t.Errorf("Version() = %q, want %q", got, "1.0.0")
	}
}

func TestAdapterInitialize(t *testing.T) {
	adapter := New(DefaultConfig())
	if err := adapter.Initialize(nil); err != nil {
		t.Errorf("Initialize() returned error: %v", err)
	}
}

func TestAdapterValidate(t *testing.T) {
	adapter := New(DefaultConfig())
	if err := adapter.Validate(); err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}
}

// ============================================================
// Analyze Function Tests
// ============================================================

func TestAdapterScan(t *testing.T) {
	adapter := New(DefaultConfig())

	// Create a temp file with a test secret
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.RiskScore == nil {
		t.Fatal("expected risk score, got nil")
	}

	t.Logf("Summary: Total=%d, Critical=%d, High=%d, Medium=%d, Low=%d",
		result.Summary.Total, result.Summary.Critical, result.Summary.High,
		result.Summary.Medium, result.Summary.Low)
	t.Logf("Risk Score: %d (%s)", result.RiskScore.Score, result.RiskScore.Level)
}

func TestAnalyzeWithStringInput(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with string input instead of []string
	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze(string) failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeWithSliceStringInput(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with []string input
	result, err := adapter.Analyze([]string{tmpDir})
	if err != nil {
		t.Fatalf("Analyze([]string) failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeWithSliceInterfaceInput(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with []interface{} input (simulating JSON unmarshal)
	input := []interface{}{tmpDir}
	result, err := adapter.Analyze(input)
	if err != nil {
		t.Fatalf("Analyze([]interface{}) failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeEmptyStringInput(t *testing.T) {
	adapter := New(DefaultConfig())
	_, err := adapter.Analyze("")
	if err == nil {
		t.Error("expected error for empty string input")
	}
}

func TestAnalyzeUnsupportedInputType(t *testing.T) {
	adapter := New(DefaultConfig())
	_, err := adapter.Analyze(12345)
	if err == nil {
		t.Error("expected error for unsupported input type")
	}
}

func TestAnalyzeNoPathsProvided(t *testing.T) {
	adapter := New(DefaultConfig())
	_, err := adapter.Analyze([]string{})
	if err == nil {
		t.Error("expected error for empty paths slice")
	}
	if err != nil && err.Error() != "no paths provided" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnalyzeNonexistentPath(t *testing.T) {
	adapter := New(DefaultConfig())
	_, err := adapter.Analyze("/nonexistent/path/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestAnalyzeFileNotFound(t *testing.T) {
	adapter := New(DefaultConfig())
	// Path exists but is a file that doesn't exist (parent dir exists)
	_, err := adapter.Analyze("/tmp/this_file_definitely_does_not_exist_12345.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestAnalyzeMultiplePaths(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir1 := t.TempDir()
	tmpFile1 := tmpDir1 + "/test1.go"
	err := os.WriteFile(tmpFile1, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir2 := t.TempDir()
	tmpFile2 := tmpDir2 + "/test2.go"
	err = os.WriteFile(tmpFile2, []byte(`
package main
const password = "secret123"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze([]string{tmpDir1, tmpDir2})
	if err != nil {
		t.Fatalf("Analyze with multiple paths failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	// Should have findings from both files
	if result.Summary.Total < 2 {
		t.Errorf("expected at least 2 findings, got %d", result.Summary.Total)
	}
}

func TestAnalyzeSingleFile(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test scanning a single file directly
	result, err := adapter.Analyze(tmpFile)
	if err != nil {
		t.Fatalf("Analyze(file) failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeDifferentFileTypes(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()

	// Create Python file with secret
	pythonFile := tmpDir + "/test.py"
	err := os.WriteFile(pythonFile, []byte(`
import os
AWS_KEY = "AKIAIOSFODNN7EXAMPLE"
print("hello")
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create JavaScript file with secret
	jsFile := tmpDir + "/test.js"
	err = os.WriteFile(jsFile, []byte(`
const AWS_KEY = "AKIAIOSFODNN7EXAMPLE";
console.log("hello");
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create YAML file with secret
	yamlFile := tmpDir + "/config.yaml"
	err = os.WriteFile(yamlFile, []byte(`
aws:
  access_key: "AKIAIOSFODNN7EXAMPLE"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeWithSeverityThreshold(t *testing.T) {
	adapter := New(Config{
		SeverityThreshold: "high",
	})

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze with severity threshold failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	// Result should be filtered by severity >= high
	t.Logf("Summary with high threshold: Total=%d", result.Summary.Total)
}

func TestAnalyzeWithCategoryFilter(t *testing.T) {
	adapter := New(Config{
		Categories: []string{"AWS"},
	})

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze with category filter failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	t.Logf("Summary with AWS category filter: Total=%d", result.Summary.Total)
}

// ============================================================
// HandleTool Tests
// ============================================================

func TestHandleToolScan(t *testing.T) {
	adapter := New(DefaultConfig())

	// Create a temp file with a test secret
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := ToolRequest{
		Command: "scan",
		Paths:   []string{tmpDir},
	}

	input, _ := json.Marshal(req)
	result, err := adapter.HandleTool(context.Background(), input)
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	t.Logf("Result: %s", string(result))
}

func TestHandleToolListCategories(t *testing.T) {
	adapter := New(DefaultConfig())

	req := ToolRequest{
		Command:  "list",
		ListType: "categories",
	}

	input, _ := json.Marshal(req)
	result, err := adapter.HandleTool(context.Background(), input)
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	t.Logf("Categories: %s", string(result))
}

func TestHandleToolListPatterns(t *testing.T) {
	adapter := New(DefaultConfig())

	req := ToolRequest{
		Command:  "list",
		ListType: "patterns",
	}

	input, _ := json.Marshal(req)
	result, err := adapter.HandleTool(context.Background(), input)
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	t.Logf("Patterns: %s", string(result))
}

func TestHandleToolVersion(t *testing.T) {
	adapter := New(DefaultConfig())

	req := ToolRequest{
		Command: "version",
	}

	input, _ := json.Marshal(req)
	result, err := adapter.HandleTool(context.Background(), input)
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	t.Logf("Version: %s", string(result))
}

func TestHandleToolUnknownCommand(t *testing.T) {
	adapter := New(DefaultConfig())

	req := ToolRequest{
		Command: "unknown_command",
	}

	input, _ := json.Marshal(req)
	_, err := adapter.HandleTool(context.Background(), input)
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestHandleToolInvalidJSON(t *testing.T) {
	adapter := New(DefaultConfig())

	_, err := adapter.HandleTool(context.Background(), json.RawMessage("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleToolScanWithFile(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	err := os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := ToolRequest{
		Command: "scan",
		Paths:   []string{tmpFile},
	}

	input, _ := json.Marshal(req)
	result, err := adapter.HandleTool(context.Background(), input)
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.RiskScore == nil {
		t.Error("expected risk score in response")
	}
}

// ============================================================
// Helper Function Tests
// ============================================================

func TestFilterBySeverity(t *testing.T) {
	findings := []core.Finding{
		{Severity: "critical", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
		{Severity: "medium", Category: "AWS"},
		{Severity: "low", Category: "AWS"},
		{Severity: "info", Category: "AWS"},
	}

	tests := []struct {
		threshold string
		want      int
	}{
		{"critical", 1},
		{"high", 2},
		{"medium", 3},
		{"low", 4},
		{"info", 5},
		{"unknown", 5}, // Unknown threshold returns all
	}

	for _, tt := range tests {
		t.Run(tt.threshold, func(t *testing.T) {
			filtered := filterBySeverity(findings, tt.threshold)
			if len(filtered) != tt.want {
				t.Errorf("filterBySeverity(%q) = %d, want %d", tt.threshold, len(filtered), tt.want)
			}
		})
	}
}

func TestFilterBySeverityEmpty(t *testing.T) {
	findings := []core.Finding{}
	filtered := filterBySeverity(findings, "high")
	if len(filtered) != 0 {
		t.Errorf("filterBySeverity(empty) = %d, want 0", len(filtered))
	}
}

func TestBuildSummary(t *testing.T) {
	findings := []core.Finding{
		{Severity: "critical", Category: "AWS"},
		{Severity: "critical", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
		{Severity: "medium", Category: "Secrets"},
		{Severity: "low", Category: "Secrets"},
		{Severity: "info", Category: "Comments"},
	}

	summary := buildSummary(findings)

	if summary.Total != 6 {
		t.Errorf("Total = %d, want 6", summary.Total)
	}
	if summary.Critical != 2 {
		t.Errorf("Critical = %d, want 2", summary.Critical)
	}
	if summary.High != 1 {
		t.Errorf("High = %d, want 1", summary.High)
	}
	if summary.Medium != 1 {
		t.Errorf("Medium = %d, want 1", summary.Medium)
	}
	if summary.Low != 1 {
		t.Errorf("Low = %d, want 1", summary.Low)
	}
	if summary.Infos != 1 {
		t.Errorf("Infos = %d, want 1", summary.Infos)
	}
	if summary.ByCategory["AWS"] != 3 {
		t.Errorf("ByCategory[AWS] = %d, want 3", summary.ByCategory["AWS"])
	}
	if summary.ByCategory["Secrets"] != 2 {
		t.Errorf("ByCategory[Secrets] = %d, want 2", summary.ByCategory["Secrets"])
	}
}

func TestBuildSummaryEmpty(t *testing.T) {
	findings := []core.Finding{}
	summary := buildSummary(findings)

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.ByCategory == nil {
		t.Error("ByCategory should not be nil")
	}
}

func TestBuildSummaryWithEmptyCategory(t *testing.T) {
	findings := []core.Finding{
		{Severity: "critical", Category: ""},
	}
	summary := buildSummary(findings)
	// Should not panic and should not add to ByCategory for empty category
	if summary.ByCategory[""] != 0 {
		t.Errorf("ByCategory[\"\"] = %d, want 0", summary.ByCategory[""])
	}
}

func TestComputeRiskScoreEmpty(t *testing.T) {
	findings := []core.Finding{}
	rs := computeRiskScore(findings)

	if rs.Score != 0 {
		t.Errorf("Score = %d, want 0", rs.Score)
	}
	if rs.Level != core.RiskLevelNone {
		t.Errorf("Level = %q, want %q", rs.Level, core.RiskLevelNone)
	}
}

func TestComputeRiskScoreLow(t *testing.T) {
	// Low score (<=30)
	findings := []core.Finding{
		{Severity: "info", Category: "Comments"},
		{Severity: "info", Category: "Comments"},
		{Severity: "info", Category: "Comments"},
	}
	rs := computeRiskScore(findings)

	if rs.Score != 15 { // 3 * 5 = 15
		t.Errorf("Score = %d, want 15", rs.Score)
	}
	if rs.Level != core.RiskLevelLow {
		t.Errorf("Level = %q, want %q", rs.Level, core.RiskLevelLow)
	}
}

func TestComputeRiskScoreMedium(t *testing.T) {
	// Medium score (>30 and <=70)
	findings := []core.Finding{
		{Severity: "low", Category: "AWS"},
		{Severity: "low", Category: "AWS"},
		{Severity: "low", Category: "AWS"},
		{Severity: "medium", Category: "AWS"},
		{Severity: "medium", Category: "AWS"},
	}
	// 3*10 + 2*20 = 70
	rs := computeRiskScore(findings)

	if rs.Score != 70 {
		t.Errorf("Score = %d, want 70", rs.Score)
	}
	if rs.Level != core.RiskLevelMedium {
		t.Errorf("Level = %q, want %q", rs.Level, core.RiskLevelMedium)
	}
}

func TestComputeRiskScoreHigh(t *testing.T) {
	// High score (>70 and <=100)
	findings := []core.Finding{
		{Severity: "medium", Category: "AWS"},
		{Severity: "medium", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
	}
	// 2*20 + 2*30 = 100
	rs := computeRiskScore(findings)

	if rs.Score != 100 {
		t.Errorf("Score = %d, want 100", rs.Score)
	}
	if rs.Level != core.RiskLevelHigh {
		t.Errorf("Level = %q, want %q", rs.Level, core.RiskLevelHigh)
	}
}

func TestComputeRiskScoreCritical(t *testing.T) {
	// Critical score (>100)
	findings := []core.Finding{
		{Severity: "high", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
		{Severity: "high", Category: "AWS"},
		{Severity: "critical", Category: "AWS"},
	}
	// 3*30 + 40 = 130, capped at different thresholds
	rs := computeRiskScore(findings)

	if rs.Level != core.RiskLevelCritical {
		t.Errorf("Level = %q, want %q", rs.Level, core.RiskLevelCritical)
	}
}

func TestComputeRiskScoreUnknownSeverity(t *testing.T) {
	findings := []core.Finding{
		{Severity: "unknown", Category: "Test"},
	}
	rs := computeRiskScore(findings)

	// Unknown severity should still be processed with default weight
	if rs.Score != 0 {
		t.Errorf("Score = %d, want 0 (unknown severity has no weight)", rs.Score)
	}
}

func TestComputeRiskScoreMixedKnownUnknown(t *testing.T) {
	findings := []core.Finding{
		{Severity: "high", Category: "AWS"},
		{Severity: "unknown", Category: "Test"},
	}
	rs := computeRiskScore(findings)

	// Only high should count: 30
	if rs.Score != 30 {
		t.Errorf("Score = %d, want 30", rs.Score)
	}
}

// ============================================================
// Edge Cases
// ============================================================

func TestAnalyzeWithEmptyCategoryFindings(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	// Create a file that will produce findings without category
	err := os.WriteFile(tmpFile, []byte(`
package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpFile)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	// Should handle gracefully even if no findings with categories
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeWithNestedDirectories(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "nested")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile := subDir + "/deep.go"
	err = os.WriteFile(tmpFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := adapter.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze with nested dirs failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestAnalyzeSymlinkDirectory(t *testing.T) {
	adapter := New(DefaultConfig())

	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	err := os.MkdirAll(realDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	realFile := filepath.Join(realDir, "test.go")
	err = os.WriteFile(realFile, []byte(`
package main
const awsKey = "AKIAIOSFODNN7EXAMPLE"
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create symlink to directory
	symlinkDir := filepath.Join(tmpDir, "link")
	err = os.Symlink(realDir, symlinkDir)
	if err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	result, err := adapter.Analyze(symlinkDir)
	if err != nil {
		t.Fatalf("Analyze with symlink dir failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestHandleToolScanInvalidPath(t *testing.T) {
	adapter := New(DefaultConfig())

	req := ToolRequest{
		Command: "scan",
		Paths:   []string{"/nonexistent/path/to/file.go"},
	}

	input, _ := json.Marshal(req)
	_, err := adapter.HandleTool(context.Background(), input)
	if err == nil {
		t.Error("expected error for nonexistent path in HandleTool")
	}
}

func TestAnalyzeNoValidPaths(t *testing.T) {
	adapter := New(DefaultConfig())
	// Provide paths that exist as files but are not valid for scanning
	// (binary files, etc.)
	tmpDir := t.TempDir()
	binaryFile := tmpDir + "/test.bin"
	err := os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02}, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// The file is regular but likely not producing findings
	_, err = adapter.Analyze(binaryFile)
	// This may or may not error depending on whether it's treated as valid path
	// Just ensure it doesn't panic
	_ = err
}

func TestNewWithConfig(t *testing.T) {
	cfg := Config{
		Categories:        []string{"AWS", "Secrets"},
		SeverityThreshold: "high",
	}
	adapter := New(cfg)

	if adapter.Name() != "pi" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "pi")
	}
	if adapter.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q", adapter.Version(), "1.0.0")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Categories != nil {
		t.Error("DefaultConfig should return nil Categories")
	}
	if cfg.SeverityThreshold != "" {
		t.Error("DefaultConfig should return empty SeverityThreshold")
	}
}
