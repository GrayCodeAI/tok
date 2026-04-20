package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

func TestModelsSectionList(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	// Populate TopModels — the base fixture only seeds TopProviders/TopProviderModels.
	ctx.Data.Dashboard.TopModels = []tracking.DashboardBreakdown{
		{Key: "claude-opus-4-7", Commands: 50, SavedTokens: 2500, ReductionPct: 60, EstimatedSavingsUSD: 0.80},
		{Key: "claude-sonnet", Commands: 30, SavedTokens: 1500, ReductionPct: 55, EstimatedSavingsUSD: 0.43},
		{Key: "gpt-4o", Commands: 40, SavedTokens: 1500, ReductionPct: 45, EstimatedSavingsUSD: 0.52},
	}

	s := newModelsSection()
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)
	for _, want := range []string{"Models", "claude-opus-4-7", "claude-sonnet", "gpt-4o"} {
		if !strings.Contains(view, want) {
			t.Fatalf("models list missing %q:\n%s", want, view)
		}
	}
}

func TestModelsSectionDrillShowsProviderPartners(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopModels = []tracking.DashboardBreakdown{
		{Key: "claude-opus-4-7", Commands: 50, SavedTokens: 2500, ReductionPct: 60},
	}
	// Provider partners for this model (from the composite map).
	ctx.Data.Dashboard.TopProviderModels = []tracking.DashboardBreakdown{
		{Key: "anthropic / claude-opus-4-7", Commands: 40, SavedTokens: 2000, ReductionPct: 62},
		{Key: "bedrock / claude-opus-4-7", Commands: 10, SavedTokens: 500, ReductionPct: 55},
		{Key: "openai / gpt-4o", Commands: 40, SavedTokens: 1500, ReductionPct: 45},
	}

	s := newModelsSection()
	s.Update(ctx, tea.KeyMsg{})
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEnter})
	if s.drill != "claude-opus-4-7" {
		t.Fatalf("drill = %q, want claude-opus-4-7", s.drill)
	}

	view := s.View(ctx)
	if !strings.Contains(view, "anthropic") || !strings.Contains(view, "bedrock") {
		t.Fatalf("detail view should list provider partners:\n%s", view)
	}
	if strings.Contains(view, "openai") {
		t.Fatalf("detail view leaked unrelated provider:\n%s", view)
	}
}

func TestFilterProviderPartners(t *testing.T) {
	items := []tracking.DashboardBreakdown{
		{Key: "anthropic / claude-opus-4-7"},
		{Key: "bedrock / claude-opus-4-7"},
		{Key: "openai / gpt-4o"},
	}
	got := filterProviderPartners(items, "claude-opus-4-7")
	if len(got) != 2 {
		t.Fatalf("expected 2 partners for claude-opus-4-7, got %d", len(got))
	}
}
