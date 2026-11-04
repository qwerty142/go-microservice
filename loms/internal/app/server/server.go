package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"route256/loms/internal/pkg/loms/model"
	"route256/loms/pkg/api/loms/v1"
)

const TracerName = "loms"

type LomsService interface {
	CreateOrder(ctx context.Context, order *model.Order) (int64, error)
	GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error)
	PayOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetStocksInfo(ctx context.Context, skuID int64) (uint64, error)
}

type LomsServer struct {
	loms.UnimplementedLomsServer
	impl LomsService
}

func NewLomsServer(impl LomsService) *LomsServer {
	return &LomsServer{impl: impl}
}

var _ loms.LomsServer = (*LomsServer)(nil)

func (s *LomsServer) RegisterPprof(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
