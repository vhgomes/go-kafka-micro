package repo

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/segmentio/kafka-go"
	"log"
	"order-service/internal/domain"
	"order-service/internal/port"
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

func (o KafkaOrderPublisher) Publish(order *domain.Order) error {
	payload, err := json.Marshal(order)

	if err != nil {
		return errors.New("KafkaPublisher | PublishOrderCreated error: " + err.Error())
	}

	message := kafka.Message{
		Key:   []byte(order.OrderId.String()),
		Value: payload,
	}

	err = o.writer.WriteMessages(context.Background(), message)
	if err != nil {
		return errors.New("KafkaPublisher | PublishOrderCreated error: " + err.Error())
	}

	log.Printf("KafkaPublisher | PublishOrderCreated - Order ID: %s\n", order.OrderId.String())
	return nil
}
