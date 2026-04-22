package session

import (
	"testing"
)

func TestSessionCmdInitialized(t *testing.T) {
	if sessionCmd == nil {
		t.Fatal("sessionCmd should be initialized")
	}
	if sessionCmd.Use != "session" {
		t.Errorf("expected command name 'session', got %q", sessionCmd.Use)
	}
}

func TestSessionSubcommands(t *testing.T) {
	subs := []string{"start", "list", "active", "compact", "adoption"}
	for _, name := range subs {
		cmd, _, err := sessionCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("failed to find subcommand %q: %v", name, err)
		}
		if cmd.Use != name {
			t.Errorf("expected subcommand %q, got %q", name, cmd.Use)
		}
	}
}
