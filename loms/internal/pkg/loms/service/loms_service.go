package service

import (
	"context"
	"fmt"

	"route256/loms/internal/pkg/loms/model"
)

// OrderRepository interface that defines the methods for managing orders.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (int64, error)
	GetById(ctx context.Context, orderID int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

// StockRepository interface that defines the methods for managing stocks.
type StockRepository interface {
	GetStockBySkuId(ctx context.Context, skuID int64) (*model.Stock, error)
	ReserveAdd(ctx context.Context, items []*model.Item) error
	ReserveRemove(ctx context.Context, items []*model.Item) error
	ReserveCancel(ctx context.Context, items []*model.Item) error
}

// LomsService implementation of server.LomsService.
type LomsService struct {
	orderRepository OrderRepository
	stockRepository StockRepository
}

// NewLomsService constructor for LomsService.
func NewLomsService(orderRepository OrderRepository, stockRepository StockRepository) *LomsService {
	return &LomsService{orderRepository: orderRepository, stockRepository: stockRepository}
}

// CreateOrder creates an order and reserves the ordered items.
func (s *LomsService) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	order.Status = model.OrderStatus_New

	orderID, err := s.orderRepository.CreateOrder(ctx, order)
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	order.OrderID = orderID

	if err := s.stockRepository.ReserveAdd(ctx, order.Items); err != nil {
		_ = s.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Failed)
		return orderID, fmt.Errorf("reserve items: %w", err)
	}

	if err := s.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Awaiting); err != nil {
		return orderID, fmt.Errorf("update order status: %w", err)
	}

	return orderID, nil
}

// GetOrderInfo retrieves the order by id.
func (s *LomsService) GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := s.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	return order, nil
}

// PayOrder pays the order and removes the reserved items from the stock.
func (s *LomsService) PayOrder(ctx context.Context, orderID int64) error {
	order, err := s.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if err := s.stockRepository.ReserveRemove(ctx, order.Items); err != nil {
		return fmt.Errorf("remove reserved items: %w", err)
	}

	if err := s.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Payed); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// CancelOrder cancels the order and releases the reserved items.
func (s *LomsService) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := s.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if err := s.stockRepository.ReserveCancel(ctx, order.Items); err != nil {
		return fmt.Errorf("cancel reserved items: %w", err)
	}

	if err := s.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Cancelled); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// GetStocksInfo returns the available (not reserved) count for the sku.
func (s *LomsService) GetStocksInfo(ctx context.Context, skuID int64) (uint64, error) {
	stock, err := s.stockRepository.GetStockBySkuId(ctx, skuID)
	if err != nil {
		return 0, fmt.Errorf("get stock: %w", err)
	}
	if stock == nil {
		return 0, model.ErrStockNotFound
	}

	return stock.TotalCount - stock.Reserved, nil
}
