package memory_repository

import (
	"context"
	"fmt"
	"sync"

	"route256/loms/internal/pkg/loms/model"
)

type orderStorage map[int64]*model.Order

// InMemoryOrderRepository in-memory implementation of service.OrderRepository, protected with RWMutex.
type InMemoryOrderRepository struct {
	maxIndex int64
	storage  orderStorage
	mutex    sync.RWMutex
}

// NewInMemoryOrderRepository constructor for InMemoryOrderRepository.
func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		maxIndex: 0,
		storage:  make(orderStorage),
	}
}

// CreateOrder stores the order and assigns a new order id.
func (repo *InMemoryOrderRepository) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.maxIndex++
	order.OrderID = repo.maxIndex
	repo.storage[repo.maxIndex] = order

	return repo.maxIndex, nil
}

// GetById retrieves the order by id.
func (repo *InMemoryOrderRepository) GetById(ctx context.Context, orderID int64) (*model.Order, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	order, exist := repo.storage[orderID]
	if !exist {
		return nil, fmt.Errorf("order %d: %w", orderID, model.ErrOrderNotFound)
	}

	return order, nil
}

// UpdateStatus updates the order status.
func (repo *InMemoryOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	order, exist := repo.storage[orderID]
	if !exist {
		return fmt.Errorf("order %d: %w", orderID, model.ErrOrderNotFound)
	}

	order.Status = status
	return nil
}
