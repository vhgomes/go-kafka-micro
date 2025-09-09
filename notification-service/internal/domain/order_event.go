package domain

import "github.com/google/uuid"

type OrderEvent struct {
	UserID  uuid.UUID
	OrderID string
	Total   int
	Items   []OrderItem
}

type OrderItem struct {
	ProductID string
	Quantity  int
	Price     int
}
