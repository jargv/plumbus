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
	mux.Handle("/thing", handlers.Thing)
	mux.Handle("/error", handlers.Error)

	log.Println("listening on port 8000")
	http.ListenAndServe(":8000", mux)
	log.Println("hello world")
}
