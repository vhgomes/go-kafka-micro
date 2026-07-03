package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/domain"
	"order-service/internal/port"

	"github.com/segmentio/kafka-go"
)

type KafkaOrderPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaOrderPublisher(brokers []string, topic string) port.OrderPublisher {
	return &KafkaOrderPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (o KafkaOrderPublisher) Publish(ctx context.Context, order *domain.Order) error {
	payload, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("KafkaPublisher | PublishOrderCreated error: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(order.OrderId.String()),
		Value: payload,
	}

	err = o.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("KafkaPublisher | PublishOrderCreated error: %w", err)
	}

	log.Printf("KafkaPublisher | PublishOrderCreated - Order ID: %s\n", order.OrderId.String())
	return nil
}
