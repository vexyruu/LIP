package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
	"github.com/vexyruu/LIP/shared/apitypes"
	"github.com/vexyruu/LIP/shared/outbox"
)

func CreateListing(ctx context.Context, pool *pgxpool.Pool, req *model.CreateListingRequest) (string, error) {
	images := req.Images
	if images == nil {
		images = []string{}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx,
		`INSERT INTO listings (user_id, title, description, price_ask, condition, category_id, status, images)
		 VALUES ($1, $2, $3, $4, $5, $6, 'DRAFT', $7)
		 RETURNING id`,
		req.UserID,
		req.Title,
		req.Description,
		req.PriceAsk,
		req.Condition,
		req.CategoryID,
		images,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(apitypes.ListingCreatedEvent{
		ListingID:   id,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Condition:   req.Condition,
		CategoryID:  req.CategoryID,
		PriceAsk:    req.PriceAsk,
	})
	if err != nil {
		return "", err
	}
	if err := outbox.Enqueue(ctx, tx, id, outbox.EventListingCreated, payload); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}

// GetListing retrieves a listing from the database
func GetListing(ctx context.Context, pool *pgxpool.Pool, listingID string) (*model.ListingStatus, error) {
	var l model.ListingStatus
	err := pool.QueryRow(ctx,
		`SELECT id, user_id, title, description, price_ask, condition, category_id, status, suggested_price, price_lower_bound, price_upper_bound,
		 risk_tier, risk_score, extracted_brand, extracted_product, extracted_size, policy_violation, images, assigned_to, assigned_at, created_at, updated_at
		 FROM listings WHERE id = $1`,
		listingID,
	).Scan(&l.ListingID, &l.UserID, &l.Title, &l.Description, &l.PriceAsk, &l.Condition, &l.CategoryID, &l.Status, &l.SuggestedPrice, &l.PriceLowerBound, &l.PriceUpperBound, &l.RiskTier, &l.RiskScore, &l.ExtractedBrand, &l.ExtractedProduct, &l.ExtractedSize, &l.PolicyViolation, &l.Images, &l.AssignedTo, &l.AssignedAt, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if l.Images == nil {
		l.Images = []string{}
	}
	l.APIStatus = model.APIStatus(l.Status)
	return &l, nil
}

func ListListings(ctx context.Context, pool *pgxpool.Pool, f model.ListListingsFilter) (*model.ListListingsResponse, error) {
	page := f.Page
	limit := f.Limit

	// Default values
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Build ORDER BY clause
	orderBy := "risk_score DESC NULLS LAST"
	switch f.Sort {
	case "risk_score":
		orderBy = "risk_score DESC NULLS LAST"
	case "created_at":
		orderBy = "created_at ASC"
	}

	var tier *string
	if f.Tier != "" {
		tier = &f.Tier
	}

	// Build WHERE clause
	where := `WHERE status = $1 AND ($2::text IS NULL OR risk_tier = $2)`

	// Count total listings
	var total int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM listings `+where, f.Status, tier).Scan(&total); err != nil {
		return nil, err
	}

	// Build query
	query := `SELECT id, title, user_id, price_ask, risk_score, risk_tier, category_id, images, assigned_to, created_at FROM listings ` +
		where + ` ORDER BY ` + orderBy + ` LIMIT $3 OFFSET $4`

	// Query listings
	rows, err := pool.Query(ctx, query, f.Status, tier, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := []model.ListingQueueItem{}
	for rows.Next() {
		var l model.ListingQueueItem
		if err := rows.Scan(&l.ListingID, &l.Title, &l.UserID, &l.PriceAsk, &l.RiskScore, &l.RiskTier, &l.CategoryID, &l.Images, &l.AssignedTo, &l.CreatedAt); err != nil {
			return nil, err
		}
		if l.Images == nil {
			l.Images = []string{}
		}
		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.ListListingsResponse{
		Listings: listings,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil

}

// ModerateListing updates listing status and records a moderation decision.
func ModerateListing(ctx context.Context, pool *pgxpool.Pool, listingID string, req *model.PatchListingRequest) (*model.PatchListingResponse, error) {
	var newStatus, dbAction, reason string

	// Validate action
	switch req.Action {
	case "APPROVE":
		newStatus = "LIVE"
		dbAction = "APPROVE"
		reason = req.Reason
		if reason == "" {
			reason = "Approved by moderator"
		}
	case "REJECT":
		newStatus = "REJECTED"
		dbAction = "REJECT"
		reason = req.Reason
	}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx,
		`SELECT status FROM listings WHERE id = $1 FOR UPDATE`,
		listingID,
	).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	if currentStatus != "UNDER_REVIEW" {
		return nil, ErrNotUnderReview
	}

	// Update listing status
	_, err = tx.Exec(ctx, `UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2`, newStatus, listingID)
	if err != nil {
		return nil, err
	}
	// Insert moderation decision
	_, err = tx.Exec(ctx, `INSERT INTO moderation_decisions (listing_id, moderator_id, action, reason) VALUES ($1, $2, $3, $4)`, listingID, req.ModeratorID, dbAction, reason)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.PatchListingResponse{
		ListingID: listingID,
		Status:    newStatus,
	}, nil
}

// assign a listing for a moderator using optimistic concurrency
func AssignListing(ctx context.Context, pool *pgxpool.Pool, listingID, moderatorID string) (*model.AssignListingResponse, error) {
	var assigned string
	err := pool.QueryRow(ctx,
		`UPDATE listings SET assigned_to = $1, assigned_at = NOW()
		 WHERE id = $2 AND status = 'UNDER_REVIEW' AND (assigned_to IS NULL OR assigned_to = $1)
		 RETURNING assigned_to`,
		moderatorID, listingID,
	).Scan(&assigned)
	if err == nil {
		return &model.AssignListingResponse{ListingID: listingID, AssignedTo: &assigned}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var status string
	var holder *string
	if qErr := pool.QueryRow(ctx,
		`SELECT status, assigned_to FROM listings WHERE id = $1`,
		listingID,
	).Scan(&status, &holder); qErr != nil {
		return nil, qErr
	}
	if status != "UNDER_REVIEW" {
		return nil, ErrNotUnderReview
	}
	return nil, ErrAlreadyAssigned
}

// release a claim, but only if the caller currently holds it
func UnassignListing(ctx context.Context, pool *pgxpool.Pool, listingID, moderatorID string) (*model.AssignListingResponse, error) {
	var id string
	err := pool.QueryRow(ctx,
		`UPDATE listings SET assigned_to = NULL, assigned_at = NULL
		 WHERE id = $1 AND assigned_to = $2
		 RETURNING id`,
		listingID, moderatorID,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotAssignedToYou
		}
		return nil, err
	}
	return &model.AssignListingResponse{ListingID: listingID, AssignedTo: nil}, nil
}
