package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProvidersSectionList(t *testing.T) {
	s := newProvidersSection()
	ctx := fixtureDashCtxWithTrends()
	// Seed table.
	s.Update(ctx, tea.KeyMsg{})
	view := s.View(ctx)
	for _, want := range []string{"Providers", "anthropic", "openai", "Reduction"} {
		if !strings.Contains(view, want) {
			t.Fatalf("providers list missing %q:\n%s", want, view)
		}
	}
}

func TestProvidersSectionDrillDown(t *testing.T) {
	s := newProvidersSection()
	ctx := fixtureDashCtxWithTrends()
	s.Update(ctx, tea.KeyMsg{}) // seed
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEnter})
	if s.drill != "anthropic" {
		t.Fatalf("drill = %q, want anthropic", s.drill)
	}
	view := s.View(ctx)
	for _, want := range []string{"Provider: anthropic", "claude-opus-4-7", "claude-sonnet", "Reduction"} {
		if !strings.Contains(view, want) {
			t.Fatalf("detail view missing %q:\n%s", want, view)
		}
	}
	// Escape back to list.
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyEsc})
	if s.drill != "" {
		t.Fatal("esc should clear drill")
	}
}

func TestFilterProviderModels(t *testing.T) {
	ctx := fixtureDashCtxWithTrends()
	models := filterProviderModels(ctx.Data.Dashboard.TopProviderModels, "anthropic")
	if len(models) != 2 {
		t.Fatalf("expected 2 anthropic models, got %d", len(models))
	}
	openai := filterProviderModels(ctx.Data.Dashboard.TopProviderModels, "openai")
	if len(openai) != 1 {
		t.Fatalf("expected 1 openai model, got %d", len(openai))
	}
}
