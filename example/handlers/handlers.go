package handlers

import "sync"

type Counter struct {
	lock  sync.Mutex
	count int
}

//go:generate plumbus Counter.Incr
func (c *Counter) Incr() *Counter {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count++
	return c
}

//go:generate plumbus Counter.Count
func (c Counter) Count() map[string]interface{} {
	return map[string]interface{}{
		"count": c.Count,
	}
}

//go:generate plumbus Thing
func Thing() interface{} {
	return map[string]interface{}{
		"name": "Jonboy",
		"age":  32,
	}
}
