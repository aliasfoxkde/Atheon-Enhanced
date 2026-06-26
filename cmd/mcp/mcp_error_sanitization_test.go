package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestSafeErrorNotExist verifies os.IsNotExist maps to "file not found"
// and does not expose the raw filesystem path.
func TestSafeErrorNotExist(t *testing.T) {
	err := os.ErrNotExist
	msg := safeError(err)
	if msg != "file not found" {
		t.Errorf("expected \"file not found\", got %q", msg)
	}
	if msg == err.Error() && strings.Contains(err.Error(), "/") {
		t.Errorf("safeError returned raw path in error message: %s", msg)
	}
}

// TestSafeErrorPermission verifies os.IsPermission maps to "permission denied".
func TestSafeErrorPermission(t *testing.T) {
	err := os.ErrPermission
	msg := safeError(err)
	if msg != "permission denied" {
		t.Errorf("expected \"permission denied\", got %q", msg)
	}
}

// TestSafeErrorGeneric verifies that unknown errors fall back to
// "internal error" without exposing the raw message.
func TestSafeErrorGeneric(t *testing.T) {
	err := errors.New("something went wrong")
	msg := safeError(err)
	if msg != "internal error" {
		t.Errorf("expected \"internal error\", got %q", msg)
	}
	if msg == err.Error() {
		t.Errorf("safeError leaked raw error message: %s", msg)
	}
}

// TestSafeErrorNil verifies safeError handles nil gracefully.
func TestSafeErrorNil(t *testing.T) {
	msg := safeError(nil)
	if msg == "" {
		t.Error("safeError(nil) must return non-empty string")
	}
	if msg == "<nil>" {
		t.Error("safeError(nil) should not return literal <nil>")
	}
}
