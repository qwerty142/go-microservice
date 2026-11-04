package model

// Cart struct that is used to pass data through service.
type Cart struct {
	UserID     int64
	Items      []Item
	TotalPrice uint32
}
