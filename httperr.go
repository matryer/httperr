package httperr

import (
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
	const truncateAfter = 50
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, httpErr{status: resp.StatusCode, message: err.Error()}
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
	if tempErr, ok := err.(temporary); ok {
		return tempErr.Temporary()
	}
	return false
}

// Body gets the complete response body from the error.
// If the repsonse provided no body or there was an error reading it,
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
