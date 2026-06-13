package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrModeratorNotFound = errors.New("moderator not found")
var ErrModeratorExists = errors.New("moderator already exists")

type ModeratorRecord struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Role         string `json:"role"`
	PasswordHash string `json:"password_hash"`
}

func ListModerators(ctx context.Context, pool *pgxpool.Pool) ([]ModeratorRecord, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, email, display_name, role, password_hash
		 FROM moderators ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []ModeratorRecord
	for rows.Next() {
		var m ModeratorRecord
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.PasswordHash); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	if mods == nil {
		mods = []ModeratorRecord{}
	}
	return mods, rows.Err()
}

func GetModeratorByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*ModeratorRecord, error) {
	var m ModeratorRecord
	err := pool.QueryRow(ctx,
		`SELECT id, email, display_name, role, password_hash
		 FROM moderators WHERE email = $1`,
		email,
	).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrModeratorNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func CreateModerator(ctx context.Context, pool *pgxpool.Pool, m *ModeratorRecord) (*ModeratorRecord, error) {
	var created ModeratorRecord
	err := pool.QueryRow(ctx,
		`INSERT INTO moderators (email, display_name, role, password_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, display_name, role, password_hash`,
		m.Email, m.DisplayName, m.Role, m.PasswordHash,
	).Scan(&created.ID, &created.Email, &created.DisplayName, &created.Role, &created.PasswordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrModeratorExists
		}
		return nil, err
	}
	return &created, nil
}

func DeleteModerator(ctx context.Context, pool *pgxpool.Pool, id string) error {
	tag, err := pool.Exec(ctx, `DELETE FROM moderators WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrModeratorNotFound
	}
	return nil
}
