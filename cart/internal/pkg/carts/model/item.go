package model

// Item struct that is used to pass data through service.
type Item struct {
	SkuID int64
	Name  string
	Price uint32
	Count uint16
}
