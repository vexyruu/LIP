package main

import(
	"os"
	"log"
	"context"
	"encoding/json"
	"net/http"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/vexyruu/LIP/shared/postgres"
)

type ListingEvent struct {
	ListingID string `json:"listing_id"`
	UserID string `json:"user_id"`
}

type RiskResponse struct {
	UserID string `json:"user_id"`
	RiskScore float64 `json:"risk_score"`
	RiskTier string `json:"risk_tier"`
	
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

	pool, err := postgres.NewPool(ctx, postgresConn)
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	// Create pubsub client
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	// Create subscription
	sub := client.Subscription(pubsubSubscription)

	// Start listening on subscription
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Received listing_id: %s", msg.Data)
		event := ListingEvent{}
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
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
			log.Printf("fraud-service returned unexpected status: %v", resp.StatusCode)
			msg.Nack()
			return
		}
		var risk RiskResponse
		err = json.NewDecoder(resp.Body).Decode(&risk)
		if err != nil {
			log.Printf("Failed to decode risk response: %v", err)
			msg.Nack()
			return
		}
		if risk.RiskTier == "HIGH" {
			_, err = pool.Exec(ctx, `UPDATE listings SET status = 'UNDER_REVIEW' WHERE id = $1`, event.ListingID)
			if err != nil {
				log.Printf("Failed to update listing status: %v", err)
				msg.Nack()
				return
			}
		}
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive message: %v", err)
	}
}