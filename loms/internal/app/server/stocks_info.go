package server

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256/loms/internal/pkg/metrics"
	"route256/loms/internal/pkg/util/logger"
	"route256/loms/pkg/api/loms/v1"
)

func (server *LomsServer) StocksInfo(ctx context.Context, req *loms.StocksInfoRequest) (resp *loms.StocksInfoResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/stocks/info", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "StocksInfo")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	count, err := server.impl.GetStocksInfo(ctx, req.SkuId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	return &loms.StocksInfoResponse{Count: count}, nil
}
