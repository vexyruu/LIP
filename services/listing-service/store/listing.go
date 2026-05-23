package store 

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/vexyruu/LIP/listing-service/model"
)

// CreateListing creates a new listing in the database
func CreateListing(ctx context.Context, pool *pgxpool.Pool, req *model.CreateListingRequest) (string, error) {
	var id string 
	err := pool.QueryRow(ctx,
		`INSERT INTO listings (user_id, title, description, price_ask, condition, category_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'DRAFT')
		 RETURNING id`,
		req.UserID,
		req.Title,
		req.Description,
		req.PriceAsk,
		req.Condition,
		req.CategoryID,
	).Scan(&id)
	
	if err != nil {
		return "", err
	}
	return id, nil

}

// GetListing retrieves a listing from the database
func GetListing(ctx context.Context, pool *pgxpool.Pool, listingID string) (*model.ListingStatus, error) {
	var l model.ListingStatus
	err := pool.QueryRow(ctx,
		`SELECT id, status, suggested_price, price_lower_bound, price_upper_bound, risk_tier, extracted_brand, extracted_product, extracted_size, policy_violation
		 FROM listings WHERE id = $1`,
		listingID,
	).Scan(&l.ListingID, &l.Status, &l.SuggestedPrice, &l.PriceLowerBound, &l.PriceUpperBound, &l.RiskTier, &l.ExtractedBrand, &l.ExtractedProduct, &l.ExtractedSize, &l.PolicyViolation)
	if err != nil {
		return nil, err
	}
	return &l, nil
}