package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPaletteForUnknownFallsBackToDark(t *testing.T) {
	dark := paletteFor(ThemeDark)
	unknown := paletteFor("nonexistent")
	if dark.bg != unknown.bg || dark.fg != unknown.fg {
		t.Fatal("unknown theme should fall back to dark palette")
	}
}

func TestAllBundledThemesResolve(t *testing.T) {
	for _, name := range AvailableThemes {
		p := paletteFor(name)
		if p.bg == "" || p.fg == "" {
			t.Errorf("theme %s missing bg/fg", name)
		}
		if len(p.accents) == 0 {
			t.Errorf("theme %s has no accent colors", name)
		}
	}
}

func TestThemeCycleAdvances(t *testing.T) {
	loader := &stubLoader{}
	m := NewModelWithLoader(Options{Theme: ThemeDark}, loader).(model)

	// Drive themeCycleMsg and verify the model's opts + theme rotate.
	seen := map[ThemeName]bool{m.opts.Theme: true}
	for i := 0; i < len(AvailableThemes)+1; i++ {
		next, _ := m.Update(themeCycleMsg{})
		m = next.(model)
		seen[m.opts.Theme] = true
	}
	// After len(AvailableThemes) cycles we must have visited every theme.
	if len(seen) != len(AvailableThemes) {
		t.Fatalf("cycle visited %d of %d themes", len(seen), len(AvailableThemes))
	}
}

func TestThemeSetActionDispatchesMsg(t *testing.T) {
	deps := ActionDeps{
		RequestTheme: func(name ThemeName) tea.Cmd {
			return func() tea.Msg { return themeChangedMsg{Name: name} }
		},
	}
	reg := DefaultActionRegistry(deps)
	action, ok := reg.Get("theme.set")
	if !ok {
		t.Fatal("theme.set not registered")
	}
	// Unknown theme should error out.
	if _, err := action.Run(nil, "neonmode"); err == nil {
		t.Fatal("expected error for unknown theme")
	}
	// Known theme runs cleanly.
	if _, err := action.Run(nil, "light"); err != nil {
		t.Fatalf("theme.set light: %v", err)
	}
}

func TestNormalizedDefaultsTheme(t *testing.T) {
	got := (Options{}).normalized()
	if got.Theme != ThemeDark {
		t.Fatalf("default theme = %q, want %q", got.Theme, ThemeDark)
	}
}
