package domain

import (
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	UserID    uuid.UUID
	OrderID   string
	Message   string
	CreatedAt time.Time
}
