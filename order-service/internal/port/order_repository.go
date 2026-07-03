package port

import (
	"context"
	"order-service/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
}
