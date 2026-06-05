package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
)

func ListUserListings(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
	page, limit int,
) (*model.ListUserListingsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE user_id = $1`,
		userID,
	).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx,
		`SELECT id, title, price_ask, status, risk_score, risk_tier, images, created_at
		 FROM listings
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := []model.UserListingItem{}
	for rows.Next() {
		var item model.UserListingItem
		if err := rows.Scan(
			&item.ListingID,
			&item.Title,
			&item.PriceAsk,
			&item.Status,
			&item.RiskScore,
			&item.RiskTier,
			&item.Images,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if item.Images == nil {
			item.Images = []string{}
		}
		listings = append(listings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.ListUserListingsResponse{
		Listings: listings,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}
