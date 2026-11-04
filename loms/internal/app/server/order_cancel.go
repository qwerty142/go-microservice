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

func (server *LomsServer) OrderCancel(ctx context.Context, req *loms.OrderCancelRequest) (resp *loms.OrderCancelResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/order/cancel", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "OrderCancel")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	err = server.impl.CancelOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &loms.OrderCancelResponse{}, nil
}
