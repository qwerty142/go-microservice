package model

type Order struct {
	OrderID int64
	UserID  int64
	Status  string
	Items   []*Item
}
