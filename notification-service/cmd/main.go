package cmd

import (
	"fmt"
	"log"
	"notification-service/internal/adapter/events"
	"notification-service/internal/adapter/sender"
	"notification-service/internal/domain"
	"notification-service/internal/usecase"
	"time"
)

func main() {
	consumer := events.NewKafkaEventConsumer(
		[]string{"localhost:9092"},
		"notification-service-group",
		"orders.created",
	)

	notificationSender := sender.NewLogNotificationSender()
	sendNotificationUC := usecase.NewSendNotification(consumer, notificationSender)

	eventsCh, err := consumer.Consume("orders.created")

	if err != nil {
		log.Fatalf("❌ Erro ao iniciar consumer: %v", err)
	}

	log.Println("🚀 Notification Service rodando e ouvindo o tópico 'orders.created'")
	for event := range eventsCh {
		notification := domain.Notification{
			OrderID:   event.OrderID,
			UserID:    event.UserID,
			CreatedAt: time.Now(),
			Message:   fmt.Sprintf("Seu pedido %s foi criado com sucesso!", event.OrderID),
		}

		// Executar caso de uso
		if err := sendNotificationUC.Execute(notification); err != nil {
			log.Printf("❌ Falha ao enviar notificação: %v", err)
		}
	}
}
