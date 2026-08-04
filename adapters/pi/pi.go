// Package pi implements the Atheon runtime adapter for the Pi CLI.
// It registers an "atheon" tool that allows Pi to scan code, analyze patterns,
// and return diagnostics directly within the Pi workflow.
package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aliasfoxkde/Atheon/core"
)

// Config holds adapter configuration.
type Config struct {
	// Categories filters scan to specific categories (nil = all).
	Categories []string
	// SeverityThreshold only shows findings at or above this level.
	SeverityThreshold string
}

// DefaultConfig returns the default adapter configuration.
func DefaultConfig() Config {
	return Config{}
}

// Adapter implements the Atheon runtime adapter for Pi.
// It satisfies the runtime.Adapter interface.
type Adapter struct {
	name    string
	version string
	config  Config
}

// New creates a new Pi adapter with the given config.
func New(cfg Config) *Adapter {
	return &Adapter{
		name:    "pi",
		version: "1.0.0",
		config:  cfg,
	}
}

// Name returns the adapter name.
func (a *Adapter) Name() string {
	return a.name
}

// Version returns the adapter version.
func (a *Adapter) Version() string {
	return a.version
}

// Initialize sets up the adapter with configuration.
func (a *Adapter) Initialize(cfg interface{}) error {
	// Pi adapter doesn't need special initialization
	return nil
}

// Validate checks if the adapter is properly configured.
func (a *Adapter) Validate() error {
	return nil
}

// Analyze runs analysis on the provided input.
// input should be a string (path) or []string (multiple paths).
func (a *Adapter) Analyze(input interface{}) (*Response, error) {
	ctx := context.Background()

	var paths []string

	switch v := input.(type) {
	case string:
		if v != "" {
			paths = []string{v}
		}
	case []string:
		paths = v
	case []interface{}:
		for _, p := range v {
			if s, ok := p.(string); ok {
				paths = append(paths, s)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths provided")
	}

	// Resolve and validate paths
	var scanPaths []string
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if info.IsDir() {
			scanPaths = append(scanPaths, abs)
		} else if info.Mode().IsRegular() {
			scanPaths = append(scanPaths, abs)
		}
	}

	if len(scanPaths) == 0 {
		return nil, fmt.Errorf("no valid paths to scan")
	}

	// Apply category filter if configured
	if len(a.config.Categories) > 0 {
		core.SetActiveCategories(a.config.Categories)
	}

	// Scan all paths
	var allFindings []core.Finding
	for _, path := range scanPaths {
		info, _ := os.Stat(path)
		if info.IsDir() {
			findings, _, err := core.ScanDir(ctx, path, core.ScanOpts{})
			if err == nil {
				allFindings = append(allFindings, findings...)
			}
		} else {
			findings, _, err := core.ScanFile(ctx, path)
			if err == nil {
				allFindings = append(allFindings, findings...)
			}
		}
	}

	// Filter by severity if configured
	if a.config.SeverityThreshold != "" {
		allFindings = filterBySeverity(allFindings, a.config.SeverityThreshold)
	}

	// Build response
	resp := &Response{
		Findings:  allFindings,
		Summary:   buildSummary(allFindings),
		RiskScore: computeRiskScore(allFindings),
	}

	return resp, nil
}

// Response represents the Atheon analysis response for Pi.
type Response struct {
	Findings  []core.Finding  `json:"findings"`
	Summary   Summary         `json:"summary"`
	RiskScore *core.RiskScore `json:"risk_score"`
}

// Summary represents finding statistics.
type Summary struct {
	Total      int            `json:"total"`
	Critical   int            `json:"critical"`
	High       int            `json:"high"`
	Medium     int            `json:"medium"`
	Low        int            `json:"low"`
	Infos      int            `json:"info"`
	ByCategory map[string]int `json:"by_category"`
}

// filterBySeverity filters findings by severity threshold.
func filterBySeverity(findings []core.Finding, threshold string) []core.Finding {
	severityOrder := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
		"info":     0,
	}

	minLevel, ok := severityOrder[threshold]
	if !ok {
		return findings
	}

	var filtered []core.Finding
	for _, f := range findings {
		if level, ok := severityOrder[f.Severity]; ok && level >= minLevel {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// buildSummary creates a summary from findings.
func buildSummary(findings []core.Finding) Summary {
	s := Summary{
		ByCategory: make(map[string]int),
	}
	s.Total = len(findings)
	for _, f := range findings {
		switch f.Severity {
		case "critical":
			s.Critical++
		case "high":
			s.High++
		case "medium":
			s.Medium++
		case "low":
			s.Low++
		case "info":
			s.Infos++
		}
		if f.Category != "" {
			s.ByCategory[f.Category]++
		}
	}
	return s
}

// computeRiskScore calculates risk score from findings.
func computeRiskScore(findings []core.Finding) *core.RiskScore {
	if len(findings) == 0 {
		return &core.RiskScore{
			Score: 0,
			Level: "none",
		}
	}

	// Simple risk scoring
	var total int
	severityScores := map[string]int{
		"critical": 40,
		"high":     30,
		"medium":   20,
		"low":      10,
		"info":     5,
	}

	for _, f := range findings {
		if score, ok := severityScores[f.Severity]; ok {
			total += score
		}
	}

	level := core.RiskLevelLow
	if s := float64(total); s > 100 {
		level = core.RiskLevelCritical
	} else if s > 70 {
		level = core.RiskLevelHigh
	} else if s > 30 {
		level = core.RiskLevelMedium
	}

	return &core.RiskScore{
		Score: total,
		Level: level,
	}
}

// HandleTool is the main entry point for Pi tool invocations.
// It handles the "atheon" tool registration in Pi.
func (a *Adapter) HandleTool(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var req ToolRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	switch req.Command {
	case "scan":
		result, err := a.Analyze(req.Paths)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case "list":
		if req.ListType == "categories" {
			cats := core.Categories()
			return json.Marshal(map[string]interface{}{"categories": cats})
		}
		patterns := core.ListEnabledPatterns()
		return json.Marshal(map[string]interface{}{"patterns": patterns})

	case "version":
		return json.Marshal(map[string]string{
			"version": "1.0.0",
		})

	default:
		return nil, fmt.Errorf("unknown command: %s", req.Command)
	}
}

// ToolRequest represents a Pi tool request.
type ToolRequest struct {
	Command  string   `json:"command"`
	Paths    []string `json:"paths"`
	ListType string   `json:"list_type,omitempty"`
	Category string   `json:"category,omitempty"`
}
