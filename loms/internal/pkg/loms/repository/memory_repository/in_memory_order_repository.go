package memory_repository

import (
	"context"
	"fmt"
	"sync"

	"route256/loms/internal/pkg/loms/model"
)

type orderStorage map[int64]*model.Order

type InMemoryOrderRepository struct {
	maxIndex int64
	storage  orderStorage
	mutex    sync.RWMutex
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		maxIndex: 0,
		storage:  make(orderStorage),
	}
}

func (repo *InMemoryOrderRepository) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.maxIndex++
	order.OrderID = repo.maxIndex
	repo.storage[repo.maxIndex] = order

	return repo.maxIndex, nil
}

func (repo *InMemoryOrderRepository) GetById(ctx context.Context, orderID int64) (*model.Order, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	order, exist := repo.storage[orderID]
	if !exist {
		return nil, fmt.Errorf("order not exist")
	}

	return order, nil
}

func (repo *InMemoryOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	order, exist := repo.storage[orderID]
	if !exist {
		return fmt.Errorf("order not exist")
	}

	order.Status = status
	return nil
}
