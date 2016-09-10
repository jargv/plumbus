package generate

import "net/http"

type HTTPError interface {
	error
	ResponseCode() int
}

type FromRequest interface {
	FromRequest(*http.Request) error
}

type ToResponse interface {
	ToResponse(http.ResponseWriter) error
}
