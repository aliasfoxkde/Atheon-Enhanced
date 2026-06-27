package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
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
	".idea": true, ".vscode": true, ".svn": true, ".hg": true, ".cache": true,
}

var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".pdf": true, ".zip": true, ".tar": true, ".gz": true,
	".exe": true, ".bin": true, ".so": true, ".dylib": true,
}

func init() {
	// ATHEON_SKIP_DIRS: comma-separated list of directory names to skip
	if v := os.Getenv("ATHEON_SKIP_DIRS"); v != "" {
		for _, d := range strings.Split(v, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				skipDirs[d] = true
			}
		}
	}
	// ATHEON_BINARY_EXTS: comma-separated list of file extensions to skip
	if v := os.Getenv("ATHEON_BINARY_EXTS"); v != "" {
		for _, e := range strings.Split(v, ",") {
			e = strings.TrimSpace(e)
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			if e != "" {
				binaryExts[e] = true
			}
		}
	}
}

// maxFileSize is the maximum file size to scan (10MB default).
// Files larger than this are skipped to prevent memory exhaustion.
const maxFileSize = 10 * 1024 * 1024

// scanBinarySniffBytes is the head-of-file slice read to detect binaries
// the extension allowlist misses (extensionless blobs, files saved
// without a final dot). A NUL byte in the first 8 KiB is the de-facto
// heuristic used by `grep -I` and most editors.
const scanBinarySniffBytes = 8 * 1024

// ErrFileTooLarge is returned by readFileCapped when a file exceeds the
// configured size cap. Distinct from a plain read error so callers can
// surface it as a per-file skip in Stats.Errors without flagging it as
// a generic I/O failure.
var ErrFileTooLarge = errors.New("core: file exceeds configured max bytes")

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

// readFileCapped reads path and returns its contents up to maxBytes. If the
// file's size exceeds maxBytes it returns ErrFileTooLarge WITHOUT reading
// the body — that's the property that bounds memory: a 10 GiB log can't
// be OOM'd into the scanner just because the caller asked for it. Files of
// exactly maxBytes are read in full (boundary inclusive). A zero-byte
// file returns ([]byte{}, nil).
//
// Extracted from ScanFile so ScanDir workers can share the same guard;
// before this helper existed ScanDir's per-file goroutines skipped the
// size check entirely, which made the cap a no-op for directory scans.
func readFileCapped(path string, maxBytes int64) ([]byte, error) {
	// Canonicalise before the size check so that a symlink to a huge file
	// (e.g. scan_root/some_link -> /proc/kcore) is sized correctly and rejected
	// rather than OOMing the process. This closes the TOCTOU race where a file
	// that passes the Stat size check is replaced by a symlink to a huge target
	// before ReadFile runs.
	canon, err := filepath.EvalSymlinks(path)
	if err != nil {
		// Broken symlink — let Open report the error.
		canon = path
	}
	info, err := os.Stat(canon)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("%w: %s is %d bytes (limit %d)", ErrFileTooLarge, path, info.Size(), maxBytes)
	}
	return os.ReadFile(canon)
}

// ScanFile reads a single file and reports every Finding produced by the
// currently active patterns. It honors .atheonignore and .gitignore when
// the file lives under the current working directory. The returned
// *Stats describes the read; findings are nil and stats are nil only when
// the file is ignored.
//
// The context controls read-side cancellation: if ctx is canceled before
// the read completes, ScanFile returns ctx.Err().
func ScanFile(ctx context.Context, path string) ([]Finding, *Stats, error) {
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
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	data, err := readFileCapped(path, maxFileSize)
	if err != nil {
		if errors.Is(err, ErrFileTooLarge) {
			info, _ := os.Stat(path)
			var size int64
			if info != nil {
				size = info.Size()
			}
			slog.Warn("skipping file exceeds size limit", "path", path, "size", size, "limit", maxFileSize)
			return []Finding{}, &Stats{Files: 1, Bytes: size}, nil
		}
		return nil, nil, err
	}
	findings := scanLines(ctx, string(data), path)
	return findings, &Stats{
		Files:     1,
		Bytes:     int64(len(data)),
		ElapsedMs: time.Since(start).Milliseconds(),
	}, nil
}

// ScanOpts tunes ScanDir behaviour for contexts where the defaults aren't
// right. The zero value preserves the historical behaviour, so callers
// that don't care can pass ScanOpts{} (or, equivalently, no opts by
// passing a fresh struct literal).
type ScanOpts struct {
	// NoFollowSymlinks, when true, skips every entry whose fs.DirEntry
	// reports fs.ModeSymlink — including dangling symlinks and symlinks
	// that resolve to files outside the scan root. The CLI keeps the
	// historical "follow symlinks" behaviour by default (preserves
	// user expectations); the MCP server defaults this to true because
	// agents scanning untrusted trees shouldn't escape the boundary.
	NoFollowSymlinks bool

	// MaxFileSize caps the number of bytes a single file may contribute
	// to the scan. Files larger than this are skipped (Stats.Errors is
	// populated with the ErrFileTooLarge sentinel). Zero means use the
	// package-level maxFileSize.
	MaxFileSize int64
}

// ScanDir walks root and scans every non-binary, non-ignored file in
// parallel using one worker per CPU (up to a sensible cap). It honors
// .atheonignore and .gitignore at root and skips well-known noise
// directories (e.g. .git, node_modules, vendor). The returned *Stats
// counts only the files whose contents were actually scanned.
//
// The context controls worker cancellation: if ctx is canceled mid-walk
// the goroutines exit promptly and ScanDir returns ctx.Err() after
// WaitGroup drains.
func ScanDir(ctx context.Context, root string, opts ScanOpts) ([]Finding, *Stats, error) {
	start := time.Now()
	slog.Debug("scan started", "root", root)
	ignoreMatcher := loadIgnorePatternsMatcher(root)
	maxBytes := opts.MaxFileSize
	if maxBytes <= 0 {
		maxBytes = maxFileSize
	}
	var paths []string
	var walkErrs []error

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Collect walk errors instead of swallowing them. Previously this
		// returned nil with a //nolint:nilerr comment — callers had no way
		// to learn that a permission error or vanished symlink had been
		// silently dropped. Surfacing them in Stats.Errors lets the CLI
		// exit non-zero and the MCP server return a useful error.
		if err != nil {
			walkErrs = append(walkErrs, fmt.Errorf("walk %s: %w", path, err))
			// SkipDir lets us keep walking the rest of the tree after an
			// unreadable directory entry; returning err would abort the
			// entire scan on the first permission error.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		// Symlink guard. WalkDir reports the link itself (not the target)
		// as the DirEntry, so d.Type() carries ModeSymlink. We skip
		// unconditionally — a symlink that escapes the scan root
		// (e.g. /tmp/leak -> /etc/passwd) would otherwise be followed by
		// os.ReadFile and leak content into the findings.
		if opts.NoFollowSymlinks && d.Type()&fs.ModeSymlink != 0 {
			slog.Debug("skipping symlink", "path", path)
			if d.IsDir() {
				return filepath.SkipDir
			}
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
	})
	if walkErr != nil && !errors.Is(walkErr, context.Canceled) && !errors.Is(walkErr, context.DeadlineExceeded) {
		// ctx cancellation is reported separately below; merge any
		// remaining walk errors into the stats so the caller sees them.
		walkErrs = append(walkErrs, walkErr)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	results := make([][]Finding, len(paths))
	sizes := make([]int64, len(paths))
	scanned := make([]bool, len(paths))
	scanErrors := make([]error, len(paths))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	// I/O-bound file reads saturate well below CPU count; cap at 2× CPUs with
	// a minimum of 4 and a ceiling of 64 to avoid overwhelming shared runners.
	workers := min(max(runtime.NumCPU()*2, 4), 64)
	sem := make(chan struct{}, workers)

	for i, p := range paths {
		wg.Add(1)
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			wg.Wait()
			return nil, nil, ctx.Err()
		}
		go func(i int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := ctx.Err(); err != nil {
				return
			}
			data, err := readFileCapped(p, maxBytes)
			if err != nil {
				errMu.Lock()
				scanErrors[i] = err
				errMu.Unlock()
				return
			}
			// Content sniff: the extension allowlist misses extensionless
			// binaries (build artefacts, minified blobs saved without a
			// final dot). The de-facto heuristic — also used by `grep -I` —
			// is a NUL byte in the first 8 KiB. We only check the head
			// to keep the scan cheap; full-file binary detection would
			// double the I/O.
			if len(data) > 0 && bytes.IndexByte(data[:min(len(data), scanBinarySniffBytes)], 0) >= 0 {
				errMu.Lock()
				scanErrors[i] = fmt.Errorf("skipping binary file (NUL byte in first %d bytes): %s", scanBinarySniffBytes, p)
				errMu.Unlock()
				return
			}
			// UTF-16 BOM detection: files encoded as UTF-16 (common on Windows
			// for config files, logs, and exported data) use a byte-order mark
			// that renders as NUL bytes when read as ASCII, and would produce
			// spurious matches on the NUL escape sequences. Detect the two
			// standard BOMs: FE FF (big-endian) and FF FE (little-endian).
			if len(data) >= 2 {
				if data[0] == 0xFE && data[1] == 0xFF {
					errMu.Lock()
					scanErrors[i] = fmt.Errorf("skipping binary file (UTF-16 BE BOM): %s", p)
					errMu.Unlock()
					return
				}
				if data[0] == 0xFF && data[1] == 0xFE {
					errMu.Lock()
					scanErrors[i] = fmt.Errorf("skipping binary file (UTF-16 LE BOM): %s", p)
					errMu.Unlock()
					return
				}
			}
			results[i] = scanLines(ctx, string(data), p)
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
	// Walk errors collected during the tree enumeration come BEFORE the
	// per-file read errors so the listing reads chronologically (what was
	// the tree doing vs what did individual reads see).
	errs = append(walkErrs, errs...)

	slog.Debug("scan complete", "root", root, "files", filesScanned, "findings", len(findings), "errors", len(errs), "elapsed_ms", time.Since(start).Milliseconds())

	return findings, &Stats{
		Files:     filesScanned,
		Bytes:     totalBytes,
		ElapsedMs: time.Since(start).Milliseconds(),
		Errors:    errs,
	}, nil
}

// ScanEnv scans the current process's environment variables for matches
// against the active patterns. Each finding uses "env:KEY" as its File
// and the matching value as its Content; Line is zero.
//
// The context is accepted for symmetry with the other Scan* entry
// points; the implementation checks ctx between iterations and returns
// early if canceled.
func ScanEnv(ctx context.Context) []Finding {
	return scanEnv(ctx, os.Environ())
}

// scanEnv is the inner implementation that accepts an explicit env list.
// Splitting this out lets tests exercise the len(parts) != 2 branch
// without having to mutate the real process environment.
func scanEnv(ctx context.Context, envs []string) []Finding {
	// Hold patternMu.RLock() for the entire scan. The per-pattern
	// bundlePattern.enabled field is mutated by Enable/Disable under the
	// write lock; a copy-by-value snapshot of the activeScanners slice
	// doesn't protect the pointed-to bundlePattern.enabled, so we have to
	// hold the read lock across the inner pattern match too. ScanEnv is
	// bounded by the env var count (typically <100), so holding the lock
	// across the iteration is not a contention concern.
	patternMu.RLock()
	defer patternMu.RUnlock()

	var findings []Finding
	for _, env := range envs {
		if err := ctx.Err(); err != nil {
			return findings
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			// Skip malformed entries and Windows-internal vars like =C:=C:\path
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
						Pattern:     p.Name(),
						File:        "env:" + key,
						Content:     val,
						Severity:    p.Severity(),
						Category:    p.Category(),
						Fingerprint: fmt.Sprintf("%s|%s|0|0", p.Name(), "env:"+key),
					})
				}
			}
		}
	}
	return findings
}

// ScanString scans a string in memory and returns every Finding produced
// by the active patterns. source is recorded as the File on each
// Finding; lines are reported with their 1-indexed line number.
//
// The context is accepted for API symmetry; the scan is in-memory and
// the implementation only checks ctx for cancellation between lines.
func ScanString(ctx context.Context, content, source string) []Finding {
	return scanLines(ctx, content, source)
}

func scanLines(ctx context.Context, content, file string) []Finding {
	// Hold patternMu.RLock() for the entire scan. The activeScanners slice
	// and the per-pattern bundlePattern.enabled fields are mutated by
	// Enable/Disable under the write lock; a snapshot of the slice doesn't
	// protect the pointed-to .enabled, so we hold the read lock across the
	// inner match too. Per-scan contention is bounded — Enable/Disable are
	// human-driven, scans are CPU-bound, so writers wait at worst one scan
	// iteration before getting the lock.
	patternMu.RLock()
	defer patternMu.RUnlock()

	var findings []Finding
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if err := ctx.Err(); err != nil {
			return findings
		}
		if strings.Contains(line, "atheon:ignore") {
			continue
		}
		for _, cs := range activeScanners {
			if !cs.combined.MatchString(line) {
				continue
			}
			for _, p := range cs.patterns {
				if !p.Matches(line) {
					continue
				}
				lineNum := i + 1
				if lineNum == 0 {
					lineNum = 1
				}
				// Column is the 1-indexed byte offset of the match's
				// first byte. Only bundlePattern can supply it (other
				// Pattern implementations don't carry a compiled
				// regex); for them we leave Column == 0 and downstream
				// consumers treat 0 as "unknown". The type-assertion is
				// the cheapest way to gate without widening the public
				// Pattern interface — adding a method would break every
				// external implementer.
				col := 0
				if bp, ok := p.(*bundlePattern); ok {
					if start, _ := bp.matchSpan(line); start >= 0 {
						col = start + 1
					}
				}
				findings = append(findings, Finding{
					Pattern:     p.Name(),
					File:        file,
					Line:        lineNum,
					Column:      col,
					Content:     strings.TrimSpace(line),
					Severity:    p.Severity(),
					Category:    p.Category(),
					Fingerprint: fmt.Sprintf("%s|%s|%d|%d", p.Name(), file, lineNum, col),
				})
			}
		}
	}
	return findings
}
