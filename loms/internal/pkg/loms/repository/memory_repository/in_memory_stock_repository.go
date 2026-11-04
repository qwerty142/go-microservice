package memory_repository

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"route256/loms/internal/pkg/loms/model"
)

//go:embed res/stock-data.json
var stockDataFile embed.FS

type stockStorage map[int64]*model.Stock

type InMemoryStockRepository struct {
	storage stockStorage
	mutex   sync.RWMutex
}

func NewInMemoryStockRepository() *InMemoryStockRepository {
	data, err := loadStockData("res/stock-data.json")
	if err != nil {
		return &InMemoryStockRepository{
			storage: make(stockStorage),
		}
	}

	repo := &InMemoryStockRepository{storage: make(map[int64]*model.Stock)}
	for _, item := range data {
		repo.storage[item.SkuID] = &item
	}

	return repo
}

func (repo *InMemoryStockRepository) GetStockBySkuId(ctx context.Context, skuID int64) (*model.Stock, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	item, exists := repo.storage[skuID]
	if !exists {
		return nil, fmt.Errorf("item with SKU ID %d not found", skuID)
	}

	return item, nil
}

func (repo *InMemoryStockRepository) ReserveAdd(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Check amount
	for _, item := range items {
		foundItem, exist := repo.storage[item.SkuID]
		if !exist {
			return fmt.Errorf("item with SKU ID %d not found", item.SkuID)
		}

		if foundItem.TotalCount < foundItem.Reserved+uint64(item.Count) {
			return fmt.Errorf("not enough stocks")
		}
	}

	// Perform reserve
	for _, item := range items {
		foundItem, _ := repo.storage[item.SkuID]
		foundItem.Reserved += uint64(item.Count)
	}

	return nil
}

func (repo *InMemoryStockRepository) ReserveRemove(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, item := range items {
		foundItem, exist := repo.storage[item.SkuID]
		if !exist {
			continue
		}

		foundItem.TotalCount -= uint64(item.Count)
		foundItem.Reserved -= uint64(item.Count)
	}

	return nil
}

func (repo *InMemoryStockRepository) ReserveCancel(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, item := range items {
		foundItem, exist := repo.storage[item.SkuID]
		if !exist {
			continue
		}

		foundItem.Reserved -= uint64(item.Count)
	}

	return nil
}

// loadStockData loads stock data from the embedded file.
func loadStockData(filePath string) ([]model.Stock, error) {
	file, err := stockDataFile.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var items []model.Stock
	if err := json.Unmarshal(file, &items); err != nil {
		return nil, err
	}

	return items, nil
}
