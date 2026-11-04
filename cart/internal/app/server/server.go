package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"route256/cart/internal/pkg/carts/model"
)

// CartService provides methods to manage items in the user's cart.
type CartService interface {
	AddItemToCart(ctx context.Context, userID, skuID int64, count uint16) error
	RemoveItemFromCart(ctx context.Context, userID, skuID int64) error
	ClearUserCart(ctx context.Context, userID int64) error
	GetUserCart(ctx context.Context, userID int64) (*model.Cart, error)
	CheckoutCart(ctx context.Context, userID int64) (int64, error)
}

// CartServer controller for handling cart operations.
type CartServer struct {
	cartService CartService
}

// NewCartServer constructor for CartServer.
func NewCartServer(cartService CartService) *CartServer {
	return &CartServer{cartService: cartService}
}

var (
	addCartItemEndpoint    = "POST /user/{user_id}/cart/{sku_id}"
	removeCartItemEndpoint = "DELETE /user/{user_id}/cart/{sku_id}"
	clearCartEndpoint      = "DELETE /user/{user_id}/cart"
	getCartEndpoint        = "GET /user/{user_id}/cart/list"
	checkoutCartEndpoint   = "POST /user/{user_id}/cart/checkout"
	metricsEndpoint        = "GET /metrics"
)

// RegisterCartEndpoints set up all handler functions.
func (s *CartServer) RegisterCartEndpoints(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc(addCartItemEndpoint, s.AddCartItem)
	mux.HandleFunc(removeCartItemEndpoint, s.RemoveCartItem)
	mux.HandleFunc(clearCartEndpoint, s.ClearCart)
	mux.HandleFunc(getCartEndpoint, s.GetCart)
	mux.HandleFunc(checkoutCartEndpoint, s.CheckoutCart)

	mux.Handle(metricsEndpoint, promhttp.Handler())

	return mux
}

func (s *CartServer) RegisterPprof(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
