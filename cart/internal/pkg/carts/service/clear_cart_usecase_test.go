package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/pkg/carts/service/mock"
)

func TestCartService_ClearUserCartTable(t *testing.T) {
	t.Parallel()

	type data struct {
		name    string
		userID  int64
		wantErr error
	}

	testData := []data{
		{
			name:    "success clear cart",
			userID:  123,
			wantErr: nil,
		},
		{
			name:    "failed clear cart",
			userID:  123,
			wantErr: errors.New("some error"),
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

			repoMock.ClearCartMock.Expect(tt.userID).Return(tt.wantErr)

			err := cartService.ClearUserCart(context.Background(), tt.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
