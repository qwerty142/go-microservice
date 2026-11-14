package server

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/metrics"
	"route256/loms/internal/pkg/util/logger"
	"route256/loms/pkg/api/loms/v1"
)

// StocksInfo handles retrieving the available stock count.
func (server *LomsServer) StocksInfo(ctx context.Context, req *loms.StocksInfoRequest) (resp *loms.StocksInfoResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/stocks/info", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "StocksInfo")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	count, err := server.impl.GetStocksInfo(ctx, req.SkuId)
	if err != nil {
		if errors.Is(err, model.ErrStockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &loms.StocksInfoResponse{Count: count}, nil
}
