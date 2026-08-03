package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewScanner(t *testing.T) {
	opts := Options{
		MaxFileSize: 1024 * 1024,
	}
	s := NewScanner(opts)
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
	if s.Options.MaxFileSize != 1024*1024 {
		t.Errorf("expected MaxFileSize 1024*1024, got %d", s.Options.MaxFileSize)
	}
}

func TestSupportedExtensions(t *testing.T) {
	s := NewScanner(Options{})
	exts := s.SupportedExtensions()

	expected := []string{".go", ".py", ".ts", ".js", ".java", ".c", ".cpp", ".rs", ".rb"}
	if len(exts) != len(expected) {
		t.Fatalf("expected %d extensions, got %d", len(expected), len(exts))
	}

	for i, ext := range exts {
		if ext != expected[i] {
			t.Errorf("expected extension %s at index %d, got %s", expected[i], i, ext)
		}
	}
}

func TestSupportedExtensionsAll(t *testing.T) {
	exts := SupportedExtensionsAll()
	expected := []string{".go", ".py", ".ts", ".js", ".java", ".c", ".cpp", ".rs", ".rb"}
	if len(exts) != len(expected) {
		t.Fatalf("expected %d extensions, got %d", len(expected), len(exts))
	}
}

func TestScanFiles_FiltersByExtension(t *testing.T) {
	s := NewScanner(Options{})

	// Create temp files with different extensions
	tmpDir := t.TempDir()

	files := map[string]string{
		"test.go":    "package main\nfunc main() {}",
		"test.py":    "print('hello')",
		"test.ts":    "const x: number = 1;",
		"test.js":    "console.log('hello');",
		"test.java":  "public class Test {}",
		"test.c":     "#include <stdio.h>",
		"test.cpp":   "#include <iostream>",
		"test.rs":    "fn main() {}",
		"test.rb":    "puts 'hello'",
		"test.txt":   "not code",
		"test.md":    "# Title",
	}

	var paths []string
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		paths = append(paths, path)
	}

	valid, err := s.ScanFiles(paths)
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// Should find only code files
	if len(valid) != 9 {
		t.Errorf("expected 9 valid files, got %d: %v", len(valid), valid)
	}
}

func TestScanFiles_SkipsNonExistent(t *testing.T) {
	s := NewScanner(Options{})
	paths := []string{
		"/nonexistent/path.go",
		filepath.Join(t.TempDir(), "real.go"),
	}

	// Create one real file
	realPath := paths[1]
	if err := os.WriteFile(realPath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	valid, err := s.ScanFiles(paths)
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	if len(valid) != 1 {
		t.Errorf("expected 1 valid file, got %d", len(valid))
	}
}

func TestScanDir_Recursive(t *testing.T) {
	s := NewScanner(Options{})

	tmpDir := t.TempDir()

	// Create nested directory structure with code files
	subDir := filepath.Join(tmpDir, "sub", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	files := map[string]string{
		filepath.Join(tmpDir, "root.go"):  "package main",
		filepath.Join(subDir, "nested.go"): "package main",
		filepath.Join(tmpDir, "root.py"):   "print('hello')",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
	}

	result, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(result), result)
	}
}

func TestScanDir_SkipsNonCodeDirs(t *testing.T) {
	s := NewScanner(Options{})

	tmpDir := t.TempDir()

	// Create files in directories that should be skipped
	skipDirs := []string{".git", "node_modules", "vendor", "__pycache__", "dist", "build", ".venv", "venv"}

	for _, dir := range skipDirs {
		skipPath := filepath.Join(tmpDir, dir, "code.go")
		if err := os.MkdirAll(filepath.Dir(skipPath), 0755); err != nil {
			t.Fatalf("failed to create directories: %v", err)
		}
		if err := os.WriteFile(skipPath, []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
	}

	// Create a valid file
	validPath := filepath.Join(tmpDir, "valid.go")
	if err := os.WriteFile(validPath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 file (skipping non-code dirs), got %d: %v", len(result), result)
	}
}

func TestScanDir_EmptyRoot(t *testing.T) {
	s := NewScanner(Options{})
	_, err := s.ScanDir("")
	if err == nil {
		t.Error("expected error for empty root")
	}
}

func TestScanDir_NonExistentRoot(t *testing.T) {
	s := NewScanner(Options{})
	_, err := s.ScanDir("/nonexistent/path/to/dir")
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}

func TestScanDir_FileNotDir(t *testing.T) {
	s := NewScanner(Options{})

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := s.ScanDir(filePath)
	if err == nil {
		t.Error("expected error when root is a file")
	}
}

func TestScanDir_MaxFileSize(t *testing.T) {
	opts := Options{
		MaxFileSize: 100, // 100 bytes
	}
	s := NewScanner(opts)

	tmpDir := t.TempDir()

	// Create a small file
	smallPath := filepath.Join(tmpDir, "small.go")
	if err := os.WriteFile(smallPath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create a large file
	largePath := filepath.Join(tmpDir, "large.go")
	largeContent := make([]byte, 200)
	for i := range largeContent {
		largeContent[i] = 'a'
	}
	if err := os.WriteFile(largePath, largeContent, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := s.ScanDir(tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 small file, got %d: %v", len(result), result)
	}
}

func TestValidateGoFile(t *testing.T) {
	s := NewScanner(Options{})

	tmpDir := t.TempDir()

	// Valid Go file
	validPath := filepath.Join(tmpDir, "valid.go")
	validContent := `package main

func main() {
	println("hello")
}
`
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if !s.validateGoFile(validPath) {
		t.Error("expected valid.go to pass validation")
	}

	// Invalid Go file (syntax error)
	invalidPath := filepath.Join(tmpDir, "invalid.go")
	invalidContent := `package main

func main() {
	println("missing closing
}
`
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if s.validateGoFile(invalidPath) {
		t.Error("expected invalid.go to fail validation")
	}
}

func TestHasSupportedExtension(t *testing.T) {
	s := NewScanner(Options{})

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.go", true},
		{"test.py", true},
		{"test.ts", true},
		{"test.js", true},
		{"test.java", true},
		{"test.c", true},
		{"test.cpp", true},
		{"test.rs", true},
		{"test.rb", true},
		{"test.txt", false},
		{"test.md", false},
		{"test", false},
		{"test.GO", true},  // Case insensitive
		{"test.PY", true},  // Case insensitive
	}

	for _, tc := range tests {
		result := s.hasSupportedExtension(tc.path)
		if result != tc.expected {
			t.Errorf("hasSupportedExtension(%s): expected %v, got %v", tc.path, tc.expected, result)
		}
	}
}
