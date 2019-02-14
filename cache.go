package goapollo

import "sync"

type cache struct {
	m *sync.Map
}

func NewCache() *cache {
	return &cache{m: &sync.Map{}}
}

func (c *cache) Load(k string) (string, bool) {
	if v, ok := c.m.Load(k); ok {
		if s, ok := v.(string); ok {
			return s, ok
		}
	}
	return "", false
}

func (c *cache) Store(k, v string) {
	c.m.Store(k, v)
}

func (c *cache) Delete(k string) {
	c.m.Delete(k)
}

func (c *cache) Range(f func(k, v string) bool) {
	c.m.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(string)

		return f(k, v)
	})
}

func (c *cache) LoadOrStore(key, value string) (actual string, loaded bool) {
	v, ok := c.m.LoadOrStore(key, value)
	loaded = ok

	if v, ok := v.(string); ok {
		actual = v
	}
	return
}
