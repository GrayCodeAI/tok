package tui

import (
	"strings"
	"testing"

	"github.com/lakshmanpatel/tok/internal/session"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

func fixtureDashCtxWithTrends() SectionContext {
	return SectionContext{
		Theme: newTheme(),
		Keys:  DefaultKeyMap(),
		Data: &tracking.WorkspaceDashboardSnapshot{
			Dashboard: &tracking.DashboardSnapshot{
				Overview: tracking.DashboardOverview{
					TotalSavedTokens: 9999,
				},
				DailyTrends: []tracking.DashboardTrendPoint{
					{Period: "2026-04-13", Commands: 10, SavedTokens: 100, ReductionPct: 20},
					{Period: "2026-04-14", Commands: 15, SavedTokens: 220, ReductionPct: 33},
					{Period: "2026-04-15", Commands: 18, SavedTokens: 340, ReductionPct: 41},
					{Period: "2026-04-16", Commands: 22, SavedTokens: 500, ReductionPct: 55},
					{Period: "2026-04-17", Commands: 30, SavedTokens: 700, ReductionPct: 60},
					{Period: "2026-04-18", Commands: 25, SavedTokens: 900, ReductionPct: 62},
					{Period: "2026-04-19", Commands: 40, SavedTokens: 1100, ReductionPct: 65},
				},
				WeeklyTrends: []tracking.DashboardTrendPoint{
					{Period: "W1", Commands: 100, SavedTokens: 4000, ReductionPct: 55},
					{Period: "W2", Commands: 140, SavedTokens: 5500, ReductionPct: 62},
				},
				TopProviders: []tracking.DashboardBreakdown{
					{Key: "anthropic", Commands: 80, SavedTokens: 4000, ReductionPct: 60, EstimatedSavingsUSD: 1.23},
					{Key: "openai", Commands: 40, SavedTokens: 1500, ReductionPct: 45, EstimatedSavingsUSD: 0.52},
				},
				TopProviderModels: []tracking.DashboardBreakdown{
					{Key: "anthropic / claude-opus-4-7", Commands: 50, SavedTokens: 2500, ReductionPct: 60},
					{Key: "anthropic / claude-sonnet", Commands: 30, SavedTokens: 1500, ReductionPct: 55},
					{Key: "openai / gpt-4o", Commands: 40, SavedTokens: 1500, ReductionPct: 45},
				},
				Streaks:   tracking.DashboardStreaks{SavingsDays: 4, GoalDays: 7, GoalReductionPct: 40, BestDay: "2026-04-19", BestDaySavedTokens: 1100, BestDayReductionPct: 65},
				Lifecycle: tracking.DashboardLifecycle{ActiveDays30d: 7, ProjectsCount: 3, AvgSavedTokensPerExec: 42.5},
				Budgets: tracking.DashboardBudgetStatus{
					Daily: tracking.DashboardBudgetWindow{FilteredTokens: 15_000, TokenBudget: 100_000, TokenUtilizationPct: 15},
				},
			},
			Sessions: &session.SessionAnalyticsSnapshot{},
		},
		Opts:   Options{Days: 30},
		Width:  120,
		Height: 40,
		Env:    Environment{UTF8: true}, // tests default to UTF-8 so glyphs match expectations
	}
}

func TestTodaySectionView(t *testing.T) {
	s := newTodaySection()
	ctx := fixtureDashCtxWithTrends()
	view := s.View(ctx)
	// Card titles render uppercased; substring match on the casefolded body.
	for _, want := range []string{"Today", "STREAK", "DAILY BUDGET", "Trailing 7 days", "vs yesterday"} {
		if !strings.Contains(view, want) {
			t.Fatalf("today view missing %q:\n%s", want, view)
		}
	}
}

func TestTodayDeltaHandlesZeroPrior(t *testing.T) {
	if got := deltaLabel(100, 0); got != "first tracked day" {
		t.Fatalf("deltaLabel(100, 0) = %q, want 'first tracked day'", got)
	}
	if got := deltaLabel(0, 0); got != "no activity" {
		t.Fatalf("deltaLabel(0, 0) = %q, want 'no activity'", got)
	}
}
