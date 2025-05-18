package main

import (
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

type Cache struct {
	items map[string]item
	ttl   time.Duration
	mu    sync.RWMutex
}

func NewCache(defaultTTl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]item),
		ttl:   defaultTTl,
	}

	go c.startEvictionLoop()
	return c
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:      value,
		expiration: time.Now().Add(c.ttl).UnixNano(),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if value, ok := c.items[key]; ok {
		return value, true
	}
	return nil, false
}

func (c *Cache) Delete(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		return true
	}
	return false
}

func (c *Cache) startEvictionLoop() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		c.evictExpired()
	}
}

func (c *Cache) evictExpired() {
	c.mu.RLock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for key, item := range c.items {
		if item.expiration == now {
			delete(c.items, key)
		}
	}
}
