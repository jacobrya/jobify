package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"hello": "world"})

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: got %q, want application/json", ct)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.Error != "" {
		t.Errorf("expected empty error, got %q", resp.Error)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Success {
		t.Error("expected Success=false")
	}
	if resp.Error != "bad input" {
		t.Errorf("error: got %q, want %q", resp.Error, "bad input")
	}
}

func TestWithMeta(t *testing.T) {
	w := httptest.NewRecorder()
	WithMeta(w, http.StatusOK, []int{1, 2, 3}, &Meta{Total: 3, Page: 1, Limit: 10, Pages: 1})

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if resp.Meta.Total != 3 {
		t.Errorf("meta.total: got %d, want 3", resp.Meta.Total)
	}
	if resp.Meta.Page != 1 {
		t.Errorf("meta.page: got %d, want 1", resp.Meta.Page)
	}
}

func TestError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}
