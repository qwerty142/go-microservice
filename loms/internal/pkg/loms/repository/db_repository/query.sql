-- name: CreateOrder :one
INSERT INTO orders (user_id, status)
VALUES ($1, $2)
RETURNING order_id;

-- name: AddOrderItem :exec
INSERT INTO order_items (order_id, sku_id, count)
VALUES ($1, $2, $3);

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2
WHERE order_id = $1;

-- name: GetOrder :one
SELECT user_id, status
FROM orders
WHERE order_id = $1;

-- name: GetOrderItems :many
SELECT sku_id, count
FROM order_items
WHERE order_id = $1;

-- name: GetStocksInfo :one
SELECT *
FROM stocks
WHERE sku_id = $1;

-- name: UpdateStocksInfo :exec
UPDATE stocks
SET total_count = $2,
    reserved    = $3
WHERE sku_id = $1;
