package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aliasfoxkde/Atheon/runtime/config"
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
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
	if err != nil {
		t.Fatalf("ValidateConfig returned error: %v", err)
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

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version returned empty string")
	}
}
