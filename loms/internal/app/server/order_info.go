package server

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/metrics"
	"route256/loms/internal/pkg/util/logger"
	"route256/loms/pkg/api/loms/v1"
)

func (server *LomsServer) OrderInfo(ctx context.Context, req *loms.OrderInfoRequest) (resp *loms.OrderInfoResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/order/info", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "OrderInfo")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	order, err := server.impl.GetOrderInfo(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	return repackOrderToInfoResponse(order), nil
}

func repackOrderToInfoResponse(order *model.Order) *loms.OrderInfoResponse {
	return &loms.OrderInfoResponse{
		Status: order.Status,
		UserId: order.UserID,
		Items:  mapItemsToProto(order.Items),
	}
}

func mapItemsToProto(items []*model.Item) []*loms.Item {
	result := make([]*loms.Item, len(items))
	for i, item := range items {
		result[i] = &loms.Item{
			SkuId: item.SkuID,
			Count: item.Count,
		}
	}

	return result
}
