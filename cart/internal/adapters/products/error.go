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

// Unwrap returns the underlying error, so RequestError participates in errors.Is/errors.As chains.
func (r *RequestError) Unwrap() error {
	return r.Err
}

// NotFound checks if this error had status code 404 Not Found.
func (r *RequestError) NotFound() bool {
	return r.StatusCode == http.StatusNotFound
}