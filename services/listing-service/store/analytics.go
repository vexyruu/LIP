package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/listing-service/model"
)

func GetAnalyticsSummary(ctx context.Context, pool *pgxpool.Pool) (*model.AnalyticsSummary, error) {
	// Initialize slices and maps so an empty DB serializes to [] / {} rather
	// than JSON null, which the dashboard treats as arrays/objects.
	summary := &model.AnalyticsSummary{
		ByStatus:         map[string]int{},
		ByRiskTier:       map[string]int{},
		DailySubmissions: []model.DailySubmissionCount{},
		RecentDecisions:  []model.RecentModerationRow{},
	}

	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM listings`).Scan(&summary.TotalListings); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `SELECT status::text, COUNT(*) FROM listings GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		summary.ByStatus[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	summary.UnderReview = summary.ByStatus["UNDER_REVIEW"]

	tierRows, err := pool.Query(ctx,
		`SELECT COALESCE(risk_tier, 'UNSCORED'), COUNT(*)
		 FROM listings
		 WHERE status != 'DRAFT'
		 GROUP BY risk_tier`,
	)
	if err != nil {
		return nil, err
	}
	defer tierRows.Close()
	for tierRows.Next() {
		var tier string
		var count int
		if err := tierRows.Scan(&tier, &count); err != nil {
			return nil, err
		}
		summary.ByRiskTier[tier] = count
	}
	if err := tierRows.Err(); err != nil {
		return nil, err
	}

	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE policy_violation = true AND status = 'REJECTED'`,
	).Scan(&summary.PolicyRejections); err != nil {
		return nil, err
	}

	var avgSeconds *float64
	if err := pool.QueryRow(ctx,
		`SELECT AVG(EXTRACT(EPOCH FROM (updated_at - created_at)))
		 FROM listings
		 WHERE status IN ('LIVE', 'UNDER_REVIEW', 'REJECTED')`,
	).Scan(&avgSeconds); err != nil {
		return nil, err
	}
	summary.AvgPipelineSeconds = avgSeconds

	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM listings WHERE created_at >= NOW() - INTERVAL '24 hours'`,
	).Scan(&summary.Last24hSubmissions); err != nil {
		return nil, err
	}

	dailyRows, err := pool.Query(ctx,
		`SELECT DATE(created_at) AS d,
		        COUNT(*) AS total,
		        COUNT(*) FILTER (WHERE risk_tier = 'HIGH') AS high_risk
		 FROM listings
		 WHERE created_at >= NOW() - INTERVAL '7 days'
		 GROUP BY DATE(created_at)
		 ORDER BY d`,
	)
	if err != nil {
		return nil, err
	}
	defer dailyRows.Close()
	for dailyRows.Next() {
		var day time.Time
		var total, high int
		if err := dailyRows.Scan(&day, &total, &high); err != nil {
			return nil, err
		}
		summary.DailySubmissions = append(summary.DailySubmissions, model.DailySubmissionCount{
			Date:  day.Format("2006-01-02"),
			Total: total,
			High:  high,
		})
	}
	if err := dailyRows.Err(); err != nil {
		return nil, err
	}

	decisionRows, err := pool.Query(ctx,
		`SELECT md.created_at, md.listing_id, l.title, md.action::text, l.risk_tier, md.reason
		 FROM moderation_decisions md
		 JOIN listings l ON l.id = md.listing_id
		 ORDER BY md.created_at DESC
		 LIMIT 10`,
	)
	if err != nil {
		return nil, err
	}
	defer decisionRows.Close()
	for decisionRows.Next() {
		var row model.RecentModerationRow
		var ts time.Time
		if err := decisionRows.Scan(&ts, &row.ListingID, &row.Title, &row.Action, &row.RiskTier, &row.Reason); err != nil {
			return nil, err
		}
		row.Timestamp = ts.UTC().Format(time.RFC3339)
		summary.RecentDecisions = append(summary.RecentDecisions, row)
	}
	if err := decisionRows.Err(); err != nil {
		return nil, err
	}

	return summary, nil
}
