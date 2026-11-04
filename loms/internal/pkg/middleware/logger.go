package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"route256/loms/internal/pkg/util/logger"
)

func Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// для превращения protbuf структур в json используем google.golang.org/protobuf/encoding/protojson пакет а не encoding/json
	raw, _ := protojson.Marshal((req).(proto.Message))
	logger.Infow(ctx, "request log", "method", info.FullMethod, "req", string(raw))

	if resp, err = handler(ctx, req); err != nil {
		logger.Infow(ctx, "response log", "method", info.FullMethod, "err", err)
		return
	}

	rawResp, _ := protojson.Marshal((resp).(proto.Message))

	// Пример академический
	// Логировать все нельзя, в объекте могут быть чувствительные к безопасности данные (персональные)
	// Также сами объекты для логирования могут быть очень большими
	logger.Infow(ctx, "response log", "method", info.FullMethod, "resp", string(rawResp))

	return resp, err
}

func WithHTTPLoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Infow(context.Background(), "logging middleware", "method", r.Method)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
