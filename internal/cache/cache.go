package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type MemoryCache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]CacheItem),
	}
	go cache.cleanupExpiredItems()
	return cache
}

// Set Добавить элемент в кэш
func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
	c.mu.Unlock()
}

// Get Получить элемент из кэша
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()
	if !found {
		return nil, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete Удаление элемента из кэша
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// cleanupExpiredItems Очистка кэша со временем
func (c *MemoryCache) cleanupExpiredItems() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now().UnixNano()
		c.mu.Lock()
		for key, item := range c.items {
			if item.Expiration > 0 && now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
	ticker.Stop()
}
