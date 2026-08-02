package core

import (
	"container/list"
	"math"
	"sync"
)

// entropyCache provides thread-safe LRU caching of entropy calculations.
// Strings that appear multiple times (e.g., repeated headers, footers)
// benefit from caching to avoid redundant computation.
// Uses LRU eviction to handle more than maxEntries unique strings.
var entropyCache = struct {
	m    map[string]*list.Element // key -> list.Element
	lru  *list.List              // front=MRU, back=LRU
	mu   sync.Mutex
	limit int
}{m: make(map[string]*list.Element), lru: list.New(), limit: 1024}

// cacheEntry stores both key and value in the list for proper LRU eviction.
type cacheEntry struct {
	key   string
	value float64
}

// shannonEntropy calculates the Shannon entropy of a string.
// Higher entropy values indicate more randomness (typical of real secrets).
// Low entropy values suggest false positives (common words, patterns).
// Results are cached for strings seen multiple times.
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}

	// Fast path: check cache with lock
	entropyCache.mu.Lock()
	if elem, ok := entropyCache.m[s]; ok {
		// Move to front (most recently used)
		entropyCache.lru.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entropyCache.mu.Unlock()
		return entry.value
	}
	entropyCache.mu.Unlock()

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

	// Store in cache with LRU eviction under lock
	entropyCache.mu.Lock()
	defer entropyCache.mu.Unlock()

	// Evict LRU entry if at limit
	if entropyCache.lru.Len() >= entropyCache.limit {
		oldest := entropyCache.lru.Back()
		if oldest != nil {
			entry := oldest.Value.(*cacheEntry)
			delete(entropyCache.m, entry.key)
			entropyCache.lru.Remove(oldest)
		}
	}

	// Add new entry
	elem := entropyCache.lru.PushFront(&cacheEntry{key: s, value: entropy})
	entropyCache.m[s] = elem

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
