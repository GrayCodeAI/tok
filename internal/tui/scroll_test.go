package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// TestHomeScrollHintShowsOnOverflow verifies the scroll hint appears
// when the Home section content exceeds the viewport.
func TestHomeScrollHintShowsOnOverflow(t *testing.T) {
	m := driveModel(t)
	view := m.View()
	if !strings.Contains(view, "▼") {
		t.Fatalf("expected downward scroll hint in Home view, got:\n%s", view)
	}
	if !strings.Contains(view, "pgdn") {
		t.Errorf("expected pgdn hint in Home view")
	}
}

// TestScrollPageDownAdvancesOffset verifies PgDn increases the scroll
// offset and moves the rendered frame forward in the section's output.
func TestScrollPageDownAdvancesOffset(t *testing.T) {
	m := driveModel(t)
	before := m.View()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m2 := next.(model)
	after := m2.View()
	if before == after {
		t.Fatalf("PgDn did not change the rendered frame")
	}
	if m2.scrollOffsets[0] == 0 {
		t.Fatalf("expected scrollOffsets[0] > 0 after PgDn, got 0")
	}
}

// TestScrollGotoBottomShowsTail verifies G (Bottom) scrolls to the end
// so the last panel is visible.
func TestScrollGotoBottomShowsTail(t *testing.T) {
	m := driveModel(t, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	view := m.View()
	// The Home view's tail contains the "Snapshot" health panel title.
	if !strings.Contains(view, "Snapshot") {
		t.Fatalf("expected 'Snapshot' panel after G, got:\n%s", view)
	}
}

// TestWelcomeRendersOnNilData verifies the onboarding welcome screen
// shows up on first launch (no data yet), not a generic loading spinner.
func TestWelcomeRendersOnNilData(t *testing.T) {
	loader := &stubLoader{snapshot: nil}
	m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	// Simulate loader finishing with no data (first-run: DB empty).
	next, _ = m.Update(snapshotLoadedMsg{snapshot: nil, loadedAt: time.Now()})
	m = next.(model)
	view := m.View()
	if !strings.Contains(view, "Welcome to tok") {
		t.Fatalf("expected welcome screen on nil data, got:\n%s", view)
	}
	if !strings.Contains(view, "tok init") {
		t.Errorf("expected setup hint on welcome screen")
	}
}

// TestCompactNavBreadcrumbSingleLine verifies the compact-mode nav fits
// on one line at 80 cols (previous behavior wrapped to 2+ lines).
func TestCompactNavBreadcrumbSingleLine(t *testing.T) {
	loader := &stubLoader{snapshot: goldenFixture()}
	m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)
	m.navIndex = 6 // Sessions
	view := m.View()

	// Find the breadcrumb line and confirm it's a single line with the
	// navigation affordance.
	if !strings.Contains(view, "[7/12]") {
		t.Fatalf("expected '[7/12]' breadcrumb in compact nav at 80 cols, got:\n%s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "[7/12]") && lipgloss.Width(line) > 80 {
			t.Errorf("compact nav line overflows 80 cols: width=%d", lipgloss.Width(line))
		}
	}
}
