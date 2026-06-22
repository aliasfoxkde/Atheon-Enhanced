package core

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed patterns.bundle
var embeddedBundle []byte

type PatternDef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Match    string `json:"match"`
	Enabled  bool   `json:"enabled"`
}

type bundlePattern struct {
	name     string
	category string
	match    string
	enabled  bool
	re       *regexp.Regexp
}

func (p *bundlePattern) Name() string             { return p.name }
func (p *bundlePattern) Category() string         { return p.category }
func (p *bundlePattern) Matches(line string) bool { return p.enabled && p.re.MatchString(line) }
func (p *bundlePattern) Enabled() bool            { return p.enabled }
func (p *bundlePattern) SetEnabled(enabled bool)  { p.enabled = enabled }

type categoryScanner struct {
	combined *regexp.Regexp
	patterns []Pattern
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var (
	allPatterns          []*bundlePattern
	activeScanners       []categoryScanner
	activeCategoryFilter []string // nil = all categories; preserved across rebuildActiveScanners
)

func init() {
	data := embeddedBundle
	home, _ := os.UserHomeDir()
	if b, err := os.ReadFile(filepath.Join(home, ".atheon", "patterns.bundle")); err == nil {
		data = b
	}
	if err := loadBundle(data); err != nil {
		fmt.Fprintf(os.Stderr, "atheon: bundle load failed: %v\n", err)
	}
	SetActiveCategories(nil)

	// Load pattern state after bundle is loaded
	if err := InitializePatternState(); err != nil {
		// Non-fatal error, just log warning
		fmt.Fprintf(os.Stderr, "atheon: pattern state initialization failed: %v\n", err)
	}
}

func loadBundle(data []byte) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer r.Close()

	var defs []PatternDef
	if err := json.NewDecoder(r).Decode(&defs); err != nil {
		return err
	}

	var external []Pattern
	for _, p := range registry {
		if _, ok := p.(*bundlePattern); !ok {
			external = append(external, p)
		}
	}
	registry = nil
	allPatterns = nil

	for _, def := range defs {
		re, err := regexp.Compile(def.Match)
		if err != nil {
			fmt.Fprintf(os.Stderr, "atheon: skipping %q: %v\n", def.Name, err)
			continue
		}
		bp := &bundlePattern{name: def.Name, category: def.Category, match: def.Match, enabled: def.Enabled, re: re}
		allPatterns = append(allPatterns, bp)
		Register(bp)
	}

	// Old bundles predate the enabled field; JSON zero-value false means all appear
	// disabled. Detect this and default everything to enabled.
	anyEnabled := false
	for _, p := range allPatterns {
		if p.enabled {
			anyEnabled = true
			break
		}
	}
	if !anyEnabled {
		for _, p := range allPatterns {
			p.enabled = true
		}
	}

	for _, p := range external {
		Register(p)
	}

	return nil
}

func SetActiveCategories(cats []string) {
	activeCategoryFilter = cats

	catSet := map[string]bool{}
	for _, c := range cats {
		catSet[strings.TrimSpace(c)] = true
	}

	byCategory := map[string][]Pattern{}
	for _, p := range allPatterns {
		if !p.enabled {
			continue
		}
		if len(cats) > 0 && !catSet[p.category] {
			continue
		}
		byCategory[p.category] = append(byCategory[p.category], p)
	}
	// Include externally registered non-bundle patterns so Register() callers are scanned
	for _, p := range registry {
		if _, ok := p.(*bundlePattern); ok {
			continue
		}
		cat := p.Category()
		if len(cats) > 0 && !catSet[cat] {
			continue
		}
		byCategory[cat] = append(byCategory[cat], p)
	}

	activeScanners = nil
	for _, patterns := range byCategory {
		// Split bundle patterns (have a match regex) from external patterns (don't)
		var bundlePs, extPs []Pattern
		for _, p := range patterns {
			if _, ok := p.(*bundlePattern); ok {
				bundlePs = append(bundlePs, p)
			} else {
				extPs = append(extPs, p)
			}
		}
		if len(bundlePs) > 0 {
			parts := make([]string, 0, len(bundlePs))
			for _, p := range bundlePs {
				parts = append(parts, "(?:"+p.(*bundlePattern).match+")")
			}
			combined, err := regexp.Compile(strings.Join(parts, "|"))
			if err == nil {
				activeScanners = append(activeScanners, categoryScanner{combined: combined, patterns: bundlePs})
			}
		}
		if len(extPs) > 0 {
			// External patterns have no regex to pre-filter with; use empty (matches all)
			combined := regexp.MustCompile("")
			activeScanners = append(activeScanners, categoryScanner{combined: combined, patterns: extPs})
		}
	}
}

func Categories() []string {
	seen := map[string]bool{}
	var cats []string
	for _, p := range allPatterns {
		if !seen[p.category] {
			seen[p.category] = true
			cats = append(cats, p.category)
		}
	}
	return cats
}

func DownloadBundle() error {
	const url = "https://github.com/HoraDomu/Atheon/releases/latest/download/patterns.bundle"

	// Get current bundle info for comparison
	var oldPatternCount int
	var oldPatterns []string
	for _, p := range allPatterns {
		oldPatternCount++
		oldPatterns = append(oldPatterns, p.name)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Parse new bundle to compare
	var newDefs []PatternDef
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse new bundle: %w", err)
	}
	defer r.Close()
	if err := json.NewDecoder(r).Decode(&newDefs); err != nil {
		return fmt.Errorf("failed to decode new bundle: %w", err)
	}

	// Report changes
	newPatterns := make(map[string]bool)
	for _, def := range newDefs {
		newPatterns[def.Name] = true
	}

	oldPatternSet := make(map[string]bool, len(oldPatterns))
	for _, name := range oldPatterns {
		oldPatternSet[name] = true
	}

	var added, removed []string

	for _, name := range oldPatterns {
		if !newPatterns[name] {
			removed = append(removed, name)
		}
	}
	for _, def := range newDefs {
		if !oldPatternSet[def.Name] {
			added = append(added, def.Name)
		}
	}

	// Print summary
	fmt.Printf("Patterns updated: %d → %d\n", oldPatternCount, len(newDefs))
	if len(added) > 0 {
		fmt.Printf("Added: %d patterns\n", len(added))
		for _, p := range added {
			fmt.Printf("  + %s\n", p)
		}
	}
	if len(removed) > 0 {
		fmt.Printf("Removed: %d patterns\n", len(removed))
		for _, p := range removed {
			fmt.Printf("  - %s\n", p)
		}
	}
	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("No pattern changes detected")
	}

	// Load into memory first; only persist to disk if that succeeds
	if err := loadBundle(data); err != nil {
		return err
	}
	SetActiveCategories(activeCategoryFilter)
	if err := os.WriteFile(filepath.Join(dir, "patterns.bundle"), data, 0o644); err != nil {
		return err
	}
	return nil
}

func EnablePattern(name string) bool {
	for _, p := range allPatterns {
		if p.name == name {
			p.enabled = true
			rebuildRegistry()
			rebuildActiveScanners()
			if err := syncPatternState(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save pattern state: %v\n", err)
			}
			return true
		}
	}
	return false
}

func DisablePattern(name string) bool {
	for _, p := range allPatterns {
		if p.name == name {
			p.enabled = false
			rebuildRegistry()
			rebuildActiveScanners()
			if err := syncPatternState(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save pattern state: %v\n", err)
			}
			return true
		}
	}
	return false
}

func SetPatternEnabled(name string, enabled bool) bool {
	for _, p := range allPatterns {
		if p.name == name {
			p.enabled = enabled
			rebuildActiveScanners()
			return true
		}
	}
	return false
}

func ListDisabledPatterns() []string {
	var disabled []string
	for _, p := range allPatterns {
		if !p.enabled {
			disabled = append(disabled, p.name)
		}
	}
	return disabled
}

func ListEnabledPatterns() []string {
	var enabled []string
	for _, p := range allPatterns {
		if p.enabled {
			enabled = append(enabled, p.name)
		}
	}
	return enabled
}

func rebuildActiveScanners() {
	SetActiveCategories(activeCategoryFilter)
}

// EnableAllPatterns enables every pattern in the bundle, overriding any prior disable calls.
func EnableAllPatterns() {
	for _, p := range allPatterns {
		p.enabled = true
	}
	rebuildActiveScanners()
}

// rebuildRegistry rebuilds the registry from allPatterns, respecting enabled state
func rebuildRegistry() {
	registry = nil
	for _, p := range allPatterns {
		if p.enabled {
			Register(p)
		}
	}
}
