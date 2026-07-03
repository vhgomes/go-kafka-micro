package usecase

import (
	"context"
	"fmt"
	"order-service/internal/domain"
	"order-service/internal/port"

	"github.com/google/uuid"

	"time"
)

type CreateOrder struct {
	repo      port.OrderRepository
	publisher port.OrderPublisher
}

type CreateOrderOutput struct {
	ID          uuid.UUID
	TotalAmount int64
	CreatedAt   time.Time
}

func CreateNewOrder(repo port.OrderRepository, publisher port.OrderPublisher) *CreateOrder {
	return &CreateOrder{
		repo:      repo,
		publisher: publisher,
	}
}

func (co *CreateOrder) SaveOrder(ctx context.Context, itens []domain.Item) (*CreateOrderOutput, error) {
	if len(itens) == 0 {
		return nil, fmt.Errorf("você precisa enviar itens")
	}

	var total int64 = 0
	for _, item := range itens {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantidade de item invalido")
		}
		total += item.Price
	}

	order := domain.Order{
		OrderId:     uuid.New(),
		Items:       itens,
		TotalAmount: total,
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}

	if err := co.repo.Save(ctx, &order); err != nil {
		return nil, fmt.Errorf("falha no repositorio: erro ao salvar")
	}

	if err := co.publisher.Publish(ctx, &order); err != nil {
		return nil, fmt.Errorf("falha no publisher: erro ao publicar a mensagem de ordem criada")
	}

	return &CreateOrderOutput{
		ID:          order.OrderId,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
	}, nil
}
