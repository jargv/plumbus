package generate

import "net/http"

type FromRequest interface {
	FromRequest(*http.Request) (int, error)
}

type WithResponse interface {
	WithResponse(http.ResponseWriter) (int, error)
}
