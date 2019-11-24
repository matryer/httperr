// +build go1.13

package httperr

import (
	"errors"
)

// Temporary checks to see if an error is temporary or whether the request
// will need to change before retrying.
func Temporary(err error) bool {
	if insideErr := errors.Unwrap(err); insideErr != nil {
		err = insideErr
	}
	type temporary interface {
		Temporary() bool
	}
	if tempErr, ok := err.(temporary); ok {
		return tempErr.Temporary()
	}
	return false
}
