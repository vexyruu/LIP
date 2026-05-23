package model

type CreateListingRequest struct {
	UserID string `json:"user_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Condition int16 `json:"condition"`
	CategoryID int `json:"category_id"`
	PriceAsk float64 `json:"price_ask"`
}

type CreateListingResponse struct {
	ListingID string `json:"listing_id"`
	Status string `json:"status"`
	ETA int `json:"eta_ms"`
}

type ListingStatus struct {
    ListingID        string   `json:"listing_id"`
    Status           string   `json:"status"`
    SuggestedPrice   *float64 `json:"suggested_price"`
    PriceLowerBound  *float64 `json:"price_lower_bound"`
    PriceUpperBound  *float64 `json:"price_upper_bound"`
    RiskTier         *string  `json:"risk_tier"`
    ExtractedBrand   *string  `json:"extracted_brand"`
    ExtractedProduct *string  `json:"extracted_product"`
    ExtractedSize    *string  `json:"extracted_size"`
	PolicyViolation  *bool    `json:"policy_violation"`
}

