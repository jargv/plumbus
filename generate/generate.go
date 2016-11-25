package generate

import (
	"fmt"
	"html/template"
	"os"
	"reflect"
	"strings"
)

func Adaptor(handler interface{}, filepath, pkg string) error {
	typ := reflect.TypeOf(handler)
	info, err := CollectInfo(typ)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("adaptor").
		Funcs(template.FuncMap{
			"typename": func(arg interface{}) string {
				typename := fmt.Sprintf("%s", arg)
				return strings.Replace(typename, pkg+".", "", 1)
			},
		}).
		Option("missingkey=error").
		Parse(adaptorTemplate)

	if err != nil {
		return err
	}

	return tmpl.Execute(file, map[string]interface{}{
		"info":       info,
		"package":    pkg,
		"lastOutput": len(info.Outputs) - 1,
	})
}

const adaptorTemplate = `
package {{.package}}

//code generated by 'go generate', do not edit

import (
	"github.com/jargv/plumbus"
	"net/http"
	"reflect"
	"encoding/json"
	"fmt"
	"log"
)

// avoid unused import errors
var _ json.Delim
var _ log.Logger
var _ fmt.Formatter

func init(){
	var dummy func(
		{{range $_, $arg := .info.Inputs}}
			{{typename $arg}},
		{{end}}
	)(
		{{range $_, $output := .info.Outputs}}
			{{typename $output}},
		{{end}}
	)

	typ := reflect.TypeOf(dummy)
	plumbus.RegisterAdaptor(typ, func(handler interface{}) http.HandlerFunc {
		callback := handler.(func(
			{{range $_, $arg := .info.Inputs}}
				{{typename $arg}},
			{{end}}
		)(
			{{range $_, $output := .info.Outputs}}
				{{typename $output}},
			{{end}}
		))

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){
			{{$info := .info}}
			{{range $i, $arg := $info.Inputs}}
				var arg{{$i}} {{typename $arg}}
				{{if eq $i $info.RequestBodyIndex}}
					{
						if err := json.NewDecoder(req.Body).Decode(&arg{{$i}}); err != nil {
							msg := fmt.Sprintf("{\"error\": \"decoding json: %s\"}", err.Error())
							http.Error(res, msg, http.StatusBadRequest)
							return
						}
					}
				{{else}}
					if err := arg{{$i}}.FromRequest(req); err != nil {
						plumbus.HandleResponseError(res, req, err)
						return
					}
				{{end}}
			{{end}}

			{{$lastOutput := .lastOutput}}
			{{range $i, $_ := .info.Outputs}}
				result{{$i}} {{if eq $i $lastOutput}} := {{else}} , {{end}}
			{{end}}

			callback(
				{{range $i, $_ := .info.Inputs}}
				arg{{$i}},
				{{end}}
			)

			{{$lastIsError := .info.LastIsError}}
			{{if $lastIsError}}
				if result{{$lastOutput}} != nil {
					plumbus.HandleResponseError(res, req, result{{$lastOutput}}.(error))
					return
				}
			{{end}}

			{{range $i, $_ := .info.Outputs}}
				{{if or (ne $i $lastOutput) (not $lastIsError)}}
					{{if eq $i $info.ResponseBodyIndex}}
						{
							if err := json.NewEncoder(res).Encode(result{{$i}}); err != nil {
								plumbus.HandleResponseError(res, req, err)
								return
							}
						}
					{{else}}
						if err := result{{$i}}.ToResponse(res); err != nil {
							plumbus.HandleResponseError(res, req, err)
							return
						}
					{{end}}
				{{end}}
			{{end}}
		})
	})
}
`
