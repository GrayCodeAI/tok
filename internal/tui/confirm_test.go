package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func armedConfirm() *ConfirmOverlay {
	c := NewConfirmOverlay()
	c.Open(Action{ID: "logs.clear", Title: "Clear logs", Description: "test"}, "")
	return c
}

func TestConfirmYesAcceptsAndEmitsActionRequest(t *testing.T) {
	c := armedConfirm()
	cmd := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if c.IsOpen() {
		t.Fatal("y should close the overlay")
	}
	if cmd == nil {
		t.Fatal("y should emit cmd")
	}
	req, ok := cmd().(actionRequestMsg)
	if !ok {
		t.Fatalf("expected actionRequestMsg, got %T", cmd())
	}
	if req.ActionID != "logs.clear" {
		t.Fatalf("ActionID = %q, want logs.clear", req.ActionID)
	}
}

func TestConfirmEscRejects(t *testing.T) {
	c := armedConfirm()
	cmd := c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if c.IsOpen() {
		t.Fatal("esc should close")
	}
	if cmd != nil {
		t.Fatalf("esc should not emit a cmd, got %v", cmd())
	}
}

func TestConfirmEnterHonorsDefaultNo(t *testing.T) {
	c := armedConfirm()
	// default_ starts false (== no). Enter cancels.
	cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if c.IsOpen() {
		t.Fatal("enter with default=no should close")
	}
	if cmd != nil {
		t.Fatalf("enter with default=no should be a cancel, got cmd: %v", cmd())
	}
}

func TestConfirmTabSwitchesDefault(t *testing.T) {
	c := armedConfirm()
	c.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !c.default_ {
		t.Fatal("tab should flip default to yes")
	}
	cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if c.IsOpen() {
		t.Fatal("enter on yes default should close")
	}
	if cmd == nil {
		t.Fatal("enter on yes default should emit actionRequestMsg")
	}
	if _, ok := cmd().(actionRequestMsg); !ok {
		t.Fatalf("expected actionRequestMsg on accept, got %T", cmd())
	}
}

func TestConfirmingDestructiveActionFlowsThroughModal(t *testing.T) {
	loader := &stubLoader{}
	m := NewModelWithLoader(Options{}, loader).(model)

	// Dispatch actionRequestMsg for the destructive logs.clear action.
	// The model should arm the confirm overlay rather than running it.
	next, _ := m.Update(actionRequestMsg{ActionID: "logs.clear"})
	m = next.(model)
	if m.confirm == nil || !m.confirm.IsOpen() {
		t.Fatal("confirm overlay should be open for logs.clear")
	}
}
