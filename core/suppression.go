package core

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

// Baseline represents a suppression baseline file.
type Baseline struct {
	Version  string            `yaml:"version"`
	Findings []BaselineFinding `yaml:"findings"`
}

// BaselineFinding represents a single suppressed finding.
type BaselineFinding struct {
	PatternID string `yaml:"pattern_id"`
	File      string `yaml:"file"`
	Line      int    `yaml:"line"`
	Hash      string `yaml:"hash,omitempty"`
}

// BaselineMatcher tracks suppressed findings and filters them out.
type BaselineMatcher struct {
	baseline        *Baseline
	findings        map[string]bool
	totalCount      int
	suppressedCount int
}

// NewBaselineMatcher creates a new baseline matcher from a baseline file.
func NewBaselineMatcher(path string) (*BaselineMatcher, error) {
	bm := &BaselineMatcher{
		findings: make(map[string]bool),
	}

	data, err := os.ReadFile(path) //nolint:gosec // Security scanner intentionally reads user-specified paths
	if err != nil {
		if os.IsNotExist(err) {
			// No baseline file - nothing to suppress
			return bm, nil
		}
		return nil, err
	}

	var baseline Baseline
	if err := yaml.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	bm.baseline = &baseline

	// Index findings by key for fast lookup
	for _, f := range baseline.Findings {
		key := baselineKey(f.PatternID, f.File, f.Line)
		bm.findings[key] = true
	}

	return bm, nil
}

// baselineKey generates a unique key for a finding.
func baselineKey(patternID, file string, line int) string {
	return patternID + "|" + file + "|" + strconv.Itoa(line)
}

// IsSuppressed checks if a finding should be suppressed.
func (bm *BaselineMatcher) IsSuppressed(f Finding) bool {
	bm.totalCount++
	key := baselineKey(f.Pattern, f.File, f.Line)
	if bm.findings[key] {
		bm.suppressedCount++
		return true
	}

	// Also check with normalized file path
	normalized := normalizePath(f.File)
	for k := range bm.findings {
		parts := strings.Split(k, "|")
		if len(parts) == 3 {
			normBaseline := normalizePath(parts[1])
			if parts[0] == f.Pattern && normBaseline == normalized && parts[2] == strconv.Itoa(f.Line) {
				bm.suppressedCount++
				return true
			}
		}
	}

	return false
}

// FilterFindings removes suppressed findings from a slice.
func (bm *BaselineMatcher) FilterFindings(findings []Finding) []Finding {
	var filtered []Finding
	for _, f := range findings {
		if !bm.IsSuppressed(f) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// Stats returns suppression statistics.
func (bm *BaselineMatcher) Stats() BaselineStats {
	return BaselineStats{
		Total:      bm.totalCount,
		Suppressed: bm.suppressedCount,
		Active:     bm.totalCount - bm.suppressedCount,
	}
}

// BaselineStats contains statistics about baseline filtering.
type BaselineStats struct {
	Total      int `json:"total"`
	Suppressed int `json:"suppressed"`
	Active     int `json:"active"`
}

// LoadDefaultBaseline attempts to load .atheon-baseline.yaml from common locations.
func LoadDefaultBaseline() (*BaselineMatcher, error) {
	// Try common locations
	locations := []string{
		".atheon-baseline.yaml",
		".atheon-baseline.yml",
		filepath.Join(".atheon", "baseline.yaml"),
		filepath.Join(".atheon", "baseline.yml"),
	}

	for _, loc := range locations {
		bm, err := NewBaselineMatcher(loc)
		if err == nil {
			return bm, nil
		}
	}

	// Return empty matcher if no baseline found
	return &BaselineMatcher{
		findings: make(map[string]bool),
	}, nil
}

// CreateBaselineFile generates a baseline file from current findings.
func CreateBaselineFile(findings []Finding, path string) error {
	baseline := Baseline{
		Version: "1.0",
	}

	for _, f := range findings {
		baseline.Findings = append(baseline.Findings, BaselineFinding{
			PatternID: f.Pattern,
			File:      f.File,
			Line:      f.Line,
			Hash:      f.Fingerprint,
		})
	}

	data, err := yaml.Marshal(baseline)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644) //nolint:gosec // Baseline files need to be readable by owner; 0o644 is standard for user config files
}

// normalizePath normalizes a file path for comparison.
func normalizePath(p string) string {
	// Convert to forward slashes and normalize
	p = filepath.ToSlash(p)
	// Remove trailing slashes
	p = strings.TrimRight(p, "/")
	return p
}
