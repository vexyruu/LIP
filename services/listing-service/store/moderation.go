package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
)

func ListModerationDecisions(ctx context.Context, pool *pgxpool.Pool, f model.ListModerationDecisionsFilter) (*model.ListModerationDecisionsResponse, error) {
	page := f.Page
	limit := f.Limit
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

	var actionFilter *string
	if f.Action != "" {
		actionFilter = &f.Action
	}

	where := `WHERE ($1::text IS NULL OR md.action::text = $1)`

	var total int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM moderation_decisions md `+where,
		actionFilter,
	).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx,
		`SELECT md.id, md.created_at, md.listing_id, l.title,
		        md.action::text, l.risk_tier, md.reason, md.moderator_id
		 FROM moderation_decisions md
		 JOIN listings l ON l.id = md.listing_id
		 `+where+`
		 ORDER BY md.created_at DESC
		 LIMIT $2 OFFSET $3`,
		actionFilter, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decisions := []model.ModerationDecisionRow{}
	for rows.Next() {
		var row model.ModerationDecisionRow
		var ts time.Time
		if err := rows.Scan(
			&row.ID,
			&ts,
			&row.ListingID,
			&row.Title,
			&row.Action,
			&row.RiskTier,
			&row.Reason,
			&row.ModeratorID,
		); err != nil {
			return nil, err
		}
		row.Timestamp = ts.UTC().Format(time.RFC3339)
		decisions = append(decisions, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.ListModerationDecisionsResponse{
		Decisions: decisions,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}
