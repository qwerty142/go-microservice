package repository

import (
	"slices"
	"sync"

	"route256/cart/internal/pkg/carts/entity"
	"route256/cart/internal/pkg/metrics"
)

type storage map[int64][]entity.Item

// InMemoryCartRepository repository implementation to save data in memory, protected with RWMutex.
type InMemoryCartRepository struct {
	storage storage
	mutex   sync.RWMutex
}

// NewInMemoryCartRepository constructor for InMemoryCartRepository.
func NewInMemoryCartRepository() *InMemoryCartRepository {
	metrics.InMemoryElementsTotal.Set(0)
	return &InMemoryCartRepository{
		storage: make(storage),
	}
}

// AddItem adds an item to the user's cart.
func (repo *InMemoryCartRepository) AddItem(userID, skuID int64, count uint16) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	items := repo.storage[userID]
	for i := range items {
		if items[i].SkuID == skuID {
			items[i].Count += count
			repo.updateElementsMetric()
			return nil
		}
	}

	repo.storage[userID] = append(items, entity.Item{SkuID: skuID, Count: count})
	repo.updateElementsMetric()

	return nil
}

// RemoveItem removes an item from the user's cart.
func (repo *InMemoryCartRepository) RemoveItem(userID, skuID int64) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	items, ok := repo.storage[userID]
	if !ok {
		return nil
	}

	repo.storage[userID] = slices.DeleteFunc(items, func(item entity.Item) bool {
		return item.SkuID == skuID
	})
	repo.updateElementsMetric()

	return nil
}

// ClearCart clears all items from the user's cart.
func (repo *InMemoryCartRepository) ClearCart(userID int64) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.storage, userID)
	repo.updateElementsMetric()

	return nil
}

// GetCart retrieves a copy of the user's cart, so the caller cannot mutate internal state.
func (repo *InMemoryCartRepository) GetCart(userID int64) (*entity.Cart, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	return &entity.Cart{
		UserID: userID,
		Items:  slices.Clone(repo.storage[userID]),
	}, nil
}

func (repo *InMemoryCartRepository) updateElementsMetric() {
	metrics.InMemoryElementsTotal.Set(float64(len(repo.storage)))
}
