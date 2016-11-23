package plumbus

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/jargv/plumbus/generate"
)

type Documentation struct {
	Endpoints []*Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Method       string `json:"method,omitempty"`
	Path         string `json:"path"`
	RequestBody  *Body  `json:"requestBody,omitempty"`
	ResponseBody *Body  `json:"responseBody,omitempty"`
}

type Body struct {
	Description string      `json:"description,omitempty"`
	Example     interface{} `json:"example"`
}

func (sm *ServeMux) Documentation() *Documentation {
	d := &Documentation{
		Endpoints: []*Endpoint{},
	}

	for path, handler := range sm.Paths.flatten() {
		es := handlerToEndpoints(handler)
		for _, e := range es {
			e.Path = path
			d.Endpoints = append(d.Endpoints, e)
		}
	}

	return d
}

func handlerToEndpoints(handler interface{}) []*Endpoint {
	switch val := handler.(type) {
	case http.HandlerFunc, func(http.ResponseWriter, *http.Request):
		return []*Endpoint{&Endpoint{}}

	case ByMethod:
		return methodHandlerToEndpoints(&val)

	case *ByMethod:
		return methodHandlerToEndpoints(val)

	default:
		return []*Endpoint{handlerFunctionToEndpoint(handler)}
	}
}

func methodHandlerToEndpoints(handlers *ByMethod) []*Endpoint {
	var result []*Endpoint

	addToResult := func(method string, handler interface{}) {
		if handler != nil {
			e := handlerFunctionToEndpoint(handler)
			e.Method = method
			result = append(result, e)
		}
	}

	addToResult("GET", handlers.GET)
	addToResult("POST", handlers.POST)
	addToResult("PUT", handlers.PUT)
	addToResult("PATCH", handlers.PATCH)
	addToResult("DELETE", handlers.DELETE)
	addToResult("OPTIONS", handlers.OPTIONS)

	return result
}

func handlerFunctionToEndpoint(handler interface{}) *Endpoint {
	typ := reflect.TypeOf(handler)
	if typ.Kind() != reflect.Func {
		return &Endpoint{}
	}

	info, err := generate.CollectInfo(typ)
	if err != nil {
		panic(fmt.Errorf("error generating documentation: %v", err))
	}

	log.Printf("info: %#v", info)

	e := &Endpoint{}

	if info.RequestBodyIndex != -1 {
		e.RequestBody = body(info.Inputs[info.RequestBodyIndex])
	}

	if info.ResponseBodyIndex != -1 {
		e.ResponseBody = body(info.Outputs[info.ResponseBodyIndex])
	}

	return e
}

func body(typ reflect.Type) *Body {
	ex := deepZero(typ).Interface()
	log.Printf("ex: %#v", ex)
	return &Body{
		Example: ex,
	}
}

func deepZero(typ reflect.Type) reflect.Value {
	needsExample := func(v reflect.Value) bool {
		isPtrOrSliceOrMap := v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Map
		canSet := isPtrOrSliceOrMap && v.CanSet()
		return isPtrOrSliceOrMap && canSet
	}

	if typ.Kind() == reflect.Ptr {
		val := deepZero(typ.Elem())
		if val.CanAddr() {
			return val.Addr()
		} else {
			return reflect.New(typ.Elem())
		}
	}

	if typ.Kind() == reflect.Slice {
		slice := reflect.MakeSlice(typ, 1, 1)
		val := slice.Index(0)
		if needsExample(val) {
			val.Set(deepZero(typ.Elem()))
		}
		return slice
	}

	if typ.Kind() == reflect.Map {
		log.Println("doing a map")
		m := reflect.MakeMap(typ)
		key := deepZero(typ.Key())
		val := deepZero(typ.Elem())
		m.SetMapIndex(key, val)
		return m
	}

	val := reflect.New(typ).Elem()

	if typ.Kind() != reflect.Struct {
		return val
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if needsExample(field) {
			val.Field(i).Set(deepZero(val.Field(i).Type()))
		}
	}
	return val
}
