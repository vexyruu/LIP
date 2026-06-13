package risk

import (
	"math"
	"os"
	"strconv"
)

type Weights struct {
	ComponentWeight float64
	PageRankWeight  float64
	VelocityWeight  float64
}

type BatchScores struct {
	ComponentScore    float64 `json:"component_score"`
	PageRankProximity float64 `json:"pagerank_proximity"`
}

func DefaultWeights() Weights {
	return Weights{ComponentWeight: 0.4, PageRankWeight: 0.4, VelocityWeight: 0.2}
}

func WeightsFromEnv() Weights {
	w := DefaultWeights()
	if v, err := strconv.ParseFloat(os.Getenv("WEIGHT_COMPONENT_SIZE"), 64); err == nil {
		w.ComponentWeight = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("WEIGHT_PAGERANK_PROXIMITY"), 64); err == nil {
		w.PageRankWeight = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("WEIGHT_VELOCITY"), 64); err == nil {
		w.VelocityWeight = v
	}
	return w
}

// TotalScore computes composite risk: weighted sum of component score, PageRank proximity, and velocity, clamped to [0, 1].
func TotalScore(b BatchScores, velocityScore float64, w Weights) float64 {
	s := w.ComponentWeight*b.ComponentScore + w.PageRankWeight*b.PageRankProximity + w.VelocityWeight*velocityScore
	return math.Min(1.0, math.Max(0.0, s))
}

func Tier(score float64) string {
	switch {
	case score > 0.8:
		return "HIGH"
	case score > 0.4:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// EffectiveTotalScore computes TotalScore and enforces a 0.9 floor when the user has a direct banned connection.
func EffectiveTotalScore(b BatchScores, velocityScore float64, w Weights, directBanned bool) float64 {
	s := TotalScore(b, velocityScore, w)
	if directBanned {
		s = math.Max(s, 0.9)
	}
	return s
}

// VelocityKey is the Redis key for a user's rolling listing-velocity counter
func VelocityKey(userID string) string {
	return "velocity:listings:" + userID
}

func VelocityScore(count int64, baseline float64) float64 {
	return float64(count) / baseline
}

func DefaultVelocityBaseline(count int64) float64 {
	return VelocityScore(count, 10)
}

func VelocityFlags(count int64) []string {
	flags := []string{}
	if count > 10 {
		flags = append(flags, "HIGH_LISTING_VELOCITY")
	}
	if count > 30 {
		flags = append(flags, "BURST_LISTING_VELOCITY")
	}
	return flags
}

func NormalizeComponentScore(size, maxSize int) float64 {
	if size <= 0 || maxSize <= 0 {
		return 0.0
	}
	// Singleton components carry no ring signal (log(1) = 0).
	if size <= 1 || maxSize <= 1 {
		return 0.0
	}
	// Blueprint: ComponentScore = log(component_size), normalized to [0, 1] by log(maxSize).
	return math.Min(1.0, math.Log(float64(size))/math.Log(float64(maxSize)))
}

func NormalizePageRank(score, maxScore float64) float64 {
	if maxScore == 0 {
		return 0.0
	}
	return math.Min(1.0, score/maxScore)
}
