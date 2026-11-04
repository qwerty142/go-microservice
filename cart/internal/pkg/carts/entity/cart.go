package entity

// Cart entity that is used in repositories.
type Cart struct {
	UserID int64
	Items  []Item
}
