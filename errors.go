package retwitch

import (
	"errors"
	"net/http"
)

var ErrNoSuchBadge = errors.New("no such badge")
var ErrHTTPStatus = errors.New("http response error")

type httpStatusError struct {
	*http.Response
}

func (e httpStatusError) Error() string {
	return e.Request.Method + " " + e.Request.URL.String() + " returned " + e.Status
}

func (e httpStatusError) Unwrap() error {
	return ErrHTTPStatus
}
