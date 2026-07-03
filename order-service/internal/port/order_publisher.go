package port

import (
	"context"
	"order-service/internal/domain"
)

type OrderPublisher interface {
	Publish(ctx context.Context, d *domain.Order) error
}
