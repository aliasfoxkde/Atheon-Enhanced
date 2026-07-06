package main

import (
	"os"
	"testing"
	"time"
)

// TestEnvInt tests the envInt helper function
func TestEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback int
		want     int
		setup    func(string)
		cleanup  func(string)
	}{
		{
			name:     "env var not set returns fallback",
			key:      "TEST_INT_NOT_SET",
			fallback: 42,
			want:     42,
			setup:    func(k string) { os.Unsetenv(k) },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to valid int",
			key:      "TEST_INT_VALID",
			fallback: 0,
			want:     123,
			setup:    func(k string) { os.Setenv(k, "123") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to invalid int returns fallback",
			key:      "TEST_INT_INVALID",
			fallback: 99,
			want:     99,
			setup:    func(k string) { os.Setenv(k, "not-a-number") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to empty string returns fallback",
			key:      "TEST_INT_EMPTY",
			fallback: 77,
			want:     77,
			setup:    func(k string) { os.Setenv(k, "") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cleanup(tc.key)
			tc.setup(tc.key)
			defer tc.cleanup(tc.key)

			got := envInt(tc.key, tc.fallback)
			if got != tc.want {
				t.Errorf("envInt(%q, %d) = %d, want %d", tc.key, tc.fallback, got, tc.want)
			}
		})
	}
}

// TestEnvBytes tests the envBytes helper function
func TestEnvBytes(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback int
		want     int
		setup    func(string)
		cleanup  func(string)
	}{
		{
			name:     "env var not set returns fallback",
			key:      "TEST_BYTES_NOT_SET",
			fallback: 1024,
			want:     1024,
			setup:    func(k string) { os.Unsetenv(k) },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to valid bytes",
			key:      "TEST_BYTES_VALID",
			fallback: 0,
			want:     4096,
			setup:    func(k string) { os.Setenv(k, "4096") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cleanup(tc.key)
			tc.setup(tc.key)
			defer tc.cleanup(tc.key)

			got := envBytes(tc.key, tc.fallback)
			if got != tc.want {
				t.Errorf("envBytes(%q, %d) = %d, want %d", tc.key, tc.fallback, got, tc.want)
			}
		})
	}
}

// TestEnvDuration tests the envDuration helper function
func TestEnvDuration(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback time.Duration
		want     time.Duration
		setup    func(string)
		cleanup  func(string)
	}{
		{
			name:     "env var not set returns fallback",
			key:      "TEST_DURATION_NOT_SET",
			fallback: 30 * time.Second,
			want:     30 * time.Second,
			setup:    func(k string) { os.Unsetenv(k) },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to valid duration",
			key:      "TEST_DURATION_VALID",
			fallback: 0,
			want:     60 * time.Second,
			setup:    func(k string) { os.Setenv(k, "60s") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
		{
			name:     "env var set to invalid duration returns fallback",
			key:      "TEST_DURATION_INVALID",
			fallback: 120 * time.Second,
			want:     120 * time.Second,
			setup:    func(k string) { os.Setenv(k, "not-a-duration") },
			cleanup:  func(k string) { os.Unsetenv(k) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cleanup(tc.key)
			tc.setup(tc.key)
			defer tc.cleanup(tc.key)

			got := envDuration(tc.key, tc.fallback)
			if got != tc.want {
				t.Errorf("envDuration(%q, %v) = %v, want %v", tc.key, tc.fallback, got, tc.want)
			}
		})
	}
}
