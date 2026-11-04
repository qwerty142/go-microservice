-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders
(
  order_id BIGSERIAL PRIMARY KEY,
  user_id  BIGINT NOT NULL,
  status   VARCHAR(24) NOT NULL
);

CREATE TABLE order_items
(
  order_id BIGINT REFERENCES orders (order_id) NOT NULL,
  sku_id   BIGINT NOT NULL,
  count    INT NOT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE order_items;
DROP TABLE orders;
-- +goose StatementEnd
