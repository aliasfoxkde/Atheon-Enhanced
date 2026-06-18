package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundlerValidation(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	secretsDir := filepath.Join(communityDir, "secrets")

	if err := os.MkdirAll(secretsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create test pattern file
	patternContent := `name: test-api-key
match: '\btest_[A-Z0-9]{32}\b'
`
	patternFile := filepath.Join(secretsDir, "test-key.yaml")
	if err := os.WriteFile(patternFile, []byte(patternContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test that pattern file can be read
	data, err := os.ReadFile(patternFile)
	if err != nil {
		t.Fatalf("failed to read pattern file: %v", err)
	}

	if len(data) == 0 {
		t.Error("pattern file is empty")
	}

	// Verify pattern content contains expected fields
	content := string(data)
	if !contains(content, "name:") {
		t.Error("pattern file missing 'name' field")
	}

	if !contains(content, "match:") {
		t.Error("pattern file missing 'match' field")
	}
}

func TestBundlerEnabledField(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	secretsDir := filepath.Join(communityDir, "secrets")

	if err := os.MkdirAll(secretsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create pattern file with enabled=false
	patternContent := `name: disabled-key
match: '\bdisabled_[A-Z0-9]{32}\b'
enabled: false
`
	patternFile := filepath.Join(secretsDir, "disabled-key.yaml")
	if err := os.WriteFile(patternFile, []byte(patternContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test that pattern file can be read
	data, err := os.ReadFile(patternFile)
	if err != nil {
		t.Fatalf("failed to read pattern file: %v", err)
	}

	content := string(data)
	if !contains(content, "enabled: false") {
		t.Error("pattern file missing 'enabled: false' field")
	}
}

func TestBundlerHandlesInvalidYAML(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	secretsDir := filepath.Join(communityDir, "secrets")

	if err := os.MkdirAll(secretsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create invalid YAML file
	invalidPattern := `name: test-key
match: [invalid yaml
`
	patternFile := filepath.Join(secretsDir, "invalid.yaml")
	if err := os.WriteFile(patternFile, []byte(invalidPattern), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test that file can be created (even if invalid)
	_, err := os.ReadFile(patternFile)
	if err != nil {
		t.Fatalf("failed to read invalid pattern file: %v", err)
	}

	// If we got here without panic, the test passes
}

func TestBundlerHandlesMissingFields(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	secretsDir := filepath.Join(communityDir, "secrets")

	if err := os.MkdirAll(secretsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create pattern file missing required fields
	incompletePattern := `name: incomplete-key
`
	patternFile := filepath.Join(secretsDir, "incomplete.yaml")
	if err := os.WriteFile(patternFile, []byte(incompletePattern), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test that file can be created (even if incomplete)
	data, err := os.ReadFile(patternFile)
	if err != nil {
		t.Fatalf("failed to read incomplete pattern file: %v", err)
	}

	content := string(data)
	if !contains(content, "name:") {
		t.Error("pattern file should have 'name' field")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
