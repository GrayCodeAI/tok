package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMapCoversShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Fatal("ShortHelp is empty")
	}
	if len(short) > 6 {
		t.Fatalf("ShortHelp = %d bindings; keep ≤6 for the footer strip", len(short))
	}
	for i, b := range short {
		h := b.Help()
		if h.Key == "" || h.Desc == "" {
			t.Fatalf("ShortHelp[%d] missing help text: %+v", i, h)
		}
	}
}

func TestKeyMapMatchesNavAndRefresh(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name    string
		keys    []string
		binding key.Binding
	}{
		{"next via tab", []string{"tab"}, km.NextSection},
		{"next via l", []string{"l"}, km.NextSection},
		{"prev via shift+tab", []string{"shift+tab"}, km.PrevSection},
		{"prev via h", []string{"h"}, km.PrevSection},
		{"refresh", []string{"r"}, km.Refresh},
		{"help", []string{"?"}, km.Help},
		{"quit via q", []string{"q"}, km.Quit},
		{"quit via ctrl+c", []string{"ctrl+c"}, km.Quit},
		{"palette", []string{":"}, km.Palette},
		{"search", []string{"/"}, km.Search},
		{"jump-1", []string{"1"}, km.JumpSection},
		{"jump-9", []string{"9"}, km.JumpSection},
	}
	for _, tc := range tests {
		msg := teaKeyFromString(tc.keys[0])
		if !key.Matches(msg, tc.binding) {
			t.Errorf("%s: key %q did not match binding %v", tc.name, tc.keys[0], tc.binding.Keys())
		}
	}
}

// teaKeyFromString constructs a tea.KeyMsg that matches `key.Matches` against
// a bubbles binding. Bubbletea parses keys via its own runes/type encoding;
// for tests we map a subset that covers every binding in DefaultKeyMap.
func teaKeyFromString(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestHandleKeyDispatchesViaRegistry(t *testing.T) {
	loader := &stubLoader{}
	m := NewModelWithLoader(Options{}, loader).(model)

	// tab → next section
	m, _ = m.handleKey(teaKeyFromString("tab"))
	if m.navIndex != 1 {
		t.Fatalf("tab: navIndex=%d, want 1", m.navIndex)
	}
	// shift+tab → prev
	m, _ = m.handleKey(teaKeyFromString("shift+tab"))
	if m.navIndex != 0 {
		t.Fatalf("shift+tab: navIndex=%d, want 0", m.navIndex)
	}
	// ? → toggle help
	m, _ = m.handleKey(teaKeyFromString("?"))
	if !m.helpOpen {
		t.Fatal("? should open help")
	}
	m, _ = m.handleKey(teaKeyFromString("?"))
	if m.helpOpen {
		t.Fatal("? should toggle help closed")
	}
	// esc → close help even if toggled open
	m, _ = m.handleKey(teaKeyFromString("?"))
	m, _ = m.handleKey(teaKeyFromString("esc"))
	if m.helpOpen {
		t.Fatal("esc should close help")
	}
	// jump shortcut
	m, _ = m.handleKey(teaKeyFromString("5"))
	if m.navIndex != 4 {
		t.Fatalf("jump 5: navIndex=%d, want 4", m.navIndex)
	}
}
