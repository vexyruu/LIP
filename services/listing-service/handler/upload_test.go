package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestCreateUpload_InvalidJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/v1/uploads", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	h.CreateUpload(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateUpload_Validation(t *testing.T) {
	const validUUID = "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing user_id",
			body: `{"filename":"a.jpg","content_type":"image/jpeg","size_bytes":1000}`,
		},
		{
			name: "unsupported content type",
			body: `{"user_id":"` + validUUID + `","filename":"a.svg","content_type":"image/svg+xml","size_bytes":1000}`,
		},
		{
			name: "size too large",
			body: `{"user_id":"` + validUUID + `","filename":"a.jpg","content_type":"image/jpeg","size_bytes":11000000}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			req := httptest.NewRequest(http.MethodPost, "/v1/uploads", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.CreateUpload(w, req)
			if w.Code != http.StatusBadRequest && w.Code != http.StatusServiceUnavailable {
				t.Errorf("expected 400 or 503, got %d", w.Code)
			}
		})
	}
}

func TestCreateUpload_StorageNotConfigured(t *testing.T) {
	h := &Handler{}
	body := `{"user_id":"123e4567-e89b-12d3-a456-426614174000","filename":"a.jpg","content_type":"image/jpeg","size_bytes":1000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/uploads", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateUpload(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
