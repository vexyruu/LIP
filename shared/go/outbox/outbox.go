// Package outbox implements the transactional-outbox pattern over the
// pending_events table. Writers enqueue an event in the SAME database
// transaction that mutates business state, guaranteeing the event is durably
// recorded iff the state change commits. A Poller then publishes pending
// events to Pub/Sub and marks them PUBLISHED, giving at-least-once delivery
// without the dual-write race of "insert row, then publish".
package outbox

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusPending   = "PENDING"
	StatusPublished = "PUBLISHED"
	StatusFailed    = "FAILED"

	// MaxRetries caps publish attempts before an event is parked as FAILED.
	MaxRetries = 5
)

// EventListingCreated is published when a listing needs (re)analysis.
const EventListingCreated = "listing.created"

// Execer is satisfied by both *pgxpool.Pool and pgx.Tx, so Enqueue can run
// inside a caller's transaction or standalone.
type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Enqueue writes a PENDING outbox row. Call it inside the same transaction as
// the state change it describes.
func Enqueue(ctx context.Context, q Execer, listingID, eventType string, payload []byte) error {
	_, err := q.Exec(ctx,
		`INSERT INTO pending_events (listing_id, event_type, payload, status)
		 VALUES ($1, $2, $3, 'PENDING')`,
		listingID, eventType, payload,
	)
	return err
}

// Publisher is the subset of shared/publisher used by the Poller.
type Publisher interface {
	Publish(ctx context.Context, topicID string, data []byte) error
}

// Poller drains pending_events to Pub/Sub on an interval.
type Poller struct {
	Pool      *pgxpool.Pool
	Pub       Publisher
	Topic     string
	Interval  time.Duration
	BatchSize int
}

// Run blocks until ctx is cancelled, draining the outbox every Interval.
func (p *Poller) Run(ctx context.Context) {
	if p.Interval <= 0 {
		p.Interval = 2 * time.Second
	}
	if p.BatchSize <= 0 {
		p.BatchSize = 50
	}

	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	log.Printf("outbox poller started (interval=%s batch=%d topic=%s)", p.Interval, p.BatchSize, p.Topic)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				n, err := p.processBatch(ctx)
				if err != nil {
					log.Printf("outbox: batch error: %v", err)
					break
				}
				if n < p.BatchSize {
					break
				}
			}
		}
	}
}

// processBatch claims up to BatchSize pending rows with FOR UPDATE SKIP LOCKED
// (so multiple instances don't double-publish), publishes them, and records
// the outcome in the same transaction. Publishing inside the transaction keeps
// claim+publish+ack atomic; the row lock is held only for the publish call.
func (p *Poller) processBatch(ctx context.Context) (int, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT id, payload, retry_count
		 FROM pending_events
		 WHERE status = 'PENDING'
		   AND retry_count < $1
		   AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		 ORDER BY created_at
		 LIMIT $2
		 FOR UPDATE SKIP LOCKED`,
		MaxRetries, p.BatchSize,
	)
	if err != nil {
		return 0, err
	}

	type pending struct {
		id      string
		payload []byte
		retry   int
	}
	var batch []pending
	for rows.Next() {
		var pe pending
		if err := rows.Scan(&pe.id, &pe.payload, &pe.retry); err != nil {
			rows.Close()
			return 0, err
		}
		batch = append(batch, pe)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(batch) == 0 {
		return 0, tx.Commit(ctx)
	}

	for _, pe := range batch {
		if err := p.Pub.Publish(ctx, p.Topic, pe.payload); err != nil {
			nextRetry := pe.retry + 1
			status := StatusPending
			if nextRetry >= MaxRetries {
				status = StatusFailed
				log.Printf("outbox: event %s parked as FAILED after %d attempts: %v", pe.id, nextRetry, err)
			} else {
				log.Printf("outbox: publish failed for event %s (attempt %d): %v", pe.id, nextRetry, err)
			}
			// Schedule the next attempt with exponential backoff: 10s * 2^retry (10s, 20s, 40s, 80s, 160s).
			if _, e := tx.Exec(ctx,
				`UPDATE pending_events
				 SET retry_count = $2, status = $3,
				     next_retry_at = NOW() + (INTERVAL '10 seconds' * power(2, $4))
				 WHERE id = $1`,
				pe.id, nextRetry, status, pe.retry,
			); e != nil {
				return 0, e
			}
			continue
		}
		if _, e := tx.Exec(ctx,
			`UPDATE pending_events SET status = 'PUBLISHED', published_at = NOW() WHERE id = $1`,
			pe.id,
		); e != nil {
			return 0, e
		}
	}

	return len(batch), tx.Commit(ctx)
}
