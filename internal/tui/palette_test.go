package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newPaletteForTest() *Palette {
	deps := ActionDeps{SectionCount: 3}
	reg := DefaultActionRegistry(deps)
	sections := []SectionRenderer{
		newHomeSection(),
		newPlaceholderSection("Today", "Easy Day"),
		newPlaceholderSection("Trends", "Analytics"),
	}
	return NewPalette(reg, sections)
}

func TestPaletteOpenClose(t *testing.T) {
	p := newPaletteForTest()
	if p.IsOpen() {
		t.Fatal("new palette should be closed")
	}
	_ = p.Open()
	if !p.IsOpen() {
		t.Fatal("Open should mark palette open")
	}
	p.Close()
	if p.IsOpen() {
		t.Fatal("Close should mark palette closed")
	}
}

func TestPaletteFuzzyMatchesSectionName(t *testing.T) {
	p := newPaletteForTest()
	_ = p.Open()

	// Type "tren" and expect "Go to Trends" near the top.
	for _, r := range "tren" {
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if len(p.matches) == 0 {
		t.Fatal("expected at least one match for 'tren'")
	}
	top := p.matches[0]
	if top.Title != "Go to Trends" {
		t.Fatalf("top match = %q, want 'Go to Trends'", top.Title)
	}
}

func TestPaletteEnterEmitsPaletteExecMsg(t *testing.T) {
	p := newPaletteForTest()
	_ = p.Open()
	for _, r := range "home" {
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should emit a cmd")
	}
	msg := cmd()
	exec, ok := msg.(paletteExecMsg)
	if !ok {
		t.Fatalf("expected paletteExecMsg, got %T", msg)
	}
	if exec.ActionID != "section.jump" {
		t.Fatalf("ActionID = %q, want section.jump", exec.ActionID)
	}
	if exec.Args != "1" { // Home is section index 1 (1-based)
		t.Fatalf("Args = %q, want '1'", exec.Args)
	}
}

func TestPaletteEscClosesWithoutExec(t *testing.T) {
	p := newPaletteForTest()
	_ = p.Open()
	cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if p.IsOpen() {
		t.Fatal("esc should close palette")
	}
	if cmd == nil {
		t.Fatal("esc should emit paletteCloseMsg")
	}
	if _, ok := cmd().(paletteCloseMsg); !ok {
		t.Fatalf("expected paletteCloseMsg, got %T", cmd())
	}
}

func TestFuzzyScoreRanking(t *testing.T) {
	tests := []struct {
		needle, id, title string
		wantPositive      bool
	}{
		{"ref", "view.refresh", "Refresh", true},
		{"jump", "section.jump", "Jump to section", true},
		{"xyz", "view.refresh", "Refresh", false},
		{"", "view.refresh", "Refresh", true},
	}
	for _, tc := range tests {
		s := fuzzyScore(tc.needle, tc.id, tc.title)
		if tc.wantPositive && s == 0 {
			t.Errorf("fuzzyScore(%q, %q, %q) = 0, want positive", tc.needle, tc.id, tc.title)
		}
		if !tc.wantPositive && s > 0 {
			t.Errorf("fuzzyScore(%q, %q, %q) = %d, want 0", tc.needle, tc.id, tc.title, s)
		}
	}
}
