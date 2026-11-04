package middleware

import (
	"net/http"
)

type ResponseWriterWrapper struct {
	writer     http.ResponseWriter
	statusCode int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{writer: w, statusCode: http.StatusOK}
}

func (writerWrapper *ResponseWriterWrapper) Header() http.Header {
	return writerWrapper.writer.Header()
}

func (writerWrapper *ResponseWriterWrapper) Write(bytes []byte) (int, error) {
	return writerWrapper.writer.Write(bytes)
}

func (writerWrapper *ResponseWriterWrapper) WriteHeader(statusCode int) {
	writerWrapper.statusCode = statusCode
	writerWrapper.writer.WriteHeader(statusCode)
}

func (writerWrapper *ResponseWriterWrapper) GetStatusCode() int {
	return writerWrapper.statusCode
}
