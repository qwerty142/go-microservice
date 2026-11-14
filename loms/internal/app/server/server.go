package server

import (
	"context"

	"route256/loms/internal/pkg/loms/model"
	"route256/loms/pkg/api/loms/v1"
)

const TracerName = "loms"

// LomsService provides methods to manage orders and stocks.
type LomsService interface {
	CreateOrder(ctx context.Context, order *model.Order) (int64, error)
	GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error)
	PayOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetStocksInfo(ctx context.Context, skuID int64) (uint64, error)
}

// LomsServer controller for handling loms operations.
type LomsServer struct {
	loms.UnimplementedLomsServer
	impl LomsService
}

// NewLomsServer constructor for LomsServer.
func NewLomsServer(impl LomsService) *LomsServer {
	return &LomsServer{impl: impl}
}

var _ loms.LomsServer = (*LomsServer)(nil)
