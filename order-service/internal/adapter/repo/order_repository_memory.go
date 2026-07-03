package repo

import (
	"context"
	"order-service/internal/domain"
	"order-service/internal/port"
	"sync"
)

type InMemoryOrderRepository struct {
	data map[string]domain.Order
	mu   sync.RWMutex
}

func NewInMemoryOrderRepository() port.OrderRepository {
	return &InMemoryOrderRepository{
		data: make(map[string]domain.Order),
	}
}

func (i *InMemoryOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.data[order.OrderId.String()] = *order
	return nil
}
