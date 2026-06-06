package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vexyruu/LIP/listing-service/handler"
	"github.com/vexyruu/LIP/shared/outbox"
	"github.com/vexyruu/LIP/shared/postgres"
	"github.com/vexyruu/LIP/shared/publisher"
	"github.com/vexyruu/LIP/shared/storage"
)

func main() {
	port := envOr("PORT", "8080")
	postgresConn := os.Getenv("POSTGRES_CONN")
	if postgresConn == "" {
		log.Fatal("POSTGRES_CONN is not set")
	}
	pubsubProjectID := os.Getenv("GCP_PROJECT_ID")
	if pubsubProjectID == "" {
		log.Fatal("GCP_PROJECT_ID is not set")
	}
	pubsubTopic := os.Getenv("PUBSUB_TOPIC_LISTING_CREATED")
	if pubsubTopic == "" {
		log.Fatal("PUBSUB_TOPIC_LISTING_CREATED is not set")
	}
	fraudServiceURL := os.Getenv("FRAUD_SERVICE_URL")
	if fraudServiceURL == "" {
		log.Fatal("FRAUD_SERVICE_URL is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	pool, err := postgres.NewPool(ctx, postgresConn)
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	pub, err := publisher.New(ctx, pubsubProjectID)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	var storageClient *storage.Client
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint != "" {
		useSSL, _ := strconv.ParseBool(envOr("S3_USE_SSL", "false"))
		storageClient, err = storage.New(ctx, storage.Config{
			Endpoint:      s3Endpoint,
			AccessKey:     envOr("S3_ACCESS_KEY", "minioadmin"),
			SecretKey:     envOr("S3_SECRET_KEY", "minioadmin"),
			Bucket:        envOr("S3_BUCKET", "mlip-listings"),
			Region:        envOr("S3_REGION", "us-east-1"),
			UseSSL:        useSSL,
			PublicBaseURL: envOr("UPLOAD_PUBLIC_BASE_URL", "http://localhost:9000/mlip-listings"),
		})
		if err != nil {
			log.Fatalf("Failed to create storage client: %v", err)
		}
		log.Println("Object storage enabled for uploads")
	}

	allowExternal, _ := strconv.ParseBool(envOr("ALLOW_EXTERNAL_IMAGE_URLS", "true"))

	h := &handler.Handler{
		Pool:                   pool,
		FraudServiceURL:        fraudServiceURL,
		Storage:                storageClient,
		UploadPublicBaseURL:    envOr("UPLOAD_PUBLIC_BASE_URL", "http://localhost:9000/mlip-listings"),
		AllowExternalImageURLs: allowExternal,
	}

	// Drain the transactional outbox to Pub/Sub in the background
	poller := &outbox.Poller{
		Pool:      pool,
		Pub:       pub,
		Topic:     pubsubTopic,
		Interval:  2 * time.Second,
		BatchSize: 50,
	}
	go poller.Run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/listings", h.CreateListing)
	mux.HandleFunc("GET /v1/listings/{id}", h.GetListing)
	mux.HandleFunc("GET /v1/listings", h.ListListings)
	mux.HandleFunc("PATCH /v1/listings/{id}", h.PatchListing)
	mux.HandleFunc("POST /v1/listings/{id}/assign", h.AssignListing)
	mux.HandleFunc("DELETE /v1/listings/{id}/assign", h.UnassignListing)
	mux.HandleFunc("GET /v1/analytics/summary", h.GetAnalytics)
	mux.HandleFunc("GET /v1/moderation/decisions", h.ListModerationDecisions)
	mux.HandleFunc("GET /v1/users/{id}", h.GetUser)
	mux.HandleFunc("GET /v1/users/{id}/listings", h.GetUserListings)
	mux.HandleFunc("POST /v1/users/{id}/ban", h.BanUser)
	mux.HandleFunc("POST /v1/uploads", h.CreateUpload)
	mux.HandleFunc("POST /v1/uploads/{id}/complete", h.CompleteUpload)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("Listing service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
	log.Println("Server shutdown complete")
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
