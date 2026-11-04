package repository

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"route256/cart/internal/pkg/carts/entity"
)

func TestInMemoryCartRepository_ConcurrentAddItem(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryCartRepository()
	userID := int64(1)
	skuID := int64(101)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.AddItem(userID, skuID, 1)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	cart, err := repo.GetCart(userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cart.Items))
	assert.Equal(t, skuID, cart.Items[0].SkuID)
	assert.Equal(t, uint16(100), cart.Items[0].Count)
}

func TestInMemoryCartRepository_ConcurrentRemoveItem(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryCartRepository()
	userID := int64(1)
	skuID := int64(101)
	var wg sync.WaitGroup

	err := repo.AddItem(userID, skuID, 100)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.RemoveItem(userID, skuID)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	cart, err := repo.GetCart(userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(cart.Items))
}

func TestInMemoryCartRepository_ConcurrentClearCart(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryCartRepository()
	userID := int64(1)
	skuID1 := int64(101)
	skuID2 := int64(102)
	var wg sync.WaitGroup

	err := repo.AddItem(userID, skuID1, 50)
	assert.NoError(t, err)
	err = repo.AddItem(userID, skuID2, 50)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.ClearCart(userID)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	cart, err := repo.GetCart(userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(cart.Items))
}

func TestInMemoryCartRepository_ConcurrentGetCart(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryCartRepository()
	userID := int64(1)
	skuID1 := int64(101)
	skuID2 := int64(102)
	var wg sync.WaitGroup

	err := repo.AddItem(userID, skuID1, 50)
	assert.NoError(t, err)
	err = repo.AddItem(userID, skuID2, 50)
	assert.NoError(t, err)

	var results []entity.Cart
	var mutex sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cart, err := repo.GetCart(userID)
			assert.NoError(t, err)

			mutex.Lock()
			defer mutex.Unlock()
			results = append(results, *cart)
		}()
	}

	wg.Wait()

	firstCart := results[0]
	for _, cart := range results {
		assert.Equal(t, firstCart, cart)
	}
}
