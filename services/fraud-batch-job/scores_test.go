package main

import (
	"math"
	"testing"

	"github.com/vexyruu/LIP/shared/risk"
)

func TestWriteBatchScores(t *testing.T) {
	components := []userComponent{
		{UserID: "user-1", ComponentID: 1},
		{UserID: "user-2", ComponentID: 2},
		{UserID: "user-3", ComponentID: 1},
	}
	pagerankScores := map[string]float64{
		"user-1": 0.5,
		"user-2": 0.75,
		"user-3": 0.25,
	}

	batches := buildBatchScores(components, pagerankScores)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batch scores, got %d", len(batches))
	}
	for userID, b := range batches {
		if b.ComponentScore < 0 || b.ComponentScore > 1 {
			t.Errorf("user %s: component score out of bounds: %v", userID, b.ComponentScore)
		}
		if b.PageRankProximity < 0 || b.PageRankProximity > 1 {
			t.Errorf("user %s: pagerank proximity out of bounds: %v", userID, b.PageRankProximity)
		}
	}
}

func TestNormalizeIntegration(t *testing.T) {
	comp := risk.NormalizeComponentScore(4, 10)
	if math.Abs(comp-math.Log(4)/math.Log(10)) > 1e-9 {
		t.Fatalf("expected log-normalized component score, got %v", comp)
	}
	pr := risk.NormalizePageRank(0.5, 1.0)
	if pr < 0.0 || pr > 1.0 {
		t.Fatalf("pagerank proximity out of bounds: %v", pr)
	}
}
