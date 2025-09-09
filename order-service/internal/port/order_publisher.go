package port

import "order-service/internal/domain"

type OrderPublisher interface {
	Publish(d *domain.Order) error
}
