package repository

import (
	"testing"
)

func BenchmarkInMemoryCartRepository_AddSameItem(b *testing.B) {
	inMemoryCartRepository := NewInMemoryCartRepository()

	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 1
	)

	for i := 1; i <= b.N; i++ {
		_ = inMemoryCartRepository.AddItem(userID, skuID, count)
	}
}

func BenchmarkInMemoryCartRepository_AddDifferentItem(b *testing.B) {
	inMemoryCartRepository := NewInMemoryCartRepository()

	var (
		userID int64  = 1
		count  uint16 = 1
	)

	for i := 1; i <= b.N; i++ {
		_ = inMemoryCartRepository.AddItem(userID, int64(i), count)
	}
}

func BenchmarkInMemoryCartRepository_AddAndRemoveItem(b *testing.B) {
	inMemoryCartRepository := NewInMemoryCartRepository()

	var (
		userID int64  = 1
		skuID  int64  = 1
		count  uint16 = 1
	)

	for i := 1; i <= b.N; i++ {
		_ = inMemoryCartRepository.AddItem(userID, skuID, count)
		_ = inMemoryCartRepository.RemoveItem(userID, skuID)
	}
}
