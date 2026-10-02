package topics

import (
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("Failed to list topics: %v", err)
		http.Error(w, "Failed to fetch topics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(topics); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var topic Topic

	if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(r.Context(), topic)
	if err != nil {
		log.Printf("Failed to create topic: %v", err)
		http.Error(w, "Failed to create topic", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(created); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
