package service

import (
	"context"

	"route256/loms/internal/pkg/loms/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (int64, error)
	GetById(ctx context.Context, orderID int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

type StockRepository interface {
	GetStockBySkuId(ctx context.Context, skuID int64) (*model.Stock, error)
	ReserveAdd(ctx context.Context, items []*model.Item) error
	ReserveRemove(ctx context.Context, items []*model.Item) error
	ReserveCancel(ctx context.Context, items []*model.Item) error
}

type LomsService struct {
	orderRepository OrderRepository
	stockRepository StockRepository
}

func NewLomsService(orderRepository OrderRepository, stockRepository StockRepository) *LomsService {
	return &LomsService{orderRepository: orderRepository, stockRepository: stockRepository}
}

func (service *LomsService) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	order.Status = model.OrderStatus_New
	orderID, err := service.orderRepository.CreateOrder(ctx, order)

	reserveErr := service.stockRepository.ReserveAdd(ctx, order.Items)
	if reserveErr != nil {
		_ = service.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Failed)
		return order.OrderID, reserveErr

	}

	err = service.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Awaiting)
	return order.OrderID, err

}

func (service *LomsService) GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := service.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (service *LomsService) PayOrder(ctx context.Context, orderID int64) error {
	order, err := service.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return err
	}

	reserveErr := service.stockRepository.ReserveRemove(ctx, order.Items)
	if reserveErr != nil {
		return reserveErr
	}

	updateErr := service.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Payed)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (service *LomsService) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := service.orderRepository.GetById(ctx, orderID)
	if err != nil {
		return err
	}

	reserveErr := service.stockRepository.ReserveCancel(ctx, order.Items)
	if reserveErr != nil {
		return reserveErr
	}

	updateErr := service.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatus_Cancelled)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (service *LomsService) GetStocksInfo(ctx context.Context, skuID int64) (uint64, error) {
	stock, err := service.stockRepository.GetStockBySkuId(ctx, skuID)
	if stock == nil || err != nil {
		return 0, err
	}

	return stock.TotalCount - stock.Reserved, err
}
