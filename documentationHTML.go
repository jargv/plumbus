package plumbus

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"
)

func (sm *ServeMux) DocumentationHTML(intro ...string) http.Handler {
	var spec *Documentation
	var lock sync.Mutex
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		if spec == nil {
			spec = sm.Documentation(intro...)
			sort.Sort(docOrder(spec.Endpoints))
		}

		res.Header().Add("content-type", "text/html")
		err := page.Execute(res, spec)
		if err != nil {
			log.Fatalf("executing template:", err)
		}
	})
}

var page = template.Must(template.New("docs page").Parse(`
<head>
  <style>
		*{
			margin: 0;
			padding: 0;
		}

		div {
		}

		.intro {
			padding: 20px;
		}

		.endpoint {
			padding: 20px;
			padding-left: 30px;
			padding-top: 5px;
			border-top: 1px solid #ddd;
		}

		.endpoint>div {
			padding: 20px;
		}
	</style>
</head>
<body>
	<div class="intro">
	  {{range .Introduction}}
			<p>{{.}}</p>
		{{end}}
	</div>
	{{range .Endpoints}}
		<div class="endpoint">
		  <h2>
				<span>{{.Method}}</span> <span>{{.Path}}</span>
			</h2>
			{{if .Description}}
			  <p>
					{{.Description}}
				</p>
			{{end}}
			{{range .Notes}}
				<p>
					{{.}}
				</p>
			{{end}}
			{{if .Params}}
			  <div>
					<h3>Params</h3>
					{{range $key, $val := .Params}}
					  <div>
							<span class="paramName">{{$key}}</span> (
							{{- if $val.Required -}}
								Required
							{{- else -}}
								Optional
							{{- end}} {{$val.Type}}): {{$val.Description}}
						</div>
					{{end}}
				</div>
			{{end}}
			{{if .RequestBody}}
			  <div>
					<h3>Requst Body</h3>
					<div>
						{{.RequestBody}}
					</div>
				</div>
			{{end}}
			{{if .ResponseBody}}
				<div>
					<h3>Response Body</h3>
					<div>
						{{.ResponseBody}}
					</div>
				</div>
			{{end}}
		</div>
	{{end}}
</body>
`))

type docOrder []*Endpoint

func (d docOrder) Len() int {
	return len(d)
}

func (d docOrder) Less(i, j int) bool {
	if d[i].Path == d[j].Path {
		return d[i].Method < d[j].Method
	}
	return d[i].Path < d[j].Path
}

func (d docOrder) Swap(i, j int) {
	tmp := d[i]
	d[i] = d[j]
	d[j] = tmp
}
