package memory_repository

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/util/logger"
)

//go:embed res/stock-data.json
var stockDataFile embed.FS

type stockStorage map[int64]*model.Stock

// InMemoryStockRepository in-memory implementation of service.StockRepository, protected with RWMutex.
type InMemoryStockRepository struct {
	storage stockStorage
	mutex   sync.RWMutex
}

// NewInMemoryStockRepository constructor for InMemoryStockRepository.
// Stock data is loaded from the embedded file; on load failure an empty repository is created.
func NewInMemoryStockRepository() *InMemoryStockRepository {
	data, err := loadStockData("res/stock-data.json")
	if err != nil {
		logger.Errorw(context.Background(), "Failed to load stock data", "error", err)
		return &InMemoryStockRepository{storage: make(stockStorage)}
	}

	repo := &InMemoryStockRepository{storage: make(stockStorage)}
	for _, item := range data {
		stock := item
		repo.storage[stock.SkuID] = &stock
	}

	return repo
}

// GetStockBySkuId retrieves the stock record by skuID.
func (repo *InMemoryStockRepository) GetStockBySkuId(ctx context.Context, skuID int64) (*model.Stock, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	stock, exists := repo.storage[skuID]
	if !exists {
		return nil, fmt.Errorf("sku %d: %w", skuID, model.ErrStockNotFound)
	}

	return stock, nil
}

// ReserveAdd reserves the items, checking first that every item exists and there is enough stock.
func (repo *InMemoryStockRepository) ReserveAdd(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Check
	for _, item := range items {
		stock, exist := repo.storage[item.SkuID]
		if !exist {
			return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrStockNotFound)
		}

		if stock.TotalCount < stock.Reserved+uint64(item.Count) {
			return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrInsufficientStocks)
		}
	}

	// Perform reserve
	for _, item := range items {
		repo.storage[item.SkuID].Reserved += uint64(item.Count)
	}

	return nil
}

// ReserveRemove removes the reserved items from the stock (order is paid).
func (repo *InMemoryStockRepository) ReserveRemove(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Check that the reserved amount is consistent to avoid unsigned underflow.
	for _, item := range items {
		stock, exist := repo.storage[item.SkuID]
		if !exist {
			continue
		}

		if stock.Reserved < uint64(item.Count) {
			return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrInsufficientStocks)
		}
	}

	for _, item := range items {
		if stock, exist := repo.storage[item.SkuID]; exist {
			stock.TotalCount -= uint64(item.Count)
			stock.Reserved -= uint64(item.Count)
		}
	}

	return nil
}

// ReserveCancel releases the reserved items back to the available stock (order is cancelled).
func (repo *InMemoryStockRepository) ReserveCancel(ctx context.Context, items []*model.Item) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Check that the reserved amount is consistent to avoid unsigned underflow.
	for _, item := range items {
		stock, exist := repo.storage[item.SkuID]
		if !exist {
			continue
		}

		if stock.Reserved < uint64(item.Count) {
			return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrInsufficientStocks)
		}
	}

	for _, item := range items {
		if stock, exist := repo.storage[item.SkuID]; exist {
			stock.Reserved -= uint64(item.Count)
		}
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
