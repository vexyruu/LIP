package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vexyruu/LIP/shared/apitypes"
	"github.com/vexyruu/LIP/shared/risk"
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

	result, err := session.Run(ctx, `
		MERGE (u:User {user_id: $userID})
		SET u.banned = true, u.risk_score = 0.9
		RETURN u.user_id`,
		map[string]any{"userID": userID},
	)
	if err != nil {
		return err
	}
	if _, err = result.Consume(ctx); err != nil {
		return err
	}

	// flush Redis so GetRisk returns HIGH immediately, not after the next batch run.
	bannedScore := risk.BatchScores{ComponentScore: 0.9}
	payload, err := json.Marshal(bannedScore)
	if err != nil {
		return err
	}
	if err := h.Redis.Set(ctx, fmt.Sprintf("risk:%s", userID), payload, time.Hour).Err(); err != nil {
		log.Printf("ban: failed to update Redis cache for user %s: %v", userID, err)
	}
	return nil
}

