package httperr

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type httperror struct {
	status  int
	message string
}

func (e httperror) Error() string {
	return fmt.Sprintf("%d: %s", e.status, e.message)
}

func (e httperror) Temporary() bool {
	return e.status >= 500
}

// Check performs additional error checking on HTTP responses.
// The response and the error from the client are passed in as inputs.
//  resp, err := httperr.Check(client.Do(req))
// On a non-2xx response, the body will be read and closed, otherwise
// you must close the response body as usual.
func Check(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, httperror{status: resp.StatusCode, message: err.Error()}
		}
		return nil, httperror{status: resp.StatusCode, message: strings.TrimSpace(string(body))}
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
