package main

import "testing"

func TestDecideStatus(t *testing.T) {
	tests := []struct {
		name            string
		policyViolation bool
		riskTier        string
		expected        string
	}{
		{name: "policy rejects regardless of risk", policyViolation: true, riskTier: "HIGH", expected: "REJECTED"},
		{name: "policy rejects low risk", policyViolation: true, riskTier: "LOW", expected: "REJECTED"},
		{name: "high seller risk queues", policyViolation: false, riskTier: "HIGH", expected: "UNDER_REVIEW"},
		{name: "medium seller risk goes live", policyViolation: false, riskTier: "MEDIUM", expected: "LIVE"},
		{name: "low seller risk goes live", policyViolation: false, riskTier: "LOW", expected: "LIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decideStatus(tt.policyViolation, tt.riskTier)
			if got != tt.expected {
				t.Errorf("decideStatus(%v, %q) = %q, want %q",
					tt.policyViolation, tt.riskTier, got, tt.expected)
			}
		})
	}
}
