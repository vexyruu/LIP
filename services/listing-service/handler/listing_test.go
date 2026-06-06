package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateListing_InvalidJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/v1/listings", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	h.CreateListing(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateListing_Validation(t *testing.T) {
	const validUUID = "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing user_id",
			body: `{"title":"Bag","price_ask":10,"condition":1,"category_id":1}`,
		},
		{
			name: "invalid user_id format",
			body: `{"user_id":"not-a-uuid","title":"Bag","price_ask":10,"condition":1,"category_id":1}`,
		},
		{
			name: "empty title",
			body: `{"user_id":"` + validUUID + `","title":"","price_ask":10,"condition":1,"category_id":1}`,
		},
		{
			name: "title too long",
			body: `{"user_id":"` + validUUID + `","title":"` + strings.Repeat("a", 129) + `","price_ask":10,"condition":1,"category_id":1}`,
		},
		{
			name: "zero price",
			body: `{"user_id":"` + validUUID + `","title":"Bag","price_ask":0,"condition":1,"category_id":1}`,
		},
		{
			name: "negative price",
			body: `{"user_id":"` + validUUID + `","title":"Bag","price_ask":-5,"condition":1,"category_id":1}`,
		},
		{
			name: "condition below range",
			body: `{"user_id":"` + validUUID + `","title":"Bag","price_ask":10,"condition":0,"category_id":1}`,
		},
		{
			name: "condition above range",
			body: `{"user_id":"` + validUUID + `","title":"Bag","price_ask":10,"condition":6,"category_id":1}`,
		},
		{
			name: "zero category_id",
			body: `{"user_id":"` + validUUID + `","title":"Bag","price_ask":10,"condition":1,"category_id":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			req := httptest.NewRequest(http.MethodPost, "/v1/listings", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.CreateListing(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestAssignListing_InvalidJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/v1/listings/x/assign", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	h.AssignListing(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAssignListing_MissingFields(t *testing.T) {
	const id = "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name   string
		pathID string
		body   string
	}{
		{name: "missing moderator id", pathID: id, body: `{"moderator_id":""}`},
		{name: "missing listing id", pathID: "", body: `{"moderator_id":"` + id + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			req := httptest.NewRequest(http.MethodPost, "/v1/listings/x/assign", strings.NewReader(tt.body))
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			w := httptest.NewRecorder()
			h.AssignListing(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestUnassignListing_MissingFields(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodDelete, "/v1/listings/x/assign", strings.NewReader(`{"moderator_id":""}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	w := httptest.NewRecorder()
	h.UnassignListing(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
