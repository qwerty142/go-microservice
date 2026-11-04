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

func TestLomsService_PayOrder(t *testing.T) {
	type data struct {
		name    string
		orderID int64
		wantErr error
	}

	testData := []data{
		{
			name:    "success payment",
			orderID: 123,
			wantErr: nil,
		},
		{
			name:    "failed payment",
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
			orderRepoMock.UpdateStatusMock.Expect(minimock.AnyContext, tt.orderID, model.OrderStatus_Payed).Return(tt.wantErr)
			stockRepoMock.ReserveRemoveMock.Return(nil)

			err := cartService.PayOrder(context.Background(), tt.orderID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
