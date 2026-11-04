package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"route256/cart/internal/pkg/util/logger"
)

// RetryableTransport transport that can be used to send request multiple times depending on response status code.
type RetryableTransport struct {
	transport   http.RoundTripper
	maxRetries  int
	retryDelay  time.Duration
	retryStatus []int
}

// NewRetryableTransport constructor for RetryableTransport.
func NewRetryableTransport(
	transport http.RoundTripper,
	maxRetries int,
	retryDelay time.Duration,
	retryStatus []int,
) *RetryableTransport {
	return &RetryableTransport{
		transport:   transport,
		maxRetries:  maxRetries,
		retryDelay:  retryDelay,
		retryStatus: retryStatus,
	}
}

// RoundTrip implementation of RoundTripper interface.
func (rt *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	resp, err := rt.transport.RoundTrip(req)
	for retries := 1; rt.shouldRetry(err, resp) && retries <= rt.maxRetries; retries++ {
		logger.Infow(context.Background(), fmt.Sprintf("Retry, attempt %d", retries))

		time.Sleep(rt.retryDelay)

		drainBody(resp)

		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = rt.transport.RoundTrip(req)
	}

	return resp, err
}

func (rt *RetryableTransport) shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	return slices.Contains(rt.retryStatus, resp.StatusCode)
}

func drainBody(resp *http.Response) {
	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
