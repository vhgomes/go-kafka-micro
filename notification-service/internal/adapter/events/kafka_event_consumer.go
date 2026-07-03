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
	done   chan struct{}
}

func NewKafkaEventConsumer(brokers []string, groupID, topic string) *KafkaEventConsumer {
	return &KafkaEventConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		done: make(chan struct{}),
	}
}

func (c *KafkaEventConsumer) Consume(ctx context.Context, topic string) (<-chan domain.OrderEvent, error) {
	events := make(chan domain.OrderEvent)

	go func() {
		defer close(events)
		defer close(c.done)

		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("🛑 Consumer Kafka encerrado via contexto")
					return
				}
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

func (c *KafkaEventConsumer) Close() error {
	<-c.done
	return c.reader.Close()
}
