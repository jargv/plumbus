# plumbus - Turn any function into a net/http handler

A Flexible ServeMux and HandlerFunc - Use any function as a
handler, as long as you implement a few interfaces to define
how that function's arguments, results, and errors are mapped to
the http request and response.

## Arguments
Arguments must implement `plumbus.FromRequest`, which looks
like:

```go
type FromRequest interface {
	FromRequest(*http.Request) error
}
```

However, if (at most) one parameter does *not* implement this
interface, then that parameter will be decoded from the
request body as json instead, by using the standard
encoding/json package (supporting other types in the future
is possible).

## Return Values
Return values must implement `plumbus.ToResponse`, which looks
like:

```go
type ToResponse interface {
	ToResponse(http.ResponseWriter) error
}
```

However, if (at most) one return value does *not* implement this
interface, then that return value will be encoded to the
response body as json instead, by using the standard
encoding/json package (supporting other types in the future
is possible).

## Errors
If a function returns an error (must be the last return
value), then the result will be a 500 internal server error
and no additional information about the error will be shown.
*However*, if the error also implements the
`plumbus.HTTPError` interface then the response code will be
obtained by calling `error.ResponseCode()`, and the response
body will be obtained from calling `error.Error()`. The same
is true for errors returned by calls to `FromRequest` and
`ToRequest`. The `plumbus.HTTPError` interface looks like:

```go
type HTTPError interface {
	error
	ResponseCode() int
}
```

## Isn't Reflection too slow?
Probabaly for some uses, *however* you can also run the
`plumbus` command line tool via `go generate` to use code
generation in place of reflection and  eliminate all
per-request reflection. (there will still be a small amount
of reflection during setup).

## Routing on Methods
To route by HTTP methods, just pass a value of type
plumbus.ByMethod as your handler. This type maps from HTTP
methods to (flexible) handler functions.
```go
mux := plumbus.NewServeMux()
mux.Handle("/user", &plumbus.ByMethod{
  GET: getUser,
  POST: createUser
  //other methods supported, but all are optional
})
```

##Path Parameters
Path parameters are also supported. Example:
```go
//handles cases such as /user/10/info
mux.Handle("/user/:userId/info", userInfo)
```
the path parameters will be translated into request
parameters (available on req.URL.Query() before your
handler is called)

##TODO
- Make generate work with methods
- Add a tutorial
- Add plumbus.Params type
- Document the automatic documentation feature
