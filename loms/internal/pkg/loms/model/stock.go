package model

type Stock struct {
	SkuID      int64  `json:"sku_id"`
	TotalCount uint64 `json:"total_count"`
	Reserved   uint64 `json:"reserved"`
}
