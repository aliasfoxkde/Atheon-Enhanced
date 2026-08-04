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
		Version:  "1.0",
		Rules:    []RuleConfig{{ID: "rule1", Enabled: true, Severity: "low"}},
		Patterns: []PatternConfig{{ID: "pattern1", Enabled: true}},
		Excludes: []string{"node_modules"},
		Output:   OutputConfig{Format: "console", Verbose: false},
		Scanner:  ScannerConfig{MaxFileSize: 1024, Extensions: []string{".go"}},
	}

	c2 := &Config{
		Version: "2.0",
		Rules:   []RuleConfig{{ID: "rule2", Enabled: true, Severity: "high"}},
		Output:  OutputConfig{Format: "json", Verbose: true},
		Scanner: ScannerConfig{MaxFileSize: 2048, Extensions: []string{".rs"}},
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

func TestMergeVerboseOverride(t *testing.T) {
	// Test that Verbose=false in c2 overrides c1's Verbose=true
	c1 := &Config{
		Output: OutputConfig{Format: "console", Verbose: true},
	}
	c2 := &Config{
		Output: OutputConfig{Verbose: false},
	}

	result := Merge(c1, c2)
	if result.Output.Verbose {
		t.Error("expected verbose to be false after override")
	}
}

func TestMergeMaxFileSizeOverride(t *testing.T) {
	// Test MaxFileSize override
	c1 := &Config{
		Scanner: ScannerConfig{MaxFileSize: 1024},
	}
	c2 := &Config{
		Scanner: ScannerConfig{MaxFileSize: 2048},
	}

	result := Merge(c1, c2)
	if result.Scanner.MaxFileSize != 2048 {
		t.Errorf("expected max file size 2048, got %d", result.Scanner.MaxFileSize)
	}
}

func TestMergeRulesMerge(t *testing.T) {
	// Test rules from both configs are merged
	c1 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: true, Severity: "low"}},
	}
	c2 := &Config{
		Rules: []RuleConfig{{ID: "rule2", Enabled: true, Severity: "high"}},
	}

	result := Merge(c1, c2)
	if len(result.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result.Rules))
	}

	ruleIDs := make(map[string]bool)
	for _, r := range result.Rules {
		ruleIDs[r.ID] = true
	}
	if !ruleIDs["rule1"] {
		t.Error("expected rule1 to be present")
	}
	if !ruleIDs["rule2"] {
		t.Error("expected rule2 to be present")
	}
}

func TestMergeRulesOverride(t *testing.T) {
	// Test that rule with same ID in c2 overrides c1
	c1 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: true, Severity: "low"}},
	}
	c2 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: false, Severity: "high"}},
	}

	result := Merge(c1, c2)
	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule after override, got %d", len(result.Rules))
	}
	if result.Rules[0].Severity != "high" {
		t.Errorf("expected severity high, got %s", result.Rules[0].Severity)
	}
	if result.Rules[0].Enabled {
		t.Error("expected rule to be disabled after override")
	}
}

func TestMergePatternsMerge(t *testing.T) {
	c1 := &Config{
		Patterns: []PatternConfig{{ID: "pattern1", Enabled: true}},
	}
	c2 := &Config{
		Patterns: []PatternConfig{{ID: "pattern2", Enabled: false}},
	}

	result := Merge(c1, c2)
	if len(result.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(result.Patterns))
	}
}

func TestMergePatternsOverride(t *testing.T) {
	c1 := &Config{
		Patterns: []PatternConfig{{ID: "pattern1", Enabled: true}},
	}
	c2 := &Config{
		Patterns: []PatternConfig{{ID: "pattern1", Enabled: false}},
	}

	result := Merge(c1, c2)
	if len(result.Patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(result.Patterns))
	}
	if result.Patterns[0].Enabled {
		t.Error("expected pattern to be disabled after override")
	}
}

func TestMergeExcludesMerge(t *testing.T) {
	c1 := &Config{
		Excludes: []string{"node_modules", ".git"},
	}
	c2 := &Config{
		Excludes: []string{"*.test", "dist"},
	}

	result := Merge(c1, c2)
	if len(result.Excludes) != 4 {
		t.Errorf("expected 4 excludes, got %d", len(result.Excludes))
	}
}

func TestMergeExcludesOverride(t *testing.T) {
	// When c2 has excludes, c1's excludes should be preserved too
	c1 := &Config{
		Excludes: []string{"node_modules"},
	}
	c2 := &Config{
		Excludes: []string{"*.test"},
	}

	result := Merge(c1, c2)
	found := false
	for _, e := range result.Excludes {
		if e == "node_modules" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected node_modules to be preserved")
	}
}

func TestMergeExtensionsMerge(t *testing.T) {
	c1 := &Config{
		Scanner: ScannerConfig{Extensions: []string{".go", ".rs"}},
	}
	c2 := &Config{
		Scanner: ScannerConfig{Extensions: []string{".py", ".js"}},
	}

	result := Merge(c1, c2)
	if len(result.Scanner.Extensions) != 4 {
		t.Errorf("expected 4 extensions, got %d", len(result.Scanner.Extensions))
	}
}

func TestMergeExtensionsDeduplicate(t *testing.T) {
	c1 := &Config{
		Scanner: ScannerConfig{Extensions: []string{".go", ".rs"}},
	}
	c2 := &Config{
		Scanner: ScannerConfig{Extensions: []string{".go", ".py"}},
	}

	result := Merge(c1, c2)
	if len(result.Scanner.Extensions) != 3 {
		t.Errorf("expected 3 extensions (deduplicated), got %d", len(result.Scanner.Extensions))
	}
}

func TestMergeOnlyC1HasRules(t *testing.T) {
	c1 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: true}},
	}
	c2 := &Config{}

	result := Merge(c1, c2)
	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result.Rules))
	}
}

func TestMergeOnlyC2HasRules(t *testing.T) {
	c1 := &Config{}
	c2 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: true}},
	}

	result := Merge(c1, c2)
	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result.Rules))
	}
}

func TestMergeEmptyRules(t *testing.T) {
	c1 := &Config{
		Rules: []RuleConfig{},
	}
	c2 := &Config{
		Rules: []RuleConfig{{ID: "rule1", Enabled: true}},
	}

	result := Merge(c1, c2)
	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result.Rules))
	}
}

func TestMergeVersionOverride(t *testing.T) {
	c1 := &Config{Version: "1.0"}
	c2 := &Config{Version: "2.0"}

	result := Merge(c1, c2)
	if result.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", result.Version)
	}
}

func TestMergeVersionEmptyC2(t *testing.T) {
	c1 := &Config{Version: "1.0"}
	c2 := &Config{Version: ""}

	result := Merge(c1, c2)
	if result.Version != "1.0" {
		t.Errorf("expected version 1.0 (c1 preserved), got %s", result.Version)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Content that is neither valid JSON nor valid YAML
	content := `{invalid yaml: [not closed`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.invalid")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML content")
	}
}

func TestLoadJSONThatFallsBackToYAML(t *testing.T) {
	// YAML content that is NOT valid JSON - tests the JSON-fail-then-YAML-succeed path
	// This is a simple YAML document that JSON parser will reject
	content := `version: "2.0"
output:
  format: yaml
  verbose: false
scanner:
  max_file_size: 12345
  extensions:
    - .yaml
    - .yml
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed for YAML content: %v", err)
	}

	if cfg.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", cfg.Version)
	}
	if cfg.Output.Format != "yaml" {
		t.Errorf("expected format yaml, got %s", cfg.Output.Format)
	}
	if cfg.Scanner.MaxFileSize != 12345 {
		t.Errorf("expected max_file_size 12345, got %d", cfg.Scanner.MaxFileSize)
	}
	if len(cfg.Scanner.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(cfg.Scanner.Extensions))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	// Empty file is actually valid for both JSON and YAML parsers
	// (they return empty/nil structures), so we test it doesn't panic
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.empty")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("empty file should not cause error: %v", err)
	}
	// Empty content results in default values (empty structs)
	if cfg.Version != "" {
		t.Errorf("expected empty version for empty file, got %s", cfg.Version)
	}
}

func TestLoadMinimalYAML(t *testing.T) {
	// Minimal YAML that is not valid JSON - just version field
	content := `version: "1.0"`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
}

func TestLoadYAMLWithOnlyOutput(t *testing.T) {
	content := `output:
  format: sarif
  verbose: true
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

	if cfg.Output.Format != "sarif" {
		t.Errorf("expected format sarif, got %s", cfg.Output.Format)
	}
	if !cfg.Output.Verbose {
		t.Error("expected verbose to be true")
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
