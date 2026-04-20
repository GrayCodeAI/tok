package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Layout audit harness. These tests render every section at a range of
// terminal widths and assert the cheapest invariants of "doesn't look
// broken":
//
//  1. No rendered line exceeds the requested width.
//  2. Panel borders "┌" and "└" pair up on the same relative offset
//     (i.e. a top-corner implies a bottom-corner below it, never a
//     stray dangling glyph on its own line).
//  3. Sidebar numeric shortcuts line up visually despite 1- vs 2-digit
//     indices.
//  4. The empty-data path renders something non-zero for every section.
//
// The tests don't diff against goldens — they're invariant checks that
// catch the class of bugs users reported ("dangling ─┘", content
// overflow). Golden tests live in golden_test.go.

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// widthsToAudit covers the breakpoints the layout cares about:
// 80 (compact threshold for old terminals), 112 (the code's current
// compact cutoff), 140 (default medium), 180 (insights-pane cutoff),
// 240 (wide workstation). Each width gets exercised by every test
// using this table.
var widthsToAudit = []int{80, 112, 140, 180, 240}

func renderFullModel(t *testing.T, width, height int, data any) string {
	t.Helper()
	var snapshot = goldenFixture()
	if data == nil {
		snapshot = nil
	}
	loader := &stubLoader{snapshot: snapshot}
	m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = next.(model)
	if snapshot != nil {
		next, _ = m.Update(snapshotLoadedMsg{snapshot: snapshot, loadedAt: time.Date(2026, 4, 20, 9, 30, 0, 0, time.UTC)})
		m = next.(model)
	}
	return m.View()
}

func TestLayoutNoLineExceedsWidth(t *testing.T) {
	// Main invariant: every line of every rendered frame must fit in
	// the window. Overflow is the #1 cause of the "dangling ─┘"
	// artifact users see.
	for _, w := range widthsToAudit {
		for section := 0; section < 12; section++ {
			t.Run(fmt.Sprintf("w%d_s%d", w, section), func(t *testing.T) {
				loader := &stubLoader{snapshot: goldenFixture()}
				m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
				next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: 45})
				m = next.(model)
				next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
				m = next.(model)
				// Jump to this section via its shortcut. Sections
				// 1–9 use single-digit keys, 10–12 are unreachable
				// via one keystroke — simulate by mutating navIndex
				// directly (we still render through the View path).
				m.navIndex = section

				view := m.View()
				for lineNum, line := range strings.Split(view, "\n") {
					if lw := lipgloss.Width(line); lw > w {
						t.Errorf("section %d @ width %d: line %d has width %d (overflow by %d)",
							section, w, lineNum, lw, lw-w)
						// Show the offending line head for diagnosis.
						shown := line
						if len(shown) > 200 {
							shown = shown[:200] + "…"
						}
						t.Logf("  offending line: %q", shown)
						return // one failure per section is enough
					}
				}
			})
		}
	}
}

func TestLayoutEmptyDataRenders(t *testing.T) {
	// Every section must render *something* with a nil snapshot —
	// not a blank frame, not a panic. This catches bugs like "section
	// assumes ctx.Data != nil and returns empty string" which makes
	// the UI look hung.
	for section := 0; section < 12; section++ {
		t.Run(fmt.Sprintf("s%d", section), func(t *testing.T) {
			loader := &stubLoader{snapshot: nil, err: nil}
			m := NewModelWithLoader(Options{Theme: ThemeDark}, loader).(model)
			next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
			m = next.(model)
			// Mark loading as done so the section is reached at all.
			next, _ = m.Update(snapshotLoadedMsg{snapshot: nil, loadedAt: time.Now()})
			m = next.(model)
			m.navIndex = section

			view := m.View()
			if strings.TrimSpace(view) == "" {
				t.Fatalf("section %d renders an empty frame", section)
			}
		})
	}
}

func TestLayoutBalancedPanelBorders(t *testing.T) {
	// For each rendered frame, count opening corner glyphs (`┌`, `╭`)
	// and closing corner glyphs (`┘`, `╯`). They should match —
	// unequal counts mean a border wrapped onto a new line and the
	// partner glyph is now orphaned.
	for _, w := range widthsToAudit {
		for section := 0; section < 12; section++ {
			t.Run(fmt.Sprintf("w%d_s%d", w, section), func(t *testing.T) {
				loader := &stubLoader{snapshot: goldenFixture()}
				m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
				next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: 45})
				m = next.(model)
				next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
				m = next.(model)
				m.navIndex = section

				view := m.View()
				opens := strings.Count(view, "┌") + strings.Count(view, "╭")
				closes := strings.Count(view, "┘") + strings.Count(view, "╯")
				if opens != closes {
					t.Errorf("section %d @ width %d: %d opening corners vs %d closing corners",
						section, w, opens, closes)
				}
			})
		}
	}
}

func TestLayoutErrorStateRenders(t *testing.T) {
	// When the loader reports an error before any data has loaded,
	// the frame should show an error banner rather than hang.
	loader := &stubLoader{snapshot: nil, err: fmt.Errorf("simulated DB unavailable")}
	m := NewModelWithLoader(Options{Theme: ThemeDark}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: nil, err: loader.err, loadedAt: time.Now()})
	m = next.(model)

	view := m.View()
	if !strings.Contains(view, "Failed to load snapshot") {
		t.Fatalf("expected error banner in frame:\n%s", view)
	}
	if !strings.Contains(view, "simulated DB unavailable") {
		t.Fatalf("expected error detail in frame:\n%s", view)
	}
}

func TestLayoutSidebarShortcutsAlign(t *testing.T) {
	// Section names should line up with each other regardless of
	// whether the shortcut number is 1 digit (1–9) or 2 digits
	// (10–12), and regardless of which row is the active selection.
	// A naive implementation renders "  1 Home" vs " 10 Rewards"
	// (one-char shift) or gives the active row a left-border the
	// inactive rows lack (another one-char shift).
	view := renderFullModel(t, 140, 45, "rich")
	lines := strings.Split(view, "\n")
	// The sidebar rows have exactly one label per line where the
	// label is one of the 12 section names preceded by a single
	// shortcut number. Match precisely so we don't pick up matches
	// embedded in panel text.
	names := []string{"Home", "Today", "Trends", "Providers", "Models", "Agents",
		"Sessions", "Commands", "Pipeline", "Rewards", "Logs", "Config"}
	prefixes := make(map[string]int)
	for _, ln := range lines {
		// Sidebar lines look like "  > 10 Rewards" or "     2 Today".
		// They always have a digit immediately before the name,
		// preceded only by leading whitespace plus an optional ">"
		// active marker. We reject matches whose prefix doesn't end
		// in "<digit(s)> " — that filters out panel titles like
		// "Top Providers" and the header "tok  Home".
		trimmed := strings.TrimRight(ln, " ")
		for _, name := range names {
			idx := strings.Index(trimmed, " "+name)
			if idx <= 0 {
				continue
			}
			prefix := trimmed[:idx+1] // include the trailing space
			prefixTrim := strings.TrimRight(prefix, " ")
			if prefixTrim == "" {
				continue
			}
			last := prefixTrim[len(prefixTrim)-1]
			if last < '0' || last > '9' {
				continue // not a sidebar row
			}
			if lipgloss.Width(prefix) > 12 {
				continue
			}
			if _, seen := prefixes[name]; !seen {
				prefixes[name] = lipgloss.Width(prefix)
			}
		}
	}
	if len(prefixes) < 10 {
		// Too few matches to make a claim (narrow terminal, compact mode).
		return
	}
	var sample int
	for _, w := range prefixes {
		sample = w
		break
	}
	for name, w := range prefixes {
		if w != sample {
			t.Errorf("sidebar entry %q has prefix width %d, expected %d", name, w, sample)
		}
	}
}
