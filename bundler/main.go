package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type patternFile struct {
	Name     string `yaml:"name"`
	Match    string `yaml:"match"`
	Enabled  *bool  `yaml:"enabled,omitempty"`
	Severity string `yaml:"severity,omitempty"`
}

type patternDef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Match    string `json:"match"`
	Enabled  bool   `json:"enabled"`
	Severity string `json:"severity,omitempty"`
}

// bundleWalkErr is a sentinel for WalkDir errors so the loop above still
// returns errors but tests can introspect.
type bundleWalkErr struct {
	path string
	err  error
}

func (e *bundleWalkErr) Error() string { return e.path + ": " + e.err.Error() }
func (e *bundleWalkErr) Unwrap() error { return e.err }

// bundleToWriter bundles the community directory and writes the result to out.
// Splitting this out from bundle() lets tests pass a failing writer.
func bundleToWriter(communityDir string, out io.Writer) (int, error) {
	defs, err := walkPatterns(communityDir)
	if err != nil {
		return 0, err
	}

	jsonBytes, err := json.Marshal(defs)
	if err != nil {
		return 0, err
	}

	if err := writeGzipped(out, jsonBytes); err != nil {
		return 0, err
	}
	return len(defs), nil
}

func bundle(communityDir, outPath string) (int, error) {
	var buf bytes.Buffer
	n, err := bundleToWriter(communityDir, &buf)
	if err != nil {
		return 0, err
	}
	// Atomic write: tempfile + rename so a crash mid-write doesn't leave
	// a partial patterns.bundle that loadBundle then rejects. The core
	// package has the same helper (core.atomicWriteFile) but bundler is a
	// separate main package and can't import it; the two implementations
	// are intentionally identical so future fixes should be mirrored.
	if err := atomicWriteFile(outPath, buf.Bytes(), 0o600); err != nil {
		return 0, err
	}
	return n, nil
}

// atomicWriteFile writes data to path via tempfile-then-rename. See
// core/atomic_file.go for the full rationale and the test that guards it.
func atomicWriteFile(path string, data []byte, perm os.FileMode) (retErr error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("atomic write: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if retErr != nil {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("atomic write: write data: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("atomic write: fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("atomic write: close: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("atomic write: chmod: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic write: rename: %w", err)
	}
	return nil
}

// walkPatterns walks communityDir and returns all parsed pattern definitions.
//
// Skip policy: a file that fails to parse, is missing required fields, has
// a duplicate name, or has an invalid regex is logged to stderr and skipped
// rather than aborting the whole build. This mirrors loadBundle's runtime
// behaviour: a single malformed pattern shouldn't poison the whole bundle.
func walkPatterns(communityDir string) ([]patternDef, error) {
	var defs []patternDef
	seen := make(map[string]string) // name → first file path
	skip := func(path, why string, args ...any) {
		fmt.Fprintf(os.Stderr, "warn: %s: %s\n", path, fmt.Sprintf(why, args...))
	}
	err := filepath.WalkDir(communityDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return &bundleWalkErr{path: path, err: err}
		}
		var pf patternFile
		if err := yaml.Unmarshal(data, &pf); err != nil {
			skip(path, "yaml parse error: %v (file skipped)", err)
			return nil
		}
		if pf.Name == "" || pf.Match == "" {
			skip(path, "missing name or match (file skipped)")
			return nil
		}
		if strings.ContainsAny(pf.Name, " \t") {
			skip(path, "pattern name %q must not contain whitespace (file skipped)", pf.Name)
			return nil
		}
		if first, dup := seen[pf.Name]; dup {
			skip(path, "duplicate pattern name %q (first defined in %s; this file skipped)", pf.Name, first)
			return nil
		}
		seen[pf.Name] = path
		if _, err := regexp.Compile(pf.Match); err != nil {
			skip(path, "invalid regex for %q: %v (file skipped)", pf.Name, err)
			return nil
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
			Severity: pf.Severity,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return defs, nil
}

// writeGzipped gzip-encodes jsonBytes into out. Accepts io.Writer so tests
// can substitute a failing writer to exercise the gzip error paths.
func writeGzipped(out io.Writer, jsonBytes []byte) error {
	gz := gzip.NewWriter(out)
	if _, err := gz.Write(jsonBytes); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

// run executes the bundler with the given args and returns the exit code.
// This is separated from main() so tests can call it without os.Exit
// terminating the test process.
func run(args []string) int {
	communityDir := "community"
	outPath := filepath.Join("core", "patterns.bundle")
	if len(args) > 0 {
		communityDir = args[0]
	}
	if len(args) > 1 {
		outPath = args[1]
	}

	n, err := bundle(communityDir, outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Printf("bundled %d patterns → %s\n", n, outPath)
	return 0
}
