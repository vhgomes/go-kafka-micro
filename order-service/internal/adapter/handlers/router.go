package handlers

import (
	"net/http"
)

func NewRouter(handler *OrderHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", handler.CreateOrder)
	return mux
}
