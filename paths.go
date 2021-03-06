package plumbus

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Paths struct {
	handler         http.Handler
	subpaths        map[string]*Paths
	variables       map[string]*Paths
	documentation   []string
	originalHandler interface{}
}

func (p *Paths) Handle(path string, handler interface{}, documentation ...string) {
	segments := getSegments(path)
	success := p.insertSegments(segments, handler, documentation)
	if !success {
		//todo: add the route to this message
		panic(fmt.Errorf("duplicate route for path %s", path))
	}
}

func (p *Paths) insertSegments(segments []string, handler interface{}, documentation []string) bool {
	if p.subpaths == nil {
		p.subpaths = map[string]*Paths{}
	}
	if p.variables == nil {
		p.variables = map[string]*Paths{}
	}

	if len(segments) == 0 {
		if p.handler != nil {
			return false
		}
		p.handler = HandlerFunc(handler)
		p.originalHandler = handler
		p.documentation = documentation
		return true
	}

	segment := segments[0]

	insertMap := p.subpaths
	//todo: check length first
	if segment[0] == ':' {
		insertMap = p.variables
		segment = segment[1:]
	}

	sub, exists := insertMap[segment]
	if !exists {
		sub = &Paths{}
		insertMap[segment] = sub
	}

	return sub.insertSegments(segments[1:], handler, documentation)
}

func (p *Paths) findHandler(url *url.URL) http.Handler {
	segments := getSegments(url.Path)
	vals := url.Query()
	handler := p.findHandlerSegments(segments, vals)
	url.RawQuery = vals.Encode()
	return handler
}

func (p *Paths) findHandlerSegments(segments []string, query url.Values) http.Handler {
	if len(segments) == 0 {
		return p.handler
	}

	segment := segments[0]

	sub, found := p.subpaths[segment]
	if found {
		//if no match, we might have a variable match instead
		if res := sub.findHandlerSegments(segments[1:], query); res != nil {
			return res
		}
	}

	//it's either a variable or not found
	for varName, sub := range p.variables {
		if handler := sub.findHandlerSegments(segments[1:], query); handler != nil {
			query.Add(varName, segment)
			return handler
		}
	}

	return nil
}

func (p *Paths) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler := p.findHandler(req.URL)
	if handler == nil {
		http.Error(res, fmt.Sprintf("not found %s", req.URL.String()), http.StatusNotFound)
		return
	}

	handler.ServeHTTP(res, req)
}

func (p *Paths) flatten() map[string]*Paths {
	res := map[string]*Paths{}
	p.flattenMap("", res)
	return res
}

func (p *Paths) flattenMap(path string, m map[string]*Paths) {
	if p.originalHandler != nil {
		m[path] = p
	}
	for p, sub := range p.subpaths {
		sub.flattenMap(path+"/"+p, m)
	}
	for p, sub := range p.variables {
		sub.flattenMap(path+"/:"+p, m)
	}
}

func getSegments(path string) []string {
	sansSlash := strings.TrimPrefix(strings.TrimSuffix(path, "/"), "/")
	return strings.Split(sansSlash, "/")
}
