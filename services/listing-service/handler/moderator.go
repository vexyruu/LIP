package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vexyruu/LIP/listing-service/store"
)

func (h *Handler) ListModerators(w http.ResponseWriter, r *http.Request) {
	mods, err := store.ListModerators(r.Context(), h.Pool)
	if err != nil {
		http.Error(w, "failed to list moderators", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mods)
}

func (h *Handler) GetModeratorByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	mod, err := store.GetModeratorByEmail(r.Context(), h.Pool, email)
	if err != nil {
		if errors.Is(err, store.ErrModeratorNotFound) {
			http.Error(w, "moderator not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get moderator", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mod)
}

func (h *Handler) CreateModerator(w http.ResponseWriter, r *http.Request) {
	var req store.ModeratorRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.DisplayName == "" || req.Role == "" || req.PasswordHash == "" {
		http.Error(w, "email, display_name, role, and password_hash are required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateModerator(r.Context(), h.Pool, &req)
	if err != nil {
		if errors.Is(err, store.ErrModeratorExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to create moderator", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) DeleteModerator(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "moderator ID is required", http.StatusBadRequest)
		return
	}
	if err := store.DeleteModerator(r.Context(), h.Pool, id); err != nil {
		if errors.Is(err, store.ErrModeratorNotFound) {
			http.Error(w, "moderator not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete moderator", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
