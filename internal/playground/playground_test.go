package playground

import "testing"

func TestPlayground(t *testing.T) {
	p := NewPlayground()

	session := p.CreateSession("s1", "gpt-4o")
	if session == nil {
		t.Fatal("Expected session")
	}
	if session.Model != "gpt-4o" {
		t.Errorf("Expected gpt-4o, got %s", session.Model)
	}

	p.AddMessage("s1", "user", "hello", 1)
	p.AddMessage("s1", "assistant", "hi there", 2)

	retrieved := p.GetSession("s1")
	if len(retrieved.History) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved.History))
	}

	cost := p.EstimateCost("s1")
	if cost == 0 {
		t.Error("Expected non-zero cost")
	}

	sessions := p.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}
}
