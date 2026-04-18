package tracking

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGetDashboardDataQualityAndPricingCoverage(t *testing.T) {
	tr := newTestTracker(t)
	project := filepath.Join(t.TempDir(), "project")
	now := time.Now()

	insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 1000,
		FilteredTokens: 400,
		SavedTokens:    600,
		ProjectPath:    project,
		SessionID:      "sess-a",
		ParseSuccess:   true,
		AgentName:      "Claude Code",
		ModelName:      "claude-4-sonnet",
		Provider:       "Anthropic",
	}, now)
	insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git diff",
		OriginalTokens: 1500,
		FilteredTokens: 900,
		SavedTokens:    600,
		ProjectPath:    project,
		ParseSuccess:   false,
		ModelName:      "mystery-model",
	}, now)

	quality, err := tr.GetDashboardDataQuality(DashboardQueryOptions{
		Days:        30,
		ProjectPath: project,
	})
	if err != nil {
		t.Fatalf("GetDashboardDataQuality(): %v", err)
	}

	if quality.TotalCommands != 2 {
		t.Fatalf("TotalCommands = %d, want 2", quality.TotalCommands)
	}
	if quality.CommandsMissingAgent != 1 {
		t.Fatalf("CommandsMissingAgent = %d, want 1", quality.CommandsMissingAgent)
	}
	if quality.CommandsMissingProvider != 1 {
		t.Fatalf("CommandsMissingProvider = %d, want 1", quality.CommandsMissingProvider)
	}
	if quality.CommandsMissingSession != 1 {
		t.Fatalf("CommandsMissingSession = %d, want 1", quality.CommandsMissingSession)
	}
	if quality.ParseFailures != 1 {
		t.Fatalf("ParseFailures = %d, want 1", quality.ParseFailures)
	}
	if quality.PricingCoverage.TotalAttributedCommands != 2 {
		t.Fatalf("TotalAttributedCommands = %d, want 2", quality.PricingCoverage.TotalAttributedCommands)
	}
	if quality.PricingCoverage.KnownPricingCommands != 1 {
		t.Fatalf("KnownPricingCommands = %d, want 1", quality.PricingCoverage.KnownPricingCommands)
	}
	if quality.PricingCoverage.FallbackPricingCommands != 1 {
		t.Fatalf("FallbackPricingCommands = %d, want 1", quality.PricingCoverage.FallbackPricingCommands)
	}
	if len(quality.PricingCoverage.UnknownModels) != 1 || quality.PricingCoverage.UnknownModels[0] != "mystery-model" {
		t.Fatalf("UnknownModels = %#v, want [mystery-model]", quality.PricingCoverage.UnknownModels)
	}
}
