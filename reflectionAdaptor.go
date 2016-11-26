package plumbus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/jargv/plumbus/generate"
)

func infoToDynamicAdaptor(info *generate.Info, handler reflect.Value) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		args := make([]reflect.Value, len(info.Inputs))
		for i, converter := range info.Inputs {
			val := reflect.New(converter.Type)
			switch converter.ConversionType {
			case generate.ConvertBody:
				if err := json.NewDecoder(req.Body).Decode(val.Interface()); err != nil {
					msg := fmt.Sprintf(`{"error": "decoding json: %s"}`, err.Error())
					http.Error(res, msg, http.StatusBadRequest)
					return
				}
			case generate.ConvertInterface:
				if converter.IsPointer {
					val.Elem().Set(reflect.New(converter.Type.Elem()))
					val = val.Elem()
				}
				err := val.Interface().(FromRequest).FromRequest(req)
				if err != nil {
					HandleResponseError(res, req, err)
					return
				}
				if converter.IsPointer {
					val = val.Addr()
				}
			}
			args[i] = val.Elem()
		}
		log.Printf("args: %#v", args)
		results := handler.Call(args)

		if info.LastIsError {
			last := results[len(results)-1]
			results = results[:len(results)-1]
			if !last.IsNil() {
				err := last.Interface().(error)
				HandleResponseError(res, req, err)
				return
			}
		}

		for i, converter := range info.Outputs {
			switch converter.ConversionType {
			case generate.ConvertBody:
				//do nothing, the response body has to be sent last
			case generate.ConvertInterface:
				err := results[i].Interface().(ToResponse).ToResponse(res)
				if err != nil {
					HandleResponseError(res, req, err)
					return
				}
			}
		}

		if info.ResponseBodyIndex != -1 {
			enc := json.NewEncoder(res)
			err := enc.Encode(results[info.ResponseBodyIndex].Interface())
			if err != nil {
				HandleResponseError(res, req, err)
				return
			}
		}
	})
}
