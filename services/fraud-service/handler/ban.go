package handler

import (
	"context"
	"net/http"
	"encoding/json"

	"github.com/vexyruu/LIP/shared/apitypes"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	if err := h.setUserBanned(r.Context(), userID); err != nil {
		http.Error(w, "failed to ban user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apitypes.BanUserResponse{
		UserID: userID,
		Status: "banned",
	})
}

func (h *Handler) setUserBanned(ctx context.Context, userID string) error {
	session := h.Neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Set the user as banned and set the risk score to 0.9
	result, err := session.Run(ctx, `
		MERGE (u:User {user_id: $userID})
		SET u.banned = true, u.risk_score = 0.9
		RETURN u.user_id`,
		map[string]any{"userID": userID},
	)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

