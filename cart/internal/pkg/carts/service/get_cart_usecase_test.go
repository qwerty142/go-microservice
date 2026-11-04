package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/pkg/carts/entity"
	"route256/cart/internal/pkg/carts/model"
	"route256/cart/internal/pkg/carts/service/mock"
)

func TestCartService_GetUserCart_CountTotalPrice(t *testing.T) {
	t.Parallel()

	userID := int64(123)

	ctrl := minimock.NewController(t)
	productMock := mock.NewProductProviderMock(ctrl)
	repoMock := mock.NewCartRepositoryMock(ctrl)
	lomsMock := mock.NewLomsClientMock(ctrl)
	cartService := NewCartService(repoMock, productMock, lomsMock)

	productMock.GetProductMock.When(minimock.AnyContext, 1).Then(model.ProductInfo{Name: "A", Price: 1}, nil)
	productMock.GetProductMock.When(minimock.AnyContext, 10).Then(model.ProductInfo{Name: "B", Price: 10}, nil)
	productMock.GetProductMock.When(minimock.AnyContext, 100).Then(model.ProductInfo{Name: "C", Price: 100}, nil)

	items := []entity.Item{
		{SkuID: int64(100), Count: 7},
		{SkuID: int64(10), Count: 8},
		{SkuID: int64(1), Count: 9},
	}
	repoMock.GetCartMock.Expect(userID).Return(&entity.Cart{UserID: userID, Items: items}, nil)

	cart, err := cartService.GetUserCart(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, cart)
	require.Equal(t, userID, cart.UserID)
	require.Equal(t, 3, len(cart.Items))
	require.Equal(t, uint32(789), cart.TotalPrice)
}

func TestCartService_GetUserCartTable(t *testing.T) {
	t.Parallel()

	type data struct {
		name    string
		userID  int64
		cart    *model.Cart
		wantErr error
	}

	testData := []data{
		{
			name:    "success get cart",
			userID:  123,
			cart:    &model.Cart{UserID: 123, Items: make([]model.Item, 0), TotalPrice: 0},
			wantErr: nil,
		},
		{
			name:    "not found cart",
			userID:  123,
			cart:    nil,
			wantErr: errors.New("cart not found"),
		},
	}

	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			productMock := mock.NewProductProviderMock(ctrl)
			repoMock := mock.NewCartRepositoryMock(ctrl)
			lomsMock := mock.NewLomsClientMock(ctrl)
			cartService := NewCartService(repoMock, productMock, lomsMock)

			repoMock.GetCartMock.Expect(tt.userID).Return(&entity.Cart{
				UserID: tt.userID,
				Items:  make([]entity.Item, 0),
			}, tt.wantErr)

			cart, err := cartService.GetUserCart(context.Background(), tt.userID)
			require.Equal(t, tt.cart, cart)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
