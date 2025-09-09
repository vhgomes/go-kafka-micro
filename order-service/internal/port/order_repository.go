package port

import "order-service/internal/domain"

type OrderRepository interface {
	Save(order *domain.Order) error
}
