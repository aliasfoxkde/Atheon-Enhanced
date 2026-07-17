package errors

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestSafeError_Nil(t *testing.T) {
	if got := SafeError(nil); got != "no error" {
		t.Errorf("SafeError(nil) = %q, want %q", got, "no error")
	}
}

func TestSafeError_NotExist(t *testing.T) {
	err := os.ErrNotExist
	if got := SafeError(err); got != "file not found" {
		t.Errorf("SafeError(ErrNotExist) = %q, want %q", got, "file not found")
	}
}

func TestSafeError_Permission(t *testing.T) {
	err := os.ErrPermission
	if got := SafeError(err); got != "permission denied" {
		t.Errorf("SafeError(ErrPermission) = %q, want %q", got, "permission denied")
	}
}

func TestSafeError_Other(t *testing.T) {
	err := errors.New("some internal error")
	if got := SafeError(err); got != "internal error" {
		t.Errorf("SafeError(other) = %q, want %q", got, "internal error")
	}
}

func TestSafeError_WrappedNotExist(t *testing.T) {
	// Note: os.IsNotExist uses message-based detection, not errors.Is,
	// so fmt.Errorf-wrapped errors are NOT recognized and return "internal error".
	// This is intentional to avoid leaking wrapped error details.
	err := fmt.Errorf("wrapper: %w", os.ErrNotExist)
	if got := SafeError(err); got != "internal error" {
		t.Errorf("SafeError(wrapped ErrNotExist) = %q, want %q", got, "internal error")
	}
}

func TestSafeError_WrappedPermission(t *testing.T) {
	// Same as above: os.IsNotExist/os.IsPermission use message-based detection
	err := fmt.Errorf("wrapper: %w", os.ErrPermission)
	if got := SafeError(err); got != "internal error" {
		t.Errorf("SafeError(wrapped ErrPermission) = %q, want %q", got, "internal error")
	}
}
