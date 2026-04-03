package undo

import (
	"errors"
	"testing"
	"time"
)

func TestNewUndoManager(t *testing.T) {
	um := NewUndoManager(50)
	if um.maxHistory != 50 {
		t.Errorf("maxHistory = %d, want 50", um.maxHistory)
	}
	if um.CanUndo() {
		t.Error("new manager should not be able to undo")
	}
	if um.CanRedo() {
		t.Error("new manager should not be able to redo")
	}
}

func TestNewUndoManagerDefaultMax(t *testing.T) {
	um := NewUndoManager(0)
	if um.maxHistory != 100 {
		t.Errorf("maxHistory = %d, want 100 (default)", um.maxHistory)
	}
}

func TestRecord(t *testing.T) {
	um := NewUndoManager(10)

	action := Action{
		Command:     "git",
		Args:        []string{"commit", "-m", "test"},
		Description: "commit changes",
	}
	um.Record(action)

	if !um.CanUndo() {
		t.Fatal("should be able to undo after recording")
	}

	history := um.GetHistory()
	if len(history) != 1 {
		t.Errorf("history length = %d, want 1", len(history))
	}
	if history[0].Command != "git" {
		t.Errorf("command = %q, want %q", history[0].Command, "git")
	}
	if history[0].ID == "" {
		t.Error("auto-generated ID should not be empty")
	}
}

func TestUndo(t *testing.T) {
	um := NewUndoManager(10)

	um.Record(Action{Command: "cmd1"})
	um.Record(Action{Command: "cmd2"})

	action, err := um.Undo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Command != "cmd2" {
		t.Errorf("undone command = %q, want %q", action.Command, "cmd2")
	}

	if !um.CanUndo() {
		t.Error("should still be able to undo (cmd1)")
	}
	if !um.CanRedo() {
		t.Error("should be able to redo")
	}
}

func TestUndoNothingToUndo(t *testing.T) {
	um := NewUndoManager(10)

	_, err := um.Undo()
	if err == nil {
		t.Error("expected error when nothing to undo")
	}
}

func TestUndoWithRollback(t *testing.T) {
	um := NewUndoManager(10)
	rollbackCalled := false

	um.Record(Action{
		Command: "cmd",
		Rollback: func() error {
			rollbackCalled = true
			return nil
		},
	})

	_, err := um.Undo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rollbackCalled {
		t.Error("rollback function should have been called")
	}
}

func TestUndoRollbackFailure(t *testing.T) {
	um := NewUndoManager(10)

	um.Record(Action{
		Command: "cmd",
		Rollback: func() error {
			return errors.New("rollback failed")
		},
	})

	_, err := um.Undo()
	if err == nil {
		t.Error("expected error when rollback fails")
	}
}

func TestRedo(t *testing.T) {
	um := NewUndoManager(10)

	um.Record(Action{Command: "cmd1"})
	um.Record(Action{Command: "cmd2"})

	// Undo both
	_, _ = um.Undo()
	_, _ = um.Undo()

	action, err := um.Redo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Command != "cmd1" {
		t.Errorf("redone command = %q, want %q", action.Command, "cmd1")
	}
}

func TestRedoNothingToRedo(t *testing.T) {
	um := NewUndoManager(10)

	_, err := um.Redo()
	if err == nil {
		t.Error("expected error when nothing to redo")
	}
}

func TestNewActionClearsRedoStack(t *testing.T) {
	um := NewUndoManager(10)

	um.Record(Action{Command: "cmd1"})
	um.Record(Action{Command: "cmd2"})

	// Undo to populate redo stack
	_, _ = um.Undo()
	if !um.CanRedo() {
		t.Fatal("should be able to redo")
	}

	// Record new action - should clear redo stack
	um.Record(Action{Command: "cmd3"})

	if um.CanRedo() {
		t.Error("recording new action should clear redo stack")
	}
}

func TestMaxHistoryTrim(t *testing.T) {
	um := NewUndoManager(3)

	um.Record(Action{Command: "cmd1"})
	um.Record(Action{Command: "cmd2"})
	um.Record(Action{Command: "cmd3"})
	um.Record(Action{Command: "cmd4"})

	if um.CanUndo() {
		history := um.GetHistory()
		if len(history) != 3 {
			t.Errorf("history length = %d, want 3 (after trim)", len(history))
		}
		if history[0].Command != "cmd2" {
			t.Errorf("oldest command = %q, want %q", history[0].Command, "cmd2")
		}
	}
}

func TestClear(t *testing.T) {
	um := NewUndoManager(10)

	um.Record(Action{Command: "cmd1"})
	um.Record(Action{Command: "cmd2"})

	um.Clear()

	if um.CanUndo() {
		t.Error("should not be able to undo after clear")
	}
	if um.CanRedo() {
		t.Error("should not be able to redo after clear")
	}
	if len(um.GetHistory()) != 0 {
		t.Error("history should be empty after clear")
	}
}

func TestGenerateReport(t *testing.T) {
	um := NewUndoManager(10)

	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)
	t3 := time.Now()

	um.Record(Action{Command: "cmd1", Description: "first", Timestamp: t1})
	um.Record(Action{Command: "cmd2", Description: "second", Timestamp: t2})
	um.Record(Action{Command: "cmd3", Description: "third", Timestamp: t3})

	report := um.GenerateReport()
	if report.TotalActions != 3 {
		t.Errorf("totalActions = %d, want 3", report.TotalActions)
	}
	if !report.OldestAction.Equal(t1) {
		t.Errorf("oldestAction = %v, want %v", report.OldestAction, t1)
	}
	if len(report.Actions) != 3 {
		t.Errorf("actions count = %d, want 3", len(report.Actions))
	}
	if report.Actions[0].Command != "cmd1" {
		t.Errorf("first action command = %q, want %q", report.Actions[0].Command, "cmd1")
	}
}

func TestGenerateReportEmpty(t *testing.T) {
	um := NewUndoManager(10)

	report := um.GenerateReport()
	if report.TotalActions != 0 {
		t.Errorf("totalActions = %d, want 0", report.TotalActions)
	}
	if report.OldestAction.IsZero() != true {
		t.Error("oldestAction should be zero for empty history")
	}
}

func TestConcurrentAccess(t *testing.T) {
	um := NewUndoManager(1000)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				um.Record(Action{Command: "cmd"})
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// No race conditions should occur
	history := um.GetHistory()
	if len(history) == 0 {
		t.Error("expected non-empty history")
	}
}
