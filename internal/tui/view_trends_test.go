package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTrendsSectionRendersDaily(t *testing.T) {
	s := newTrendsSection()
	ctx := fixtureDashCtxWithTrends()
	view := s.View(ctx)
	for _, want := range []string{"Trends", "Saved tokens", "Reduction", "Commands"} {
		if !strings.Contains(view, want) {
			t.Fatalf("trends view missing %q:\n%s", want, view)
		}
	}
	if !strings.ContainsRune(view, '⠀') && !strings.ContainsAny(view, "⠁⠂⠃⠄⠅⠆⠇⠈⠉⠊") {
		// Any Braille cell in the output is enough — check it's non-empty.
		t.Fatalf("expected Braille characters in rendered chart; got:\n%s", view)
	}
}

func TestTrendsSectionToggleGranularity(t *testing.T) {
	s := newTrendsSection()
	if s.granularity != trendDaily {
		t.Fatal("default granularity should be daily")
	}
	s.Update(fixtureDashCtxWithTrends(), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	if s.granularity != trendWeekly {
		t.Fatal("'w' should switch to weekly")
	}
	s.Update(fixtureDashCtxWithTrends(), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if s.granularity != trendDaily {
		t.Fatal("'d' should switch back to daily")
	}
}
