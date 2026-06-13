package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/vexyruu/LIP/listing-service/store"
	"github.com/vexyruu/LIP/shared/apitypes"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Email == "" || body.DisplayName == "" {
		http.Error(w, "email and display_name are required", http.StatusBadRequest)
		return
	}

	user, err := store.CreateUser(r.Context(), h.Pool, store.CreateUserRequest{
		Email:       body.Email,
		DisplayName: body.DisplayName,
	})
	if err != nil {
		if err.Error() == "email already registered" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	user, err := store.GetUser(r.Context(), h.Pool, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUserListings(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	resp, err := store.ListUserListings(r.Context(), h.Pool, userID, page, limit)
	if err != nil {
		http.Error(w, "failed to list user listings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	var req apitypes.BanUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ModeratorID == "" {
		http.Error(w, "moderator ID is required", http.StatusBadRequest)
		return
	}
	if req.Reason == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}

	// If Postgres fails after Neo4j succeeds, the seller is banned in the graph only
	banURL := fmt.Sprintf("%s/v1/users/%s/ban", h.FraudServiceURL, userID)
	resp, err := http.Post(banURL, "application/json", nil)
	if err != nil {
		http.Error(w, "failed to update fraud graph", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "failed to update fraud graph", http.StatusBadGateway)
		return
	}

	result, err := store.BanUserWithAudit(r.Context(), h.Pool, userID, req.ModeratorID, req.Reason)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrAlreadyBanned) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to ban user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apitypes.BanUserResponse{
		UserID:          userID,
		Status:          "BANNED",
		ListingsFlagged: result.ListingsFlagged,
	})
}
