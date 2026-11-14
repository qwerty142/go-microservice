package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256/loms/internal/pkg/util/logger"
)

// Panic converts a handler panic into an Internal grpc error.
func Panic(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			logger.Errorw(ctx, "panic recovery", "error", e)
			err = status.Errorf(codes.Internal, "panic: %v", e)
		}
	}()

	return handler(ctx, req)
}
