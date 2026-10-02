package content

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// List returns the feed of content items.
// GET /api/feed
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("Failed to list feed: %v", err)
		http.Error(w, "Failed to fetch feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Get returns a single content item.
// GET /api/items/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // requires Go 1.22+
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get item %d: %v", id, err)
		// Usually we'd check for a Not Found error specifically, but for V0 this is ok
		http.Error(w, "Item not found or server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) interact(w http.ResponseWriter, r *http.Request, action string) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	err = h.service.Interact(r.Context(), id, action)
	if err != nil {
		log.Printf("Failed to record %s interaction for item %d: %v", action, id, err)
		http.Error(w, "Failed to record interaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// InteractRead marks an item as read.
// POST /api/items/{id}/read
func (h *Handler) InteractRead(w http.ResponseWriter, r *http.Request) {
	h.interact(w, r, "read")
}

// InteractSave marks an item as saved.
// POST /api/items/{id}/save
func (h *Handler) InteractSave(w http.ResponseWriter, r *http.Request) {
	h.interact(w, r, "saved")
}

// InteractDismiss marks an item as dismissed.
// POST /api/items/{id}/dismiss
func (h *Handler) InteractDismiss(w http.ResponseWriter, r *http.Request) {
	h.interact(w, r, "dismissed")
}
