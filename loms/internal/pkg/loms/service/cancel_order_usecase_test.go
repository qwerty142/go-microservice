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

func TestLomsService_CancelOrder(t *testing.T) {
	type data struct {
		name    string
		orderID int64
		wantErr error
	}

	testData := []data{
		{
			name:    "success cancel",
			orderID: 123,
			wantErr: nil,
		},
		{
			name:    "failed cancel",
			orderID: 123,
			wantErr: errors.New("some error"),
		},
	}

	ctrl := minimock.NewController(t)
	orderRepoMock := mock.NewOrderRepositoryMock(ctrl)
	stockRepoMock := mock.NewStockRepositoryMock(ctrl)
	cartService := NewLomsService(orderRepoMock, stockRepoMock)

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			orderRepoMock.GetByIdMock.Expect(minimock.AnyContext, tt.orderID).Return(&model.Order{OrderID: tt.orderID}, tt.wantErr)
			orderRepoMock.UpdateStatusMock.Expect(minimock.AnyContext, tt.orderID, model.OrderStatus_Cancelled).Return(tt.wantErr)
			stockRepoMock.ReserveCancelMock.Return(nil)

			err := cartService.CancelOrder(context.Background(), tt.orderID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
