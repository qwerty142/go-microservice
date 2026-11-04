package loms

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"route256/cart/internal/pkg/metrics"
	"route256/cart/pkg/api/loms/v1"
)

type ClientWrapper struct {
	client loms.LomsClient
}

func NewClientWrapper(client loms.LomsClient) *ClientWrapper {
	return &ClientWrapper{client: client}
}

func (cw *ClientWrapper) OrderCreate(ctx context.Context, in *loms.OrderCreateRequest, opts ...grpc.CallOption) (resp *loms.OrderCreateResponse, err error) {
	defer func(startTime time.Time) {
		s, _ := status.FromError(err)
		metrics.SaveExternalCallDuration(startTime, "/order/create", int(s.Code()))
	}(time.Now())

	return cw.client.OrderCreate(ctx, in, opts...)
}

func (cw *ClientWrapper) OrderInfo(ctx context.Context, in *loms.OrderInfoRequest, opts ...grpc.CallOption) (resp *loms.OrderInfoResponse, err error) {
	defer func(startTime time.Time) {
		s, _ := status.FromError(err)
		metrics.SaveExternalCallDuration(startTime, "/order/info", int(s.Code()))
	}(time.Now())

	return cw.client.OrderInfo(ctx, in, opts...)
}

func (cw *ClientWrapper) OrderPay(ctx context.Context, in *loms.OrderPayRequest, opts ...grpc.CallOption) (resp *loms.OrderPayResponse, err error) {
	defer func(startTime time.Time) {
		s, _ := status.FromError(err)
		metrics.SaveExternalCallDuration(startTime, "/order/pay", int(s.Code()))
	}(time.Now())

	return cw.client.OrderPay(ctx, in, opts...)
}

func (cw *ClientWrapper) OrderCancel(ctx context.Context, in *loms.OrderCancelRequest, opts ...grpc.CallOption) (resp *loms.OrderCancelResponse, err error) {
	defer func(startTime time.Time) {
		s, _ := status.FromError(err)
		metrics.SaveExternalCallDuration(startTime, "/order/cancel", int(s.Code()))
	}(time.Now())

	return cw.client.OrderCancel(ctx, in, opts...)
}

func (cw *ClientWrapper) StocksInfo(ctx context.Context, in *loms.StocksInfoRequest, opts ...grpc.CallOption) (resp *loms.StocksInfoResponse, err error) {
	defer func(startTime time.Time) {
		s, _ := status.FromError(err)
		metrics.SaveExternalCallDuration(startTime, "/stocks/info", int(s.Code()))
	}(time.Now())

	return cw.client.StocksInfo(ctx, in, opts...)
}
