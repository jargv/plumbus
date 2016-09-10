package midus

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/jargv/midus/generate"
)

type adaptorFunc func(interface{}) http.HandlerFunc

var adaptors map[reflect.Type]adaptorFunc

type FromRequest generate.FromRequest
type ToResponse generate.ToResponse
type HTTPError generate.HTTPError

type ServeMux struct {
	*http.ServeMux
}

func RegisterAdaptor(typ reflect.Type, adaptor adaptorFunc) {
	if adaptors == nil {
		adaptors = make(map[reflect.Type]adaptorFunc)
	}
	adaptors[typ] = adaptor
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		ServeMux: http.NewServeMux(),
	}
}

func (sm *ServeMux) Handle(route string, fn interface{}) {
	defer func() {
		err := recover()
		if err, ok := err.(error); ok {
			panic(fmt.Errorf("Error while routing %s: %s", route, err.Error()))
		}
	}()
	switch fn := fn.(type) {
	case func(http.ResponseWriter, *http.Request):
		sm.ServeMux.HandleFunc(route, fn)
	case http.Handler:
		sm.ServeMux.Handle(route, fn)
	default:
		sm.ServeMux.Handle(route, HandlerFunc(fn))
	}
}

func HandlerFunc(handler interface{}) http.Handler {
	typ := reflect.TypeOf(handler)
	adaptor, exists := adaptors[typ]
	if !exists {
		log.Printf("WARNING: function of type `%v` using slow reflection adaptor", typ)
		log.Printf("NOTE   : run go generate")
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
			panic(errors.New("internal midus error. Mismatch of types."))
		}
		info, err := generate.CollectInfo(typ)
		if err != nil {
			panic(err)
		}
		return infoToDynamicAdaptor(info, val)
	}
}

func responseError(res http.ResponseWriter, err error) {
	if err, ok := err.(HTTPError); ok {
		http.Error(res, err.Error(), err.ResponseCode())
	} else {
		http.Error(res, "", http.StatusInternalServerError)
	}
}

func infoToDynamicAdaptor(info *generate.Info, handler reflect.Value) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		args := make([]reflect.Value, len(info.Inputs))
		for i, typ := range info.Inputs {
			arg := reflect.New(typ)
			if i == info.RequestBodyIndex {
				dec := json.NewDecoder(req.Body)
				err := dec.Decode(arg.Interface())
				if err != nil {
					msg := fmt.Sprintf("error decoding json: %s", err.Error())
					http.Error(res, msg, http.StatusBadRequest)
					return
				}
			} else if info.IsPointer[i] {
				arg.Elem().Set(reflect.New(typ.Elem()))
				err := arg.Elem().Interface().(FromRequest).FromRequest(req)
				if err != nil {
					responseError(res, err)
					return
				}
			} else {
				err := arg.Interface().(FromRequest).FromRequest(req)
				if err != nil {
					responseError(res, err)
					return
				}
			}
			args[i] = arg.Elem()
		}
		results := handler.Call(args)

		if info.LastIsError {
			last := results[len(results)-1]
			results = results[:len(results)-1]
			if !last.IsNil() {
				err := last.Interface().(error)
				responseError(res, err)
				return
			}
		}

		for i, result := range results {
			if i == info.ResponseBodyIndex {
				continue
			}

			err := result.Interface().(ToResponse).ToResponse(res)
			if err != nil {
				responseError(res, err)
				return
			}
		}

		if info.ResponseBodyIndex != -1 {
			enc := json.NewEncoder(res)
			err := enc.Encode(results[info.ResponseBodyIndex].Interface())
			if err != nil {
				log.Printf("json encoding error: %s", err.Error())
				http.Error(res, "", http.StatusInternalServerError)
				return
			}
		}
	})
}
