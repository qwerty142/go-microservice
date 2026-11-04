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

func TestLomsService_GetOrderInfo(t *testing.T) {
	type data struct {
		name    string
		orderID int64
		order   *model.Order
		wantErr error
	}

	testData := []data{
		{
			name:    "success get order",
			orderID: 123,
			order:   &model.Order{OrderID: 123},
			wantErr: nil,
		},
		{
			name:    "failed get order",
			orderID: 123,
			order:   nil,
			wantErr: errors.New("some error"),
		},
	}

	ctrl := minimock.NewController(t)
	orderRepoMock := mock.NewOrderRepositoryMock(ctrl)
	stockRepoMock := mock.NewStockRepositoryMock(ctrl)
	cartService := NewLomsService(orderRepoMock, stockRepoMock)

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {

			orderRepoMock.GetByIdMock.Expect(minimock.AnyContext, tt.orderID).Return(tt.order, tt.wantErr)

			order, err := cartService.GetOrderInfo(context.Background(), tt.orderID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.order, order)
		})
	}
}
