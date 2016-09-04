package midus

import (
	"encoding/json"
	"errors"
	"github.com/jargv/midus/generate"
	"log"
	"net/http"
	"reflect"
)

type adaptorFunc func(interface{}) http.HandlerFunc

var adaptors map[reflect.Type]adaptorFunc

type FromRequest generate.FromRequest
type WithResponse generate.WithResponse

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

func infoToDynamicAdaptor(info *generate.Info, handler reflect.Value) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		args := make([]reflect.Value, len(info.Inputs))
		for i, typ := range info.Inputs {
			arg := reflect.New(typ)
			if i == info.RequestBodyIndex {
				dec := json.NewDecoder(req.Body)
				err := dec.Decode(arg.Interface())
				if err != nil {
					http.Error(res, "couldn't Parse Json: "+err.Error(), http.StatusBadRequest)
					return
				}
			} else if info.IsPointer[i] {
				arg.Elem().Set(reflect.New(typ.Elem()))
				code, err := arg.Elem().Interface().(FromRequest).FromRequest(req)
				if err != nil {
					http.Error(res, err.Error(), code)
					return
				}
			} else {
				code, err := arg.Interface().(FromRequest).FromRequest(req)
				if err != nil {
					http.Error(res, err.Error(), code)
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
				msg := last.Interface().(error).Error()
				http.Error(res, msg, http.StatusInternalServerError)
				return
			}
		}

		for i, result := range results {
			if i == info.ResponseBodyIndex {
				continue
			}

			code, err := result.Interface().(WithResponse).WithResponse(res)
			if err != nil {
				http.Error(res, err.Error(), code)
				return
			}
		}

		if info.ResponseBodyIndex != -1 {
			enc := json.NewEncoder(res)
			err := enc.Encode(results[info.ResponseBodyIndex].Interface())
			if err != nil {
				http.Error(res, "couldn't serialize result into json", http.StatusInternalServerError)
				return
			}
		}
	})
}
