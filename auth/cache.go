package auth

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

// ICache defines the interface for a generic cache
type ICache interface {
	Add(key, value interface{}) bool
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{}) bool
	Purge()
}

// LRUCache implements the ICache interface using an LRU cache
type LRUCache struct {
	cache *lru.Cache
}

// NewLRUCache creates a new LRUCache with the specified size
func NewLRUCache(size int) (ICache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %v", err)
	}
	return &LRUCache{cache: cache}, nil
}

// Add adds a value to the cache
func (c *LRUCache) Add(key, value interface{}) bool {
	return c.cache.Add(key, value)
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	return c.cache.Get(key)
}

// Remove removes a value from the cache
func (c *LRUCache) Remove(key interface{}) bool {
	c.cache.Remove(key)
	return true
}

// Purge clears all entries in the cache
func (c *LRUCache) Purge() {
	c.cache.Purge()
}
