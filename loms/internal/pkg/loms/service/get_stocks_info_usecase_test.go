package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/loms/service/mock"
)

func TestLomsService_GetStocksInfo(t *testing.T) {
	type data struct {
		name       string
		skuID      int64
		stock      *model.Stock
		wantReturn uint64
		wantErr    error
	}

	testData := []data{
		{
			name:       "success get stock",
			skuID:      123,
			stock:      &model.Stock{SkuID: 123, TotalCount: 1000, Reserved: 1},
			wantReturn: 999,
			wantErr:    nil,
		},
		{
			name:       "failed get stock",
			skuID:      123,
			stock:      nil,
			wantReturn: 0,
			wantErr:    errors.New("some error"),
		},
	}

	ctrl := minimock.NewController(t)
	orderRepoMock := mock.NewOrderRepositoryMock(ctrl)
	stockRepoMock := mock.NewStockRepositoryMock(ctrl)
	cartService := NewLomsService(orderRepoMock, stockRepoMock)

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			stockRepoMock.GetStockBySkuIdMock.Expect(minimock.AnyContext, tt.skuID).Return(tt.stock, tt.wantErr)

			freeStock, err := cartService.GetStocksInfo(context.Background(), tt.skuID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.wantReturn, freeStock)
		})
	}
}
