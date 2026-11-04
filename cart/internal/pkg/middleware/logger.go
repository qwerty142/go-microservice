package middleware

import (
	"context"
	"net/http"
	"time"

	"route256/cart/internal/pkg/util/logger"
)

// Logger wrapper for logging all incoming requests.
type Logger struct {
	handler http.Handler
}

func (l *Logger) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()

	l.handler.ServeHTTP(writer, request)
	logger.Infow(context.Background(), "Execution of HTTP request", "method", request.Method, "path", request.URL.Path, "duration", time.Since(start))
}

// NewLogger function to construct Logger instance.
func NewLogger(handlerToWrap http.Handler) *Logger {
	return &Logger{handlerToWrap}
}
