package risk

import (
	"math"
	"testing"
)

func TestNormalizeComponentScore_LogScale(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		maxSize int
		want    float64
	}{
		{name: "zero size", size: 0, maxSize: 10, want: 0},
		{name: "singleton", size: 1, maxSize: 100, want: 0},
		{name: "singleton graph", size: 5, maxSize: 1, want: 0},
		{name: "linear would be 0.4", size: 4, maxSize: 10, want: math.Log(4) / math.Log(10)},
		{name: "max component", size: 10, maxSize: 10, want: 1},
		{name: "medium ring vs large graph", size: 50, maxSize: 1000, want: math.Log(50) / math.Log(1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeComponentScore(tt.size, tt.maxSize)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("NormalizeComponentScore(%d, %d) = %v, want %v", tt.size, tt.maxSize, got, tt.want)
			}
			if got < 0 || got > 1 {
				t.Fatalf("score out of bounds: %v", got)
			}
		})
	}
}

func TestNormalizeComponentScore_BeatsLinearOnMediumRings(t *testing.T) {
	size, maxSize := 50, 10_000
	logScore := NormalizeComponentScore(size, maxSize)
	linearScore := float64(size) / float64(maxSize)
	if logScore <= linearScore {
		t.Fatalf("expected log score (%v) > linear score (%v)", logScore, linearScore)
	}
}

func TestTotalScore_ClampedToUnitInterval(t *testing.T) {
	w := DefaultWeights()
	high := BatchScores{ComponentScore: 1, PageRankProximity: 1}
	got := TotalScore(high, 100, w)
	if got != 1.0 {
		t.Fatalf("expected clamp at 1.0, got %v", got)
	}

	zero := BatchScores{}
	if TotalScore(zero, 0, w) != 0 {
		t.Fatalf("expected zero score for empty batch")
	}
}

func TestTier_Boundaries(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0.81, "HIGH"},
		{0.80, "MEDIUM"},
		{0.41, "MEDIUM"},
		{0.40, "LOW"},
		{0.0, "LOW"},
	}
	for _, tt := range tests {
		if got := Tier(tt.score); got != tt.want {
			t.Fatalf("Tier(%v) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestEffectiveTotalScore_DirectBannedFloor(t *testing.T) {
	w := DefaultWeights()
	lowBatch := BatchScores{ComponentScore: 0.1, PageRankProximity: 0.1}

	withoutBan := EffectiveTotalScore(lowBatch, 0, w, false)
	if Tier(withoutBan) != "LOW" {
		t.Fatalf("expected LOW without banned link, got tier %s (score %v)", Tier(withoutBan), withoutBan)
	}

	withBan := EffectiveTotalScore(lowBatch, 0, w, true)
	if withBan < 0.9 {
		t.Fatalf("expected >= 0.9 with direct banned connection, got %v", withBan)
	}
	if Tier(withBan) != "HIGH" {
		t.Fatalf("expected HIGH tier after banned floor, got %s", Tier(withBan))
	}
}

func TestEffectiveTotalScore_DirectBannedDoesNotLowerHighScore(t *testing.T) {
	w := DefaultWeights()
	highBatch := BatchScores{ComponentScore: 0.5, PageRankProximity: 0.5}
	// Velocity burst already pushes composite to 1.0; banned floor must not reduce it.
	got := EffectiveTotalScore(highBatch, 3.0, w, true)
	if got != 1.0 {
		t.Fatalf("floor must not reduce an already-high score, got %v", got)
	}
}

func TestNormalizePageRank(t *testing.T) {
	if got := NormalizePageRank(0.5, 0); got != 0 {
		t.Fatalf("expected 0 when maxScore is 0, got %v", got)
	}
	if got := NormalizePageRank(0.75, 1.0); math.Abs(got-0.75) > 1e-9 {
		t.Fatalf("expected 0.75, got %v", got)
	}
	if got := NormalizePageRank(2.0, 1.0); got != 1.0 {
		t.Fatalf("expected clamp at 1.0, got %v", got)
	}
}

func TestDefaultVelocityBaseline(t *testing.T) {
	if got := DefaultVelocityBaseline(10); got != 1.0 {
		t.Fatalf("expected baseline 1.0 at count=10, got %v", got)
	}
	if got := DefaultVelocityBaseline(20); got != 2.0 {
		t.Fatalf("expected 2.0 at count=20, got %v", got)
	}
}

func TestBatchServeSplit_HighVelocityElevatesTier(t *testing.T) {
	// Batch path: moderate graph scores. Serve path: burst velocity pushes composite HIGH.
	w := DefaultWeights()
	batch := BatchScores{ComponentScore: 0.3, PageRankProximity: 0.3}
	velocity := DefaultVelocityBaseline(30) // 3.0
	score := TotalScore(batch, velocity, w)
	if Tier(score) != "HIGH" {
		t.Fatalf("expected HIGH from velocity burst, got %s (score %v)", Tier(score), score)
	}
}
