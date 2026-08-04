package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aliasfoxkde/Atheon/runtime/cache"
	"github.com/aliasfoxkde/Atheon/runtime/config"
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
	"github.com/aliasfoxkde/Atheon/runtime/patterns"
	"github.com/aliasfoxkde/Atheon/runtime/rules"
	"github.com/aliasfoxkde/Atheon/runtime/scanner"
)

func TestNewRuntime(t *testing.T) {
	r := NewRuntime()
	if r == nil {
		t.Fatal("NewRuntime returned nil")
	}
	if r.scanner == nil {
		t.Error("Runtime.scanner is nil")
	}
	if r.rules == nil {
		t.Error("Runtime.rules is nil")
	}
	if r.patterns == nil {
		t.Error("Runtime.patterns is nil")
	}
	if r.config == nil {
		t.Error("Runtime.config is nil")
	}
	if r.cache == nil {
		t.Error("Runtime.cache is nil")
	}
	if len(r.formats) == 0 {
		t.Error("Runtime.formats is empty")
	}
}

func TestNewRuntimeWithOptions(t *testing.T) {
	cfg := config.LoadDefault()
	r := NewRuntime(
		WithConfig(cfg),
	)
	if r == nil {
		t.Fatal("NewRuntime returned nil")
	}
	if r.config != cfg {
		t.Error("WithConfig option did not set config")
	}
}

func TestNewRuntimeWithScanner(t *testing.T) {
	s := scanner.NewScanner(scanner.Options{})
	r := NewRuntime(WithScanner(s))
	if r.scanner != s {
		t.Error("WithScanner did not set scanner")
	}
}

func TestNewRuntimeWithRules(t *testing.T) {
	reg := rules.NewRegistry()
	r := NewRuntime(WithRules(reg))
	if r.rules != reg {
		t.Error("WithRules did not set rules")
	}
}

func TestNewRuntimeWithPatterns(t *testing.T) {
	m := patterns.NewMatcher(patterns.BuiltinPatterns())
	r := NewRuntime(WithPatterns(m))
	if r.patterns != m {
		t.Error("WithPatterns did not set patterns")
	}
}

func TestNewRuntimeWithCache(t *testing.T) {
	c := cache.NewCache(0)
	r := NewRuntime(WithCache(c))
	if r.cache != c {
		t.Error("WithCache did not set cache")
	}
}

func TestRuntimeAnalyze(t *testing.T) {
	// Create temp directory with a Go file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	password := "secret123"
	println(password)
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	d, err := r.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if d == nil {
		t.Fatal("Analyze returned nil diagnostics")
	}
}

func TestRuntimeAnalyzeError(t *testing.T) {
	r := NewRuntime()
	// Analyze with nonexistent directory should return error
	_, err := r.Analyze("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Analyze should return error for nonexistent path")
	}
}

func TestRuntimeAnalyzeFiles(t *testing.T) {
	// Create temp directory with a Go file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	password := "secret123"
	println(password)
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	d, err := r.AnalyzeFiles([]string{goFile})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
	if d.Statistics.FilesScanned != 1 {
		t.Errorf("expected FilesScanned=1, got %d", d.Statistics.FilesScanned)
	}
}

func TestRuntimeAnalyzeFilesParseError(t *testing.T) {
	// Create temp file that is not valid Go
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.go")
	err := os.WriteFile(badFile, []byte(`not valid go code {{{`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	d, err := r.AnalyzeFiles([]string{badFile})
	if err != nil {
		t.Fatalf("AnalyzeFiles should not return error for parse failure: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
}

func TestRuntimeFormatResult(t *testing.T) {
	r := NewRuntime()

	// Test console format
	d := &diagnostics.Diagnostics{}
	data, err := r.FormatResult(d, "console")
	if err != nil {
		t.Fatalf("FormatResult failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("FormatResult returned empty data")
	}
}

func TestRuntimeFormatResultJSON(t *testing.T) {
	r := NewRuntime()
	d := &diagnostics.Diagnostics{}
	data, err := r.FormatResult(d, "json")
	if err != nil {
		t.Fatalf("FormatResult json failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("FormatResult json returned empty data")
	}
}

func TestRuntimeFormatResultSARIF(t *testing.T) {
	r := NewRuntime()
	d := &diagnostics.Diagnostics{}
	data, err := r.FormatResult(d, "sarif")
	if err != nil {
		t.Fatalf("FormatResult sarif failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("FormatResult sarif returned empty data")
	}
}

func TestRuntimeFormatResultMarkdown(t *testing.T) {
	r := NewRuntime()
	d := &diagnostics.Diagnostics{}
	data, err := r.FormatResult(d, "markdown")
	if err != nil {
		t.Fatalf("FormatResult markdown failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("FormatResult markdown returned empty data")
	}
}

func TestRuntimeFormatResultUnknown(t *testing.T) {
	r := NewRuntime()
	d := &diagnostics.Diagnostics{}
	// Unknown format should fall back to console
	data, err := r.FormatResult(d, "unknown")
	if err != nil {
		t.Fatalf("FormatResult unknown failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("FormatResult unknown returned empty data")
	}
}

func TestRuntimeListFormatters(t *testing.T) {
	r := NewRuntime()
	formatters := r.ListFormatters()
	if len(formatters) == 0 {
		t.Error("ListFormatters returned empty list")
	}
	// Check expected formatters
	expected := []string{"console", "json", "sarif", "markdown"}
	for _, exp := range expected {
		found := false
		for _, f := range formatters {
			if f == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected formatter %q not found", exp)
		}
	}
}

func TestRuntimeGetConfig(t *testing.T) {
	r := NewRuntime()
	cfg := r.GetConfig()
	if cfg == nil {
		t.Error("GetConfig returned nil")
	}
}

func TestRuntimeUpdateConfig(t *testing.T) {
	r := NewRuntime()
	newCfg := config.LoadDefault()
	r.UpdateConfig(newCfg)
	if r.config != newCfg {
		t.Error("UpdateConfig did not set config")
	}
}

func TestRuntimeBenchmark(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	println("hello")
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	result, err := r.Benchmark(tmpDir)
	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}
	if result == nil {
		t.Fatal("Benchmark returned nil")
	}
	if result.FilesScanned == 0 {
		t.Error("Benchmark should scan at least 1 file")
	}
}

func TestRuntimeBenchmarkError(t *testing.T) {
	r := NewRuntime()
	// Benchmark with nonexistent directory should return error
	_, err := r.Benchmark("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Benchmark should return error for nonexistent path")
	}
}

func TestRuntimeValidateConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"version": "1.0",
		"rules": [
			{"id": "G401", "enabled": true}
		]
	}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	r := NewRuntime()
	result, err := r.ValidateConfig(configFile)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	if result == nil {
		t.Fatal("ValidateConfig returned nil")
	}
	if !result.Valid {
		t.Error("ValidateConfig should be valid")
	}
}

func TestRuntimeValidateConfigInvalid(t *testing.T) {
	r := NewRuntime()
	result, err := r.ValidateConfig("/nonexistent/config.json")
	if err == nil {
		t.Fatal("ValidateConfig expected error for nonexistent file")
	}
	if result == nil {
		t.Fatal("ValidateConfig returned nil")
	}
	if result.Valid {
		t.Error("ValidateConfig should be invalid for nonexistent file")
	}
	if len(result.Errors) == 0 {
		t.Error("ValidateConfig should have errors for nonexistent file")
	}
}

func TestRuntimeValidateConfigUnknownRule(t *testing.T) {
	// Create temp config file with unknown rule
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"version": "1.0",
		"rules": [
			{"id": "UNKNOWN_RULE_123", "enabled": true}
		]
	}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	r := NewRuntime()
	result, err := r.ValidateConfig(configFile)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	if result == nil {
		t.Fatal("ValidateConfig returned nil")
	}
	if !result.Valid {
		t.Error("ValidateConfig should still be valid with unknown rule")
	}
	if len(result.Warnings) == 0 {
		t.Error("ValidateConfig should have warning for unknown rule")
	}
}

func TestRuntimeValidateConfigDisabledRule(t *testing.T) {
	// Create temp config file with disabled rule
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"version": "1.0",
		"rules": [
			{"id": "G401", "enabled": false}
		]
	}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	r := NewRuntime()
	result, err := r.ValidateConfig(configFile)
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	if result == nil {
		t.Fatal("ValidateConfig returned nil")
	}
	if !result.Valid {
		t.Error("ValidateConfig should be valid with disabled rule")
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version returned empty string")
	}
}

func TestRuntimeAnalyzeFilesEmpty(t *testing.T) {
	r := NewRuntime()
	d, err := r.AnalyzeFiles([]string{})
	if err != nil {
		t.Fatalf("AnalyzeFiles with empty list failed: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics for empty input")
	}
	if d.Statistics.FilesScanned != 0 {
		t.Errorf("expected FilesScanned=0, got %d", d.Statistics.FilesScanned)
	}
}

func TestRuntimeAnalyzeFilesNonexistent(t *testing.T) {
	r := NewRuntime()
	d, err := r.AnalyzeFiles([]string{"/nonexistent/file.go"})
	if err != nil {
		t.Fatalf("AnalyzeFiles should not error on nonexistent file: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
}

func TestRuntimeAnalyzeMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	files := []string{}
	for i := 0; i < 3; i++ {
		goFile := filepath.Join(tmpDir, "test.go")
		err := os.WriteFile(goFile, []byte(`package main
func main() {
	password := "secret123"
	println(password)
}
`), 0644)
		if err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		files = append(files, goFile)
	}

	r := NewRuntime()
	d, err := r.AnalyzeFiles(files)
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
}

func TestRuntimeBenchmarkResultFields(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	println("hello")
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	result, err := r.Benchmark(tmpDir)
	if err != nil {
		t.Fatalf("Benchmark failed: %v", err)
	}

	// Verify all result fields are populated
	if result.FilesScanned == 0 {
		t.Error("FilesScanned should not be 0")
	}
	if result.FilesAnalyzed == 0 {
		t.Error("FilesAnalyzed should not be 0")
	}
	if result.Duration < 0 {
		t.Error("Duration should not be negative")
	}
	if result.ThroughputFiles < 0 {
		t.Error("ThroughputFiles should not be negative")
	}
}

func TestFormatterRegistration(t *testing.T) {
	r := NewRuntime()

	// Verify all formatters are registered
	formatters := []string{"console", "json", "sarif", "markdown"}
	for _, f := range formatters {
		if _, ok := r.formats[f]; !ok {
			t.Errorf("formatter %q not registered", f)
		}
	}
}

func TestFormatResultWithIssues(t *testing.T) {
	r := NewRuntime()

	// Create diagnostics with some issues
	d := &diagnostics.Diagnostics{
		Issues: []diagnostics.Issue{
			{
				RuleID:  "G401",
				Message: "Sensitive data found",
				File:    "test.go",
				Line:    10,
			},
		},
	}

	// Test all formats with actual issue data
	for _, format := range []string{"console", "json", "sarif", "markdown"} {
		data, err := r.FormatResult(d, format)
		if err != nil {
			t.Errorf("FormatResult(%s) failed: %v", format, err)
		}
		if len(data) == 0 {
			t.Errorf("FormatResult(%s) returned empty data", format)
		}
	}
}

func TestRuntimeAnalyzeWithIssue(t *testing.T) {
	// Create a custom registry with a rule that always returns an issue
	reg := rules.NewRegistry()
	reg.Register(&rules.Rule{
		ID:          "TEST001",
		Description: "Test rule",
		Severity:    "high",
		Check: func(file string, node interface{}) *diagnostics.Issue {
			// Always return an issue to trigger AddIssue
			return &diagnostics.Issue{
				RuleID:   "TEST001",
				Message:  "Test issue",
				Severity: "high",
				File:     file,
			}
		},
	})

	// Create temp directory with a Go file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	println("hello")
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime(WithRules(reg))
	d, err := r.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if d == nil {
		t.Fatal("Analyze returned nil diagnostics")
	}
	if len(d.Issues) == 0 {
		t.Error("Analyze should have found issues with custom rule")
	}
}

func TestRuntimeAnalyzeFilesWithIssue(t *testing.T) {
	// Create a custom registry with a rule that always returns an issue
	reg := rules.NewRegistry()
	reg.Register(&rules.Rule{
		ID:          "TEST001",
		Description: "Test rule",
		Severity:    "high",
		Check: func(file string, node interface{}) *diagnostics.Issue {
			return &diagnostics.Issue{
				RuleID:   "TEST001",
				Message:  "Test issue",
				Severity: "high",
				File:     file,
			}
		},
	})

	// Create temp directory with a Go file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte(`package main
func main() {
	println("hello")
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime(WithRules(reg))
	d, err := r.AnalyzeFiles([]string{goFile})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
	if len(d.Issues) == 0 {
		t.Error("AnalyzeFiles should have found issues with custom rule")
	}
}

func TestRuntimeAnalyzeFilesWithParseError(t *testing.T) {
	// Create temp file that will fail to parse
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.go")
	// Write invalid Go code
	err := os.WriteFile(badFile, []byte(`package main
func main() {
	invalid syntax here
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	r := NewRuntime()
	d, err := r.AnalyzeFiles([]string{badFile})
	if err != nil {
		t.Fatalf("AnalyzeFiles should not return error for parse failure: %v", err)
	}
	if d == nil {
		t.Fatal("AnalyzeFiles returned nil diagnostics")
	}
	// Verify the file was counted as scanned despite parse failure
	if d.Statistics.FilesScanned != 1 {
		t.Errorf("expected FilesScanned=1, got %d", d.Statistics.FilesScanned)
	}
}
