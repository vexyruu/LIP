package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
)

func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (*model.UserProfile, error) {
	var u model.UserProfile
	err := pool.QueryRow(ctx, `
		SELECT id, email, display_name, status, created_at 
		FROM users WHERE id = $1`,
		userID,
	).Scan(&u.UserID, &u.Email, &u.DisplayName, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	if u.Status == "BANNED" {
		audit, err := GetLatestUserBanAudit(ctx, pool, userID)
		if err != nil {
			return nil, err
		}
		u.BanAudit = audit
	}
	return &u, nil
}
