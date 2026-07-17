package core

import (
	"math"
)

// shannonEntropy calculates the Shannon entropy of a string.
// Higher entropy values indicate more randomness (typical of real secrets).
// Low entropy values suggest false positives (common words, patterns).
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}

	// Count byte frequencies
	freq := make(map[byte]int)
	for i := 0; i < len(s); i++ {
		freq[s[i]]++
	}

	// Calculate Shannon entropy
	var entropy float64
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / float64(len(s))
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// MinEntropy is the minimum entropy threshold for high-entropy matches.
// Below this value, matches are considered likely false positives.
const MinEntropy = 3.0

// IsHighEntropy returns true if the string has sufficient entropy
// to be considered a potential secret (not a false positive).
func IsHighEntropy(s string) bool {
	return shannonEntropy(s) >= MinEntropy
}
