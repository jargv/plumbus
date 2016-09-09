# midus - Use any function as a net/http handler

## How?
Reflection. But keep reading... there's more to it.

## Isn't Reflection too slow?
Probabaly not for many uses, *however* you can also
run the `midus` command line tool via `go generate` to
use code generation in place of reflection and  eliminate
all per-request reflection. (there will still be a small
amount of reflection during setup).

## Are there any constraints on the functions?
Yes there are:
1. All but one parameter must implement the FromRequest
   interface
2. All but one result must implement to ToResponse interface
3. If the last result is an error, and is not nil, the http
   response code will be 500 (internal server error)
   *unless* the error also implements HTTPError. Then the
   response code will be retrieved from err.ResponseCode()
   and the Response body will be err.Error()

If at most one argument, and at most one result do not
implement these interfaces, they will be treated as the http
request and http response bodies, respectively. They will be
decoded/encoded from JSON. Supporting other formats in the
future is possible.

##TODO
- Create the HTTPError interface, make that work
- Make the Code Gen match the reflection
- Allow any parameter to be the *http.Request or
  http.ResponseWriter of a normal handler
- Add a tutorial
