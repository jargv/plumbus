package midus

import (
	"fmt"
	"net/http"
	"strings"
)

type Method struct {
	GET    interface{}
	POST   interface{}
	PUT    interface{}
	PATCH  interface{}
	DELETE interface{}
}

type method struct {
	GET    http.Handler
	POST   http.Handler
	PUT    http.Handler
	PATCH  http.Handler
	DELETE http.Handler

	acceptedMethods string
}

func (m *Method) compile() *method {
	result := &method{}
	accepted := []string{}

	if m.GET != nil {
		result.GET = HandlerFunc(m.GET)
		accepted = append(accepted, "GET")
	}

	if m.POST != nil {
		result.POST = HandlerFunc(m.POST)
		accepted = append(accepted, "POST")
	}

	if m.PUT != nil {
		result.PUT = HandlerFunc(m.PUT)
		accepted = append(accepted, "PUT")
	}

	if m.PATCH != nil {
		result.PATCH = HandlerFunc(m.PATCH)
		accepted = append(accepted, "PATCH")
	}

	if m.DELETE != nil {
		result.DELETE = HandlerFunc(m.DELETE)
		accepted = append(accepted, "DELETE")
	}

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
	}

	if handler == nil {
		msg := fmt.Sprintf("method %s not allowed, expected {%s}", m.acceptedMethods)
		http.Error(res, msg, http.StatusMethodNotAllowed)
	}

	handler.ServeHTTP(res, req)
}
