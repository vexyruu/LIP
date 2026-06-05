package apitypes

// ListingCreatedEvent is the Pub/Sub payload published when a listing is
// submitted. Attempt counts analysis tries: 0 on initial submit, incremented
// when analysis-worker re-enqueues after an ml-service failure (fail-closed
// re-analysis). It bounds re-analysis loops while ml-service is unavailable.
type ListingCreatedEvent struct {
	ListingID   string  `json:"listing_id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Condition   int16   `json:"condition"`
	CategoryID  int     `json:"category_id"`
	PriceAsk    float64 `json:"price_ask"`
	Attempt     int     `json:"attempt,omitempty"`
}

// RiskResponse is returned by GET /v1/risk/{user_id} on fraud-service.
type RiskResponse struct {
	UserID            string   `json:"user_id"`
	RiskScore         float64  `json:"risk_score"`
	RiskTier          string   `json:"risk_tier"`
	ComponentScore    float64  `json:"component_score"`
	PageRankProximity float64  `json:"pagerank_proximity"`
	CacheHit          bool     `json:"cache_hit"`
	ComputedAt        string   `json:"computed_at"`
	VelocityCount     int64    `json:"velocity_count"`
	VelocityScore     float64  `json:"velocity_score"`
	VelocityFlags     []string `json:"velocity_flags"`
}

type GraphNode struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Label     string   `json:"label"`
	Banned    *bool    `json:"banned,omitempty"`
	RiskScore *float64 `json:"risk_score,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type GraphResponse struct {
	UserID string      `json:"user_id"`
	Nodes  []GraphNode `json:"nodes"`
	Edges  []GraphEdge `json:"edges"`
}

type UserBanAudit struct {
	Reason          string `json:"reason"`
	ModeratorID     string `json:"moderator_id"`
	ListingsFlagged int    `json:"listings_flagged"`
	CreatedAt       string `json:"created_at"`
}

type UserProfile struct {
	UserID      string        `json:"user_id"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name"`
	Status      string        `json:"status"`
	CreatedAt   string        `json:"created_at"`
	BanAudit    *UserBanAudit `json:"ban_audit,omitempty"`
}

type BanUserRequest struct {
	ModeratorID string `json:"moderator_id"`
	Reason      string `json:"reason"`
}

type BanUserResponse struct {
	UserID          string `json:"user_id"`
	Status          string `json:"status"`
	ListingsFlagged int    `json:"listings_flagged"`
}
