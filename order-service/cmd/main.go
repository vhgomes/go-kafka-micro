package main

import (
	"context"
	"log"
	"net/http"
	"order-service/internal/adapter/handlers"
	"order-service/internal/adapter/repo"
	"order-service/internal/usecase"
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

	orderRepo := repo.NewInMemoryOrderRepository()
	orderPublisher := repo.NewKafkaOrderPublisher(
		[]string{kafkaBroker},
		"orders.created",
	)

	log.Println("✅ Repo e publisher criados:", orderRepo, orderPublisher)

	createOrderUseCase := usecase.CreateNewOrder(orderRepo, orderPublisher)

	log.Println("✅ Usecase inicializados:", orderRepo, orderPublisher)

	orderHandler := handlers.NewOrderHandler(createOrderUseCase)
	router := handlers.NewRouter(orderHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("🚀 Order Service running on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 Recebido sinal de interrupção, iniciando graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro durante o shutdown: %v", err)
	}

	log.Println("✅ Servidor finalizado com sucesso.")
}
