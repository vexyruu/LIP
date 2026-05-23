package main

import "testing"

func TestDecideStatus(t *testing.T) {
	tests := []struct {
		policyViolation bool
		riskTier        string
		expected        string
	}{
		{true, "HIGH", "REJECTED"},
		{false, "HIGH", "UNDER_REVIEW"},
		{false, "LOW", "LIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := decideStatus(tt.policyViolation, tt.riskTier)
			if got != tt.expected {
				t.Errorf("decideStatus(%v, %q) = %q, want %q",
					tt.policyViolation, tt.riskTier, got, tt.expected)
			}
		})
	}
}