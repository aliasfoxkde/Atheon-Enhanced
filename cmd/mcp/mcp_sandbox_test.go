package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSandboxPathRelativeInsideAllowed verifies a relative path inside cwd is allowed.
func TestSandboxPathRelativeInsideAllowed(t *testing.T) {
	// "cmd/mcp" is a relative path from the repo root.
	got, err := sandboxPath("cmd/mcp")
	if err != nil {
		t.Errorf("sandboxPath(%q) returned error: %v", "cmd/mcp", err)
	}
	if got == "" {
		t.Error("expected non-empty resolved path")
	}
}

// TestSandboxPathTraversalBlocked verifies a relative traversal attempt is rejected.
func TestSandboxPathTraversalBlocked(t *testing.T) {
	// "../../../etc/passwd" from the repo root is outside any plausible cwd.
	got, err := sandboxPath("../../../etc/passwd")
	if err == nil {
		t.Errorf("sandboxPath(%q) = %q, want error", "../../../etc/passwd", got)
	}
}

// TestSandboxPathBrokenSymlinkAllowed verifies broken symlinks pass through
// to the handler (which will report file not found).
func TestSandboxPathBrokenSymlinkAllowed(t *testing.T) {
	tmp := t.TempDir()
	broken := filepath.Join(tmp, "broken")
	os.Symlink("/nonexistent/filepath", broken)

	got, err := sandboxPath(broken)
	if err != nil {
		t.Errorf("sandboxPath(%q) returned error: %v", broken, err)
	}
	_ = got
}

// TestSandboxPathAbsoluteAllowed verifies absolute paths are always allowed
// (the user explicitly requested them).
func TestSandboxPathAbsoluteAllowed(t *testing.T) {
	cwd, _ := os.Getwd()
	absPath := filepath.Join(cwd, "cmd/mcp", "main.go")

	got, err := sandboxPath(absPath)
	if err != nil {
		t.Errorf("sandboxPath(%q) returned error: %v", absPath, err)
	}
	if got == "" {
		t.Error("expected non-empty resolved path")
	}
}

// TestSandboxPathSymlinkUnderCwdAllowed verifies a relative symlink that
// stays inside cwd is allowed.
func TestSandboxPathSymlinkUnderCwdAllowed(t *testing.T) {
	cwd, _ := os.Getwd()
	// Create a symlink inside cwd that points to another file inside cwd.
	// Use an absolute symlink target so it resolves correctly regardless of CWD.
	subdir := filepath.Join(cwd, "cmd")
	link := filepath.Join(subdir, "testlink")
	target := filepath.Join(cwd, "README.md") // exists, inside cwd
	os.Symlink(target, link)
	defer os.Remove(link)

	got, err := sandboxPath(filepath.Join("cmd", "testlink"))
	if err != nil {
		t.Errorf("sandboxPath(testlink) returned error: %v", err)
	}
	_ = got
}

// TestSandboxPathSymlinkTraversalBlocked verifies a relative symlink inside cwd
// that points outside cwd is blocked.
func TestSandboxPathSymlinkTraversalBlocked(t *testing.T) {
	cwd, _ := os.Getwd()
	// Create a symlink in cwd that points to /etc/passwd (outside cwd).
	// This is a relative symlink so it needs to be accessed via relative path.
	link := filepath.Join(cwd, ".traversallink")
	if err := os.Symlink("/etc/passwd", link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
	defer os.Remove(link)

	// Access via relative path - the symlink points outside cwd so should be blocked
	got, err := sandboxPath(".traversallink")
	if err == nil {
		t.Errorf("sandboxPath(%q) = %q, want error for symlink traversal", ".traversallink", got)
	}
}
