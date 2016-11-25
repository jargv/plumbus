package main

import (
	"html/template"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type generator struct {
	Path   string
	Pkg    string
	Func   string
	Type   string
	Target string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatalf("plumbus: requires an argument to generate adapter for")
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var typ, f string
	target := os.Args[1]

	if parts := strings.Split(target, "."); len(parts) == 2 {
		f = parts[1]
		typ = parts[0]
	} else if len(parts) == 1 {
		f = target
	} else {
		log.Fatalf("plumbus: target %s should have at most one '.' character", target)
	}

	pkg := os.Getenv("GOPACKAGE")

	if pkg == "main" {
		log.Fatalf("plumbus: can't generate for package main, move handlers into another package")
	}

	g := generator{
		Path: path.Join(dir, target+".adaptor-generated.go"),
		Pkg:  pkg,
		Func: f,
		Type: typ,
	}

	g.generate()
}

func (g *generator) generate() {
	path := path.Join(os.TempDir(), "plumbus-generator.go")
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("creating file: %v", err)
	}

	err = generatorTemplate.Execute(file, g)
	if err != nil {
		log.Fatalf("failure: %v", err)
	}

	cmd := exec.Command("goimports", "-w", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf(string(out))
		panic(err)
	}

	cmd = exec.Command("go", "run", path)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("running go generate: %s", string(out))
	}
}

var generatorTemplate = template.Must(
	template.New("generator").
		Option("missingkey=error").
		Parse(templateString),
)

var templateString string = `
	package main

	import (
		"github.com/jargv/plumbus/generate"
		"os"
		"log"
	)

	func main(){
		{{if (len .Type)}}
			v := {{.Pkg}}.{{.Type}}{}
			f := v.{{.Func}}
		{{else}}
			f := {{.Pkg}}.{{.Func}}
		{{end}}
		err := generate.Adaptor(f, "{{.Path}}", "{{.Pkg}}")
		if err != nil {
			log.Printf("couldn't generate: %s", err)
			os.Exit(1)
		}
	}
`
