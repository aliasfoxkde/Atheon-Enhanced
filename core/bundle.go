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
)

//go:embed patterns.bundle
var embeddedBundle []byte

type PatternDef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Match    string `json:"match"`
}

type bundlePattern struct {
	name     string
	category string
	match    string
	re       *regexp.Regexp
}

func (p *bundlePattern) Name() string             { return p.name }
func (p *bundlePattern) Matches(line string) bool { return p.re.MatchString(line) }

type categoryScanner struct {
	combined *regexp.Regexp
	patterns []Pattern
}

var (
	allPatterns    []*bundlePattern
	activeScanners []categoryScanner
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

	registry = nil
	allPatterns = nil

	for _, def := range defs {
		re, err := regexp.Compile(def.Match)
		if err != nil {
			fmt.Fprintf(os.Stderr, "atheon: skipping %q: %v\n", def.Name, err)
			continue
		}
		bp := &bundlePattern{name: def.Name, category: def.Category, match: def.Match, re: re}
		allPatterns = append(allPatterns, bp)
		Register(bp)
	}
	return nil
}

func SetActiveCategories(cats []string) {
	catSet := map[string]bool{}
	for _, c := range cats {
		catSet[strings.TrimSpace(c)] = true
	}

	byCategory := map[string][]Pattern{}
	for _, p := range allPatterns {
		if len(cats) > 0 && !catSet[p.category] {
			continue
		}
		byCategory[p.category] = append(byCategory[p.category], p)
	}

	activeScanners = nil
	for _, patterns := range byCategory {
		parts := make([]string, 0, len(patterns))
		for _, p := range patterns {
			if bp, ok := p.(*bundlePattern); ok {
				parts = append(parts, "(?:"+bp.match+")")
			}
		}
		combined, err := regexp.Compile(strings.Join(parts, "|"))
		if err != nil {
			continue
		}
		activeScanners = append(activeScanners, categoryScanner{
			combined: combined,
			patterns: patterns,
		})
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
	resp, err := http.Get(url) //nolint:gosec
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
	return os.WriteFile(filepath.Join(dir, "patterns.bundle"), data, 0o644)
}
