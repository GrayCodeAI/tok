package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchOverlayOpenClose(t *testing.T) {
	s := NewSearchOverlay()
	if s.IsOpen() {
		t.Fatal("new overlay should be closed")
	}
	_ = s.Open()
	if !s.IsOpen() {
		t.Fatal("Open() should mark overlay as open")
	}
	s.Close()
	if s.IsOpen() {
		t.Fatal("Close() should mark overlay as closed")
	}
	if s.Query() != "" {
		t.Fatalf("Close() should reset query; got %q", s.Query())
	}
}

func TestSearchOverlayTypingDispatchesSearchMsg(t *testing.T) {
	s := NewSearchOverlay()
	_ = s.Open()

	cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatal("expected a cmd after typing")
	}
	// Drain the batch and find a searchMsg.
	found := drainForSearchMsg(cmd)
	if found == nil {
		t.Fatal("expected searchMsg in the emitted cmd batch")
	}
	if found.Query != "a" {
		t.Fatalf("searchMsg.Query = %q, want %q", found.Query, "a")
	}
}

func TestSearchOverlayEscCloses(t *testing.T) {
	s := NewSearchOverlay()
	_ = s.Open()
	cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.IsOpen() {
		t.Fatal("esc should close overlay")
	}
	if cmd == nil {
		t.Fatal("esc should emit searchCloseMsg")
	}
	msg := cmd()
	if _, ok := msg.(searchCloseMsg); !ok {
		t.Fatalf("expected searchCloseMsg, got %T", msg)
	}
}

// drainForSearchMsg runs the cmd and recursively unwraps tea.BatchMsg to
// find the first searchMsg. Returns nil if none is produced.
func drainForSearchMsg(cmd tea.Cmd) *searchMsg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	return findSearchMsg(msg)
}

func findSearchMsg(msg tea.Msg) *searchMsg {
	switch v := msg.(type) {
	case searchMsg:
		return &v
	case tea.BatchMsg:
		for _, c := range v {
			if found := drainForSearchMsg(c); found != nil {
				return found
			}
		}
	}
	return nil
}
