package cache

import "sync"

type Cache struct {
	sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(key, value string) error {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
	return nil
}

func (c *Cache) Get(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	val, found := c.data[key]
	return val, found
}

func (c *Cache) Has(key string) bool {
	c.RLock()
	defer c.RUnlock()
	_, found := c.data[key]
	return found
}

func (c *Cache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()
	delete(c.data, key)
	return nil
}
