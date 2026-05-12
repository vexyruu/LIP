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
	"github.com/go-chi/chi/v5"
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
	pubsubProjectID := os.Getenv("PUBSUB_PROJECT_ID")
	if pubsubProjectID == "" {
		log.Fatal("PUBSUB_PROJECT_ID is not set")
	}
	pubsubTopic := os.Getenv("PUBSUB_TOPIC")
	if pubsubTopic == "" {
		log.Fatal("PUBSUB_TOPIC is not set")
	}
	// Context creation
	ctx := context.Background()

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

	r := chi.NewRouter()
	h := &handler.Handler{
		Pool: pool,
		Pub: pub,
		Topic: pubsubTopic,
	}

	r.Post("/v1/listings", h.CreateListing)
	r.Get("/v1/listings/{id}", h.GetListing)

	// Server creation
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Server startup
	go func(){
		log.Printf("Listing service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	
	}()

	// Signal Handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	// Server shutdown
	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
	log.Println("Server shutdown complete")
}


