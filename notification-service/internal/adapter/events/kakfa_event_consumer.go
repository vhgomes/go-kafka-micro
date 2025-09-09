package events

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"notification-service/internal/domain"

	"log"
)

type KafkaEventConsumer struct {
	reader *kafka.Reader
}

func NewKafkaEventConsumer(brokers []string, groupID, topic string) *KafkaEventConsumer {
	return &KafkaEventConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}
func (c *KafkaEventConsumer) Consume(topic string) (<-chan domain.OrderEvent, error) {
	events := make(chan domain.OrderEvent)

	go func() {
		defer close(events)

		for {
			msg, err := c.reader.ReadMessage(context.Background())
			if err != nil {
				log.Fatal("❌ Erro | Consumer Kafka: ", err)
			}

			var event domain.OrderEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Fatal("❌ Erro | Unmarshall: ", err)
			}

			log.Printf("📥 Evento recebido do Kafka: %+v", event)
			events <- event
		}
	}()

	return events, nil
}
