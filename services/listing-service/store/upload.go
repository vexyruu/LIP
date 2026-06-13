package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUploadNotFound = errors.New("upload not found")

type UploadRecord struct {
	ID               string
	UserID           string
	ObjectKey        string
	PublicURL        string
	ContentType      string
	SizeBytes        int64
	Status           string
	ExpiresAt        time.Time
	OriginalFilename string
}

func CreateUpload(ctx context.Context, pool *pgxpool.Pool, id string, rec *UploadRecord) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO uploads (id, user_id, object_key, public_url, content_type, size_bytes, status, expires_at, original_filename)
		 VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', $7, $8)`,
		id,
		rec.UserID,
		rec.ObjectKey,
		rec.PublicURL,
		rec.ContentType,
		rec.SizeBytes,
		rec.ExpiresAt,
		rec.OriginalFilename,
	)
	return err
}

func GetUpload(ctx context.Context, pool *pgxpool.Pool, uploadID string) (*UploadRecord, error) {
	var rec UploadRecord
	err := pool.QueryRow(ctx,
		`SELECT id, user_id, object_key, public_url, content_type, size_bytes, status, expires_at, original_filename
		 FROM uploads WHERE id = $1`,
		uploadID,
	).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.ObjectKey,
		&rec.PublicURL,
		&rec.ContentType,
		&rec.SizeBytes,
		&rec.Status,
		&rec.ExpiresAt,
		&rec.OriginalFilename,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUploadNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func MarkUploadReady(ctx context.Context, pool *pgxpool.Pool, uploadID string) error {
	tag, err := pool.Exec(ctx,
		`UPDATE uploads SET status = 'READY' WHERE id = $1 AND status = 'PENDING'`,
		uploadID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUploadNotFound
	}
	return nil
}

// returns object keys so the caller is responsible for deleting the backing S3 objects.
func PurgeExpiredUploads(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx,
		`DELETE FROM uploads
		 WHERE status = 'PENDING' AND expires_at < NOW()
		 RETURNING object_key`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func VerifyUploadReady(ctx context.Context, pool *pgxpool.Pool, userID, publicURL string) (bool, error) {
	var ok bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM uploads
			WHERE user_id = $1 AND public_url = $2 AND status = 'READY'
		)`,
		userID,
		publicURL,
	).Scan(&ok)
	return ok, err
}
