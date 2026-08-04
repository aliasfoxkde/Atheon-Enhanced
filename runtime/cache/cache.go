package cache

import "time"

type Cache struct {
	entries map[string]*CacheEntry
	maxAge  time.Duration
}

type CacheEntry struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

func NewCache(maxAge time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		maxAge:  maxAge,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		return nil, false
	}
	return entry.Value, true
}

func (c *Cache) Set(key string, value []byte) {
	c.entries[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(c.maxAge),
	}
}

func (c *Cache) Delete(key string) {
	delete(c.entries, key)
}

func (c *Cache) Clear() {
	c.entries = make(map[string]*CacheEntry)
}

func (c *Cache) Size() int {
	return len(c.entries)
}

func (c *Cache) Cleanup() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}
