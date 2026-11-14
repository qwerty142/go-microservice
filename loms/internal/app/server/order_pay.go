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

// OrderPay handles order payment.
func (server *LomsServer) OrderPay(ctx context.Context, req *loms.OrderPayRequest) (resp *loms.OrderPayResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/order/pay", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "OrderPay")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	err = server.impl.PayOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &loms.OrderPayResponse{}, nil
}
