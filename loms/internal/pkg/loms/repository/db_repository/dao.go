package db_repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/pkg/kafka/producer"
	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/util/logger"
)

// LomsDao implementation of server.LomsService on top of PostgreSQL and Kafka.
type LomsDao struct {
	conn        *pgx.Conn
	cmd         *QueriesWrapper
	kafkaSender producer.KafkaSender
}

// NewLomsDao constructor for LomsDao.
func NewLomsDao(conn *pgx.Conn, cmd *QueriesWrapper, producer producer.KafkaSender) *LomsDao {
	return &LomsDao{
		conn:        conn,
		cmd:         cmd,
		kafkaSender: producer,
	}
}

// CreateOrder creates the order with its items, then reserves the ordered items.
// If the reservation fails, the order is marked as failed.
func (dao *LomsDao) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	tx, err := dao.conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin create order tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := dao.cmd.WithTx(tx)
	orderID, err := qtx.CreateOrder(ctx, CreateOrderParams{
		UserID: order.UserID,
		Status: order.Status,
	})
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}

	order.OrderID = orderID
	if err := changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_New); err != nil {
		return 0, fmt.Errorf("change order status: %w", err)
	}

	for _, item := range order.Items {
		err := qtx.AddOrderItem(ctx, AddOrderItemParams{
			OrderID: orderID,
			SkuID:   item.SkuID,
			Count:   int32(item.Count),
		})
		if err != nil {
			return 0, fmt.Errorf("add order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit create order tx: %w", err)
	}

	if err := dao.finalizeCreateOrder(ctx, order); err != nil {
		_ = changeOrderStatus(ctx, dao.cmd, dao.kafkaSender, order, model.OrderStatus_Failed)
		return orderID, err
	}

	if err := changeOrderStatus(ctx, dao.cmd, dao.kafkaSender, order, model.OrderStatus_Awaiting); err != nil {
		return orderID, fmt.Errorf("change order status: %w", err)
	}

	return orderID, nil
}

// finalizeCreateOrder reserves the ordered items in a separate transaction.
func (dao *LomsDao) finalizeCreateOrder(ctx context.Context, order *model.Order) error {
	txReserve, err := dao.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reserve tx: %w", err)
	}
	defer txReserve.Rollback(ctx)

	reserveQtx := dao.cmd.WithTx(txReserve)
	for _, item := range order.Items {
		stockInfo, err := reserveQtx.GetStocksInfo(ctx, item.SkuID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrStockNotFound)
			}
			return fmt.Errorf("get stocks info: %w", err)
		}

		if (stockInfo.TotalCount - stockInfo.Reserved) < int64(item.Count) {
			return fmt.Errorf("sku %d: %w", item.SkuID, model.ErrInsufficientStocks)
		}

		updateErr := reserveQtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: stockInfo.TotalCount,
			Reserved:   stockInfo.Reserved + int64(item.Count),
		})
		if updateErr != nil {
			return fmt.Errorf("update stocks info: %w", updateErr)
		}
	}

	if err := txReserve.Commit(ctx); err != nil {
		return fmt.Errorf("commit reserve tx: %w", err)
	}

	return nil
}

// changeOrderStatus persists the new order status and emits the corresponding Kafka event.
// A Kafka delivery failure is only logged, so events are delivered at-most-once per status change.
func changeOrderStatus(ctx context.Context, qtx *QueriesWrapper, kafkaSender producer.KafkaSender, order *model.Order, status string) error {
	err := qtx.UpdateOrderStatus(ctx, UpdateOrderStatusParams{
		OrderID: order.OrderID,
		Status:  status,
	})
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic:     kafkaSender.TopicName,
		Key:       sarama.StringEncoder(strconv.FormatInt(order.OrderID, 10)),
		Value:     sarama.ByteEncoder(status),
		Timestamp: time.Now(),
	}
	if _, _, err = kafkaSender.Producer.SendMessage(message); err != nil {
		logger.Errorw(ctx, "Failed to send message", "err", err)
	}

	return nil
}

// GetOrderInfo retrieves the order with its items.
func (dao *LomsDao) GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error) {
	orderRow, err := dao.cmd.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order %d: %w", orderID, model.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	itemRows, err := dao.cmd.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}

	return &model.Order{
		OrderID: orderID,
		UserID:  orderRow.UserID,
		Status:  orderRow.Status,
		Items:   repackItemRowsIntoModel(itemRows),
	}, nil
}

func repackItemRowsIntoModel(rows []GetOrderItemsRow) []*model.Item {
	items := make([]*model.Item, len(rows))
	for i, row := range rows {
		items[i] = &model.Item{
			SkuID: row.SkuID,
			Count: uint32(row.Count),
		}
	}

	return items
}

// PayOrder pays the order and removes the reserved items from the stock.
func (dao *LomsDao) PayOrder(ctx context.Context, orderID int64) error {
	order, err := dao.GetOrderInfo(ctx, orderID)
	if err != nil {
		return err
	}

	tx, err := dao.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin pay order tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := dao.cmd.WithTx(tx)
	for _, item := range order.Items {
		info, getErr := qtx.GetStocksInfo(ctx, item.SkuID)
		if getErr != nil {
			return fmt.Errorf("get stocks info: %w", getErr)
		}

		reserveErr := qtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: info.TotalCount - int64(item.Count),
			Reserved:   info.Reserved - int64(item.Count),
		})
		if reserveErr != nil {
			return fmt.Errorf("update stocks info: %w", reserveErr)
		}
	}

	if err := changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_Payed); err != nil {
		return fmt.Errorf("change order status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit pay order tx: %w", err)
	}

	return nil
}

// CancelOrder cancels the order and releases the reserved items.
func (dao *LomsDao) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := dao.GetOrderInfo(ctx, orderID)
	if err != nil {
		return err
	}

	tx, err := dao.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin cancel order tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := dao.cmd.WithTx(tx)
	for _, item := range order.Items {
		info, getErr := qtx.GetStocksInfo(ctx, item.SkuID)
		if getErr != nil {
			return fmt.Errorf("get stocks info: %w", getErr)
		}

		reserveErr := qtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: info.TotalCount,
			Reserved:   info.Reserved - int64(item.Count),
		})
		if reserveErr != nil {
			return fmt.Errorf("update stocks info: %w", reserveErr)
		}
	}

	if err := changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_Cancelled); err != nil {
		return fmt.Errorf("change order status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cancel order tx: %w", err)
	}

	return nil
}

// GetStocksInfo returns the available (not reserved) count for the sku.
func (dao *LomsDao) GetStocksInfo(ctx context.Context, skuID int64) (uint64, error) {
	stock, err := dao.cmd.GetStocksInfo(ctx, skuID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("sku %d: %w", skuID, model.ErrStockNotFound)
		}
		return 0, fmt.Errorf("get stocks info: %w", err)
	}

	return uint64(stock.TotalCount - stock.Reserved), nil
}
