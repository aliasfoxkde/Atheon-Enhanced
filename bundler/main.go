package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type patternFile struct {
	Name    string `yaml:"name"`
	Match   string `yaml:"match"`
	Enabled *bool  `yaml:"enabled,omitempty"`
}

type patternDef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Match    string `json:"match"`
}

func main() {
	communityDir := "community"
	outPath := filepath.Join("core", "patterns.bundle")
	if len(os.Args) > 1 {
		communityDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}

	var defs []patternDef
	err := filepath.WalkDir(communityDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var pf patternFile
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if pf.Name == "" || pf.Match == "" {
			return fmt.Errorf("%s: missing name or match", path)
		}
		category := filepath.Base(filepath.Dir(path))
		enabled := true
		if pf.Enabled != nil {
			enabled = *pf.Enabled
		}
		defs = append(defs, patternDef{
			Name:     pf.Name,
			Category: category,
			Match:    pf.Match,
			Enabled:  enabled,
		})
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	jsonBytes, err := json.Marshal(defs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(jsonBytes)
	gz.Close()

	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("bundled %d patterns → %s\n", len(defs), outPath)
}
