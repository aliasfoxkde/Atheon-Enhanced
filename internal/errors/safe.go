// Package errors provides safe error handling utilities for Atheon.
//
// This package is separate from the standard library's errors package to avoid
// import conflicts with packages that use both stdlib errors and Atheon's own
// error types.
package errors

import (
	"os"
)

// SafeError returns a user-safe error message that does not leak filesystem
// paths or internal details. Used by both the CLI and MCP server to map
// OS-level errors to human-readable strings.
func SafeError(err error) string {
	if err == nil {
		return "no error"
	}
	switch {
	case os.IsNotExist(err):
		return "file not found"
	case os.IsPermission(err):
		return "permission denied"
	default:
		return "internal error"
	}
}
