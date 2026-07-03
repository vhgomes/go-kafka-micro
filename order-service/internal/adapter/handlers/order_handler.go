package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type OrderHandler struct {
	createOrder *usecase.CreateOrder
}

func NewOrderHandler(co *usecase.CreateOrder) *OrderHandler {
	return &OrderHandler{co}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var items []domain.Item

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&items); err != nil {
		log.Printf("Erro ao decodificar request: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	output, err := h.createOrder.SaveOrder(r.Context(), items)
	if err != nil {
		log.Printf("Erro interno: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(output)
	if err != nil {
		log.Printf("Erro ao encodar resposta: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
