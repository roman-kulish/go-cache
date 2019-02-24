package cache

import "sync"

type mapCache struct {
	cache map[string][]byte
	mu    *sync.RWMutex
}

func (c mapCache) Set(key string, value []byte) (err error) {
	c.mu.Lock()
	c.cache[key] = value
	c.mu.Unlock()

	return
}

func (c mapCache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	v := c.cache[key]
	c.mu.RUnlock()

	return v, nil
}

func NewMap(capacity uint) Cache {
	return &mapCache{
		cache: make(map[string][]byte, capacity),
		mu:    &sync.RWMutex{},
	}
}
