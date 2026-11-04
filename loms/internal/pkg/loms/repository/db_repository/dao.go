package db_repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/pkg/kafka/producer"
	"route256/loms/internal/pkg/loms/model"
	"route256/loms/internal/pkg/util/logger"
)

type LomsDao struct {
	conn        *pgx.Conn
	cmd         *QueriesWrapper
	kafkaSender producer.KafkaSender
}

func NewLomsDao(conn *pgx.Conn, cmd *QueriesWrapper, producer producer.KafkaSender) *LomsDao {
	return &LomsDao{
		conn:        conn,
		cmd:         cmd,
		kafkaSender: producer,
	}
}

func (dao *LomsDao) CreateOrder(ctx context.Context, order *model.Order) (int64, error) {
	tx, err := dao.conn.Begin(ctx)
	if err != nil {
		return 0, err
	}

	qtx := dao.cmd.WithTx(tx)
	orderID, err := qtx.CreateOrder(ctx, CreateOrderParams{
		UserID: order.UserID,
		Status: order.Status,
	})
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	order.OrderID = orderID
	_ = changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_New)

	for _, item := range order.Items {
		err := qtx.AddOrderItem(ctx, AddOrderItemParams{
			OrderID: orderID,
			SkuID:   item.SkuID,
			Count:   int32(item.Count),
		})

		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}
	}
	txErr := tx.Commit(ctx)
	if txErr != nil {
		return 0, txErr
	}

	err = dao.finalizeCreateOrder(ctx, order)
	if err != nil {
		_ = changeOrderStatus(ctx, dao.cmd, dao.kafkaSender, order, model.OrderStatus_Failed)
	}
	return orderID, err
}

func (dao *LomsDao) finalizeCreateOrder(ctx context.Context, order *model.Order) error {
	txReserve, err := dao.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer txReserve.Rollback(ctx)

	reserveQtx := dao.cmd.WithTx(txReserve)
	for _, item := range order.Items {
		stockInfo, err := reserveQtx.GetStocksInfo(ctx, item.SkuID)
		if err != nil {
			return err
		}

		if (stockInfo.TotalCount - stockInfo.Reserved) < int64(item.Count) {
			_ = changeOrderStatus(ctx, reserveQtx, dao.kafkaSender, order, model.OrderStatus_Failed)
			return fmt.Errorf("not enough items")
		}

		updateErr := reserveQtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: stockInfo.TotalCount,
			Reserved:   stockInfo.Reserved + int64(item.Count),
		})
		if updateErr != nil {
			return updateErr
		}
	}
	_ = changeOrderStatus(ctx, reserveQtx, dao.kafkaSender, order, model.OrderStatus_Awaiting)
	txReserveErr := txReserve.Commit(ctx)
	return txReserveErr
}

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
	_, _, err = kafkaSender.Producer.SendMessage(message)
	if err != nil {
		logger.Errorw(ctx, "Failed to send message", "err", err)
	}
	return nil
}

func (dao *LomsDao) GetOrderInfo(ctx context.Context, orderID int64) (*model.Order, error) {
	orderRow, err := dao.cmd.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	itemRows, err := dao.cmd.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order items not found: %w", err)
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

func (dao *LomsDao) PayOrder(ctx context.Context, orderID int64) error {
	order, err := dao.GetOrderInfo(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	tx, txErr := dao.conn.Begin(ctx)
	if txErr != nil {
		return txErr
	}
	defer tx.Rollback(ctx)

	qtx := dao.cmd.WithTx(tx)
	for _, item := range order.Items {
		info, getErr := qtx.GetStocksInfo(ctx, item.SkuID)
		if getErr != nil {
			return getErr
		}

		reserveErr := qtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: info.TotalCount - int64(item.Count),
			Reserved:   info.Reserved - int64(item.Count),
		})
		if reserveErr != nil {
			return reserveErr
		}
	}

	err = changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_Payed)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (dao *LomsDao) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := dao.GetOrderInfo(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	tx, txErr := dao.conn.Begin(ctx)
	if txErr != nil {
		return txErr
	}
	defer tx.Rollback(ctx)

	qtx := dao.cmd.WithTx(tx)
	for _, item := range order.Items {
		info, getErr := qtx.GetStocksInfo(ctx, item.SkuID)
		if getErr != nil {
			return getErr
		}

		reserveErr := qtx.UpdateStocksInfo(ctx, UpdateStocksInfoParams{
			SkuID:      item.SkuID,
			TotalCount: info.TotalCount,
			Reserved:   info.Reserved - int64(item.Count),
		})
		if reserveErr != nil {
			return reserveErr
		}
	}

	err = changeOrderStatus(ctx, qtx, dao.kafkaSender, order, model.OrderStatus_Cancelled)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (dao *LomsDao) GetStocksInfo(ctx context.Context, skuID int64) (uint64, error) {
	stock, err := dao.cmd.GetStocksInfo(ctx, skuID)
	if err != nil {
		return 0, fmt.Errorf("stock not found: %w", err)
	}

	return uint64(stock.TotalCount - stock.Reserved), nil
}
