package main

import (
	"net/http"
	"log"
	"os"
	"context"
	"syscall"
	"time"
	"os/signal"

	"github.com/vexyruu/LIP/shared/postgres"
	"github.com/vexyruu/LIP/shared/publisher"
	"github.com/vexyruu/LIP/listing-service/handler"
)

func main() {
	// Environment variable reads
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
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

	// Signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	// Postgres pool creation
	pool, err := postgres.NewPool(ctx, postgresConn)
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pool.Close()

	// Publisher creation
	pub, err := publisher.New(ctx, pubsubProjectID)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	h := &handler.Handler{
		Pool: pool,
		Pub: pub,
		Topic: pubsubTopic,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/listings", h.CreateListing)
	mux.HandleFunc("GET /v1/listings/{id}", h.GetListing)

	// Server creation
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Server startup
	go func(){
		log.Printf("Listing service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	
	}()

	// Signal Handling
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)	
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
	log.Println("Server shutdown complete")

}


