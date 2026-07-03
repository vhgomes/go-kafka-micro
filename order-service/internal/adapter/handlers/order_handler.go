package handlers

import (
	"encoding/json"
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
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := h.createOrder.SaveOrder(r.Context(), items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
