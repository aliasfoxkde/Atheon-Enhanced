package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg := LoadDefault()

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
	if cfg.Output.Format != "console" {
		t.Errorf("expected output format console, got %s", cfg.Output.Format)
	}
	if cfg.Output.Verbose {
		t.Error("expected verbose to be false by default")
	}
	if cfg.Scanner.MaxFileSize != 10*1024*1024 {
		t.Errorf("expected max file size 10MB, got %d", cfg.Scanner.MaxFileSize)
	}
	if len(cfg.Scanner.Extensions) == 0 {
		t.Error("expected default extensions to be set")
	}
}

func TestLoadJSON(t *testing.T) {
	content := `{
		"version": "2.0",
		"rules": [{"id": "rule1", "enabled": true, "severity": "high"}],
		"patterns": [{"id": "pattern1", "enabled": false}],
		"excludes": ["node_modules"],
		"output": {"format": "json", "verbose": true},
		"scanner": {"max_file_size": 5000000, "extensions": [".go", ".rs"]}
	}`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", cfg.Version)
	}
	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].ID != "rule1" {
		t.Errorf("expected rule id rule1, got %s", cfg.Rules[0].ID)
	}
	if !cfg.Rules[0].Enabled {
		t.Error("expected rule1 to be enabled")
	}
	if cfg.Output.Format != "json" {
		t.Errorf("expected format json, got %s", cfg.Output.Format)
	}
	if !cfg.Output.Verbose {
		t.Error("expected verbose to be true")
	}
	if cfg.Scanner.MaxFileSize != 5000000 {
		t.Errorf("expected max file size 5000000, got %d", cfg.Scanner.MaxFileSize)
	}
}

func TestLoadYAML(t *testing.T) {
	content := `version: "2.0"
rules:
  - id: rule1
    enabled: true
    severity: high
patterns:
  - id: pattern1
    enabled: false
excludes:
  - node_modules
output:
  format: json
  verbose: true
scanner:
  max_file_size: 5000000
  extensions:
    - .go
    - .rs
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", cfg.Version)
	}
	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].ID != "rule1" {
		t.Errorf("expected rule id rule1, got %s", cfg.Rules[0].ID)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMerge(t *testing.T) {
	c1 := &Config{
		Version: "1.0",
		Rules:   []RuleConfig{{ID: "rule1", Enabled: true, Severity: "low"}},
		Patterns: []PatternConfig{{ID: "pattern1", Enabled: true}},
		Excludes: []string{"node_modules"},
		Output:   OutputConfig{Format: "console", Verbose: false},
		Scanner:  ScannerConfig{MaxFileSize: 1024, Extensions: []string{".go"}},
	}

	c2 := &Config{
		Version: "2.0",
		Rules:   []RuleConfig{{ID: "rule2", Enabled: true, Severity: "high"}},
		Output:   OutputConfig{Format: "json", Verbose: true},
		Scanner:  ScannerConfig{MaxFileSize: 2048, Extensions: []string{".rs"}},
	}

	result := Merge(c1, c2)

	if result.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", result.Version)
	}
	if result.Output.Format != "json" {
		t.Errorf("expected format json, got %s", result.Output.Format)
	}
	if !result.Output.Verbose {
		t.Error("expected verbose to be true")
	}
	if result.Scanner.MaxFileSize != 2048 {
		t.Errorf("expected max file size 2048, got %d", result.Scanner.MaxFileSize)
	}

	// Check rules merge - both rule1 and rule2 should be present
	if len(result.Rules) != 2 {
		t.Errorf("expected 2 rules after merge, got %d", len(result.Rules))
	}

	// Check scanner extensions merge
	if len(result.Scanner.Extensions) != 2 {
		t.Errorf("expected 2 extensions after merge, got %d", len(result.Scanner.Extensions))
	}

	// Check excludes merge - c1 excludes should be preserved
	foundExcludes := false
	for _, e := range result.Excludes {
		if e == "node_modules" {
			foundExcludes = true
			break
		}
	}
	if !foundExcludes {
		t.Error("expected node_modules to be preserved in excludes")
	}
}

func TestMergeNilConfigs(t *testing.T) {
	result := Merge(nil, nil)
	if result == nil {
		t.Error("expected non-nil result when both args are nil")
	}

	c1 := &Config{Version: "1.0"}
	result = Merge(nil, c1)
	if result.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", result.Version)
	}

	result = Merge(c1, nil)
	if result.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", result.Version)
	}
}

func TestMergeOverrideBehavior(t *testing.T) {
	c1 := &Config{
		Output: OutputConfig{Format: "console", Verbose: false},
	}
	c2 := &Config{
		Output: OutputConfig{Format: "json"},
	}

	result := Merge(c1, c2)
	if result.Output.Format != "json" {
		t.Errorf("expected format json (override), got %s", result.Output.Format)
	}
	// Verbose should remain false since c2 didn't specify it
	if result.Output.Verbose != false {
		t.Errorf("expected verbose to remain false, got %v", result.Output.Verbose)
	}
}

func TestConfigJSONRoundTrip(t *testing.T) {
	original := &Config{
		Version:  "1.5",
		Rules:    []RuleConfig{{ID: "test-rule", Enabled: true, Severity: "medium"}},
		Patterns: []PatternConfig{{ID: "test-pattern", Enabled: false}},
		Excludes: []string{"*.log"},
		Output:   OutputConfig{Format: "sarif", Verbose: true},
		Scanner:  ScannerConfig{MaxFileSize: 8192, Extensions: []string{".ts"}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	var restored Config
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if restored.Version != original.Version {
		t.Errorf("version mismatch: got %s, want %s", restored.Version, original.Version)
	}
	if len(restored.Rules) != len(original.Rules) {
		t.Errorf("rules length mismatch: got %d, want %d", len(restored.Rules), len(original.Rules))
	}
	if restored.Output.Format != original.Output.Format {
		t.Errorf("output format mismatch: got %s, want %s", restored.Output.Format, original.Output.Format)
	}
}
