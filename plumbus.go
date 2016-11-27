// A Flexible ServeMux and HandlerFunc - Implement interfaces to
// determine how function arguments, results, and errors are mapped to
// the http request and response. Then write functions instead of
// http.Handlers or http.HandlerFunc's
package plumbus

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"

	"github.com/jargv/plumbus/generate"
)

type adaptorFunc func(interface{}) http.HandlerFunc

var adaptors map[reflect.Type]adaptorFunc

type FromRequest generate.FromRequest
type ToResponse generate.ToResponse
type HTTPError generate.HTTPError

func RegisterAdaptor(typ reflect.Type, adaptor adaptorFunc) {
	if adaptors == nil {
		adaptors = make(map[reflect.Type]adaptorFunc)
	}
	adaptors[typ] = adaptor
}

type ServeMux struct {
	*Paths
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		Paths: &Paths{},
	}
}

func (sm *ServeMux) Handle(route string, fn interface{}, documentation ...string) {
	defer func() {
		err := recover()
		if err, ok := err.(error); ok {
			panic(fmt.Errorf("Error while routing %s: %s", route, err.Error()))
		}
	}()

	sm.Paths.Handle(route, fn, documentation...)
}

func HandlerFunc(handler interface{}) http.Handler {
	switch val := handler.(type) {
	case func(http.ResponseWriter, *http.Request):
		return http.HandlerFunc(val)
	case http.Handler:
		return val
	case ByMethod:
		return val.compile()
	case *ByMethod:
		return val.compile()
	}

	typ := reflect.TypeOf(handler)
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf(
			"plumbus.HandlerFunc called on non-function type %v",
			typ,
		))
	}

	adaptor, exists := adaptors[typ]
	if !exists {
		name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		log.Printf("WARNING: using slow reflection adaptor for function: %s", name)
		log.Printf("NOTE   : annotate with `//go:generate plumbus <function name>` and run `go generate`")
		adaptor = makeDynamicAdaptor(typ)
		if adaptors == nil {
			adaptors = make(map[reflect.Type]adaptorFunc)
		}
		adaptors[typ] = adaptor
	}
	return adaptor(handler)
}

func makeDynamicAdaptor(typ reflect.Type) adaptorFunc {
	return func(handler interface{}) http.HandlerFunc {
		val := reflect.ValueOf(handler)
		if typ != val.Type() {
			panic(errors.New("internal plumbus error. Mismatch of types."))
		}
		info, err := generate.CollectInfo(typ)
		if err != nil {
			panic(err)
		}
		return infoToDynamicAdaptor(info, val)
	}
}

func HandleResponseError(res http.ResponseWriter, req *http.Request, err error) {
	if httperr, ok := err.(HTTPError); ok {
		res.WriteHeader(httperr.ResponseCode())
		json.NewEncoder(res).Encode(map[string]interface{}{
			"error": httperr.Error(),
		})
	} else {
		log.Printf(
			"error handling request: %s %s: %v",
			req.Method,
			req.URL.Path,
			err,
		)
		body := `{"error":"internal server error"}`
		http.Error(res, body, http.StatusInternalServerError)
	}
}
