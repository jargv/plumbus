package main

import (
	"log"
	"net/http"

	"github.com/jargv/plumbus"
	"github.com/jargv/plumbus/example/handlers"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	mux := plumbus.NewServeMux()

	counter := &handlers.Counter{}
	mux.Handle("/incr", counter.Incr)
	mux.Handle("/count", counter.Count)
	mux.Handle("/error", handlers.Error)
	mux.Handle("/echo", handlers.EchoParam)
	mux.Handle("/user/:userId", plumbus.ByMethod{
		PATCH: handlers.EditUser,
		GET:   handlers.GetUser,
	}, `
	  this should be the same on patch and get
	`)
	mux.Handle("/custom", handlers.HandleCustom)
	mux.Handle("/nachos", handlers.GetNachos)
	mux.Handle("/spec", func() interface{} {
		return mux.Documentation()
	})
	mux.Handle("/docs", mux.DocumentationHTML(`
	  This is the top-level documentation. It will be a paragraph
	`, `
	  The next paragraph can come next. And so on.
	`))

	log.Println("listening on port 8000")
	http.ListenAndServe(":8000", mux)
	log.Println("hello world")
}
