package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/pubsub"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/shared/pb"
	"github.com/vexyruu/LIP/shared/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ListingEvent struct {
	ListingID   string  `json:"listing_id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Condition   int16   `json:"condition"`
	CategoryID  int     `json:"category_id"`
	PriceAsk    float64 `json:"price_ask"`
}

type RiskResponse struct {
	UserID    string  `json:"user_id"`
	RiskScore float64 `json:"risk_score"`
	RiskTier  string  `json:"risk_tier"`
}

func decideStatus(policyViolation bool, riskTier string) string {
	if policyViolation {
		return "REJECTED"
	}
	if riskTier == "HIGH" {
		return "UNDER_REVIEW"
	}
	return "LIVE"
}

func lookupCategoryPath(ctx context.Context, pool *pgxpool.Pool, categoryID int) (string, error) {
	var path string
	err := pool.QueryRow(ctx, `SELECT path FROM categories WHERE id = $1`, categoryID).Scan(&path)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return path, err
}

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID is not set")
	}

	pubsubSubscription := os.Getenv("PUBSUB_SUBSCRIPTION")
	if pubsubSubscription == "" {
		log.Fatal("PUBSUB_SUBSCRIPTION is not set")
	}

	postgresConn := os.Getenv("POSTGRES_CONN")
	if postgresConn == "" {
		log.Fatal("POSTGRES_CONN is not set")
	}

	fraudServiceURL := os.Getenv("FRAUD_SERVICE_URL")
	if fraudServiceURL == "" {
		log.Fatal("FRAUD_SERVICE_URL is not set")
	}

	mlServiceAddr := os.Getenv("ML_SERVICE_ADDR")
	if mlServiceAddr == "" {
		log.Fatal("ML_SERVICE_ADDR is not set")
	}

	pool, err := postgres.NewPool(ctx, postgresConn)
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	mlConn, err := grpc.NewClient(mlServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to ml-service: %v", err)
	}
	defer mlConn.Close()
	mlClient := pb.NewMLServiceClient(mlConn)

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(pubsubSubscription)

	log.Println("analysis-worker listening on subscription:", pubsubSubscription)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Received event: %s", msg.Data)

		var event ListingEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal listing event: %v", err)
			msg.Nack()
			return
		}

		resp, err := http.Get(fmt.Sprintf("%s/v1/risk/%s", fraudServiceURL, event.UserID))
		if err != nil {
			log.Printf("Failed to get risk: %v", err)
			msg.Nack()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("fraud-service returned unexpected status: %d", resp.StatusCode)
			msg.Nack()
			return
		}

		var risk RiskResponse
		if err := json.NewDecoder(resp.Body).Decode(&risk); err != nil {
			log.Printf("Failed to decode risk response: %v", err)
			msg.Nack()
			return
		}

		categoryPath, err := lookupCategoryPath(ctx, pool, event.CategoryID)
		if err != nil {
			log.Printf("Failed to lookup category path: %v", err)
			msg.Nack()
			return
		}

		mlResp, err := mlClient.AnalyzeListing(ctx, &pb.ListingRequest{
			ListingId:   event.ListingID,
			Title:       event.Title,
			Description: event.Description,
			Condition:   strconv.Itoa(int(event.Condition)),
			CategoryId:  categoryPath,
		})
		if err != nil {
			log.Printf("ml-service call failed: %v", err)
			msg.Nack()
			return
		}

		status := decideStatus(mlResp.GetPolicyViolation(), risk.RiskTier)

		_, err = pool.Exec(ctx, `
			UPDATE listings SET
				status = $1,
				suggested_price = $2,
				price_lower_bound = $3,
				price_upper_bound = $4,
				extracted_brand = $5,
				extracted_product = $6,
				extracted_size = $7,
				policy_violation = $8,
				risk_score = $9,
				risk_tier = $10,
				updated_at = NOW()
			WHERE id = $11`,
			status,
			mlResp.GetSuggestedPrice(),
			mlResp.GetPriceLowerBound(),
			mlResp.GetPriceUpperBound(),
			mlResp.GetBrand(),
			mlResp.GetProduct(),
			mlResp.GetSize(),
			mlResp.GetPolicyViolation(),
			risk.RiskScore,
			risk.RiskTier,
			event.ListingID,
		)
		if err != nil {
			log.Printf("Failed to update listing: %v", err)
			msg.Nack()
			return
		}

		log.Printf("listing %s -> %s (brand=%q price=%.2f violation=%v risk=%s)",
			event.ListingID, status, mlResp.GetBrand(), mlResp.GetSuggestedPrice(),
			mlResp.GetPolicyViolation(), risk.RiskTier)
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive messages: %v", err)
	}
}
