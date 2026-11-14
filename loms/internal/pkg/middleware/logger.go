package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"route256/loms/internal/pkg/util/logger"
)

// Logger logs the incoming request and the outgoing response or error.
func Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// protobuf structures are marshaled with protojson, not encoding/json
	if msg, ok := req.(proto.Message); ok {
		if raw, marshalErr := protojson.Marshal(msg); marshalErr == nil {
			logger.Infow(ctx, "request log", "method", info.FullMethod, "req", string(raw))
		}
	}

	if resp, err = handler(ctx, req); err != nil {
		logger.Infow(ctx, "response log", "method", info.FullMethod, "err", err)
		return
	}

	// Academic example: logging everything is unsafe (personal data)
	// and the objects themselves may be very large.
	if msg, ok := resp.(proto.Message); ok {
		if raw, marshalErr := protojson.Marshal(msg); marshalErr == nil {
			logger.Infow(ctx, "response log", "method", info.FullMethod, "resp", string(raw))
		}
	}

	return resp, err
}

// WithHTTPLoggingMiddleware logs incoming HTTP requests.
func WithHTTPLoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Infow(r.Context(), "logging middleware", "method", r.Method)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
