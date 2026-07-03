package main

import (
	"context"
	"fmt"
	"log"
	"notification-service/internal/adapter/events"
	"notification-service/internal/adapter/sender"
	"notification-service/internal/domain"
	"notification-service/internal/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER_ADDRESS")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}

	consumer := events.NewKafkaEventConsumer(
		[]string{kafkaBroker},
		"notification-service-group",
		"orders.created",
	)

	notificationSender := sender.NewLogNotificationSender()
	sendNotificationUC := usecase.NewSendNotification(notificationSender)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventsCh, err := consumer.Consume(ctx, "orders.created")
	if err != nil {
		log.Fatalf("❌ Erro ao iniciar consumer: %v", err)
	}

	log.Println("🚀 Notification Service rodando e ouvindo o tópico 'orders.created'")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		for event := range eventsCh {
			notification := domain.Notification{
				OrderID:   event.OrderID,
				UserID:    event.UserID,
				CreatedAt: time.Now(),
				Message:   fmt.Sprintf("Seu pedido %s foi criado com sucesso!", event.OrderID),
			}

			if err := sendNotificationUC.Execute(notification); err != nil {
				log.Printf("❌ Falha ao enviar notificação: %v", err)
			}
		}
	}()

	<-quit
	log.Println("🛑 Sinal de parada recebido. Iniciando Graceful Shutdown...")
	cancel()

	if err := consumer.Close(); err != nil {
		log.Printf("⚠️ Erro ao fechar o consumer: %v", err)
	}

	log.Println("👋 Notification Service desligado com segurança.")
}
