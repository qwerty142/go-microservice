package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/carts/service/mock"
	"route256/cart/pkg/api/loms/v1"
)

func TestCartService_AddItemToCartTable(t *testing.T) {
	t.Parallel()

	var userID int64 = 123
	item := model.Item{
		SkuID: 100,
		Count: 2,
	}
	productNotFoundError := errors.New("not found")
	repositoryError := errors.New("repository error")

	type fields struct {
		productProviderMock *mock.ProductProviderMock
		cartRepositoryMock  *mock.CartRepositoryMock
		lomsClientMock      *mock.LomsClientMock
	}
	type data struct {
		name    string
		prepare func(f *fields)
		wantErr error
	}

	testData := []data{
		{
			name: "valid add item",
			prepare: func(f *fields) {
				f.productProviderMock.GetProductMock.Expect(minimock.AnyContext, item.SkuID).Return(model.ProductInfo{
					Name:  "Книга",
					Price: 300,
				}, nil)
				f.cartRepositoryMock.AddItemMock.Expect(userID, item.SkuID, item.Count).Return(nil)
				f.lomsClientMock.StocksInfoMock.Return(&loms.StocksInfoResponse{Count: 999}, nil)
			},
			wantErr: nil,
		},
		{
			name: "product not found",
			prepare: func(f *fields) {
				f.productProviderMock.GetProductMock.Expect(minimock.AnyContext, 100).Return(model.ProductInfo{}, productNotFoundError)
			},
			wantErr: productNotFoundError,
		},
		{
			name: "repository error",
			prepare: func(f *fields) {
				f.productProviderMock.GetProductMock.Expect(minimock.AnyContext, 100).Return(model.ProductInfo{}, nil)
				f.cartRepositoryMock.AddItemMock.Expect(userID, item.SkuID, item.Count).Return(repositoryError)
				f.lomsClientMock.StocksInfoMock.Return(&loms.StocksInfoResponse{Count: 999}, nil)
			},
			wantErr: repositoryError,
		},
	}

	for _, tt := range testData {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			fieldsForTableTest := fields{
				productProviderMock: mock.NewProductProviderMock(ctrl),
				cartRepositoryMock:  mock.NewCartRepositoryMock(ctrl),
				lomsClientMock:      mock.NewLomsClientMock(ctrl),
			}

			addHandler := NewCartService(
				fieldsForTableTest.cartRepositoryMock,
				fieldsForTableTest.productProviderMock,
				fieldsForTableTest.lomsClientMock,
			)

			tt.prepare(&fieldsForTableTest)
			err := addHandler.AddItemToCart(context.Background(), userID, item.SkuID, item.Count)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
