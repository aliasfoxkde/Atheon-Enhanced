package core

import (
	"os"
	"testing"
)

func TestNewYARAScanner(t *testing.T) {
	ys := NewYARAScanner("")
	if ys == nil {
		t.Fatal("NewYARAScanner returned nil")
	}
	if ys.rules == nil {
		t.Error("rules map is nil")
	}
}

func TestYARAScanner_AddRule(t *testing.T) {
	ys := NewYARAScanner("")
	ys.AddRule(&YARARule{
		Name:        "test-rule",
		Description: "Test description",
		Tags:        []string{"suspicious"},
		Severity:    "high",
		Category:    "test",
	})
	if len(ys.rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(ys.rules))
	}
}

func TestYARAScanner_ScanFile(t *testing.T) {
	ys := NewYARAScanner("")
	for _, rule := range DefaultYARARules() {
		ys.AddRule(rule)
	}

	// Create temp file with suspicious content
	tmpfile, err := os.CreateTemp("", "yara-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString("eval(base64_decode('aGVsbG8='))"); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	findings, err := ys.ScanFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}
	if len(findings) == 0 {
		t.Error("expected findings for suspicious content")
	}
}

func TestYARAScanner_ScanFile_Nonexistent(t *testing.T) {
	ys := NewYARAScanner("")
	_, err := ys.ScanFile("/nonexistent/path/to/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestConvertToFinding(t *testing.T) {
	yf := YARAFinding{
		Rule:        "test-rule",
		File:        "test.go",
		Line:        10,
		Matched:     "suspicious code",
		Severity:    "high",
		Category:    "test",
		Description: "Test description",
	}
	f := ConvertToFinding(yf)
	if f.Pattern != "yara_test-rule" {
		t.Errorf("expected pattern yara_test-rule, got %s", f.Pattern)
	}
	if f.Severity != "high" {
		t.Errorf("expected severity high, got %s", f.Severity)
	}
}

func TestDefaultYARARules(t *testing.T) {
	rules := DefaultYARARules()
	if len(rules) == 0 {
		t.Error("DefaultYARARules returned empty")
	}
	// Check a known rule exists
	found := false
	for _, r := range rules {
		if r.Name == "crypto_miner" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected crypto_miner rule to exist")
	}
}

func TestYARAScanner_ScanDir(t *testing.T) {
	ys := NewYARAScanner("")
	for _, rule := range DefaultYARARules() {
		ys.AddRule(rule)
	}

	// Create temp dir with files
	tmpdir, err := os.MkdirTemp("", "yara-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create a file with suspicious content
	tmpfile, err := os.CreateTemp(tmpdir, "test-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.WriteString("eval(base64_decode('aGVsbG8='))"); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	findings, err := ys.ScanDir(tmpdir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}
	if len(findings) == 0 {
		t.Error("expected findings in scan dir")
	}
}
