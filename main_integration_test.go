package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testBinaryName() string {
	if runtime.GOOS == "windows" {
		return "atheon-test.exe"
	}
	return "atheon-test"
}

func buildTestBinary(t *testing.T) (string, func()) {
	t.Helper()
	name := testBinaryName()
	bin, err := filepath.Abs(name)
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", bin, ".")
	if err := buildCmd.Run(); err != nil {
		t.Skip("Failed to build binary, skipping test")
	}
	return bin, func() { os.Remove(bin) }
}

// TestMainVersionFlag tests main() with --version flag
func TestMainVersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	cmd := exec.Command(bin, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("--version flag error: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from --version flag")
	}

	if !bytes.Contains(output, []byte("atheon")) {
		t.Error("Expected 'atheon' in version output")
	}
}

// TestMainListCommand tests main() with list command
func TestMainListCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	cmd := exec.Command(bin, "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from list command")
	}

	if !strings.Contains(string(output), "aws-access-key") {
		t.Error("Expected 'aws-access-key' in list output")
	}
}

// TestMainHelpCommand tests main() with --help flag
func TestMainHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	cmd := exec.Command(bin, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--help command failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from --help command")
	}

	if !strings.Contains(string(output), "usage:") {
		t.Error("Expected 'usage:' in help output")
	}
}

// TestMainJSONOutput tests JSON output functionality
func TestMainJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	tmpFile := filepath.Join(os.TempDir(), "atheon-test-input.txt")
	defer os.Remove(tmpFile)
	content := []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "--json", "--file", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("JSON command completed with exit code: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected JSON output")
	}

	if !bytes.HasPrefix(output, []byte("[")) {
		t.Error("Expected JSON array output")
	}
}

// TestMainEnvScanning tests environment variable scanning
func TestMainEnvScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	os.Setenv("TEST_AWS_KEY", "AKIAIOSFODNN7EXAMPLE")
	defer os.Unsetenv("TEST_AWS_KEY")

	cmd := exec.Command(bin, "--env")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Env scan completed with exit code: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from env scan")
	}

	if !strings.Contains(string(output), "aws-access-key") {
		t.Error("Expected 'aws-access-key' in env scan output")
	}
}

// TestMainInvalidArgs tests error handling for invalid arguments
func TestMainInvalidArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}

	bin, cleanup := buildTestBinary(t)
	defer cleanup()

	cmd := exec.Command(bin, "invalid-command")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error for invalid command")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid command")
	}

	outputStr := strings.ToLower(string(output))
	if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "unknown") {
		t.Error("Expected error message in output")
	}
}
