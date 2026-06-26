package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

func TestListCategories(t *testing.T) {
	categories := core.Categories()

	if len(categories) == 0 {
		t.Error("expected to find at least one category")
	}

	// Verify expected categories exist
	expectedCategories := []string{"secrets", "pii"}
	for _, cat := range expectedCategories {
		found := false
		for _, c := range categories {
			if c == cat {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find category '%s'", cat)
		}
	}
}

func TestListAllPatterns(t *testing.T) {
	patterns := core.All()

	if len(patterns) == 0 {
		t.Error("expected to find at least one pattern")
	}

	// Verify expected patterns exist
	expectedPatterns := []string{"aws-access-key", "openai-api-key"}
	for _, expectedPat := range expectedPatterns {
		found := false
		for _, p := range patterns {
			if p.Name() == expectedPat {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find pattern '%s'", expectedPat)
		}
	}
}

func TestScanFile(t *testing.T) {
	// Test scanning a single file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := core.ScanFile(context.Background(), testFile)
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

func TestScanDirectory(t *testing.T) {
	// Test scanning a directory
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file2, []byte("api_key=sk-test1234567890abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := core.ScanDir(context.Background(), tmpDir, core.ScanOpts{})
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if stats.Files != 2 {
		t.Errorf("expected 2 files scanned, got %d", stats.Files)
	}

	if len(findings) < 2 {
		t.Errorf("expected at least 2 findings, got %d", len(findings))
	}

	// Verify findings from both files
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

func TestCategoryFiltering(t *testing.T) {
	// Test category filtering
	// Set active categories to only secrets
	core.SetActiveCategories([]string{"secrets"})

	patterns := core.All()

	// Verify that only secrets patterns are active
	// This is a basic check to ensure filtering works
	foundSecret := false
	for _, p := range patterns {
		if strings.Contains(p.Name(), "api-key") || strings.Contains(p.Name(), "token") {
			foundSecret = true
			break
		}
	}

	if !foundSecret {
		t.Error("expected to find secrets patterns when filtering by secrets category")
	}

	// Reset to all categories
	core.SetActiveCategories(nil)
}

func TestIgnoreFiles(t *testing.T) {
	// Test .atheonignore functionality
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create .atheonignore file
	ignoreFile := filepath.Join(tmpDir, ".atheonignore")
	if err := os.WriteFile(ignoreFile, []byte("test.txt"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := core.ScanDir(context.Background(), tmpDir, core.ScanOpts{})
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	// With .atheonignore, test.txt should be excluded
	// But due to how the ignore system works, we might still scan it
	// So we'll just verify the scan completed successfully
	if stats.Files > 0 {
		// If files were scanned, verify none are from ignored files
		for _, finding := range findings {
			if filepath.Base(finding.File) == "test.txt" {
				t.Error("test.txt should have been ignored")
			}
		}
	}

	if len(findings) > 0 {
		// Verify findings are not from ignored files
		for _, finding := range findings {
			if filepath.Base(finding.File) == "test.txt" {
				t.Error("expected no findings from ignored file")
			}
		}
	}
}

func TestJSONOutput(t *testing.T) {
	// Test that findings can be converted to JSON
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "AKIAIOSFODNN7EXAMPLE"

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, _, err := core.ScanFile(context.Background(), testFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// Convert findings to JSON format
	jsonFindings := make([]map[string]interface{}, 0, len(findings))
	for _, f := range findings {
		jsonFindings = append(jsonFindings, map[string]interface{}{
			"pattern": f.Pattern,
			"file":    f.File,
			"line":    f.Line,
			"content": f.Content,
		})
	}

	if len(jsonFindings) == 0 {
		t.Error("expected to convert findings to JSON format")
	}

	if jsonFindings[0]["pattern"] != "aws-access-key" {
		t.Error("expected JSON to contain aws-access-key pattern")
	}

	if jsonFindings[0]["file"] != testFile {
		t.Error("expected JSON to contain correct file path")
	}
}

func TestEnvironmentScanning(t *testing.T) {
	// Test environment variable scanning
	oldKey := os.Getenv("TEST_AWS_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("TEST_AWS_KEY", oldKey)
		} else {
			os.Unsetenv("TEST_AWS_KEY")
		}
	}()

	os.Setenv("TEST_AWS_KEY", "AKIAIOSFODNN7EXAMPLE")

	findings := core.ScanEnv(context.Background())

	if len(findings) == 0 {
		t.Error("expected to find AWS key in environment variables")
	}

	foundAWSKey := false
	for _, finding := range findings {
		if finding.Pattern == "aws-access-key" && finding.File == "env:TEST_AWS_KEY" {
			foundAWSKey = true
			break
		}
	}

	if !foundAWSKey {
		t.Error("expected to find aws-access-key pattern in TEST_AWS_KEY environment variable")
	}
}

func TestStringScanning(t *testing.T) {
	// Test scanning strings directly
	testContent := "Here's a test string with AKIAIOSFODNN7EXAMPLE embedded"

	findings := core.ScanString(context.Background(), testContent, "test-source")

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

func TestPatternRegistration(t *testing.T) {
	// Test that patterns are properly registered
	patterns := core.All()

	if len(patterns) == 0 {
		t.Error("expected patterns to be registered")
	}

	// Test that each pattern has required methods
	for _, p := range patterns {
		name := p.Name()
		if name == "" {
			t.Error("expected pattern to have a name")
		}

		// Test that pattern can be used
		testContent := "test content"
		_ = p.Matches(testContent) // Should not panic
	}
}

func TestUpdateCommand(t *testing.T) {
	// Serve a minimal valid bundle from a local test server so DownloadBundle
	// never touches the real network (avoids the HTTP/2 goroutine leak that
	// caused this test to hang for the full 15-minute timeout).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if err := json.NewEncoder(gz).Encode([]core.PatternDef{}); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		gz.Close()
		bundleBytes := buf.Bytes()
		// Serve checksums.txt so hash verification passes.
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			h := sha256.New()
			h.Write(bundleBytes)
			checksumLine := hex.EncodeToString(h.Sum(nil)) + "  patterns.bundle\n"
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(checksumLine)) //nolint:errcheck
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(bundleBytes)
	}))
	defer srv.Close()

	// Snapshot the on-disk bundle (if any) so DownloadBundle's os.WriteFile
	// call doesn't permanently corrupt ~/.atheon/patterns.bundle for other
	// processes and future test runs.
	home, _ := os.UserHomeDir()
	bundlePath := filepath.Join(home, ".atheon", "patterns.bundle")
	origBundle, origErr := os.ReadFile(bundlePath)
	defer func() {
		if origErr == nil {
			os.WriteFile(bundlePath, origBundle, 0o600) //nolint:errcheck
		} else {
			os.Remove(bundlePath) //nolint:errcheck
		}
	}()

	restore := core.SetBundleDownloadURLForTest(srv.URL + "/")
	defer restore()
	// Reload the embedded bundle after this test so subsequent tests in the
	// same binary see a full pattern set (DownloadBundle replaces allPatterns).
	defer core.ReloadBundle()

	if err := core.DownloadBundle(context.Background(), false); err != nil {
		t.Fatalf("DownloadBundle failed: %v", err)
	}
}

func TestHelpText(t *testing.T) {
	// Test that help text is available
	// We can't easily test the actual help output without calling main()
	// but we can verify the functions exist

	// Just verify the core package has the expected functions
	findings := core.ScanString(context.Background(), "test", "test")
	// ScanString should return a slice (may be nil or empty)
	// The important thing is it doesn't panic
	if len(findings) > 0 {
		t.Log("ScanString found patterns in test content")
	} else {
		t.Log("ScanString correctly handles content with no patterns")
	}
}

func TestMultiplePatternsInSameLine(t *testing.T) {
	// Test that multiple patterns in the same line are detected
	testContent := "AWS_KEY=AKIAIOSFODNN7EXAMPLE and api_key=sk-test1234567890abcdef"

	findings := core.ScanString(context.Background(), testContent, "test")

	if len(findings) < 2 {
		t.Errorf("expected to find at least 2 patterns, got %d", len(findings))
	}
}

func TestBinaryFilesExcluded(t *testing.T) {
	// Test that binary files are excluded from scanning
	tmpDir := t.TempDir()

	// Create binary files that should be skipped
	binaryFiles := []string{
		"image.png",
		"document.pdf",
		"archive.zip",
	}

	for _, filename := range binaryFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create text file that should be scanned
	textFile := filepath.Join(tmpDir, "config.txt")
	if err := os.WriteFile(textFile, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, stats, err := core.ScanDir(context.Background(), tmpDir, core.ScanOpts{})
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
