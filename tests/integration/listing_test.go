//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func listingServiceURL() string {
	if u := os.Getenv("LISTING_SERVICE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:8080"
}

func TestPostListing_BadRequest(t *testing.T) {
	url := listingServiceURL()

	cases := []struct {
		name string
		body map[string]any
	}{
		{"empty title", map[string]any{"user_id": "123e4567-e89b-12d3-a456-426614174000", "title": "", "price_ask": 10, "condition": 1, "category_id": 1}},
		{"zero price", map[string]any{"user_id": "123e4567-e89b-12d3-a456-426614174000", "title": "Bag", "price_ask": 0, "condition": 1, "category_id": 1}},
		{"bad condition", map[string]any{"user_id": "123e4567-e89b-12d3-a456-426614174000", "title": "Bag", "price_ask": 10, "condition": 9, "category_id": 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			resp, err := http.Post(url+"/v1/listings", "application/json", bytes.NewReader(b))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

// TestPostListingE2E submits a listing and polls until it leaves DRAFT status.
func TestPostListingE2E(t *testing.T) {
	url := listingServiceURL()

	payload, _ := json.Marshal(map[string]any{
		"user_id":	"00000000-0000-0000-0000-000000000002",
		"title":	"Integration test hoodie",
		"description":	"Like new, worn twice",
		"price_ask":	35.00,
		"condition":	2,
		"category_id":	1,
	})

	resp, err := http.Post(url+"/v1/listings", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /v1/listings: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	var created struct {
		ListingID string `json:"listing_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil || created.ListingID == "" {
		t.Fatalf("no listing_id in response: %v", err)
	}
	t.Logf("created listing %s, polling for status change", created.ListingID)

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)

		getResp, err := http.Get(url + "/v1/listings/" + created.ListingID)
		if err != nil {
			t.Fatalf("GET /v1/listings/%s: %v", created.ListingID, err)
		}
		var listing struct {
			Status string `json:"status"`
		}
		json.NewDecoder(getResp.Body).Decode(&listing)
		getResp.Body.Close()

		if listing.Status != "DRAFT" {
			t.Logf("listing %s reached status %s", created.ListingID, listing.Status)
			return
		}
	}
	t.Fatalf("listing %s still in DRAFT after 30s", created.ListingID)
}
