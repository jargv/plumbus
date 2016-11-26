package handlers

import (
	"sync"

	"github.com/jargv/plumbus"
)

type Counter struct {
	lock     sync.Mutex
	HitCount int `json:"count"`
}

//go:generate plumbus Counter.Incr
func (c *Counter) Incr() *Counter {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.HitCount++
	return c
}

//go:generate plumbus Counter.Count
func (c Counter) Count() map[string]interface{} {
	return map[string]interface{}{
		"count": c.HitCount,
	}
}

type messageQueryParam string

func EchoParam(message messageQueryParam) map[string]string {
	return map[string]string{
		"message": string(message),
	}
}

//go:generate plumbus Error
func Error() error {
	return plumbus.Errorf(404, "this is an error")
}
