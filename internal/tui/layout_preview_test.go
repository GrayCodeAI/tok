package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Preview tests dump the rendered frame at edge widths so a human
// reviewer can eyeball layout. Skipped unless -v is set.
func TestLayoutPreviewDump(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("run with -v to dump frames at audit widths")
	}
	configs := []struct {
		w, h    int
		section int
		name    string
	}{
		{80, 24, 0, "narrow-home"},
		{80, 24, 6, "narrow-sessions"},
		{80, 24, 10, "narrow-logs"},
		{140, 40, 0, "default-home"},
		{140, 40, 2, "default-trends"},
		{180, 48, 8, "wide-pipeline"},
		{240, 60, 1, "ultrawide-today"},
	}
	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			m := driveToSection(t, c.w, c.h, c.section)
			view := m.View()
			header := fmt.Sprintf("=== %s (%dx%d, section=%d) ===", c.name, c.w, c.h, c.section)
			t.Log("\n" + header + "\n" + view)
		})
	}
}

// driveToSection returns a model positioned at the given section via the
// real input-dispatch path, not by mutating navIndex. This matters
// because sections sync their row data inside Update — a manual
// m.navIndex = N skips the sync and hides bugs like "section shows
// 'no rows' at first render until the refresh tick lands".
func driveToSection(t *testing.T, width, height, section int) model {
	t.Helper()
	loader := &stubLoader{snapshot: goldenFixture()}
	m := NewModelWithLoader(Options{Theme: ThemeDark, Days: 7}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)
	// Simulate pressing shortcut keys; sections 1–9 use their digit,
	// 10–12 need two digits — dispatch each digit as a separate
	// keypress through Update so the section's own Update runs.
	target := section + 1
	presses := []rune(fmt.Sprintf("%d", target))
	for _, r := range presses {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(model)
	}
	// For 10–12 the single keystroke produces the right index already
	// (keymap.JumpSection only binds 1–9). Fall back to manual jump +
	// synthetic Update for those.
	if m.navIndex != section {
		m.navIndex = section
		ctx := m.sectionContext()
		if next, _ := m.sections[section].Update(ctx, nil); next != nil {
			m.sections[section] = next
		}
	}
	return m
}
