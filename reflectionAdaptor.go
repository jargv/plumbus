package plumbus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"

	"github.com/jargv/plumbus/generate"
)

func infoToDynamicAdaptor(info *generate.Info, handler reflect.Value) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var queryParams url.Values
		args := make([]reflect.Value, len(info.Inputs))
		for i, converter := range info.Inputs {
			val := reflect.New(converter.Type)
			switch t := converter.ConversionType; t {
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
			case generate.ConvertStringQueryParam:
				if queryParams == nil {
					queryParams = req.URL.Query()
				}

				if _, sent := queryParams[converter.Name]; !sent {
					HandleResponseError(res, req, Errorf(
						http.StatusBadRequest,
						"missing required query parameter '%s'",
						converter.Name,
					))
					return
				}

				val.Elem().SetString(queryParams.Get(converter.Name))
			case generate.ConvertOptionalStringQueryParam:
				if queryParams == nil {
					queryParams = req.URL.Query()
				}

				if _, sent := queryParams[converter.Name]; sent {
					val.Elem().Set(reflect.New(converter.Type.Elem()))
					val.Elem().Elem().SetString(queryParams.Get(converter.Name))
				}
			default:
				log.Fatalf("unexpected Convert Type: %s", t)
			}
			args[i] = val.Elem()
		}
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
			switch t := converter.ConversionType; t {
			case generate.ConvertBody:
				//do nothing, the response body has to be sent last
			case generate.ConvertInterface:
				err := results[i].Interface().(ToResponse).ToResponse(res)
				if err != nil {
					HandleResponseError(res, req, err)
					return
				}
			default:
				log.Fatalf("unexpected Convert Type: %s", t)
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
