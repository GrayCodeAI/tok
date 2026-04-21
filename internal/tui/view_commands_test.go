package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

func TestCommandsSectionTopListAndToggleToWeak(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopCommands = []tracking.DashboardBreakdown{
		{Key: "git diff", Commands: 20, SavedTokens: 12000, ReductionPct: 92},
		{Key: "npm test", Commands: 15, SavedTokens: 9000, ReductionPct: 88},
	}
	ctx.Data.Dashboard.LowSavingsCommands = []tracking.DashboardBreakdown{
		{Key: "gh pr view 3 --json", Commands: 6, SavedTokens: 10, ReductionPct: 4},
	}

	s := newCommandsSection()
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)
	if !strings.Contains(view, "git diff") {
		t.Fatalf("top view should list git diff:\n%s", view)
	}
	if strings.Contains(view, "gh pr view") {
		t.Fatalf("top view should not leak weak commands:\n%s", view)
	}

	// Toggle to weak mode
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	view = s.View(ctx)
	if !strings.Contains(view, "gh pr view") {
		t.Fatalf("weak view should list gh pr view:\n%s", view)
	}
	if strings.Contains(view, "git diff") {
		t.Fatalf("weak view should not include top commands:\n%s", view)
	}
}

func TestCommandsSectionDrillShowsLayers(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopCommands = []tracking.DashboardBreakdown{
		{Key: "git diff", Commands: 20, SavedTokens: 12000, ReductionPct: 92},
	}
	ctx.Data.Dashboard.TopLayers = []tracking.DashboardLayerSummary{
		{LayerName: "entropy_prune", CallCount: 100, TotalSaved: 5000, AvgSaved: 50},
		{LayerName: "dedup", CallCount: 80, TotalSaved: 4000, AvgSaved: 50},
	}

	s := newCommandsSection()
	s.Update(ctx, tea.KeyMsg{})
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEnter})
	if s.drill != "git diff" {
		t.Fatalf("drill = %q, want 'git diff'", s.drill)
	}

	view := s.View(ctx)
	for _, want := range []string{"git diff", "Top pipeline layers", "entropy_prune", "dedup"} {
		if !strings.Contains(view, want) {
			t.Fatalf("commands drill missing %q:\n%s", want, view)
		}
	}
}
