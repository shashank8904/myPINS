package interests

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
	interests, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("Failed to list interests: %v", err)
		http.Error(w, "Failed to fetch interests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(interests); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var interest Interest

	if err := json.NewDecoder(r.Body).Decode(&interest); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(r.Context(), interest)
	if err != nil {
		log.Printf("Failed to create interest: %v", err)
		http.Error(w, "Failed to create interest", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(created); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
