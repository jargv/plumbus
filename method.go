package plumbus

import (
	"fmt"
	"net/http"
	"strings"
)

type ByMethod struct {
	GET, POST, PUT, PATCH, DELETE, OPTIONS interface{}
}

type method struct {
	GET, POST, PUT, PATCH, DELETE, OPTIONS http.Handler
	acceptedMethods                        string
}

func (m *ByMethod) compile() *method {
	result := &method{}
	accepted := []string{}

	handle := func(name string, handler interface{}) http.Handler {
		if handler == nil {
			return nil
		}

		accepted = append(accepted, name)
		return HandlerFunc(handler)
	}

	result.GET = handle("GET", m.GET)
	result.POST = handle("POST", m.POST)
	result.PUT = handle("PUT", m.PUT)
	result.PATCH = handle("PATCH", m.PATCH)
	result.DELETE = handle("DELETE", m.DELETE)
	result.OPTIONS = handle("OPTIONS", m.OPTIONS)

	if len(accepted) == 0 {
		result.acceptedMethods = "<none>"
	} else {
		result.acceptedMethods = strings.Join(accepted, ", ")
	}

	return result
}

func (m *method) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var handler http.Handler
	switch strings.ToUpper(req.Method) {
	case "GET":
		handler = m.GET
	case "POST":
		handler = m.POST
	case "PUT":
		handler = m.PUT
	case "PATCH":
		handler = m.PATCH
	case "DELETE":
		handler = m.DELETE
	case "OPTIONS":
		handler = m.OPTIONS
	}

	if handler == nil {
		msg := fmt.Sprintf("method %s not allowed, expected {%s}", m.acceptedMethods)
		http.Error(res, msg, http.StatusMethodNotAllowed)
		return
	}

	handler.ServeHTTP(res, req)
}
