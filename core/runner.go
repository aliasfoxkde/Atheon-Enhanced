package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	".terraform": true, "dist": true, "build": true, "__pycache__": true,
}

var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".pdf": true, ".zip": true, ".tar": true, ".gz": true,
	".exe": true, ".bin": true, ".so": true, ".dylib": true,
}

func loadIgnorePatternsMatcher(root string) []*ignoreMatcher {
	var matchers []*ignoreMatcher
	for _, name := range []string{".atheonignore", ".gitignore"} {
		m, _ := compileIgnoreFile(filepath.Join(root, name))
		if m != nil {
			matchers = append(matchers, m)
		}
	}
	return matchers
}

func isIgnored(path string, matchers []*ignoreMatcher) bool {
	clean := filepath.ToSlash(path)
	for _, m := range matchers {
		if m.matchesPath(clean) {
			return true
		}
	}
	return false
}

func ScanFile(path string) ([]Finding, *Stats, error) {
	start := time.Now()
	// Respect .atheonignore and .gitignore for files under the working directory,
	// so that `atheon file.go` and `atheon .` agree on what gets scanned.
	if absPath, err := filepath.Abs(path); err == nil {
		if root, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(root, absPath); err == nil && !strings.HasPrefix(rel, "..") {
				if matchers := loadIgnorePatternsMatcher(root); isIgnored(filepath.ToSlash(rel), matchers) {
					return []Finding{}, nil, nil
				}
			}
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	findings := scanLines(string(data), path)
	return findings, &Stats{
		Files:     1,
		Bytes:     int64(len(data)),
		ElapsedMs: time.Since(start).Milliseconds(),
	}, nil
}

func ScanDir(root string) ([]Finding, *Stats, error) {
	start := time.Now()
	ignoreMatcher := loadIgnorePatternsMatcher(root)
	var paths []string

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			// Don't SkipDir for user ignore rules — walk the dir and check files
			// individually so negation rules (e.g. !dist/keep.yaml) can un-ignore
			// specific files inside an otherwise-ignored directory.
			return nil
		}
		if isIgnored(rel, ignoreMatcher) {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !binaryExts[ext] {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	results := make([][]Finding, len(paths))
	sizes := make([]int64, len(paths))
	scanned := make([]bool, len(paths))
	scanErrors := make([]error, len(paths))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	workers := max(8, runtime.NumCPU()*8)
	sem := make(chan struct{}, workers)

	for i, p := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			data, err := os.ReadFile(p)
			if err != nil {
				errMu.Lock()
				scanErrors[i] = err
				errMu.Unlock()
				return
			}
			results[i] = scanLines(string(data), p)
			sizes[i] = int64(len(data))
			scanned[i] = true
		}(i, p)
	}
	wg.Wait()

	var findings []Finding
	var totalBytes int64
	var filesScanned int
	var errs []error
	for i := range results {
		if scanned[i] {
			filesScanned++
		}
		if scanErrors[i] != nil {
			errs = append(errs, scanErrors[i])
		}
		findings = append(findings, results[i]...)
		totalBytes += sizes[i]
	}

	return findings, &Stats{
		Files:     filesScanned,
		Bytes:     totalBytes,
		ElapsedMs: time.Since(start).Milliseconds(),
		Errors:    errs,
	}, nil
}

func ScanEnv() []Finding {
	var findings []Finding
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		for _, cs := range activeScanners {
			if !cs.combined.MatchString(val) {
				continue
			}
			for _, p := range cs.patterns {
				if p.Matches(val) {
					findings = append(findings, Finding{
						Pattern: p.Name(),
						File:    "env:" + key,
						Content: val,
					})
				}
			}
		}
	}
	return findings
}

func ScanString(content, source string) []Finding {
	return scanLines(content, source)
}

func scanLines(content, file string) []Finding {
	var findings []Finding
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "atheon:ignore") {
			continue
		}
		for _, cs := range activeScanners {
			if !cs.combined.MatchString(line) {
				continue
			}
			for _, p := range cs.patterns {
				if p.Matches(line) {
					findings = append(findings, Finding{
						Pattern: p.Name(),
						File:    file,
						Line:    i + 1,
						Content: strings.TrimSpace(line),
					})
				}
			}
		}
	}
	return findings
}
