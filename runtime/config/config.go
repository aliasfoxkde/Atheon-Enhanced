package config

import (
	"encoding/json"
	"os"

	"github.com/goccy/go-yaml"
)

// Config represents the runtime configuration.
type Config struct {
	Version  string           `json:"version"`
	Rules    []RuleConfig     `json:"rules,omitempty"`
	Patterns []PatternConfig  `json:"patterns,omitempty"`
	Excludes []string         `json:"excludes,omitempty"`
	Output   OutputConfig     `json:"output"`
	Scanner  ScannerConfig    `json:"scanner"`
}

// RuleConfig represents a rule configuration.
type RuleConfig struct {
	ID       string `json:"id"`
	Enabled  bool   `json:"enabled"`
	Severity string `json:"severity,omitempty"`
}

// PatternConfig represents a pattern configuration.
type PatternConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// OutputConfig configures output formatting.
type OutputConfig struct {
	Format  string `json:"format"` // json, sarif, markdown, console
	Verbose bool   `json:"verbose"`
}

// ScannerConfig configures scanner behavior.
type ScannerConfig struct {
	MaxFileSize int64    `json:"max_file_size"`
	Extensions  []string `json:"extensions,omitempty"`
}

// Load reads configuration from a JSON or YAML file.
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Try JSON first
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}

	// Try YAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadDefault returns a Config with sensible defaults.
func LoadDefault() *Config {
	return &Config{
		Version: "1.0",
		Rules:   []RuleConfig{},
		Patterns: []PatternConfig{},
		Excludes: []string{
			"node_modules",
			".git",
			"*.test",
		},
		Output: OutputConfig{
			Format:  "console",
			Verbose: false,
		},
		Scanner: ScannerConfig{
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			Extensions:  []string{".go", ".rs", ".py", ".js", ".ts"},
		},
	}
}

// Merge combines two configs, with c2 overriding c1.
// Values in c2 take precedence over c1.
func Merge(c1, c2 *Config) *Config {
	if c1 == nil && c2 == nil {
		return LoadDefault()
	}
	if c1 == nil {
		return c2
	}
	if c2 == nil {
		return c1
	}

	result := *c1

	// Override with c2 values if they differ from zero values
	if c2.Version != "" {
		result.Version = c2.Version
	}
	if c2.Output.Format != "" {
		result.Output.Format = c2.Output.Format
	}
	result.Output.Verbose = c2.Output.Verbose
	result.Scanner.MaxFileSize = c2.Scanner.MaxFileSize

	// Merge slices: c2 items override c1 items with same ID, otherwise append
	if len(c2.Rules) > 0 {
		ruleMap := make(map[string]RuleConfig)
		for _, r := range result.Rules {
			ruleMap[r.ID] = r
		}
		for _, r := range c2.Rules {
			ruleMap[r.ID] = r
		}
		result.Rules = make([]RuleConfig, 0, len(ruleMap))
		for _, r := range ruleMap {
			result.Rules = append(result.Rules, r)
		}
	}

	if len(c2.Patterns) > 0 {
		patternMap := make(map[string]PatternConfig)
		for _, p := range result.Patterns {
			patternMap[p.ID] = p
		}
		for _, p := range c2.Patterns {
			patternMap[p.ID] = p
		}
		result.Patterns = make([]PatternConfig, 0, len(patternMap))
		for _, p := range patternMap {
			result.Patterns = append(result.Patterns, p)
		}
	}

	if len(c2.Excludes) > 0 {
		excludeSet := make(map[string]struct{})
		for _, e := range result.Excludes {
			excludeSet[e] = struct{}{}
		}
		for _, e := range c2.Excludes {
			excludeSet[e] = struct{}{}
		}
		result.Excludes = make([]string, 0, len(excludeSet))
		for e := range excludeSet {
			result.Excludes = append(result.Excludes, e)
		}
	}

	if len(c2.Scanner.Extensions) > 0 {
		extSet := make(map[string]struct{})
		for _, ext := range result.Scanner.Extensions {
			extSet[ext] = struct{}{}
		}
		for _, ext := range c2.Scanner.Extensions {
			extSet[ext] = struct{}{}
		}
		result.Scanner.Extensions = make([]string, 0, len(extSet))
		for ext := range extSet {
			result.Scanner.Extensions = append(result.Scanner.Extensions, ext)
		}
	}

	return &result
}