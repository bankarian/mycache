package core

import (
	"github/mycache/lru"
	"sync"
)

// cache is a wrapper of lru to ensure thread safety
type cache struct {
	mu       sync.Mutex
	lru      *lru.Cache
	maxBytes int64
}

func (c *cache) add(k string, v ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.maxBytes, nil)
	}
	c.lru.Add(k, v)
}

func (c *cache) get(k string) (v ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(k); ok {
		return v.(ByteView), ok
	}
	return
}
