package playground

import "testing"

func TestPlaygroundEngine(t *testing.T) {
	engine := NewPlaygroundEngine()

	session := engine.CreateSession("gpt-4o")
	if session == nil {
		t.Fatal("Expected non-nil session")
	}
	if session.Model != "gpt-4o" {
		t.Errorf("Expected gpt-4o, got %s", session.Model)
	}

	engine.AddMessage(session.ID, "user", "hello world")
	engine.AddMessage(session.ID, "assistant", "hi there!")

	retrieved := engine.GetSession(session.ID)
	if len(retrieved.History) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(retrieved.History))
	}

	cost := engine.EstimateCost(session.ID, 2.5, 10.0)
	if cost < 0 {
		t.Error("Expected non-negative cost")
	}
}

func TestPlaygroundEngineCompareModels(t *testing.T) {
	engine := NewPlaygroundEngine()
	session := engine.CreateSession("gpt-4o")
	engine.AddMessage(session.ID, "user", "test input")
	engine.AddMessage(session.ID, "assistant", "test output")

	results := engine.CompareModels(session.ID, []string{"gpt-4o", "claude-3-haiku"})
	if len(results) != 2 {
		t.Errorf("Expected 2 model results, got %d", len(results))
	}
}

func TestPlaygroundEngineExport(t *testing.T) {
	engine := NewPlaygroundEngine()
	session := engine.CreateSession("gpt-4o")
	engine.AddMessage(session.ID, "user", "hello")

	export := engine.ExportSession(session.ID)
	if export == "" {
		t.Error("Expected non-empty export")
	}
}

func TestPlaygroundEngineList(t *testing.T) {
	engine := NewPlaygroundEngine()
	engine.CreateSession("gpt-4o")
	engine.CreateSession("claude-3-haiku")

	sessions := engine.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}
