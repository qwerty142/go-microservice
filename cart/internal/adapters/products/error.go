package products

import (
	"net/http"
)

// RequestError error that is used to track response status.
type RequestError struct {
	StatusCode int
	Err        error
}

// Error get error as string.
func (r *RequestError) Error() string {
	return r.Err.Error()
}

// NotFound checks if this error had status code 404 Not Found.
func (r *RequestError) NotFound() bool {
	return r.StatusCode == http.StatusNotFound
}
