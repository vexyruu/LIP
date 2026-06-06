package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vexyruu/LIP/shared/apitypes"
	"github.com/vexyruu/LIP/shared/outbox"
	"github.com/vexyruu/LIP/shared/pb"
	"github.com/vexyruu/LIP/shared/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const maxReanalysisAttempts = 5
const maxDeliveryAttempts = 5

// decideStatus decides the status of a listing based on the policy violation and risk tier
func decideStatus(policyViolation bool, riskTier string) string {
	if policyViolation {
		return "REJECTED"
	}
	if riskTier == "HIGH" {
		return "UNDER_REVIEW"
	}
	return "LIVE"
}

// lookupCategoryPath looks up the path of a category in the categories table
func lookupCategoryPath(ctx context.Context, pool *pgxpool.Pool, categoryID int) (string, error) {
	var path string
	err := pool.QueryRow(ctx, `SELECT path FROM categories WHERE id = $1`, categoryID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
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

	// connect to the postgres database
	pool, err := postgres.NewPool(ctx, postgresConn)
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	// connect to the ml-service
	mlConn, err := grpc.NewClient(mlServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to ml-service: %v", err)
	}
	defer mlConn.Close()
	mlClient := pb.NewMLServiceClient(mlConn)

	// connect to the pubsub subscription
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(pubsubSubscription)

	var dltTopic *pubsub.Topic
	if dltTopicID := os.Getenv("PUBSUB_TOPIC_DEADLETTER"); dltTopicID != "" {
		dltTopic = client.Topic(dltTopicID)
		defer dltTopic.Stop()
		log.Printf("dead-letter topic enabled: %s", dltTopicID)
	}

	deadLetter := func(ctx context.Context, data []byte, reason string) bool {
		if dltTopic == nil {
			log.Printf("no dead-letter topic configured, dropping message: %s", reason)
			return true
		}
		result := dltTopic.Publish(ctx, &pubsub.Message{
			Data:       data,
			Attributes: map[string]string{"dead_letter_reason": reason},
		})
		if _, err := result.Get(ctx); err != nil {
			log.Printf("failed to publish to dead-letter topic: %v", err)
			return false
		}
		log.Printf("message dead-lettered: %s", reason)
		return true
	}

	log.Println("analysis-worker listening on subscription:", pubsubSubscription)

	// receive messages from the pubsub subscription
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Received event: %s", msg.Data)
		retryOrDeadLetter := func(reason string) {
			if msg.DeliveryAttempt != nil && *msg.DeliveryAttempt >= maxDeliveryAttempts {
				if deadLetter(ctx, msg.Data, reason) {
					msg.Ack()
					return
				}
			}
			msg.Nack()
		}

		// unmarshal the listing created event and check for malformed payload
		var event apitypes.ListingCreatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("malformed payload: %v", err)
			if deadLetter(ctx, msg.Data, fmt.Sprintf("malformed payload: %v", err)) {
				msg.Ack()
			} else {
				msg.Nack()
			}
			return
		}
	
		// get the risk tier and score from the fraud-service
		fraudResp := apitypes.RiskResponse{RiskTier: "LOW", RiskScore: 0}
		if resp, err := http.Get(fmt.Sprintf("%s/v1/risk/%s", fraudServiceURL, event.UserID)); err != nil {
			log.Printf("fraud-service unavailable, degrading to LOW: %v", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("fraud-service returned %d, degrading to LOW", resp.StatusCode)
			} else if err := json.NewDecoder(resp.Body).Decode(&fraudResp); err != nil {
				log.Printf("failed to decode risk response, degrading to LOW: %v", err)
			}
		}

		// lookup the category path from the categories table
		categoryPath, err := lookupCategoryPath(ctx, pool, event.CategoryID)
		if err != nil {
			log.Printf("failed to lookup category path: %v", err)
			retryOrDeadLetter(fmt.Sprintf("category lookup failed: %v", err))
			return
		}

		// call the ml-service to analyze the listing
		mlCtx, mlCancel := context.WithTimeout(ctx, 30*time.Second)
		defer mlCancel()
		mlResp, mlErr := mlClient.AnalyzeListing(mlCtx, &pb.ListingRequest{
			ListingId:   event.ListingID,
			Title:       event.Title,
			Description: event.Description,
			Condition:   strconv.Itoa(int(event.Condition)),
			CategoryId:  categoryPath,
		})

		// set the listing status and extracted values
		var (
			suggestedPrice  *float64
			priceLower      *float64
			priceUpper      *float64
			brand           *string
			product         *string
			size            *string
			policyViolation *bool
			status          string
		)

		// if the ml-service call failed, set the listing status to UNDER_REVIEW
		if mlErr != nil {
			log.Printf("ml-service call failed, holding for review: %v", mlErr)
			status = "UNDER_REVIEW"
		} else {
			status = decideStatus(mlResp.GetPolicyViolation(), fraudResp.RiskTier)
			sp := mlResp.GetSuggestedPrice()
			pl := mlResp.GetPriceLowerBound()
			pu := mlResp.GetPriceUpperBound()
			b := mlResp.GetBrand()
			p := mlResp.GetProduct()
			sz := mlResp.GetSize()
			pv := mlResp.GetPolicyViolation()
			suggestedPrice = &sp
			priceLower = &pl
			priceUpper = &pu
			brand = &b
			product = &p
			size = &sz
			policyViolation = &pv
		}

		// update the listing in the database
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
			suggestedPrice,
			priceLower,
			priceUpper,
			brand,
			product,
			size,
			policyViolation,
			fraudResp.RiskScore,
			fraudResp.RiskTier,
			event.ListingID,
		)
		if err != nil {
			log.Printf("failed to update listing: %v", err)
			retryOrDeadLetter(fmt.Sprintf("listing update failed: %v", err))
			return
		}

		// Enqueue a fresh listing.created event via the outbox
		if mlErr != nil && event.Attempt < maxReanalysisAttempts {
			next := event
			next.Attempt = event.Attempt + 1
			if payload, mErr := json.Marshal(next); mErr != nil {
				log.Printf("failed to marshal re-analysis event for %s: %v", event.ListingID, mErr)
			} else if err := outbox.Enqueue(ctx, pool, event.ListingID, outbox.EventListingCreated, payload); err != nil {
				log.Printf("failed to enqueue re-analysis for %s: %v", event.ListingID, err)
			} else {
				log.Printf("scheduled re-analysis for %s (attempt %d)", event.ListingID, next.Attempt)
			}
		}

		log.Printf("listing %s -> %s (risk=%s ml_ok=%v)",
			event.ListingID, status, fraudResp.RiskTier, mlErr == nil)
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive messages: %v", err)
	}
}
