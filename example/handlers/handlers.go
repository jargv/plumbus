package handlers

import (
	"net/http"
	"strconv"
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

type body struct {
	Name   string
	Age    int
	Nachos int
}

type index int

func (i *index) FromRequest(req *http.Request) error {
	str := req.URL.Query().Get("index")
	val, err := strconv.Atoi(str)
	*i = index(val)
	return err
}

//go:generate plumbus Thing
func Thing(b body) interface{} {
	return map[string]interface{}{
		"name": "Jonboy",
		"age":  32,
	}
}

//go:generate plumbus Error
func Error() error {
	return plumbus.Errorf(404, "this is an error")
}
