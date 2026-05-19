package main

import (
	"os"
	"log"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
)

type userComponent struct {
	UserID string
	ComponentID int64
}

func main() {
	// Environment variable reads
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
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is not set")
	}

	// Context creation
	ctx := context.Background()

	// Neo4j connection
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	log.Println("Connected to Neo4j")

	// Redis connection
	redisDB := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisDB.Close()

	if err := redisDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Run WCC
	components, err := runWCC(ctx, driver)
	if err != nil {
		log.Fatalf("Failed to run WCC: %v", err)
	}
	log.Printf("WCC Complete: Found %d components", len(components))

	// Compute scores
	if err := computeScores(ctx, redisDB, components); err != nil {
		log.Fatalf("Failed to compute scores: %v", err)
	}
	log.Printf("Batch job complete: %d users scored", len(components))


}

func runWCC(ctx context.Context, driver neo4j.DriverWithContext) ([]userComponent, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Drop the graph if it exists
	session.Run(ctx, "CALL gds.graph.drop('fraud-graph', false) YIELD graphName", nil)

	// Project the graph
	_, err := session.Run(ctx,
		"CALL gds.graph.project('fraud-graph', ['User'], ['USED_DEVICE', 'SHARED_IP', 'USED_PAYMENT'])",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("graph projection failed: %w", err)
	}

	// Run the WCC query
	result, err := session.Run(ctx,
		`CALL gds.wcc.stream('fraud-graph')
		 YIELD nodeId, componentId
		 MATCH (u:User) WHERE id(u) = nodeId
		 RETURN u.user_id AS user_id, componentId`,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("WCC query failed: %w", err)
	}

	// Process the results
	var components []userComponent

	// Iterate through the results
	for result.Next(ctx) {
		record := result.Record()
		userIDRaw, _ := record.Get("user_id")
		componentIDRaw, _ := record.Get("componentId")
		userID := userIDRaw.(string)
		componentID := componentIDRaw.(int64)
		components = append(components, userComponent{UserID: userID, ComponentID: componentID})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("WCC query failed: %w", err)
	}

	session.Run(ctx, "CALL gds.graph.drop('fraud-graph', false) YIELD graphName", nil)
	return components, nil
}

func computeScores(ctx context.Context, redisDB *redis.Client, components []userComponent) error {
	componentSizes := make(map[int64]int)
	
	// Count the size of each component
	for _, c := range components {
		componentSizes[c.ComponentID]++
	}
	
	// Set the risk score for each user (iterates through components again)
	for _, c := range components {
		key := fmt.Sprintf("risk:%s", c.UserID)
		score := math.Log(float64(componentSizes[c.ComponentID]))
		err := redisDB.Set(ctx, key, score, time.Hour).Err()
		if err != nil {
			return fmt.Errorf("failed to set risk score: %w", err)
		}
	}
	return nil
}