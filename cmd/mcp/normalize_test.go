package main

import (
	"encoding/json"
	"testing"
)

// TestNormalizeID tests the normalizeID helper function
func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name string
		id   any
		want string
	}{
		{
			name: "nil returns empty",
			id:   nil,
			want: "",
		},
		{
			name: "string returns as-is",
			id:   "test-id-123",
			want: "test-id-123",
		},
		{
			name: "float64 converts to string",
			id:   float64(123.456),
			want: "123.456",
		},
		{
			name: "int converts to string",
			id:   42,
			want: "42",
		},
		{
			name: "json.Number returns string",
			id:   json.Number("999"),
			want: "999",
		},
		{
			name: "unknown type uses format",
			id:   []byte{'a', 'b'},
			want: "[97 98]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeID(tc.id)
			if got != tc.want {
				t.Errorf("normalizeID(%v) = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
}
