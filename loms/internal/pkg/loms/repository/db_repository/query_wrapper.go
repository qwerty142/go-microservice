package db_repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"route256/loms/internal/pkg/metrics"
)

var (
	selectOperation = "SELECT"
	insertOperation = "INSERT"
	updateOperation = "UPDATE"
	deleteOperation = "DELETE"
)

type QueriesWrapper struct {
	queries *Queries
}

func NewQueryWrapper(queries *Queries) *QueriesWrapper {
	return &QueriesWrapper{queries: queries}
}

func (qw *QueriesWrapper) WithTx(tx pgx.Tx) *QueriesWrapper {
	return NewQueryWrapper(qw.queries.WithTx(tx))
}

func (qw *QueriesWrapper) AddOrderItem(ctx context.Context, arg AddOrderItemParams) (err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, insertOperation)
	}(time.Now())

	return qw.queries.AddOrderItem(ctx, arg)
}

func (qw *QueriesWrapper) CreateOrder(ctx context.Context, arg CreateOrderParams) (result int64, err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, insertOperation)
	}(time.Now())

	return qw.queries.CreateOrder(ctx, arg)
}

func (qw *QueriesWrapper) GetOrder(ctx context.Context, orderID int64) (result GetOrderRow, err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, selectOperation)
	}(time.Now())

	return qw.queries.GetOrder(ctx, orderID)
}

func (qw *QueriesWrapper) GetOrderItems(ctx context.Context, orderID int64) (result []GetOrderItemsRow, err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, selectOperation)
	}(time.Now())

	return qw.queries.GetOrderItems(ctx, orderID)
}

func (qw *QueriesWrapper) GetStocksInfo(ctx context.Context, skuID int64) (result Stock, err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, selectOperation)
	}(time.Now())

	return qw.queries.GetStocksInfo(ctx, skuID)
}

func (qw *QueriesWrapper) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, updateOperation)
	}(time.Now())

	return qw.queries.UpdateOrderStatus(ctx, arg)
}

func (qw *QueriesWrapper) UpdateStocksInfo(ctx context.Context, arg UpdateStocksInfoParams) (err error) {
	defer func(startTime time.Time) {
		metrics.SaveDatabaseRequestDuration(startTime, err, updateOperation)
	}(time.Now())

	return qw.queries.UpdateStocksInfo(ctx, arg)
}
