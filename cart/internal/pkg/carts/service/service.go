package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"route256/cart/internal/pkg/carts/entity"
	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/util"
	"route256/cart/internal/pkg/util/logger"
	"route256/cart/pkg/api/loms/v1"
)

// Errors returned by the cart service.
var (
	// ErrCartNotFound is returned when the user cart does not exist or is empty.
	ErrCartNotFound = errors.New("cart not found")

	// ErrInsufficientStocks is returned when the available stock is less than the requested count.
	ErrInsufficientStocks = errors.New("insufficient stocks")
)

// CartRepository interface that defines the methods for managing a cart.
type CartRepository interface {
	AddItem(userID, skuID int64, count uint16) error
	RemoveItem(userID, skuID int64) error
	ClearCart(userID int64) error
	GetCart(userID int64) (*entity.Cart, error)
}

// ProductProvider defines the method to fetch product details by skuID.
type ProductProvider interface {
	GetProduct(ctx context.Context, skuID int64) (model.ProductInfo, error)
}

// LomsClient defines the loms methods required by the cart service.
type LomsClient interface {
	OrderCreate(ctx context.Context, in *loms.OrderCreateRequest, opts ...grpc.CallOption) (*loms.OrderCreateResponse, error)
	StocksInfo(ctx context.Context, in *loms.StocksInfoRequest, opts ...grpc.CallOption) (*loms.StocksInfoResponse, error)
}

const (
	TracerName     = "cart"
	defaultRPS     = 10
	refillInterval = time.Second
)

// CartService implementation of server.CartService.
type CartService struct {
	cartRepository  CartRepository
	productProvider ProductProvider
	productRPS      int
	lomsClient      LomsClient
}

// Option functional option for CartService.
type Option func(*CartService)

// WithProductRPS sets the requests-per-second limit for the product provider.
func WithProductRPS(rps int) Option {
	return func(s *CartService) {
		if rps > 0 {
			s.productRPS = rps
		}
	}
}

// NewCartService constructor for CartService.
func NewCartService(cartRepository CartRepository, productProvider ProductProvider, lomsClient LomsClient, opts ...Option) *CartService {
	s := &CartService{
		cartRepository:  cartRepository,
		productProvider: productProvider,
		productRPS:      defaultRPS,
		lomsClient:      lomsClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddItemToCart adds a specified quantity of a product to the user's cart.
func (s *CartService) AddItemToCart(ctx context.Context, userID, skuID int64, count uint16) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.AddItemToCart")
	defer span.End()

	if _, err := s.productProvider.GetProduct(ctx, skuID); err != nil {
		logger.Errorw(ctx, "Product provider error", "error", err)
		return fmt.Errorf("get product %d: %w", skuID, err)
	}

	stockResponse, err := s.lomsClient.StocksInfo(ctx, &loms.StocksInfoRequest{SkuId: skuID})
	if err != nil {
		return fmt.Errorf("failed to get stock info: %w", err)
	}
	if stockResponse.Count < uint64(count) {
		return fmt.Errorf("sku %d: %w", skuID, ErrInsufficientStocks)
	}

	if err := s.cartRepository.AddItem(userID, skuID, count); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to add item to cart for user %d: %w", userID, err)
	}

	return nil
}

// RemoveItemFromCart removes a product from the user's cart.
func (s *CartService) RemoveItemFromCart(ctx context.Context, userID, skuID int64) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.RemoveItemFromCart")
	defer span.End()

	if err := s.cartRepository.RemoveItem(userID, skuID); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to remove item from cart for user %d: %w", userID, err)
	}

	return nil
}

// ClearUserCart clears all items from the user's cart.
func (s *CartService) ClearUserCart(ctx context.Context, userID int64) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.ClearUserCart")
	defer span.End()

	if err := s.cartRepository.ClearCart(userID); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to clear cart for user %d: %w", userID, err)
	}

	return nil
}

// GetUserCart retrieves the current state of the user's cart.
func (s *CartService) GetUserCart(ctx context.Context, userID int64) (*model.Cart, error) {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.GetUserCart")
	defer span.End()

	cartEntity, err := s.cartRepository.GetCart(userID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	if cartEntity == nil {
		return nil, ErrCartNotFound
	}

	cartModel := model.Cart{
		UserID: cartEntity.UserID,
		Items:  make([]model.Item, len(cartEntity.Items)),
	}

	eg, ctx := util.ErrGroupWithContext(ctx)
	rateLimiter := util.NewRateLimiter(s.productRPS, refillInterval)

	// Each goroutine writes to its own slice element, so no additional synchronization is needed.
	for i, cartItem := range cartEntity.Items {
		eg.Go(func() error {
			if err := rateLimiter.Wait(ctx); err != nil {
				return err
			}

			product, err := s.productProvider.GetProduct(ctx, cartItem.SkuID)
			if err != nil {
				return fmt.Errorf("product provider error: %w", err)
			}

			cartModel.Items[i] = model.Item{
				SkuID: cartItem.SkuID,
				Name:  product.Name,
				Count: cartItem.Count,
				Price: product.Price,
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}

	for _, item := range cartModel.Items {
		cartModel.TotalPrice += item.Price * uint32(item.Count)
	}

	return &cartModel, nil
}

// CheckoutCart transfer cart to order.
func (s *CartService) CheckoutCart(ctx context.Context, userID int64) (int64, error) {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.CheckoutCart")
	defer span.End()

	cart, err := s.cartRepository.GetCart(userID)
	if err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return 0, fmt.Errorf("repository error: %w", err)
	}
	if cart == nil || len(cart.Items) == 0 {
		return 0, ErrCartNotFound
	}

	requestItems := make([]*loms.Item, 0, len(cart.Items))
	for _, item := range cart.Items {
		requestItems = append(requestItems, &loms.Item{SkuId: item.SkuID, Count: uint32(item.Count)})
	}

	response, err := s.lomsClient.OrderCreate(ctx, &loms.OrderCreateRequest{
		UserId: userID,
		Items:  requestItems,
	})
	if err != nil {
		return 0, fmt.Errorf("loms client error: %w", err)
	}

	return response.OrderId, nil
}
