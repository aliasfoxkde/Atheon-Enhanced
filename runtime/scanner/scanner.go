package scanner

import (
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
