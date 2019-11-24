// +build !go1.13

package httperr

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
