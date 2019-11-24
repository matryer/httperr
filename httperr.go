package httperr

// from https://github.com/matryer/httperr
// license: MIT https://github.com/matryer/httperr/blob/master/LICENSE

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Check performs additional error checking on HTTP responses.
// The response and the error from the client are passed as inputs.
// If an error is returned the body will be read and closed, otherwise
// you must close the response body as usual.
//  resp, err := httperr.Check(client.Do(req))
//  if err != nil {
//   	// handle error
//		return err
//  }
//  defer resp.Body.Close()
// You do not need to explicitly check the error from client.Do, Check
// will do that for you.
func Check(resp *http.Response, err error) (*http.Response, error) {
	// truncateAfter is the maximum length of the body to include.
	const truncateAfter = 100
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, httpErr{err: err, status: resp.StatusCode, message: err.Error()}
		}
		s := strings.TrimSpace(string(body))
		if len(s) > truncateAfter {
			s = s[:truncateAfter] + "..."
		}
		return nil, httpErr{status: resp.StatusCode, message: s, body: body}
	}
	return resp, nil
}

// Temporary checks to see if an error is temporary or whether the request
// will need to change before retrying.
func Temporary(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	if insideErr := errors.Unwrap(err); err != nil {
		err = insideErr
	}
	if tempErr, ok := err.(temporary); ok {
		return tempErr.Temporary()
	}
	return false
}

// Body gets the complete response body from the error.
// If the response provided no body or there was an error reading it,
// nil is returned.
// To get the body of successful requests, access it in the usual way
// from the http.Response object.
func Body(err error) []byte {
	if h, ok := err.(httpErr); ok {
		return h.body
	}
	return nil
}

type httpErr struct {
	err     error
	status  int
	message string
	body    []byte
}

func (e httpErr) Error() string {
	return fmt.Sprintf("%d: %s", e.status, e.message)
}

// Temporary returns true for error status codes above 500.
func (e httpErr) Temporary() bool {
	return e.status >= 500
}

// Unwrap gets the underlying error that was returned when
// attempting to make this request. May be nil if it was a higher
// level (i.e. bad status code) error.
func (e httpErr) Unwrap() error {
	return e.err
}
