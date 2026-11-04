package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"route256/cart/internal/pkg/carts/entity"
)

func TestInMemoryCartRepository_AddItemBlind(t *testing.T) {
	t.Parallel()

	repository := NewInMemoryCartRepository()

	err := repository.AddItem(1, 1, 1)

	require.NoError(t, err)
}

func TestInMemoryCartRepository_AddItemWithGet(t *testing.T) {
	t.Parallel()

	repository := NewInMemoryCartRepository()
	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 1
	)

	errAdd := repository.AddItem(userID, skuID, count)
	cart, errGet := repository.GetCart(userID)

	require.NoError(t, errAdd)
	require.NoError(t, errGet)
	require.Equal(t, 1, len(cart.Items))
	require.Equal(t, entity.Item{SkuID: skuID, Count: count}, cart.Items[0])
}

func TestInMemoryCartRepository_AddExistingItem(t *testing.T) {
	t.Parallel()

	repository := NewInMemoryCartRepository()
	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 4
	)

	errAdd1 := repository.AddItem(userID, skuID, count)
	errAdd2 := repository.AddItem(userID, skuID, count)
	cart, errGet := repository.GetCart(userID)

	require.NoError(t, errAdd1)
	require.NoError(t, errAdd2)
	require.NoError(t, errGet)
	require.Equal(t, 1, len(cart.Items))
	require.Equal(t, entity.Item{SkuID: skuID, Count: count + count}, cart.Items[0])
}

func TestInMemoryCartRepository_AddItemWithRemove(t *testing.T) {
	t.Parallel()

	repository := NewInMemoryCartRepository()
	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 1
	)

	errAdd := repository.AddItem(userID, skuID, count)
	errRemove := repository.RemoveItem(userID, skuID)
	cart, errGet := repository.GetCart(userID)

	require.NoError(t, errAdd)
	require.NoError(t, errRemove)
	require.NoError(t, errGet)
	require.Equal(t, 0, len(cart.Items))
}

func TestInMemoryCartRepository_ClearCart(t *testing.T) {
	t.Parallel()

	repository := NewInMemoryCartRepository()
	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 1
	)

	errAdd1 := repository.AddItem(userID, skuID, count)
	errAdd2 := repository.AddItem(userID, skuID+1, count)
	errAdd3 := repository.AddItem(userID, skuID+2, count)
	errClear := repository.ClearCart(userID)
	cart, errGet := repository.GetCart(userID)

	require.NoError(t, errAdd1)
	require.NoError(t, errAdd2)
	require.NoError(t, errAdd3)
	require.NoError(t, errClear)
	require.NoError(t, errGet)
	require.Equal(t, 0, len(cart.Items))
}
