package generate

import "net/http"

type FromRequest interface {
	FromRequest(*http.Request) (int, error)
}

type ToResponse interface {
	ToResponse(http.ResponseWriter) (int, error)
}
