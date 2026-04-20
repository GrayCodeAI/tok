package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigSectionShowsHookStatus(t *testing.T) {
	// Point TOK_CONFIG_DIR at a clean temp dir so the test is hermetic
	// and doesn't flip the user's actual tok hook flag file.
	dir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", dir)

	ctx := fixtureDashCtxWithTrends()
	s := newConfigSection()
	view := s.View(ctx)
	for _, want := range []string{"Config", "Tok hook", "Paths", "Data quality", "Active filters"} {
		if !strings.Contains(view, want) {
			t.Fatalf("config view missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "inactive") {
		t.Fatalf("fresh temp dir: hook should be inactive:\n%s", view)
	}
}

func TestConfigToggleKeyEmitsActionRequest(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOK_CONFIG_DIR", dir)

	s := newConfigSection()
	ctx := fixtureDashCtxWithTrends()

	_, cmd := s.Update(ctx, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if cmd == nil {
		t.Fatal("expected cmd from 't' press")
	}
	msg := cmd()
	req, ok := msg.(actionRequestMsg)
	if !ok {
		t.Fatalf("expected actionRequestMsg, got %T", msg)
	}
	if req.ActionID != "hooks.toggle" {
		t.Fatalf("ActionID = %q, want hooks.toggle", req.ActionID)
	}

	// Direct-run the registered action and verify the flag file landed.
	deps := ActionDeps{
		RequestToast: func(ToastKind, string) tea.Cmd { return nil },
	}
	reg := DefaultActionRegistry(deps)
	action, ok := reg.Get("hooks.toggle")
	if !ok {
		t.Fatal("hooks.toggle not registered")
	}
	if _, err := action.Run(nil, ""); err != nil {
		t.Fatalf("toggle failed: %v", err)
	}
	// Flag path uses TOK_CONFIG_DIR — check for the file.
	flag := filepath.Join(dir, ".tok-active")
	if _, err := os.Stat(flag); err != nil {
		t.Fatalf("expected flag file at %s after toggle: %v", flag, err)
	}
}
