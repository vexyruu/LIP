package handler

import (
	"reflect"
	"testing"

	"github.com/vexyruu/LIP/shared/risk"
)

func TestRiskTier(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"clear high", 0.9, "HIGH"},
		{"clear medium", 0.6, "MEDIUM"},
		{"clear low", 0.2, "LOW"},
		{"boundary high", 0.80, "MEDIUM"},
		{"boundary medium", 0.40, "LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := risk.Tier(tt.score); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestVelocityFlags(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected []string
	}{
		{"no flags at 8", 8, []string{}},
		{"no flags at 10", 10, []string{}},
		{"HIGH at 11", 11, []string{"HIGH_LISTING_VELOCITY"}},
		{"HIGH at 30", 30, []string{"HIGH_LISTING_VELOCITY"}},
		{"HIGH+BURST at 31", 31, []string{"HIGH_LISTING_VELOCITY", "BURST_LISTING_VELOCITY"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := risk.VelocityFlags(tt.count); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestEffectiveScoreWithVelocity(t *testing.T) {
	w := risk.DefaultWeights()
	batch := risk.BatchScores{ComponentScore: 0.5, PageRankProximity: 0.75}
	total := risk.TotalScore(batch, 3.0, w)
	if risk.Tier(total) != "HIGH" {
		t.Fatalf("expected HIGH, got %s", risk.Tier(total))
	}
}

func TestEffectiveScore_DirectBannedLink(t *testing.T) {
	w := risk.DefaultWeights()
	batch := risk.BatchScores{ComponentScore: 0.1, PageRankProximity: 0.1}
	score := risk.EffectiveTotalScore(batch, 0, w, true)
	if score < 0.9 {
		t.Fatalf("direct banned connection should floor score at 0.9, got %v", score)
	}
	if risk.Tier(score) != "HIGH" {
		t.Fatalf("expected HIGH tier, got %s", risk.Tier(score))
	}
}
