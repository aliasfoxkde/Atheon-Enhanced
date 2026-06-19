package core

import (
	"os"
	"path/filepath"
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

	findings, stats, err := ScanFile(testFile)
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

	findings, stats, err := ScanDir(tmpDir)
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
	os.Setenv("TEST_TOKEN", "regular-token")

	findings := ScanEnv()

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

	findings := ScanString(testContent, "test-source")

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
		if err := os.WriteFile(path, []byte("fake content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create text file that should be scanned
	textFile := filepath.Join(tmpDir, "config.txt")
	if err := os.WriteFile(textFile, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := ScanDir(tmpDir)
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

	findings, stats, err := ScanDir(tmpDir)
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
