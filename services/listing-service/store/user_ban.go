package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
)

type BanUserResult struct {
	ListingsFlagged int
}

// BanUserWithAudit bans the user, moves LIVE listings to UNDER_REVIEW, and records audit rows.
func BanUserWithAudit(ctx context.Context, pool *pgxpool.Pool, userID, moderatorID, reason string) (*BanUserResult, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE users
		SET status = 'BANNED', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status <> 'BANNED'`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		var status string
		err := tx.QueryRow(ctx, `SELECT status FROM users WHERE id = $1`, userID).Scan(&status)
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		if err != nil {
			return nil, err
		}
		if status == "BANNED" {
			return nil, ErrAlreadyBanned
		}
		return nil, ErrUserNotFound
	}

	rows, err := tx.Query(ctx, `
		UPDATE listings
		SET status = 'UNDER_REVIEW', updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND status = 'LIVE'
		RETURNING id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listingIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		listingIDs = append(listingIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, listingID := range listingIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO moderation_decisions (listing_id, moderator_id, action, reason)
			VALUES ($1, $2, 'BAN', $3)`,
			listingID, moderatorID, reason,
		)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_ban_audits (user_id, moderator_id, reason, listings_flagged)
		VALUES ($1, $2, $3, $4)`,
		userID, moderatorID, reason, len(listingIDs),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &BanUserResult{ListingsFlagged: len(listingIDs)}, nil
}

// GetLatestUserBanAudit returns the most recent ban audit row for a user, or nil if none.
func GetLatestUserBanAudit(ctx context.Context, pool *pgxpool.Pool, userID string) (*model.UserBanAudit, error) {
	var audit model.UserBanAudit
	err := pool.QueryRow(ctx, `
		SELECT reason, moderator_id, listings_flagged, created_at
		FROM user_ban_audits
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`,
		userID,
	).Scan(&audit.Reason, &audit.ModeratorID, &audit.ListingsFlagged, &audit.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &audit, nil
}
