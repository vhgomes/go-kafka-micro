package domain

import (
	"github.com/google/uuid"
	"time"
)

type Item struct {
	Name     string
	Quantity int
	Price    float32
}

type Order struct {
	OrderId     uuid.UUID
	Items       []Item
	TotalAmount float32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
