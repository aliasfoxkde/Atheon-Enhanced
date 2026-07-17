package core

import (
	"math"
	"sync"
)

// entropyCache provides thread-safe caching of entropy calculations.
// Strings that appear multiple times (e.g., repeated headers, footers)
// benefit from caching to avoid redundant computation.
var entropyCache = struct {
	m    map[string]float64
	mu   sync.RWMutex
	max  int
	size int
}{m: make(map[string]float64), max: 1024}

// shannonEntropy calculates the Shannon entropy of a string.
// Higher entropy values indicate more randomness (typical of real secrets).
// Low entropy values suggest false positives (common words, patterns).
// Results are cached for strings seen multiple times.
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}

	// Check cache first (read lock)
	entropyCache.mu.RLock()
	if e, ok := entropyCache.m[s]; ok {
		entropyCache.mu.RUnlock()
		return e
	}
	entropyCache.mu.RUnlock()

	// Calculate entropy
	var entropy float64
	freq := make(map[byte]int)
	for i := 0; i < len(s); i++ {
		freq[s[i]]++
	}
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / float64(len(s))
			entropy -= p * math.Log2(p)
		}
	}

	// Store in cache (write lock)
	entropyCache.mu.Lock()
	if entropyCache.size < entropyCache.max {
		entropyCache.m[s] = entropy
		entropyCache.size++
	}
	entropyCache.mu.Unlock()

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
