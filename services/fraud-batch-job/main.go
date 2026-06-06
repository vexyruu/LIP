package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"github.com/vexyruu/LIP/shared/risk"
)

type userComponent struct {
	UserID      string
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

	// parse the batch interval
	var interval time.Duration
	if raw := os.Getenv("BATCH_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			log.Fatalf("invalid BATCH_INTERVAL %q: %v", raw, err)
		}
		interval = parsed
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	// create a new neo4j driver
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	log.Println("Connected to Neo4j")

	// create a new redis client
	redisDB := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisDB.Close()

	if err := redisDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	log.Println("Connected to Redis")

	// if the batch interval is 0, run a single cycle and exit
	if interval == 0 {
		if err := runBatchCycle(ctx, driver, redisDB); err != nil {
			log.Fatalf("batch job failed: %v", err)
		}
		return
	}

	// if the batch interval is not 0, run immediately, then on a fixed interval until signalled
	log.Printf("scheduler mode: recomputing fraud scores every %s", interval)
	if err := runBatchCycle(ctx, driver, redisDB); err != nil {
		log.Printf("batch cycle failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping scheduler")
			return
		case <-ticker.C:
			if err := runBatchCycle(ctx, driver, redisDB); err != nil {
				log.Printf("batch cycle failed: %v", err)
			}
		}
	}
}


// runBatchCycle executes one full WCC + PageRank pass and writes the resulting risk scores to Redis
func runBatchCycle(ctx context.Context, driver neo4j.DriverWithContext, redisDB *redis.Client) error {
	start := time.Now()

	components, err := runWCC(ctx, driver)
	if err != nil {
		return fmt.Errorf("WCC: %w", err)
	}
	log.Printf("WCC complete: found %d components", len(components))

	pagerankScores, err := runPageRank(ctx, driver)
	if err != nil {
		return fmt.Errorf("pagerank: %w", err)
	}
	log.Printf("PageRank complete: found %d scores", len(pagerankScores))

	if err := dropGraph(ctx, driver); err != nil {
		log.Printf("warning: failed to drop graph projection: %v", err)
	}

	if err := writeBatchScores(ctx, redisDB, components, pagerankScores); err != nil {
		return fmt.Errorf("write scores: %w", err)
	}

	log.Printf("batch cycle complete: %d users scored in %s", len(components), time.Since(start))
	return nil
}

// dropGraph drops the fraud-graph projection from the Neo4j database
func dropGraph(ctx context.Context, driver neo4j.DriverWithContext) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	session.Run(ctx, "CALL gds.graph.drop('fraud-graph', false) YIELD graphName", nil)
	return nil
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

	return components, nil
}

func writeBatchScores(ctx context.Context, redisDB *redis.Client, components []userComponent, pagerankScores map[string]float64) error {
	batches := buildBatchScores(components, pagerankScores)

	// Write the batches to Redis
	for userID, batch := range batches {
		payload, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("marshal batch for %s: %w", userID, err)
		}
		key := fmt.Sprintf("risk:%s", userID)
		if err := redisDB.Set(ctx, key, payload, time.Hour).Err(); err != nil {
			return fmt.Errorf("redis set %s: %w", key, err)
		}
	}
	return nil
}

// Run PageRank
func runPageRank(ctx context.Context, driver neo4j.DriverWithContext) (map[string]float64, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (b:User {banned: true})
		WITH collect(id(b)) AS sourceNodeIds
		WHERE size(sourceNodeIds) > 0
		CALL gds.pageRank.stream('fraud-graph', {
			sourceNodes: sourceNodeIds,
			maxIterations: 20,
			dampingFactor: 0.85
		})
		YIELD nodeId, score
		MATCH (u:User) WHERE id(u) = nodeId
		RETURN u.user_id AS user_id, score AS pagerank_score
	`, nil)
	if err != nil {
		return nil, fmt.Errorf("PageRank query failed: %w", err)
	}

	scores := make(map[string]float64)
	for result.Next(ctx) {
		record := result.Record()
		userID, _ := record.Get("user_id")
		prScore, _ := record.Get("pagerank_score")
		scores[userID.(string)] = prScore.(float64)
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return scores, nil
}

// Build the batch scores
func buildBatchScores(components []userComponent, pagerankScores map[string]float64) map[string]risk.BatchScores {
	sizes := make(map[int64]int)
	for _, c := range components {
		sizes[c.ComponentID]++
	}

	maxSize := 0
	for _, s := range sizes {
		if s > maxSize {
			maxSize = s
		}
	}

	maxPagerank := 0.0
	for _, s := range pagerankScores {
		if s > maxPagerank {
			maxPagerank = s
		}
	}

	// Build the output
	out := make(map[string]risk.BatchScores, len(components))
	for _, c := range components {
		out[c.UserID] = risk.BatchScores{
			ComponentScore:    risk.NormalizeComponentScore(sizes[c.ComponentID], maxSize),
			PageRankProximity: risk.NormalizePageRank(pagerankScores[c.UserID], maxPagerank),
		}
	}
	return out
}
