package plumbus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/jargv/plumbus/generate"
)

func infoToDynamicAdaptor(info *generate.Info, handler reflect.Value) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var queryParams url.Values
		if info.UsesQueryParams {
			queryParams = req.URL.Query()
		}

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
				interfaceVal := val
				if converter.IsPointer {
					val.Elem().Set(reflect.New(converter.Type.Elem()))
					interfaceVal = val.Elem()
				}
				err := interfaceVal.Interface().(FromRequest).FromRequest(req)
				if err != nil {
					HandleResponseError(res, req, err)
					return
				}
			case generate.ConvertStringQueryParam,
				generate.ConvertOptionalStringQueryParam,
				generate.ConvertIntQueryParam,
				generate.ConvertOptionalIntQueryParam:

				_, sent := queryParams[converter.Name]

				if !sent {
					if t.IsOptional() {
						break
					}
					HandleResponseError(res, req, Errorf(
						http.StatusBadRequest,
						"missing required query parameter '%s'",
						converter.Name,
					))
					return
				}

				paramString := queryParams.Get(converter.Name)

				setVal := val
				if t.IsOptional() {
					val.Elem().Set(reflect.New(converter.Type.Elem()))
					setVal = val.Elem()
				}

				if !t.IsInt() {
					setVal.Elem().SetString(paramString)
				} else {
					paramInt, err := strconv.Atoi(paramString)
					if err != nil {
						HandleResponseError(res, req, Errorf(
							http.StatusBadRequest,
							"query param '%s' expected to be integer value",
							converter.Name,
						))
						return
					}
					setVal.Elem().SetInt(int64(paramInt))
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
