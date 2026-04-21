package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/session"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

func fixtureSessionCtx(width int) SectionContext {
	now := time.Now()
	return SectionContext{
		Theme: newTheme(),
		Keys:  DefaultKeyMap(),
		Data: &tracking.WorkspaceDashboardSnapshot{
			Sessions: &session.SessionAnalyticsSnapshot{
				StoreSummary: session.SessionStoreSummary{
					TotalSessions:  3,
					ActiveSessions: 1,
					TopAgent:       "claude",
				},
				RecentSessions: []session.SessionOverview{
					{
						ID:                "sess-01",
						Agent:             "claude",
						ProjectPath:       "/home/user/project-alpha",
						StartedAt:         now.Add(-2 * time.Hour),
						LastActivity:      now.Add(-5 * time.Minute),
						IsActive:          true,
						TotalTurns:        42,
						TotalTokens:       12345,
						CompressionRatio:  1.7,
						ContextBlockCount: 8,
						SnapshotCount:     3,
					},
					{
						ID:                "sess-02",
						Agent:             "copilot",
						ProjectPath:       "/home/user/project-beta",
						StartedAt:         now.Add(-36 * time.Hour),
						LastActivity:      now.Add(-20 * time.Hour),
						TotalTurns:        7,
						TotalTokens:       998,
						CompressionRatio:  1.1,
						ContextBlockCount: 2,
						SnapshotCount:     0,
					},
				},
			},
		},
		Opts:    Options{Days: 30},
		Width:   width,
		Height:  30,
		Compact: false,
		Focused: true,
		Env:     Environment{UTF8: true},
	}
}

func TestSessionsSectionList(t *testing.T) {
	s := newSessionsSection()
	ctx := fixtureSessionCtx(120)
	// Sync the table by running Update with a no-op message.
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)

	for _, want := range []string{"Sessions", "claude", "copilot", "project-alpha", "project-beta"} {
		if !strings.Contains(view, want) {
			t.Fatalf("list view missing %q:\n%s", want, view)
		}
	}
}

func TestSessionsSectionDrillDown(t *testing.T) {
	s := newSessionsSection()
	ctx := fixtureSessionCtx(120)
	s.Update(ctx, tea.KeyMsg{}) // seed the table

	// Move down one and press Enter to drill into the second session.
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEnter})
	if s.drillID != "sess-02" {
		t.Fatalf("drillID = %q, want sess-02", s.drillID)
	}

	view := s.View(ctx)
	for _, want := range []string{"Session detail", "sess-02", "Compression ratio", "copilot"} {
		if !strings.Contains(view, want) {
			t.Fatalf("detail view missing %q:\n%s", want, view)
		}
	}

	// Esc should return to the list.
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEsc})
	if s.drillID != "" {
		t.Fatalf("drillID after esc = %q, want empty", s.drillID)
	}
}

func TestSessionsSectionFilterViaSearchMsg(t *testing.T) {
	s := newSessionsSection()
	ctx := fixtureSessionCtx(120)
	s.Update(ctx, tea.KeyMsg{}) // seed
	if s.table.Len() != 2 {
		t.Fatalf("initial rows = %d, want 2", s.table.Len())
	}

	s.Update(ctx, searchMsg{Query: "alpha"})
	if got := s.table.Len(); got != 1 {
		t.Fatalf("after filter 'alpha' rows = %d, want 1", got)
	}

	s.Update(ctx, searchMsg{Query: ""})
	if got := s.table.Len(); got != 2 {
		t.Fatalf("cleared filter rows = %d, want 2", got)
	}
}
