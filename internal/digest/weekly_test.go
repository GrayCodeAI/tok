package digest

import (
	"testing"
	"time"
)

func TestNewDigestGenerator(t *testing.T) {
	dg := NewDigestGenerator()
	if dg == nil {
		t.Fatal("NewDigestGenerator returned nil")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	dg := NewDigestGenerator()
	now := time.Now()
	digest := &WeeklyDigest{
		WeekStart:      now.AddDate(0, 0, -7),
		WeekEnd:        now,
		TotalCost:      42.50,
		TotalTokens:    1000000,
		TotalRequests:  5000,
		AvgCostPerReq:  0.0085,
		CostChangePct:  10.0,
		TokenChangePct: -5.0,
		TopModels: []ModelUsage{
			{Model: "claude-3-opus", Requests: 2000, Tokens: 500000, Cost: 25.0, Percentage: 58.8},
		},
		TopTeams: []TeamUsage{
			{Team: "engineering", Requests: 3000, Cost: 20.0, Percentage: 47.0},
		},
		BudgetStatus: BudgetStatus{
			MonthlyBudget:  100.0,
			MonthlySpend:   42.50,
			UsagePercent:   42.5,
			ProjectedSpend: 85.0,
			DaysRemaining:  14,
			IsAtRisk:       false,
		},
		Recommendations: []string{"Consider using cheaper models for simple tasks"},
		Anomalies: []AnomalyEvent{
			{Type: "spike", Description: "Cost spike on Tuesday", Severity: "warning", Date: now},
		},
	}

	md := dg.GenerateMarkdown(digest)
	if md == "" {
		t.Error("GenerateMarkdown returned empty string")
	}
	// Check key content appears
	if !containsStr(md, "# TokMan Weekly Cost Digest") {
		t.Errorf("markdown should contain header, got: %s", md[:min(100, len(md))])
	}
	if !containsStr(md, "42.50") {
		t.Error("markdown should contain total cost")
	}
}

func TestGenerateJSON(t *testing.T) {
	dg := NewDigestGenerator()
	now := time.Now()
	digest := &WeeklyDigest{
		WeekStart:   now.AddDate(0, 0, -7),
		WeekEnd:     now,
		TotalCost:   42.50,
		TotalTokens: 1000000,
	}

	data, err := dg.GenerateJSON(digest)
	if err != nil {
		t.Fatalf("GenerateJSON error = %v", err)
	}
	if len(data) == 0 {
		t.Error("GenerateJSON returned empty data")
	}
}

func TestGenerateMarkdown_Empty(t *testing.T) {
	dg := NewDigestGenerator()
	digest := &WeeklyDigest{
		WeekStart: time.Now().AddDate(0, 0, -7),
		WeekEnd:   time.Now(),
	}

	md := dg.GenerateMarkdown(digest)
	if md == "" {
		t.Error("GenerateMarkdown should return something even for empty digest")
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
