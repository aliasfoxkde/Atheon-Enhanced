package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// atheonTestBin is the absolute path to the built atheon binary used by
// integration tests that exec the CLI. It is built once in TestMain and
// removed when the test process exits. Using an absolute path under
// t.TempDir keeps tests deterministic regardless of the package working
// directory, which on macOS in particular can differ from the package
// directory because the shell that runs `go test ./...` may relocate it.
var atheonTestBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "atheon-test-bin-*")
	if err != nil {
		// Fall back to the empty string; every test that needs the binary
		// will skip with a clear message.
		os.Exit(m.Run())
	}
	bin := filepath.Join(dir, "atheon-test")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	// Build from the current package directory (this file is in cmd/atheon).
	if out, err := cmd.CombinedOutput(); err != nil {
		// Drop the temp dir so it does not leak; tests will skip.
		_ = os.RemoveAll(dir)
		os.Stderr.Write(out)
		os.Exit(m.Run())
	}
	atheonTestBin = bin

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// TestMainVersionFlag tests the binary's --version flag.
func TestMainVersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	output, err := exec.Command(atheonTestBin, "--version").CombinedOutput()
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

// TestMainListCommand tests the binary's list command.
func TestMainListCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	output, err := exec.Command(atheonTestBin, "list").CombinedOutput()
	if err != nil {
		t.Fatalf("list command failed: %v\noutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected output from list command")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "aws-access-key") {
		t.Error("Expected 'aws-access-key' in list output")
	}
}

// TestMainHelpCommand tests the binary's --help flag.
func TestMainHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	output, err := exec.Command(atheonTestBin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("--help command failed: %v\noutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected output from help command")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "usage:") {
		t.Error("Expected 'usage:' in help output")
	}
}

// TestMainJSONOutput tests JSON output functionality.
func TestMainJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	tmpFile := filepath.Join(t.TempDir(), "input.txt")
	content := []byte("AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := exec.Command(atheonTestBin, "--json", "--file", tmpFile).CombinedOutput()
	if err != nil {
		// Non-zero exit code is expected when findings are present.
		t.Logf("JSON command completed with exit code: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected JSON output")
	}

	if !bytes.HasPrefix(output, []byte("[")) {
		t.Errorf("Expected JSON array output, got: %s", output)
	}
}

// TestMainEnvScanning tests environment variable scanning.
func TestMainEnvScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	t.Setenv("TEST_AWS_KEY", "AKIAIOSFODNN7EXAMPLE")

	output, err := exec.Command(atheonTestBin, "--env").CombinedOutput()
	if err != nil {
		// Non-zero exit code is expected when findings are present.
		t.Logf("Env scan completed with exit code: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from env scan")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "aws-access-key") {
		t.Errorf("Expected 'aws-access-key' in env scan output, got: %s", output)
	}
}

// TestMainInvalidArgs tests error handling for invalid arguments.
func TestMainInvalidArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping main() test in short mode")
	}
	if atheonTestBin == "" {
		t.Skip("atheon-test binary was not built")
	}

	output, err := exec.Command(atheonTestBin, "invalid-command").CombinedOutput()
	if err == nil {
		t.Error("Expected error for invalid command")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid command")
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "error") && !strings.Contains(strings.ToLower(outputStr), "unknown") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}
