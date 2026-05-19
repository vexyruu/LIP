package handler 

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Risk response struct
type RiskResponse struct {
	UserID string `json:"user_id"`
	RiskScore float64 `json:"risk_score"`
	RiskTier string `json:"risk_tier"`
	CacheHit bool `json:"cache_hit"`
	ComputedAt string `json:"computed_at"`
	VelocityCount int64 `json:"velocity_count"`
	VelocityScore float64 `json:"velocity_score"`
	VelocityFlags []string `json:"velocity_flags"`
}

// Handler struct
type Handler struct {
	Redis *redis.Client
	Neo4j neo4j.DriverWithContext
}	

// Get risk handler
func (h *Handler) GetRisk(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	// Increment velocity counter
	key := fmt.Sprintf("velocity:listings:%s", userID)
	count, err := h.Redis.Incr(r.Context(), key).Result()
	if err != nil {
		http.Error(w, "failed to increment velocity counter", http.StatusInternalServerError)
		return
	}
	if count == 1{
		h.Redis.Expire(r.Context(), key, time.Hour)
	}

	// Calculate velocity score and flags
	const baseline = 10.0
	velocityScore := float64(count) / baseline
	flags := velocityFlags(count)

	// Check if risk is in Redis
	risk, err := h.Redis.Get(r.Context(), fmt.Sprintf("risk:%s", userID)).Result()

	// if risk is not in Redis, query Neo4j, parse the result and return it
	if err == redis.Nil {
		session := h.Neo4j.NewSession(r.Context(), neo4j.SessionConfig{})
		defer session.Close(r.Context())

		result, err := session.Run(r.Context(),
        	`MATCH (u:User {user_id: $userID})-[:USED_DEVICE|SHARED_IP|USED_PAYMENT]-(banned:User {banned: true})
         	RETURN count(banned) > 0 AS connected`,
        	map[string]any{"userID": userID},
    	)
		if err != nil {
			http.Error(w, "failed to query Neo4j", http.StatusInternalServerError)
			return
		}
		score := 0.0
		if result.Next(r.Context()) {
			connected, _ := result.Record().Get("connected")
			if connected.(bool) {
				score = 0.9
			}
		}

		resp := RiskResponse{
			UserID: userID,
			RiskScore: score,
			RiskTier: riskTier(score),
			CacheHit: false,
			VelocityCount: count,
			VelocityScore: velocityScore,
			VelocityFlags: flags,
			ComputedAt: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	// if there is an actual error, return it
	} else if err != nil{
		http.Error(w, "failed to get risk", http.StatusInternalServerError)
		return

	// if risk is in Redis, parse it and return it
	} else {
		score, err := strconv.ParseFloat(risk, 64)
		if err != nil {
			http.Error(w, "failed to parse risk score", http.StatusInternalServerError)
			return
		}
		resp := RiskResponse{
			UserID: userID,
			RiskScore: score,
			RiskTier: riskTier(score),
			CacheHit: true,
			VelocityCount: count,
			VelocityScore: velocityScore,
			VelocityFlags: flags,
			ComputedAt: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}


}

// Risk classification function
func riskTier(score float64) string {
	if score > 0.8 {
		return "HIGH"
	} else if score > 0.4 {
		return "MEDIUM"
	}
	return "LOW"
}

func velocityFlags(count int64) []string {
	flags := []string{}
	if count > 10 {
		flags = append(flags, "HIGH_LISTING_VELOCITY")
	}
	if count > 30 {
		flags = append(flags, "BURST_LISTING_VELOCITY")
	}
	return flags
}