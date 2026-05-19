package main

import(
	"os"
	"log"
	"context"
	"syscall"
	"os/signal"
	"net/http"
	"time"


	"github.com/redis/go-redis/v9"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vexyruu/LIP/fraud-service/handler"
)

func main() {
	// Environment variable reads
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is not set")
	}
	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		log.Fatal("NEO4J_URI is not set")
	}
	neo4jUser := os.Getenv("NEO4J_USER")
	if neo4jUser == "" {
		log.Fatal("NEO4J_USER is not set")
	}
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	if neo4jPassword == "" {
		log.Fatal("NEO4J_PASSWORD is not set")
	}

	// Signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	// Redis connection
	redisDB := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisDB.Close()

	// Neo4j connection
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	// Verify Neo4j connection
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	log.Println("Connected to Neo4j")

	// Handler creation
	h := &handler.Handler{
		Redis: redisDB,
		Neo4j: driver,
	}

	// Server creation + routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/risk/{user_id}", h.GetRisk)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Server startup
	go func(){
		log.Printf("Fraud service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	<-ctx.Done()
	stop()

	// Server shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
	log.Println("Server shutdown complete")
}