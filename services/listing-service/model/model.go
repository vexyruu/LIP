package model

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type CreateListingRequest struct {
	UserID      string   `json:"user_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Condition   int16    `json:"condition"`
	CategoryID  int      `json:"category_id"`
	PriceAsk    float64  `json:"price_ask"`
	Images      []string `json:"images"`
}

func (r *CreateListingRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if !uuidRE.MatchString(r.UserID) {
		return errors.New("user_id must be a valid UUID")
	}
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 128 {
		return errors.New("title must be 128 characters or fewer")
	}
	if r.PriceAsk <= 0 {
		return errors.New("price_ask must be greater than 0")
	}
	if r.Condition < 1 || r.Condition > 5 {
		return errors.New("condition must be between 1 and 5")
	}
	if r.CategoryID <= 0 {
		return errors.New("category_id must be greater than 0")
	}
	if len(r.Images) > 10 {
		return errors.New("images must contain at most 10 URLs")
	}
	for i, raw := range r.Images {
		u := strings.TrimSpace(raw)
		if u == "" {
			return errors.New("images must not contain empty URLs")
		}
		if len(u) > 2048 {
			return errors.New("image URL is too long")
		}
		parsed, err := url.Parse(u)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return errors.New("images must be valid http or https URLs")
		}
		r.Images[i] = u
	}
	return nil
}

type CreateListingResponse struct {
	ListingID string `json:"listing_id"`
	Status    string `json:"status"`
	ETA       int    `json:"eta_ms"`
}

// API-facing listing states
const (
	APIStatusProcessing  = "processing"
	APIStatusLive        = "live"
	APIStatusUnderReview = "under_review"
	APIStatusRejected    = "rejected"
)

// APIStatus maps an internal DB listing_status to the client-facing status
func APIStatus(dbStatus string) string {
	switch dbStatus {
	case "LIVE":
		return APIStatusLive
	case "UNDER_REVIEW":
		return APIStatusUnderReview
	case "REJECTED":
		return APIStatusRejected
	default:
		return APIStatusProcessing
	}
}

type ListingStatus struct {
	ListingID        string     `json:"listing_id"`
	UserID           string     `json:"user_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	PriceAsk         float64    `json:"price_ask"`
	Condition        int16      `json:"condition"`
	CategoryID       int        `json:"category_id"`
	Status           string     `json:"status"`
	APIStatus        string     `json:"api_status"`
	SuggestedPrice   *float64   `json:"suggested_price"`
	PriceLowerBound  *float64   `json:"price_lower_bound"`
	PriceUpperBound  *float64   `json:"price_upper_bound"`
	RiskTier         *string    `json:"risk_tier"`
	RiskScore        *float64   `json:"risk_score"`
	ExtractedBrand   *string    `json:"extracted_brand"`
	ExtractedProduct *string    `json:"extracted_product"`
	ExtractedSize    *string    `json:"extracted_size"`
	PolicyViolation  *bool      `json:"policy_violation"`
	Images           []string   `json:"images"`
	AssignedTo       *string    `json:"assigned_to"`
	AssignedAt       *time.Time `json:"assigned_at"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}

type ListingQueueItem struct {
	ListingID  string    `json:"listing_id"`
	Title      string    `json:"title"`
	UserID     string    `json:"user_id"`
	PriceAsk   float64   `json:"price_ask"`
	RiskScore  *float64  `json:"risk_score"`
	RiskTier   *string   `json:"risk_tier"`
	CategoryID int       `json:"category_id"`
	Images     []string  `json:"images"`
	AssignedTo *string   `json:"assigned_to"`
	CreatedAt  time.Time `json:"created_at"`
}

type AssignListingRequest struct {
	ModeratorID string `json:"moderator_id"`
}

type AssignListingResponse struct {
	ListingID  string  `json:"listing_id"`
	AssignedTo *string `json:"assigned_to"`
}

type ListListingsResponse struct {
	Listings []ListingQueueItem `json:"listings"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

type PatchListingRequest struct {
	Action      string `json:"action"`
	ModeratorID string `json:"moderator_id"`
	Reason      string `json:"reason"`
}

type PatchListingResponse struct {
	ListingID string `json:"listing_id"`
	Status    string `json:"status"`
}

type ListListingsFilter struct {
	Status string
	Tier   string
	Sort   string
	Page   int
	Limit  int
}

type DailySubmissionCount struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
	High  int    `json:"high_risk"`
}

type RecentModerationRow struct {
	Timestamp string  `json:"timestamp"`
	ListingID string  `json:"listing_id"`
	Title     string  `json:"title"`
	Action    string  `json:"action"`
	RiskTier  *string `json:"risk_tier"`
	Reason    string  `json:"reason"`
}

type AnalyticsSummary struct {
	TotalListings      int                    `json:"total_listings"`
	UnderReview        int                    `json:"under_review"`
	ByStatus           map[string]int         `json:"by_status"`
	ByRiskTier         map[string]int         `json:"by_risk_tier"`
	PolicyRejections   int                    `json:"policy_rejections"`
	AvgPipelineSeconds *float64               `json:"avg_pipeline_seconds"`
	Last24hSubmissions int                    `json:"last_24h_submissions"`
	DailySubmissions   []DailySubmissionCount `json:"daily_submissions"`
	RecentDecisions    []RecentModerationRow  `json:"recent_decisions"`
}

type ModerationDecisionRow struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	ListingID   string  `json:"listing_id"`
	Title       string  `json:"title"`
	Action      string  `json:"action"`
	RiskTier    *string `json:"risk_tier"`
	Reason      string  `json:"reason"`
	ModeratorID string  `json:"moderator_id"`
}

type ListModerationDecisionsFilter struct {
	Action string
	Page   int
	Limit  int
}

type ListModerationDecisionsResponse struct {
	Decisions []ModerationDecisionRow `json:"decisions"`
	Total     int                     `json:"total"`
	Page      int                     `json:"page"`
	Limit     int                     `json:"limit"`
}

type UserBanAudit struct {
	Reason          string    `json:"reason"`
	ModeratorID     string    `json:"moderator_id"`
	ListingsFlagged int       `json:"listings_flagged"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserProfile struct {
	UserID      string        `json:"user_id"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	BanAudit    *UserBanAudit `json:"ban_audit,omitempty"`
}

type UserListingItem struct {
	ListingID string    `json:"listing_id"`
	Title     string    `json:"title"`
	PriceAsk  float64   `json:"price_ask"`
	Status    string    `json:"status"`
	RiskScore *float64  `json:"risk_score"`
	RiskTier  *string   `json:"risk_tier"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
}

type ListUserListingsResponse struct {
	Listings []UserListingItem `json:"listings"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}
