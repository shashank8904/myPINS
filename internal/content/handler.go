package content

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// ContentGetter is the subset of Service used by the Get handler.
// Introduced to allow handler unit testing without a live database.
type ContentGetter interface {
	GetByID(ctx context.Context, id int64) (ContentItem, error)
}

// Handler provides HTTP handlers for content items and feed interactions.
// It holds a reference to the Service, which coordinates between the HTTP
// layer and the Repository.
type Handler struct {
	service *Service
	getter  ContentGetter
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
		getter:  service,
	}
}

// List returns the personalized feed of content items.
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

// Get returns a single content item by ID.
// GET /api/items/{id}
//
// Status codes:
//   - 200: item found and returned
//   - 400: the {id} path value is not a valid integer
//   - 404: no item exists with that ID
//   - 500: database or encoding error
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // requires Go 1.22+
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.getter.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errItemNotFound) {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get item %d: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
