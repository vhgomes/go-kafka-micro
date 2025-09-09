package port

import "notification-service/internal/domain"

type EventConsumer interface {
	Consume(topic string) (<-chan domain.OrderEvent, error)
}
