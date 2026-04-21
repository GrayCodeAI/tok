package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/session"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

func TestAgentsSectionList(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopAgents = []tracking.DashboardBreakdown{
		{Key: "claude", Commands: 80, SavedTokens: 4000, ReductionPct: 62, EstimatedSavingsUSD: 1.10},
		{Key: "copilot", Commands: 40, SavedTokens: 1500, ReductionPct: 45, EstimatedSavingsUSD: 0.50},
	}

	s := newAgentsSection()
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)
	for _, want := range []string{"Agents", "claude", "copilot"} {
		if !strings.Contains(view, want) {
			t.Fatalf("agents list missing %q:\n%s", want, view)
		}
	}
}

func TestAgentsSectionDrillShowsSessions(t *testing.T) {
	now := time.Now()
	ctx := fixtureDashCtxWithTrends()
	ctx.Data.Dashboard.TopAgents = []tracking.DashboardBreakdown{
		{Key: "claude", Commands: 80, SavedTokens: 4000, ReductionPct: 62},
	}
	ctx.Data.Sessions.RecentSessions = []session.SessionOverview{
		{ID: "s1", Agent: "claude", ProjectPath: "/home/u/alpha", TotalTokens: 500, LastActivity: now.Add(-time.Hour)},
		{ID: "s2", Agent: "copilot", ProjectPath: "/home/u/beta", TotalTokens: 800, LastActivity: now.Add(-2 * time.Hour)},
		{ID: "s3", Agent: "claude", ProjectPath: "/home/u/gamma", TotalTokens: 700, LastActivity: now.Add(-3 * time.Hour)},
	}

	s := newAgentsSection()
	s.Update(ctx, tea.KeyMsg{})
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEnter})
	if s.drill != "claude" {
		t.Fatalf("drill = %q, want claude", s.drill)
	}

	view := s.View(ctx)
	if !strings.Contains(view, "alpha") || !strings.Contains(view, "gamma") {
		t.Fatalf("agents drill should include claude sessions:\n%s", view)
	}
	if strings.Contains(view, "beta") {
		t.Fatalf("agents drill leaked non-matching agent session:\n%s", view)
	}
}
