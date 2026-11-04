package repository

import (
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

	if _, exists := repo.storage[userID]; !exists {
		repo.storage[userID] = make([]entity.Item, 0)
	}

	found := false
	for i := 0; i < len(repo.storage[userID]) && !found; i++ {
		if repo.storage[userID][i].SkuID == skuID {
			repo.storage[userID][i].Count += count
			found = true
		}
	}

	if !found {
		repo.storage[userID] = append(repo.storage[userID], entity.Item{SkuID: skuID, Count: count})
	}

	metrics.InMemoryElementsTotal.Set(float64(len(repo.storage)))

	return nil
}

// RemoveItem removes an item from the user's cart.
func (repo *InMemoryCartRepository) RemoveItem(userID, sku int64) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if items, ok := repo.storage[userID]; ok {
		for i, item := range items {
			if item.SkuID == sku {
				repo.storage[userID] = append(items[:i], items[i+1:]...)
				break
			}
		}
	}

	metrics.InMemoryElementsTotal.Set(float64(len(repo.storage)))

	return nil
}

// ClearCart clears all items from the user's cart.
func (repo *InMemoryCartRepository) ClearCart(userID int64) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.storage, userID)

	metrics.InMemoryElementsTotal.Set(float64(len(repo.storage)))

	return nil
}

// GetCart retrieves the user's cart.
func (repo *InMemoryCartRepository) GetCart(userID int64) (*entity.Cart, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	cart := entity.Cart{}
	cart.UserID = userID
	cart.Items = repo.storage[userID]

	return &cart, nil
}
