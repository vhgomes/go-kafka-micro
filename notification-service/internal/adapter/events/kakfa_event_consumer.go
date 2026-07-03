package events

import (
	"context"
	"encoding/json"
	"notification-service/internal/domain"

	"github.com/segmentio/kafka-go"

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
				log.Println("❌ Erro | Consumer Kafka: ", err)
				continue
			}

			var event domain.OrderEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Println("❌ Erro | Unmarshall: ", err)
				continue
			}

			log.Printf("📥 Evento recebido do Kafka: %+v", event)
			events <- event
		}
	}()

	return events, nil
}
