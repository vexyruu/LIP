package handler

import(
	"testing"
	"reflect"
)


func TestRiskTier(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"clear high",      0.9,  "HIGH"},
		{"clear medium",    0.6,  "MEDIUM"},
		{"clear low",       0.2,  "LOW"},
		{"boundary high",   0.80, "MEDIUM"},
		{"boundary medium", 0.40, "LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := riskTier(tt.score)
			if resp != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resp)
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
		{"clear 8",	8,	[]string{}},
		{"clear 10",	10,	[]string{}},
		{"clear 11",	11,	[]string{"HIGH_LISTING_VELOCITY"}},
		{"clear 30",	30,	[]string{"HIGH_LISTING_VELOCITY"}},
		{"clear 31",	31,	[]string{"HIGH_LISTING_VELOCITY", "BURST_LISTING_VELOCITY"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := velocityFlags(tt.count)
			if !reflect.DeepEqual(resp, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, resp)
			}
		})
	}
}
