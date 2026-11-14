package model

import "errors"

// Errors shared across the loms module layers.
var (
	// ErrOrderNotFound is returned when the requested order does not exist.
	ErrOrderNotFound = errors.New("order not found")

	// ErrStockNotFound is returned when the requested stock record does not exist.
	ErrStockNotFound = errors.New("stock not found")

	// ErrInsufficientStocks is returned when the available stock is less than requested.
	ErrInsufficientStocks = errors.New("insufficient stocks")
)
