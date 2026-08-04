package scanner

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Supported extensions for code scanning
var supportedExtensions = []string{
	".go", ".py", ".ts", ".js", ".java", ".c", ".cpp", ".rs", ".rb",
}

// Options configures Scanner behavior
type Options struct {
	// IncludePatterns are glob patterns for files to include
	IncludePatterns []string
	// ExcludePatterns are glob patterns for files to exclude
	ExcludePatterns []string
	// MaxFileSize is the maximum file size in bytes to consider
	MaxFileSize int64
	// SkipUnchanged enables incremental scanning - skip files that haven't changed
	SkipUnchanged bool
	// CacheDir specifies where to store the scan cache for incremental scanning
	CacheDir string
}

// Scanner discovers code files for analysis
type Scanner struct {
	Options Options
}

// NewScanner creates a new Scanner with the given options
func NewScanner(opts Options) *Scanner {
	return &Scanner{Options: opts}
}

// ScanDir recursively finds all code files in the given directory
// Returns a list of absolute file paths matching supported extensions
func (s *Scanner) ScanDir(root string) ([]string, error) {
	if root == "" {
		return nil, errors.New("root path cannot be empty")
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root: %w", err)
	}

	if !info.IsDir() {
		return nil, errors.New("root must be a directory")
	}

	var files []string

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't access
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			// Skip common non-code directories
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == "__pycache__" || name == ".pytest_cache" ||
				name == "dist" || name == "build" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be included
		if !s.shouldInclude(path) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// ScanFiles filters the given paths to only those with supported extensions
func (s *Scanner) ScanFiles(paths []string) ([]string, error) {
	var valid []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			// Skip files that can't be accessed
			continue
		}

		if info.IsDir() {
			continue
		}

		if !s.hasSupportedExtension(path) {
			continue
		}

		valid = append(valid, path)
	}

	return valid, nil
}

// SupportedExtensions returns the list of supported file extensions
func (s *Scanner) SupportedExtensions() []string {
	result := make([]string, len(supportedExtensions))
	copy(result, supportedExtensions)
	return result
}

// shouldInclude determines if a file should be included in scanning
func (s *Scanner) shouldInclude(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Check if extension is supported
	if !s.hasSupportedExtension(path) {
		return false
	}

	// Check max file size
	if s.Options.MaxFileSize > 0 {
		info, err := os.Stat(path)
		if err != nil || info.Size() > s.Options.MaxFileSize {
			return false
		}
	}

	// Validate Go files using go/parser
	if ext == ".go" {
		if !s.validateGoFile(path) {
			return false
		}
	}

	return true
}

// hasSupportedExtension checks if a file has a supported extension
func (s *Scanner) hasSupportedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range supportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// validateGoFile uses go/parser to validate Go source files
func (s *Scanner) validateGoFile(path string) bool {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	return err == nil
}

// SupportedExtensionsAll returns all supported extensions (package-level)
func SupportedExtensionsAll() []string {
	result := make([]string, len(supportedExtensions))
	copy(result, supportedExtensions)
	return result
}

// ScanCache stores file modification times and hashes for incremental scanning
type ScanCache struct {
	Entries map[string]CacheEntry `json:"entries"`
}

// CacheEntry stores a single file's scan metadata
type CacheEntry struct {
	ModTime  int64  `json:"mod_time"`
	FileSize int64  `json:"file_size"`
	Hash     string `json:"hash"`
}

// LoadCache loads a scan cache from the given path
func LoadCache(cachePath string) (*ScanCache, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ScanCache{Entries: make(map[string]CacheEntry)}, nil
		}
		return nil, err
	}
	var cache ScanCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

// Save saves the scan cache to the given path
func (c *ScanCache) Save(cachePath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0644)
}

// ShouldScan checks if a file should be scanned based on the cache
// Returns true if the file should be scanned, false if it can be skipped
func (c *ScanCache) ShouldScan(path string, info os.FileInfo) bool {
	entry, exists := c.Entries[path]
	if !exists {
		return true // new file or not in cache
	}
	// Check modification time and size
	if info.Size() != entry.FileSize {
		return true // size changed
	}
	if info.ModTime().Unix() != entry.ModTime {
		return true // mod time changed
	}
	return false // unchanged
}

// Update updates the cache entry for a file
func (c *ScanCache) Update(path string, info os.FileInfo, hash string) {
	c.Entries[path] = CacheEntry{
		ModTime:  info.ModTime().Unix(),
		FileSize: info.Size(),
		Hash:     hash,
	}
}

// FileHash computes a fast hash of a file for change detection
func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// For large files, only hash the first and last 4KB + size
	hash := sha256.New()
	if len(data) > 8192 {
		hash.Write(data[:4096])
		hash.Write(data[len(data)-4096:])
		hash.Write([]byte(fmt.Sprintf("%d", len(data))))
	} else {
		hash.Write(data)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetCacheDir returns the default cache directory for Atheon
func GetCacheDir() string {
	// Use standard cache location
	home, err := os.UserCacheDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, "atheon")
}

// CachePath returns the path to the scan cache for a given root
func CachePath(root string) string {
	// Create a hash of the root path for unique cache per-project
	hash := sha256.Sum256([]byte(root))
	cacheDir := GetCacheDir()
	return filepath.Join(cacheDir, fmt.Sprintf("scan-%x.cache", hash[:8]))
}

// ScanDirIncremental scans a directory with incremental scanning support
// Files that haven't changed since last scan are skipped
func (s *Scanner) ScanDirIncremental(root string) ([]string, error) {
	if root == "" {
		return nil, errors.New("root path cannot be empty")
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root: %w", err)
	}

	if !info.IsDir() {
		return nil, errors.New("root must be a directory")
	}

	cachePath := CachePath(root)
	cache, err := LoadCache(cachePath)
	if err != nil {
		// If we can't load cache, scan everything
		cache = &ScanCache{Entries: make(map[string]CacheEntry)}
	}

	var files []string
	var changedCount, unchangedCount int

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == "__pycache__" || name == ".pytest_cache" ||
				name == "dist" || name == "build" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			return nil
		}

		if !s.shouldInclude(path) {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return nil
		}

		// If SkipUnchanged is disabled, always scan
		if !s.Options.SkipUnchanged {
			hash, _ := FileHash(path)
			cache.Update(path, fileInfo, hash)
			changedCount++
			files = append(files, path)
			return nil
		}

		// Check if file should be scanned based on cache
		if cache.ShouldScan(path, fileInfo) {
			// File is new or changed - scan it
			hash, _ := FileHash(path)
			cache.Update(path, fileInfo, hash)
			changedCount++
			files = append(files, path)
		} else {
			// File is unchanged - skip it
			unchangedCount++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Save updated cache if SkipUnchanged is enabled
	if s.Options.SkipUnchanged {
		cacheDir := GetCacheDir()
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create cache dir: %v\n", err)
		} else if err := cache.Save(cachePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save cache: %v\n", err)
		} else {
			fmt.Printf("Incremental scan: %d files changed, %d unchanged (cache saved to %s)\n", changedCount, unchangedCount, cachePath)
		}
	}

	return files, nil
}
