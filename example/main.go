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
	mux.Handle("/docs", func() interface{} {
		return mux.Documentation()
	}, `
	  this endpoint gives you back the documentation
		metatdata. Note that this is just a giant json
		object with the data. Real documentation would
		require some rendering.
	`)

	log.Println("listening on port 8000")
	http.ListenAndServe(":8000", mux)
	log.Println("hello world")
}
