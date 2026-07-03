package port

import (
	"context"

	"notification-service/internal/domain"
)

type EventConsumer interface {
	Consume(ctx context.Context, topic string) (<-chan domain.OrderEvent, error)
}
