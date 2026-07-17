package core

import (
	"testing"
)

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		minEntropy float64
		maxEntropy float64
	}{
		{"empty string", "", 0, 0},
		{"single char repeated", "aaaaaaaaaa", 0, 0.5},
		{"low entropy common word", "hello", 1.5, 2.5},
		{"medium entropy", "hello world", 2.5, 3.5},
		{"high entropy random", "xK8#9@mP$", 3, 4},
		{"very high entropy", "AKIAIOSFODNN7EXAMPLE", 3.5, 4.5},
		{"hex string", "deadbeef12345678", 3, 4.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entropy := shannonEntropy(tc.input)
			if entropy < tc.minEntropy || entropy > tc.maxEntropy {
				t.Errorf("shannonEntropy(%q) = %v, expected between %v and %v",
					tc.input, entropy, tc.minEntropy, tc.maxEntropy)
			}
		})
	}
}

func TestIsHighEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"short low entropy", "abc", false},
		{"common word", "password", false},
		{"short random", "xK8#9", false},
		{"AWS access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"github token", "ghp_abcdefghijklmnopqrstuvwxyz1234567890", true},
		{"random hex", "deadbeefcafebabe12345678abcdef00", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsHighEntropy(tc.input)
			if result != tc.expected {
				t.Errorf("IsHighEntropy(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMinEntropyConstant(t *testing.T) {
	if MinEntropy != 3.0 {
		t.Errorf("MinEntropy = %v, expected 3.0", MinEntropy)
	}
}

// BenchmarkShannonEntropy measures the throughput of the entropy calculation
// against strings of varying lengths and entropy profiles.
func BenchmarkShannonEntropy(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"aws_access_key", "AKIAIOSFODNN7EXAMPLE"},
		{"github_token", "ghp_abcdefghijklmnopqrstuvwxyz1234567890"},
		{"hex_32bytes", "deadbeefcafebabe12345678abcdef0099"},
		{"hex_64bytes", "deadbeefcafebabe12345678abcdef0099deadbeefcafebabe12345678abcdef0099"},
		{"password_16", "MyStr0ng!P@ssw0rd"},
		{"password_32", "MyStr0ng!P@ssw0rdMyStr0ng!P@ssw0rd"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"jwt_token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = shannonEntropy(tc.input)
			}
		})
	}
}

// BenchmarkIsHighEntropy measures the throughput of the high-entropy check
// for strings that are commonly used as secret values.
func BenchmarkIsHighEntropy(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"aws_access_key", "AKIAIOSFODNN7EXAMPLE"},
		{"github_token", "ghp_abcdefghijklmnopqrstuvwxyz1234567890"},
		{"hex_32bytes", "deadbeefcafebabe12345678abcdef0099"},
		{"random_32", "xK8#9@mP$L2n$P4q$R6t$T8v"},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = IsHighEntropy(tc.input)
			}
		})
	}
}

// BenchmarkShannonEntropyLarge measures entropy calculation for large strings
// (e.g., certificates, large keys).
func BenchmarkShannonEntropyLarge(b *testing.B) {
	// 1KB of pseudo-random data
	largeInput := make([]byte, 1024)
	for i := range largeInput {
		largeInput[i] = byte((i * 17) % 256)
	}
	input := string(largeInput)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = shannonEntropy(input)
	}
}
