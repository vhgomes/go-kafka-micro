package main

import (
	"log"
	"net/http"
	"order-service/internal/adapter/handlers"
	"order-service/internal/adapter/repo"
	"order-service/internal/usecase"
)

func main() {
	orderRepo := repo.NewInMemoryOrderRepository()

	orderPublisher := repo.NewKafkaOrderPublisher(
		[]string{"localhost:9092"},
		"orders.created",
	)

	log.Println("✅ Repo e publisher criados:", orderRepo, orderPublisher)

	createOrderUseCase := usecase.CreateNewOrder(orderRepo, orderPublisher)

	log.Println("✅ Usecase inicializados:", orderRepo, orderPublisher)

	orderHandler := handlers.NewOrderHandler(createOrderUseCase)

	router := handlers.NewRouter(orderHandler)

	log.Println("🚀 Order Service running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
