package cache

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewCache(5 * time.Second)
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	if c.maxAge != 5*time.Second {
		t.Errorf("expected maxAge 5s, got %v", c.maxAge)
	}
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("key1", []byte("value1"))

	got, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if string(got) != "value1" {
		t.Errorf("expected value1, got %s", string(got))
	}
}

func TestCacheGetNotFound(t *testing.T) {
	c := NewCache(1 * time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent key")
	}
}

func TestCacheDelete(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("key1", []byte("value1"))
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("key1", []byte("value1"))
	c.Set("key2", []byte("value2"))
	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	c.Set("key1", []byte("value1"))

	// Should still be valid immediately
	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestCacheCleanup(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	c.Set("key1", []byte("value1"))

	// Wait for key1 to expire
	time.Sleep(60 * time.Millisecond)

	// key2 is set after key1 expires
	c.Set("key2", []byte("value2"))
	c.Cleanup()

	_, ok1 := c.Get("key1")
	_, ok2 := c.Get("key2")
	if ok1 {
		t.Error("expected key1 to be cleaned up")
	}
	if !ok2 {
		t.Error("expected key2 to still exist")
	}
}

func TestCacheSize(t *testing.T) {
	c := NewCache(1 * time.Hour)
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}

	c.Set("key1", []byte("value1"))
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}

	c.Set("key2", []byte("value2"))
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	c.Delete("key1")
	if c.Size() != 1 {
		t.Errorf("expected size 1 after delete, got %d", c.Size())
	}
}

func TestCacheOverwrite(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("key1", []byte("value1"))
	c.Set("key1", []byte("value2"))

	got, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if string(got) != "value2" {
		t.Errorf("expected value2, got %s", string(got))
	}
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}
}
