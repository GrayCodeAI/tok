package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"

	"github.com/GrayCodeAI/tok/internal/session"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// These tests capture the rendered frames after a scripted sequence of
// updates and compare them to files in testdata/. Run with -update to
// refresh when you've intentionally changed the output shape:
//
//	go test -update ./internal/tui/... -run TestGolden
//
// The golden-file harness strips ANSI escape sequences before diffing,
// so terminal-color variance between machines does not break tests.
//
// We pin the lipgloss color profile to ASCII for these tests so the
// captured output is stable across hosts that would otherwise detect
// different truecolor/256-color support.
func init() {
	// termenv.Ascii strips all color sequences from rendered output so
	// golden files only contain layout, not host-specific truecolor
	// escapes. Must run before any theme is built.
	lipgloss.SetColorProfile(termenv.Ascii)
}

// goldenFixtureNow is the pinned wall-clock for golden tests. Real
// rendering code calls nowFunc() (overridden to this in init below).
// Without pinning, `time.Since(...)` drifts between runs and goldens
// flake on "13h ago" vs "14h ago" etc.
var goldenFixtureNow = time.Date(2026, 4, 20, 9, 30, 0, 0, time.UTC)

func init() {
	nowFunc = func() time.Time { return goldenFixtureNow }
}

// goldenFixture is the canonical "busy workspace" dataset used by
// every golden test so output sizes are meaningful (placeholder values
// tend to hide layout bugs because they don't wrap).
func goldenFixture() *tracking.WorkspaceDashboardSnapshot {
	return &tracking.WorkspaceDashboardSnapshot{
		Dashboard: &tracking.DashboardSnapshot{
			Overview: tracking.DashboardOverview{
				TotalSavedTokens:    82345,
				EstimatedSavingsUSD: 2.1734,
				ReductionPct:        68.5,
				TotalCommands:       142,
				UniqueProviders:     3,
				UniqueModels:        5,
				UniqueAgents:        2,
			},
			DailyTrends: []tracking.DashboardTrendPoint{
				{Period: "2026-04-14", Commands: 10, SavedTokens: 1200, ReductionPct: 52},
				{Period: "2026-04-15", Commands: 12, SavedTokens: 1800, ReductionPct: 58},
				{Period: "2026-04-16", Commands: 15, SavedTokens: 2300, ReductionPct: 62},
				{Period: "2026-04-17", Commands: 20, SavedTokens: 3100, ReductionPct: 65},
				{Period: "2026-04-18", Commands: 22, SavedTokens: 3800, ReductionPct: 67},
				{Period: "2026-04-19", Commands: 28, SavedTokens: 4500, ReductionPct: 68},
				{Period: "2026-04-20", Commands: 35, SavedTokens: 5400, ReductionPct: 70},
			},
			WeeklyTrends: []tracking.DashboardTrendPoint{
				{Period: "W1", Commands: 80, SavedTokens: 12000, ReductionPct: 58},
				{Period: "W2", Commands: 142, SavedTokens: 22300, ReductionPct: 68},
			},
			TopProviders: []tracking.DashboardBreakdown{
				{Key: "anthropic", Commands: 80, SavedTokens: 48000, ReductionPct: 72, EstimatedSavingsUSD: 1.30},
				{Key: "openai", Commands: 40, SavedTokens: 18000, ReductionPct: 54, EstimatedSavingsUSD: 0.50},
				{Key: "local", Commands: 22, SavedTokens: 16345, ReductionPct: 60, EstimatedSavingsUSD: 0.37},
			},
			LowSavingsCommands: []tracking.DashboardBreakdown{
				{Key: "gh pr view 3 --json", Commands: 6, SavedTokens: 40, ReductionPct: 4},
				{Key: "git log --oneline -1", Commands: 4, SavedTokens: 20, ReductionPct: 3},
			},
			Streaks:      tracking.DashboardStreaks{SavingsDays: 5, GoalDays: 7, GoalReductionPct: 40, BestDay: "2026-04-20", BestDaySavedTokens: 5400, BestDayReductionPct: 70},
			Lifecycle:    tracking.DashboardLifecycle{ActiveDays30d: 7, ProjectsCount: 3, AvgSavedTokensPerExec: 50.2},
			Gamification: tracking.DashboardGamification{Points: 1240, Level: 3, NextLevelPoints: 2000, Badges: []string{"early-bird"}},
			Budgets: tracking.DashboardBudgetStatus{
				Daily: tracking.DashboardBudgetWindow{FilteredTokens: 15000, TokenBudget: 100000, TokenUtilizationPct: 15},
			},
		},
		DataQuality: tracking.DashboardDataQuality{
			TotalCommands: 142, ParseFailures: 1,
		},
		Sessions: &session.SessionAnalyticsSnapshot{
			StoreSummary: session.SessionStoreSummary{TotalSessions: 3, ActiveSessions: 1, TopAgent: "claude"},
			RecentSessions: []session.SessionOverview{
				{ID: "sess-01", Agent: "claude", ProjectPath: "/home/user/project-alpha", TotalTokens: 12000, TotalTurns: 22, SnapshotCount: 2, IsActive: true, StartedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC), LastActivity: time.Date(2026, 4, 20, 9, 30, 0, 0, time.UTC)},
				{ID: "sess-02", Agent: "copilot", ProjectPath: "/home/user/project-beta", TotalTokens: 3000, TotalTurns: 7, SnapshotCount: 0, StartedAt: time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC), LastActivity: time.Date(2026, 4, 19, 18, 0, 0, 0, time.UTC)},
			},
		},
	}
}

// driveModel applies a sequence of messages to a fresh model. Returns
// the final model so the caller can call View() for the snapshot.
// Messages go through Update verbatim — no tea.Program, no timers.
// The fixed snapshot arrives via snapshotLoadedMsg so the render path
// is exactly what a live user would see after the first refresh tick.
func driveModel(t *testing.T, msgs ...tea.Msg) model {
	t.Helper()
	loader := &stubLoader{snapshot: goldenFixture()}
	m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Date(2026, 4, 20, 9, 30, 0, 0, time.UTC)})
	m = next.(model)
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}

func TestGoldenHomeFrame(t *testing.T) {
	m := driveModel(t)
	golden.RequireEqual(t, m.View())
}

func TestGoldenSessionsList(t *testing.T) {
	// Jump to Sessions (section index 6 → 1-based key "7") and render.
	m := driveModel(t, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	golden.RequireEqual(t, m.View())
}

func TestGoldenSessionsDrillDown(t *testing.T) {
	m := driveModel(t,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")},
		tea.KeyMsg{Type: tea.KeyEnter}, // drill first row
	)
	golden.RequireEqual(t, m.View())
}

func TestGoldenConfirmOverlayOpens(t *testing.T) {
	m := driveModel(t,
		// Send actionRequestMsg for the destructive logs.clear; the
		// root model arms the confirm modal instead of running.
		actionRequestMsg{ActionID: "logs.clear"},
	)
	golden.RequireEqual(t, m.View())
}
