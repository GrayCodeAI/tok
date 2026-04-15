package session

import (
	"context"
	"os"
	"testing"
	"time"
)

func setupTestManager(t *testing.T) (*SessionManager, func()) {
	tmpDir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Set data path for test
	oldDataPath := os.Getenv("TOKMAN_DATA_PATH")
	os.Setenv("TOKMAN_DATA_PATH", tmpDir)

	// Update config data path
	dataDir = tmpDir

	sm, err := NewSessionManager()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create session manager: %v", err)
	}

	cleanup := func() {
		sm.Close()
		os.RemoveAll(tmpDir)
		os.Setenv("TOKMAN_DATA_PATH", oldDataPath)
	}

	return sm, cleanup
}

// dataDir is used to override config data path in tests
var dataDir string

func TestNewSessionManager(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	if sm == nil {
		t.Fatal("expected non-nil session manager")
	}
	if sm.db == nil {
		t.Error("expected non-nil database")
	}
	if sm.sessions == nil {
		t.Error("expected initialized sessions map")
	}
	if sm.hooks == nil {
		t.Error("expected initialized hooks map")
	}
	if sm.compressor == nil {
		t.Error("expected non-nil compressor")
	}
}

func TestSessionManager_CreateSession(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	session, err := sm.CreateSession("claude", "/test/project")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.ID == "" {
		t.Error("expected session ID to be set")
	}
	if session.Agent != "claude" {
		t.Errorf("expected agent='claude', got '%s'", session.Agent)
	}
	if session.ProjectPath != "/test/project" {
		t.Errorf("expected projectPath='/test/project', got '%s'", session.ProjectPath)
	}
	if !session.IsActive {
		t.Error("expected session to be active")
	}
	if len(session.ContextBlocks) != 0 {
		t.Error("expected empty context blocks")
	}
	if session.State.Variables == nil {
		t.Error("expected initialized variables map")
	}
}

func TestSessionManager_GetSession(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Create a session
	created, err := sm.CreateSession("test-agent", "/test/path")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Retrieve it
	retrieved, err := sm.GetSession(created.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Error("expected to retrieve same session")
	}

	// Try to get non-existent session
	_, err = sm.GetSession("non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestSessionManager_GetActiveSession(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Initially no active session
	if sm.GetActiveSession() != nil {
		t.Error("expected no active session initially")
	}

	// Create session - should become active
	session, _ := sm.CreateSession("agent", "/project")

	active := sm.GetActiveSession()
	if active == nil {
		t.Fatal("expected active session after creation")
	}
	if active.ID != session.ID {
		t.Error("expected active session to match created session")
	}
}

func TestSessionManager_SetActiveSession(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Create two sessions
	session1, _ := sm.CreateSession("agent1", "/project1")
	session2, _ := sm.CreateSession("agent2", "/project2")

	// Switch to session1
	err := sm.SetActiveSession(session1.ID)
	if err != nil {
		t.Errorf("failed to set active session: %v", err)
	}

	if sm.GetActiveSession().ID != session1.ID {
		t.Error("expected active session to be session1")
	}

	// Switch to session2
	err = sm.SetActiveSession(session2.ID)
	if err != nil {
		t.Errorf("failed to set active session: %v", err)
	}

	if sm.GetActiveSession().ID != session2.ID {
		t.Error("expected active session to be session2")
	}
}

func TestSessionManager_AddContextBlock(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Create session
	sm.CreateSession("agent", "/project")

	// Add context block
	err := sm.AddContextBlock("code", "package main", 100)
	if err != nil {
		t.Errorf("failed to add context block: %v", err)
	}

	// Add another
	err = sm.AddContextBlock("test", "test output here", 50)
	if err != nil {
		t.Errorf("failed to add context block: %v", err)
	}

	session := sm.GetActiveSession()
	if len(session.ContextBlocks) != 2 {
		t.Errorf("expected 2 context blocks, got %d", len(session.ContextBlocks))
	}
	if session.Metadata.TotalTurns != 2 {
		t.Errorf("expected TotalTurns=2, got %d", session.Metadata.TotalTurns)
	}
	if session.Metadata.TotalTokens != 150 {
		t.Errorf("expected TotalTokens=150, got %d", session.Metadata.TotalTokens)
	}
}

func TestSessionManager_AddContextBlock_NoActiveSession(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Don't create session, try to add block
	err := sm.AddContextBlock("code", "content", 100)
	if err == nil {
		t.Error("expected error when no active session")
	}
}

func TestSessionManager_RegisterHook(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	hook := func(ctx context.Context, session *Session, data interface{}) error {
		return nil
	}

	sm.RegisterHook(HookSessionStart, hook)

	// Create session - should trigger hook
	_, _ = sm.CreateSession("agent", "/project")

	// Give hook time to execute
	time.Sleep(100 * time.Millisecond)

	// Note: Hook runs in goroutine with timeout, may or may not complete
	// Just verify registration doesn't panic
}

func TestSessionManager_CleanupExpired(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	// Cleanup on empty database should not error
	err := sm.CleanupExpired()
	if err != nil {
		t.Errorf("cleanup failed: %v", err)
	}
}

func TestSessionManager_Close(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	// Don't defer cleanup, we'll close manually

	err := sm.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}

	// Clean up temp dir
	cleanup()
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("expected non-empty session ID")
	}
	if id1 == id2 {
		t.Error("expected unique session IDs")
	}
	if len(id1) != 32 {
		t.Errorf("expected ID length 32, got %d", len(id1))
	}
}
