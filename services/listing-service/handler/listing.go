package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/vexyruu/LIP/listing-service/model"
	"github.com/vexyruu/LIP/listing-service/store"
	"github.com/vexyruu/LIP/shared/risk"
	"github.com/vexyruu/LIP/shared/storage"
)

type Handler struct {
	Pool                   *pgxpool.Pool
	FraudServiceURL        string
	Storage                *storage.Client
	Redis                  *redis.Client
	UploadPublicBaseURL    string
	AllowExternalImageURLs bool
}

func (h *Handler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req model.CreateListingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validateListingImages(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := store.CreateListing(r.Context(), h.Pool, &req)
	if err != nil {
		http.Error(w, "failed to create listing", http.StatusInternalServerError)
		return
	}

	h.recordListingVelocity(r.Context(), req.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(model.CreateListingResponse{
		ListingID: id,
		Status:    model.APIStatusProcessing,
		ETA:       500,
	})
}

// best-effort: a Redis failure must never block or fail listing creation.
func (h *Handler) recordListingVelocity(ctx context.Context, userID string) {
	if h.Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	key := risk.VelocityKey(userID)
	count, err := h.Redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("velocity incr failed for user %s: %v", userID, err)
		return
	}
	// Open a fresh 1h window on the first listing in the window.
	if count == 1 {
		if err := h.Redis.Expire(ctx, key, time.Hour).Err(); err != nil {
			log.Printf("velocity expire failed for user %s: %v", userID, err)
		}
	}
}

// GetListing retrieves a listing from the database
func (h *Handler) GetListing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	listing, err := store.GetListing(r.Context(), h.Pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "listing not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get listing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

// ListListings lists listings in the database
func (h *Handler) ListListings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	status := q.Get("status")
	if status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	resp, err := store.ListListings(r.Context(), h.Pool, model.ListListingsFilter{
		Status: status,
		Tier:   q.Get("tier"),
		Sort:   q.Get("sort"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		http.Error(w, "failed to list listings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) PatchListing(w http.ResponseWriter, r *http.Request) {
	listingID := r.PathValue("id")
	var req model.PatchListingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if listingID == "" {
		http.Error(w, "listing ID is required", http.StatusBadRequest)
		return
	}
	if req.ModeratorID == "" {
		http.Error(w, "moderator ID is required", http.StatusBadRequest)
		return
	}
	if req.Action != "APPROVE" && req.Action != "REJECT" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}
	if req.Action == "REJECT" && req.Reason == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}

	// Moderate listing
	resp, err := store.ModerateListing(r.Context(), h.Pool, listingID, &req)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "listing not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrNotUnderReview) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, store.ErrAssignedToOther) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "failed to moderate listing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// claim a listing for the requesting moderator
func (h *Handler) AssignListing(w http.ResponseWriter, r *http.Request) {
	listingID := r.PathValue("id")
	var req model.AssignListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if listingID == "" || req.ModeratorID == "" {
		http.Error(w, "listing ID and moderator ID are required", http.StatusBadRequest)
		return
	}

	resp, err := store.AssignListing(r.Context(), h.Pool, listingID, req.ModeratorID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "listing not found", http.StatusNotFound)
		case errors.Is(err, store.ErrNotUnderReview), errors.Is(err, store.ErrAlreadyAssigned):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "failed to assign listing", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// release the requesting moderator's claim on a listing
func (h *Handler) UnassignListing(w http.ResponseWriter, r *http.Request) {
	listingID := r.PathValue("id")
	var req model.AssignListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if listingID == "" || req.ModeratorID == "" {
		http.Error(w, "listing ID and moderator ID are required", http.StatusBadRequest)
		return
	}

	resp, err := store.UnassignListing(r.Context(), h.Pool, listingID, req.ModeratorID)
	if err != nil {
		if errors.Is(err, store.ErrNotAssignedToYou) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to release listing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetAnalytics gets the analytics summary for the listing service
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	summary, err := store.GetAnalyticsSummary(r.Context(), h.Pool)
	if err != nil {
		http.Error(w, "failed to get analytics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// ListModerationDecisions lists the moderation decisions for the listing service
func (h *Handler) ListModerationDecisions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	resp, err := store.ListModerationDecisions(r.Context(), h.Pool, model.ListModerationDecisionsFilter{
		Action: q.Get("action"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		http.Error(w, "failed to list moderation decisions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
