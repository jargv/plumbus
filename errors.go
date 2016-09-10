package midus

import "fmt"

type wrappedError struct {
	error
	code int
}

func (we *wrappedError) ResponseCode() int {
	return we.code
}

func WrapError(code int, err error) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		error: err,
		code:  code,
	}
}

type simpleError struct {
	msg  string
	code int
}

func (se *simpleError) Error() string {
	return se.msg
}

func (se *simpleError) ResponseCode() int {
	return se.code
}

func Error(code int, msg string) error {
	return &simpleError{
		msg:  msg,
		code: code,
	}
}

func Errorf(code int, msg string, args ...interface{}) error {
	return &simpleError{
		code: code,
		msg:  fmt.Sprintf(msg, args...),
	}
}
