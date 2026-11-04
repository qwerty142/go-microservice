package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"route256/cart/internal/pkg/carts/entity"
	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/util"
	"route256/cart/internal/pkg/util/logger"
	"route256/cart/pkg/api/loms/v1"
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

// CartService implementation of server.CartService.
type CartService struct {
	cartRepository  CartRepository
	productProvider ProductProvider
	productRPS      int
	lomsClient      loms.LomsClient
}

// NewCartService constructor for CartService.
// param[0] is used to set RPS for product provider.
func NewCartService(
	cartRepository CartRepository,
	productProvider ProductProvider,
	lomsClient loms.LomsClient,
	param ...interface{},
) *CartService {
	rps := 10
	if len(param) > 0 {
		rps = param[0].(int)
	}

	return &CartService{
		cartRepository:  cartRepository,
		productProvider: productProvider,
		productRPS:      rps,
		lomsClient:      lomsClient,
	}
}

const TracerName = "cart"

// AddItemToCart adds a specified quantity of a product to the user's cart.
func (service *CartService) AddItemToCart(ctx context.Context, userID, skuID int64, count uint16) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.AddItemToCart")
	defer span.End()

	if _, err := service.productProvider.GetProduct(ctx, skuID); err != nil {
		logger.Errorw(ctx, "Product provider error", "error", err)
		return err
	}

	stockResponse, err := service.lomsClient.StocksInfo(ctx, &loms.StocksInfoRequest{SkuId: skuID})
	if err != nil {
		return fmt.Errorf("failed to get stock info: %v", err)
	}
	if stockResponse.Count < uint64(count) {
		return fmt.Errorf("insufficient stocks")
	}

	if err := service.cartRepository.AddItem(userID, skuID, count); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to add item to cart for user %d: %w", userID, err)
	}

	return nil
}

// RemoveItemFromCart removes a product from the user's cart.
func (service *CartService) RemoveItemFromCart(ctx context.Context, userID, skuID int64) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.RemoveItemFromCart")
	defer span.End()

	if err := service.cartRepository.RemoveItem(userID, skuID); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to remove item from cart for user %d: %w", userID, err)
	}

	return nil
}

// ClearUserCart clears all items from the user's cart.
func (service *CartService) ClearUserCart(ctx context.Context, userID int64) error {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.ClearUserCart")
	defer span.End()

	if err := service.cartRepository.ClearCart(userID); err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return fmt.Errorf("failed to clear cart for user %d: %w", userID, err)
	}

	return nil
}

// GetUserCart retrieves the current state of the user's cart.
func (service *CartService) GetUserCart(ctx context.Context, userID int64) (*model.Cart, error) {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.GetUserCart")
	defer span.End()

	cartEntity, err := service.cartRepository.GetCart(userID)
	if cartEntity == nil || err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}

	cartModel := model.Cart{
		UserID:     cartEntity.UserID,
		Items:      make([]model.Item, 0),
		TotalPrice: 0,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, ctx := util.ErrGroupWithContext(ctx)
	rl := util.NewRateLimiter(service.productRPS, time.Second)

	updateWg := &sync.WaitGroup{}
	updateChan := make(chan model.Item)
	go func() {
		updateWg.Add(1)
		defer updateWg.Done()
		for item := range updateChan {
			cartModel.Items = append(cartModel.Items, item)
			cartModel.TotalPrice += item.Price * uint32(item.Count)
		}
	}()

	for _, iterItem := range cartEntity.Items {
		item := iterItem
		eg.Go(func() error {
			rl.Wait()
			product, err := service.productProvider.GetProduct(ctx, item.SkuID)
			if err != nil {
				return fmt.Errorf("product provider error %s", err)
			}

			updateChan <- model.Item{
				SkuID: item.SkuID,
				Name:  product.Name,
				Count: item.Count,
				Price: product.Price,
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		close(updateChan)
		updateWg.Wait()
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}

	close(updateChan)
	updateWg.Wait()
	return &cartModel, nil
}

// CheckoutCart transfer cart to order.
func (service *CartService) CheckoutCart(ctx context.Context, userID int64) (int64, error) {
	ctx, span := otel.Tracer(TracerName).Start(ctx, "service.CheckoutCart")
	defer span.End()

	cart, err := service.cartRepository.GetCart(userID)
	if err != nil {
		logger.Errorw(ctx, "Cart repository error", "error", err)
		return -1, fmt.Errorf("repository error: %w", err)
	}

	requestItems := make([]*loms.Item, len(cart.Items))
	for i := 0; i < len(cart.Items); i++ {
		item := cart.Items[i]
		requestItems[i] = &loms.Item{SkuId: item.SkuID, Count: uint32(item.Count)}
	}
	request := loms.OrderCreateRequest{
		UserId: userID,
		Items:  requestItems,
	}
	response, err := service.lomsClient.OrderCreate(ctx, &request)
	if err != nil {
		return -1, fmt.Errorf("loms client error: %w", err)
	}

	return response.OrderId, nil
}
