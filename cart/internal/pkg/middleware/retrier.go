package middleware

import (
	"bytes"
	"context"
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
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	resp, err := rt.transport.RoundTrip(req)
	for retries := 1; rt.shouldRetry(err, resp) && retries <= rt.maxRetries; retries++ {
		logger.Infow(req.Context(), "Retry request", "attempt", retries, "url", req.URL.String())

		if !waitForDelay(req.Context(), rt.retryDelay) {
			drainBody(resp)
			return nil, req.Context().Err()
		}

		drainBody(resp)

		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
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

// waitForDelay waits for the given duration and reports whether the wait completed.
func waitForDelay(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func drainBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
