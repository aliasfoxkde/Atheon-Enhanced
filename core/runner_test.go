package core

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestScanFile(t *testing.T) {
	// Create temporary file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Some test content\nAWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\nMore content\n"

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanFile(context.Background(), testFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if stats.Files != 1 {
		t.Errorf("expected 1 file scanned, got %d", stats.Files)
	}

	if len(findings) == 0 {
		t.Error("expected to find AWS key pattern")
	}

	foundAWSKey := false
	for _, finding := range findings {
		if finding.Pattern == "aws-access-key" {
			foundAWSKey = true
			break
		}
	}

	if !foundAWSKey {
		t.Error("did not find expected aws-access-key pattern")
	}
}

func TestScanDir(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	ignoredFile := filepath.Join(tmpDir, "ignored.log")

	if err := os.WriteFile(file1, []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file2, []byte("api_key=sk-test1234567890abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(ignoredFile, []byte("token=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create .gitignore to exclude .log files
	gitignore := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("*.log"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should have scanned at least the txt files
	if stats.Files < 2 {
		t.Errorf("expected at least 2 files scanned, got %d", stats.Files)
	}

	if len(findings) < 2 {
		t.Errorf("expected at least 2 findings, got %d", len(findings))
	}

	// Verify we found patterns in the txt files
	foundAWSKey := false
	foundOpenAIKey := false
	for _, finding := range findings {
		if finding.Pattern == "aws-access-key" {
			foundAWSKey = true
		}
		if finding.Pattern == "openai-api-key" {
			foundOpenAIKey = true
		}
	}

	if !foundAWSKey {
		t.Error("expected to find aws-access-key pattern")
	}

	if !foundOpenAIKey {
		t.Error("expected to find openai-api-key pattern")
	}
}

func TestScanEnv(t *testing.T) {
	// Set test environment variables
	oldKey := os.Getenv("TEST_AWS_KEY")
	oldToken := os.Getenv("TEST_TOKEN")

	defer func() {
		if oldKey != "" {
			os.Setenv("TEST_AWS_KEY", oldKey)
		} else {
			os.Unsetenv("TEST_AWS_KEY")
		}
		if oldToken != "" {
			os.Setenv("TEST_TOKEN", oldToken)
		} else {
			os.Unsetenv("TEST_TOKEN")
		}
	}()

	os.Setenv("TEST_AWS_KEY", "AKIAIOSFODNN7EXAMPLE")
	// Use realistic token format that should NOT match AWS key pattern
	os.Setenv("TEST_TOKEN", "tok_1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p")

	findings := ScanEnv(context.Background())

	if len(findings) == 0 {
		t.Error("expected to find AWS key in environment")
	}

	foundAWSKey := false
	for _, finding := range findings {
		if finding.Pattern == "aws-access-key" && finding.File == "env:TEST_AWS_KEY" {
			foundAWSKey = true
			break
		}
	}

	if !foundAWSKey {
		t.Error("did not find expected aws-access-key pattern in environment")
	}
}

func TestScanString(t *testing.T) {
	testContent := "Here's a test string with AKIAIOSFODNN7EXAMPLE embedded"

	findings := ScanString(context.Background(), testContent, "test-source")

	if len(findings) == 0 {
		t.Error("expected to find AWS key pattern")
	}

	if findings[0].Pattern != "aws-access-key" {
		t.Errorf("expected pattern name 'aws-access-key', got '%s'", findings[0].Pattern)
	}

	if findings[0].File != "test-source" {
		t.Errorf("expected file 'test-source', got '%s'", findings[0].File)
	}
}

func TestIsIgnored(t *testing.T) {
	// Test that the isIgnored function works with nil matcher
	// When matcher is nil, nothing should be ignored
	tests := []struct {
		path     string
		expected bool
	}{
		{"debug.log", false}, // nil matcher returns false
		{"temp_file.txt", false},
		{"build/output.txt", false},
		{"src/main.go", false},
		{"debug.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isIgnored(tt.path, nil)
			if result != tt.expected {
				t.Errorf("isIgnored(%q, nil) = %v, want %v", tt.path, result, tt.expected) //nolint
			}
		})
	}
}

func TestLoadIgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	gitignoreContent := "# Comment line\n*.log\ntemp_*\n**/*.generated.go\n\n!important.log\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	matchers := loadIgnorePatternsMatcher(tmpDir)
	if len(matchers) == 0 {
		t.Fatal("expected at least one matcher")
	}

	if !isIgnored("foo.log", matchers) {
		t.Error("expected foo.log to be ignored")
	}
	if !isIgnored("temp_file.go", matchers) {
		t.Error("expected temp_file.go to be ignored")
	}
	if isIgnored("main.go", matchers) {
		t.Error("expected main.go to not be ignored")
	}
	if !isIgnored("pkg/api/handler.generated.go", matchers) {
		t.Error("expected **/*.generated.go to match nested path")
	}
}

func TestBinaryExts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create binary files that should be skipped
	binaryFiles := []string{
		"image.png",
		"document.pdf",
		"archive.zip",
		"program.exe",
	}

	for _, filename := range binaryFiles {
		path := filepath.Join(tmpDir, filename)
		// Use actual binary-like content that should still be skipped
		// PNG magic bytes + pattern data that would match if not skipped
		binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		binaryContent = append(binaryContent, []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")...)
		if err := os.WriteFile(path, binaryContent, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create text file that should be scanned
	textFile := filepath.Join(tmpDir, "config.txt")
	if err := os.WriteFile(textFile, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only scan the text file, not binary files
	if stats.Files != 1 {
		t.Errorf("expected 1 file scanned (binary files should be skipped), got %d", stats.Files)
	}

	if len(findings) == 0 {
		t.Error("expected to find pattern in text file")
	}
}

func TestSkipDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories that should be skipped
	skipDirs := []string{
		".git",
		"node_modules",
		"vendor",
		"__pycache__",
	}

	for _, dirName := range skipDirs {
		dirPath := filepath.Join(tmpDir, dirName)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatal(err)
		}

		// Add file inside skipped directory
		filePath := filepath.Join(dirPath, "secret.txt")
		if err := os.WriteFile(filePath, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create regular file that should be scanned
	regularFile := filepath.Join(tmpDir, "config.txt")
	if err := os.WriteFile(regularFile, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Should only scan the regular file
	if stats.Files != 1 {
		t.Errorf("expected 1 file scanned (skip directories should be ignored), got %d", stats.Files)
	}

	if len(findings) == 0 {
		t.Error("expected to find pattern in regular file")
	}
}

// TestScanDir_NonExistent tests ScanDir with non-existent directory
func TestScanDir_NonExistent(t *testing.T) {
	findings, stats, err := ScanDir(context.Background(), "/nonexistent/directory/path")
	// ScanDir might not error on non-existent directories, just return empty results
	if err != nil {
		// Error is acceptable
		t.Logf("ScanDir returned error for non-existent directory: %v", err)
	}

	if stats.Files != 0 {
		t.Errorf("expected 0 files scanned for non-existent directory, got %d", stats.Files)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-existent directory, got %d", len(findings))
	}
}

// TestScanFile_LargeFile tests that files larger than maxFileSize are skipped
func TestScanFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.bin")

	// Create a file larger than maxFileSize (10MB)
	// Use sparse-like approach: write a large chunk
	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatal(err)
	}

	// Write 11MB of data
	data := make([]byte, 11*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	findings, stats, err := ScanFile(context.Background(), largeFile)
	if err != nil {
		t.Fatalf("ScanFile should not error for large file, got: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for large file, got %d", len(findings))
	}
	if stats.Files != 1 {
		t.Errorf("expected Files=1 (counted even if skipped), got %d", stats.Files)
	}
}

// TestScanDir_PermissionError tests ScanDir with permission errors
func TestScanDir_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root, permission test ineffective")
	}

	tmpDir := t.TempDir()
	// Create a subdirectory with no permissions
	noPermDir := filepath.Join(tmpDir, "noperm")
	if err := os.Mkdir(noPermDir, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(noPermDir, 0o755) // Cleanup

	// Also create a readable file so ScanDir has something to process
	readableFile := filepath.Join(tmpDir, "readable.txt")
	if err := os.WriteFile(readableFile, []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	// ScanDir should still work, just skip the inaccessible directory
	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		// Permission errors might be returned, that's acceptable
		t.Logf("ScanDir returned error (acceptable): %v", err)
	}

	// Should have completed without crashing and scanned the readable file
	if stats.Files < 1 {
		t.Errorf("expected at least 1 file scanned (readable), got %d", stats.Files)
	}

	if findings == nil {
		t.Error("findings should not be nil")
	}
}

// TestScanDir_EmptyDirectory tests ScanDir with empty directory
func TestScanDir_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed on empty directory: %v", err)
	}

	if stats.Files != 0 {
		t.Errorf("expected 0 files scanned, got %d", stats.Files)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

// TestScanDir_WithSubdirectories tests ScanDir with nested directory structure
func TestScanDir_WithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(tmpDir, "subdir2")
	nestedDir := filepath.Join(subDir1, "nested")

	for _, dir := range []string{subDir1, subDir2, nestedDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create files in different directories
	file1 := filepath.Join(tmpDir, "root.txt")
	file2 := filepath.Join(subDir1, "sub1.txt")
	file3 := filepath.Join(subDir2, "sub2.txt")
	file4 := filepath.Join(nestedDir, "nested.txt")

	for _, file := range []string{file1, file2, file3, file4} {
		if err := os.WriteFile(file, []byte("TEST_KEY=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if stats.Files != 4 {
		t.Errorf("expected 4 files scanned, got %d", stats.Files)
	}

	if len(findings) < 4 {
		t.Errorf("expected at least 4 findings, got %d", len(findings))
	}
}

// TestScanDir_BinaryFiles tests ScanDir with binary file filtering
func TestScanDir_BinaryFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create text file
	textFile := filepath.Join(tmpDir, "text.txt")
	if err := os.WriteFile(textFile, []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create binary file (executable)
	binaryFile := filepath.Join(tmpDir, "binary.exe")
	if err := os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// Binary files should be skipped
	if stats.Files != 1 {
		t.Errorf("expected 1 file scanned (text only), got %d", stats.Files)
	}

	if len(findings) < 1 {
		t.Errorf("expected at least 1 finding from text file, got %d", len(findings))
	}
}

// TestScanDirContextCancelledBeforeStart cancels the context before ScanDir walks any files.
func TestScanDirContextCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := ScanDir(ctx, tmpDir)
	if err != nil && err != context.Canceled {
		t.Errorf("expected nil or context.Canceled, got: %v", err)
	}
}

// TestScanDirFileReadErrorSkipped exercises the goroutine path where os.ReadFile fails.
// Skipped on Windows (chmod is not enforced) and when running as root (ignores mode bits).
func TestScanDirFileReadErrorSkipped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod file permissions are not enforced on Windows")
	}

	tmpDir := t.TempDir()
	unreadable := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(unreadable, []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(unreadable, 0o644)

	findings, stats, err := ScanDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("unexpected ScanDir error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	// Unreadable file is skipped — no findings expected.
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for unreadable file, got %d", len(findings))
	}
}

// TestScanLinesContextCancelled exercises the early-exit path in scanLines.
func TestScanLinesContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	findings := scanLines(ctx, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\nline2\nline3\n", "test.txt")
	// Pre-cancelled context: the early-exit select fires before any line is processed.
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with cancelled context, got %d", len(findings))
	}
}
