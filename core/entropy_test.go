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
