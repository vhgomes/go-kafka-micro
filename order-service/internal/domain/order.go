package domain

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Name     string
	Quantity int
	Price    int64 // Lembrando que esse valor é em centavos
}

type Order struct {
	OrderId     uuid.UUID
	Items       []Item
	TotalAmount int64 // Lembrando que esse valor é em centavos
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
