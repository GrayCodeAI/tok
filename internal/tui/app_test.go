package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lakshmanpatel/tok/internal/session"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

type stubLoader struct {
	snapshot *tracking.WorkspaceDashboardSnapshot
	err      error
	closed   bool
}

func (s *stubLoader) Load(context.Context, Options) (*tracking.WorkspaceDashboardSnapshot, error) {
	return s.snapshot, s.err
}

func (s *stubLoader) Close() error {
	s.closed = true
	return nil
}

func TestOptionsNormalized(t *testing.T) {
	opts := (Options{}).normalized()
	if opts.RefreshInterval <= 0 {
		t.Fatal("expected positive refresh interval")
	}
	if opts.Days != 30 {
		t.Fatalf("Days = %d, want 30", opts.Days)
	}
}

func TestSectionShortcutIndex(t *testing.T) {
	idx, ok := sectionShortcutIndex("3", 12)
	if !ok || idx != 2 {
		t.Fatalf("sectionShortcutIndex(\"3\") = (%d, %v), want (2, true)", idx, ok)
	}
	idx, ok = sectionShortcutIndex("11", 12)
	if !ok || idx != 10 {
		t.Fatalf("sectionShortcutIndex(\"11\") = (%d, %v), want (10, true)", idx, ok)
	}
	// "0" jumps to section 10 as a single-keystroke shortcut past
	// section 9. Only valid when there are at least 10 sections.
	idx, ok = sectionShortcutIndex("0", 12)
	if !ok || idx != 9 {
		t.Fatalf("sectionShortcutIndex(\"0\") = (%d, %v), want (9, true)", idx, ok)
	}
	if _, ok := sectionShortcutIndex("0", 8); ok {
		t.Fatal("expected 0 to be invalid when section count < 10")
	}
	if _, ok := sectionShortcutIndex("b", 12); ok {
		t.Fatal("expected alpha shortcuts to be invalid")
	}
}

func TestModelRefreshAndNavigation(t *testing.T) {
	loader := &stubLoader{
		snapshot: &tracking.WorkspaceDashboardSnapshot{
			Dashboard: &tracking.DashboardSnapshot{
				Overview: tracking.DashboardOverview{
					TotalSavedTokens:    1200,
					EstimatedSavingsUSD: 0.02,
					ReductionPct:        52.5,
					TotalCommands:       8,
				},
				Lifecycle: tracking.DashboardLifecycle{ActiveDays30d: 4},
				Streaks:   tracking.DashboardStreaks{SavingsDays: 3},
				Gamification: tracking.DashboardGamification{
					Points: 420,
					Level:  1,
				},
			},
			Sessions: &session.SessionAnalyticsSnapshot{
				StoreSummary: session.SessionStoreSummary{TotalSessions: 2, ActiveSessions: 1},
			},
		},
	}
	m := NewModelWithLoader(Options{}, loader).(model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	if !m.ready || m.compact {
		t.Fatalf("expected ready non-compact model, got ready=%v compact=%v", m.ready, m.compact)
	}

	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)
	if m.data == nil {
		t.Fatal("expected snapshot data")
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m = next.(model)
	if m.navIndex != 1 {
		t.Fatalf("navIndex = %d, want 1", m.navIndex)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = next.(model)
	if !m.helpOpen {
		t.Fatal("expected helpOpen")
	}
}

func TestHomeViewDoesNotOverflowWidth(t *testing.T) {
	loader := &stubLoader{
		snapshot: &tracking.WorkspaceDashboardSnapshot{
			Dashboard: &tracking.DashboardSnapshot{
				Overview: tracking.DashboardOverview{
					TotalSavedTokens:    1200,
					EstimatedSavingsUSD: 0.02,
					ReductionPct:        52.5,
					TotalCommands:       8,
				},
				DailyTrends:        []tracking.DashboardTrendPoint{{SavedTokens: 10}, {SavedTokens: 20}},
				WeeklyTrends:       []tracking.DashboardTrendPoint{{SavedTokens: 100}, {SavedTokens: 150}},
				TopProviders:       []tracking.DashboardBreakdown{{Key: "(unknown)", SavedTokens: 900, ReductionPct: 80}},
				LowSavingsCommands: []tracking.DashboardBreakdown{{Key: "gh pr view 3 --json", SavedTokens: 10, ReductionPct: 4}},
				Lifecycle:          tracking.DashboardLifecycle{ActiveDays30d: 4},
				Streaks:            tracking.DashboardStreaks{SavingsDays: 3},
				Gamification: tracking.DashboardGamification{
					Points: 420,
					Level:  1,
				},
			},
			DataQuality: tracking.DashboardDataQuality{
				TotalCommands:           8,
				CommandsMissingAgent:    2,
				CommandsMissingProvider: 2,
				CommandsMissingModel:    2,
				CommandsMissingSession:  1,
				ParseFailures:           1,
			},
			Sessions: &session.SessionAnalyticsSnapshot{
				StoreSummary: session.SessionStoreSummary{TotalSessions: 2, ActiveSessions: 1, TopAgent: "agent"},
			},
		},
	}
	m := NewModelWithLoader(Options{}, loader).(model)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	view := m.View()
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 120 {
			t.Fatalf("line width %d exceeds 120: %q", lipgloss.Width(line), line)
		}
	}
}

func TestQuitKeyClosesLoader(t *testing.T) {
	loader := &stubLoader{}
	m := NewModelWithLoader(Options{}, loader).(model)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = next.(model)
	if !m.quitting {
		t.Fatal("expected quitting=true after q")
	}
	if cmd == nil {
		t.Fatal("expected shutdown command")
	}
	msg := cmd()
	if _, ok := msg.(quitMsg); !ok {
		t.Fatalf("expected quitMsg, got %T", msg)
	}
	if !loader.closed {
		t.Fatal("loader was not closed on shutdown")
	}
}

func TestCancelledLoadIsSuppressed(t *testing.T) {
	loader := &stubLoader{}
	m := NewModelWithLoader(Options{}, loader).(model)

	next, _ := m.Update(snapshotLoadedMsg{err: context.Canceled, loadedAt: time.Now()})
	m = next.(model)
	if m.err != nil {
		t.Fatalf("expected cancellation to be suppressed, got err=%v", m.err)
	}
}
