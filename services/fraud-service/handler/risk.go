package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"github.com/vexyruu/LIP/shared/apitypes"
	"github.com/vexyruu/LIP/shared/risk"
)

type Handler struct {
	Redis   *redis.Client
	Neo4j   neo4j.DriverWithContext
	Weights risk.Weights
}

func (h *Handler) GetRisk(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	key := fmt.Sprintf("velocity:listings:%s", userID)
	count, err := h.Redis.Incr(r.Context(), key).Result()
	if err != nil {
		http.Error(w, "failed to increment velocity counter", http.StatusInternalServerError)
		return
	}
	if count == 1 {
		h.Redis.Expire(r.Context(), key, time.Hour)
	}

	velScore := risk.DefaultVelocityBaseline(count)
	flags := risk.VelocityFlags(count)

	batch, cacheHit, err := h.loadBatchScores(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get risk", http.StatusInternalServerError)
		return
	}

	directBanned := h.queryNeo4jBannedConnection(r.Context(), userID)
	score := risk.EffectiveTotalScore(batch, velScore, h.Weights, directBanned)

	resp := apitypes.RiskResponse{
		UserID:            userID,
		RiskScore:         score,
		RiskTier:          risk.Tier(score),
		ComponentScore:    batch.ComponentScore,
		PageRankProximity: batch.PageRankProximity,
		CacheHit:          cacheHit,
		VelocityCount:     count,
		VelocityScore:     velScore,
		VelocityFlags:     flags,
		ComputedAt:        time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) loadBatchScores(ctx context.Context, userID string) (risk.BatchScores, bool, error) {
	raw, err := h.Redis.Get(ctx, fmt.Sprintf("risk:%s", userID)).Result()
	if err == redis.Nil {
		return risk.BatchScores{}, false, nil
	}
	if err != nil {
		return risk.BatchScores{}, false, err
	}

	var batch risk.BatchScores
	if json.Unmarshal([]byte(raw), &batch) == nil {
		return batch, true, nil
	}

	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return risk.BatchScores{ComponentScore: f}, true, nil
	}
	return risk.BatchScores{}, false, fmt.Errorf("failed to parse batch scores for user %s", userID)
}

func (h *Handler) queryNeo4jBannedConnection(ctx context.Context, userID string) bool {
	session := h.Neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (u:User {user_id: $userID})-[:USED_DEVICE|SHARED_IP|USED_PAYMENT]-(banned:User {banned: true})
		RETURN count(banned) > 0 AS connected`,
		map[string]any{"userID": userID},
	)
	if err != nil || !result.Next(ctx) {
		return false
	}
	connected, _ := result.Record().Get("connected")
	b, _ := connected.(bool)
	return b
}
