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

func TestLomsService_CreateOrder(t *testing.T) {
	sampleOrder := model.Order{
		UserID: 123,
		Status: model.OrderStatus_New,
		Items:  []*model.Item{{SkuID: 1, Count: 1}},
	}
	reserveError := errors.New("not enough items")

	type fields struct {
		orderRepoMock *mock.OrderRepositoryMock
		stockRepoMock *mock.StockRepositoryMock
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
				ctx := context.Background()
				f.orderRepoMock.CreateOrderMock.Expect(ctx, &sampleOrder).Return(1, nil)
				f.stockRepoMock.ReserveAddMock.Expect(ctx, sampleOrder.Items).Return(nil)
				f.orderRepoMock.UpdateStatusMock.Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "product not found",
			prepare: func(f *fields) {
				ctx := context.Background()
				f.orderRepoMock.CreateOrderMock.Expect(ctx, &sampleOrder).Return(1, nil)
				f.stockRepoMock.ReserveAddMock.Expect(ctx, sampleOrder.Items).Return(reserveError)
				f.orderRepoMock.UpdateStatusMock.Return(nil)
			},
			wantErr: reserveError,
		},
	}

	ctrl := minimock.NewController(t)
	fieldsForTableTest := fields{
		orderRepoMock: mock.NewOrderRepositoryMock(ctrl),
		stockRepoMock: mock.NewStockRepositoryMock(ctrl),
	}

	addHandler := NewLomsService(fieldsForTableTest.orderRepoMock, fieldsForTableTest.stockRepoMock)

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(&fieldsForTableTest)
			_, err := addHandler.CreateOrder(context.Background(), &sampleOrder)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
