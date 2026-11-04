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

func (server *LomsServer) OrderCreate(ctx context.Context, req *loms.OrderCreateRequest) (resp *loms.OrderCreateResponse, err error) {
	defer func(startTime time.Time) {
		metrics.SaveRequestDuration(startTime, "/order/create", err)
	}(time.Now())

	ctx, span := otel.Tracer(TracerName).Start(ctx, "OrderCreate")
	defer span.End()
	ctx = logger.ContextWithSpanValues(ctx, span)

	id, err := server.impl.CreateOrder(ctx, repackCreateRequestToOrder(req))
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	return &loms.OrderCreateResponse{OrderId: id}, nil
}

func repackCreateRequestToOrder(in *loms.OrderCreateRequest) *model.Order {
	return &model.Order{
		UserID: in.UserId,
		Status: model.OrderStatus_New,
		Items:  mapItemsToModel(in.Items),
	}
}

func mapItemsToModel(items []*loms.Item) []*model.Item {
	result := make([]*model.Item, len(items))
	for i, item := range items {
		result[i] = &model.Item{
			SkuID: item.SkuId,
			Count: item.Count,
		}
	}

	return result
}
