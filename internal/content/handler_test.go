package content

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockGetter struct {
	item ContentItem
	err  error
}

func (m *mockGetter) GetByID(ctx context.Context, id int64) (ContentItem, error) {
	return m.item, m.err
}

func TestGetItem_NotFound(t *testing.T) {
	handler := &Handler{
		getter: &mockGetter{
			err: errItemNotFound,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/items/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d; got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestGetItem_OK(t *testing.T) {
	expectedItem := ContentItem{ID: 1, Title: "Test Item"}
	handler := &Handler{
		getter: &mockGetter{
			item: expectedItem,
			err:  nil,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/items/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, res.StatusCode)
	}
}

func TestGetItem_InternalError(t *testing.T) {
	handler := &Handler{
		getter: &mockGetter{
			err: errors.New("some DB error"),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/items/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, res.StatusCode)
	}
}
