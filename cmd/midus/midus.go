package main

import (
	"html/template"
	"log"
	"os"
	"os/exec"
	"path"
)

type generator struct {
	Path   string
	Pkg    string
	Target string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Println("midus requires an argument to generate adapter for")
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	target := os.Args[1]
	g := generator{
		Path:   path.Join(dir, target+".midus-generated.go"),
		Pkg:    os.Getenv("GOPACKAGE"),
		Target: target,
	}

	g.generate()
}

func (g *generator) generate() {
	goFile := path.Join(os.TempDir(), "midus-generator.go")
	file, err := os.Create(goFile)
	if err != nil {
		panic(err)
	}
	generatorTemplate.Execute(file, g)

	cmd := exec.Command("goimports", "-w", goFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf(string(out))
		panic(err)
	}

	cmd = exec.Command("go", "run", goFile)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Print(string(out))
		panic(err)
	}
}

var generatorTemplate = template.Must(template.New("generator").Parse(`
package main

import (
	"github.com/jargv/midus/generate"
	"os"
	"log"
)

func main(){
	err := generate.Adaptor({{.Pkg}}.{{.Target}}, "{{.Path}}", "{{.Pkg}}", "{{.Target}}")
	if err != nil {
		log.Printf("couldn't generate: %s", err)
		os.Exit(1)
	}
}
`))
