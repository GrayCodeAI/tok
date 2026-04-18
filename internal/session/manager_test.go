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
	oldDataPath := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)

	sm, err := NewSessionManager()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create session manager: %v", err)
	}

	cleanup := func() {
		sm.Close()
		os.RemoveAll(tmpDir)
		os.Setenv("XDG_DATA_HOME", oldDataPath)
	}

	return sm, cleanup
}

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

	sm.mu.Lock()
	delete(sm.sessions, created.ID)
	sm.mu.Unlock()

	reloaded, err := sm.GetSession(created.ID)
	if err != nil {
		t.Fatalf("failed to reload session from database: %v", err)
	}
	if reloaded.ID != created.ID {
		t.Error("expected reloaded session to match created session")
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

func TestSessionManager_PersistsExpiresAt(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	session, err := sm.CreateSession("agent", "/project")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	expiresAt := time.Now().Add(2 * time.Hour).Round(time.Second)
	session.ExpiresAt = &expiresAt
	if err := sm.persistSession(session); err != nil {
		t.Fatalf("persistSession: %v", err)
	}

	sm.mu.Lock()
	delete(sm.sessions, session.ID)
	sm.mu.Unlock()

	reloaded, err := sm.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if reloaded.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be persisted")
	}
	if !reloaded.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", reloaded.ExpiresAt, expiresAt)
	}
}

func TestSessionManager_GetSessionFailsOnCorruptStoredJSON(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	session, err := sm.CreateSession("agent", "/project")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	_, err = sm.db.Exec(`UPDATE sessions SET metadata = ? WHERE id = ?`, "{not-json", session.ID)
	if err != nil {
		t.Fatalf("corrupt metadata update: %v", err)
	}

	sm.mu.Lock()
	delete(sm.sessions, session.ID)
	sm.mu.Unlock()

	_, err = sm.GetSession(session.ID)
	if err == nil {
		t.Fatal("expected GetSession to fail on corrupt stored JSON")
	}
}

func TestSessionManager_GetSummary(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	first, err := sm.CreateSession("Claude Code", "/project-a")
	if err != nil {
		t.Fatalf("CreateSession(first): %v", err)
	}
	second, err := sm.CreateSession("Claude Code", "/project-b")
	if err != nil {
		t.Fatalf("CreateSession(second): %v", err)
	}
	second.IsActive = false
	if err := sm.persistSession(second); err != nil {
		t.Fatalf("persistSession(second): %v", err)
	}
	if _, err := sm.CreateSnapshot(first.ID); err != nil {
		t.Fatalf("CreateSnapshot(): %v", err)
	}

	summary, err := sm.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary(): %v", err)
	}
	if summary.TotalSessions != 2 {
		t.Fatalf("TotalSessions = %d, want 2", summary.TotalSessions)
	}
	if summary.ActiveSessions != 1 {
		t.Fatalf("ActiveSessions = %d, want 1", summary.ActiveSessions)
	}
	if summary.SnapshotCount != 1 {
		t.Fatalf("SnapshotCount = %d, want 1", summary.SnapshotCount)
	}
	if summary.TopAgent != "Claude Code" {
		t.Fatalf("TopAgent = %q, want Claude Code", summary.TopAgent)
	}
	if summary.TopAgentCount != 2 {
		t.Fatalf("TopAgentCount = %d, want 2", summary.TopAgentCount)
	}
	if summary.ActiveSessionID == "" {
		t.Fatal("expected ActiveSessionID to be set")
	}
}

func TestSessionManager_GetAnalyticsSnapshot(t *testing.T) {
	sm, cleanup := setupTestManager(t)
	defer cleanup()

	first, err := sm.CreateSession("Claude Code", "/project-a")
	if err != nil {
		t.Fatalf("CreateSession(first): %v", err)
	}
	if err := sm.AddContextBlock(BlockTypeUserQuery, "open file", 40); err != nil {
		t.Fatalf("AddContextBlock(user): %v", err)
	}
	if err := sm.AddContextBlock(BlockTypeToolResult, "file contents", 60); err != nil {
		t.Fatalf("AddContextBlock(tool): %v", err)
	}
	if _, err := sm.CreateSnapshot(first.ID); err != nil {
		t.Fatalf("CreateSnapshot(first): %v", err)
	}

	second, err := sm.CreateSession("Cursor", "/project-b")
	if err != nil {
		t.Fatalf("CreateSession(second): %v", err)
	}
	second.IsActive = false
	if err := sm.persistSession(second); err != nil {
		t.Fatalf("persistSession(second): %v", err)
	}

	snapshot, err := sm.GetAnalyticsSnapshot(SessionListOptions{Limit: 10}, 10)
	if err != nil {
		t.Fatalf("GetAnalyticsSnapshot(): %v", err)
	}
	if snapshot.StoreSummary.TotalSessions != 2 {
		t.Fatalf("StoreSummary.TotalSessions = %d, want 2", snapshot.StoreSummary.TotalSessions)
	}
	if len(snapshot.RecentSessions) != 2 {
		t.Fatalf("len(RecentSessions) = %d, want 2", len(snapshot.RecentSessions))
	}
	foundSnapshot := false
	for _, item := range snapshot.RecentSessions {
		if item.ID == first.ID && item.SnapshotCount == 1 {
			foundSnapshot = true
		}
	}
	if !foundSnapshot {
		t.Fatalf("expected recent session overview to include snapshot count for %q: %+v", first.ID, snapshot.RecentSessions)
	}
	if len(snapshot.SnapshotHistory) != 1 {
		t.Fatalf("len(SnapshotHistory) = %d, want 1", len(snapshot.SnapshotHistory))
	}
	if snapshot.SnapshotHistory[0].SessionID != first.ID {
		t.Fatalf("SnapshotHistory[0].SessionID = %q, want %q", snapshot.SnapshotHistory[0].SessionID, first.ID)
	}
	if snapshot.ActiveContext == nil {
		t.Fatal("expected ActiveContext")
	}
	if snapshot.ActiveContext.SessionID != second.ID {
		t.Fatalf("ActiveContext.SessionID = %q, want %q", snapshot.ActiveContext.SessionID, second.ID)
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
