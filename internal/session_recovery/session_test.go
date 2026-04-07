package session_recovery

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestStore(t *testing.T) (*RecoveryStore, func()) {
	tmpDir, err := os.MkdirTemp("", "recovery-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	cfg := Config{
		BaseDir: tmpDir,
		Enabled: true,
	}

	store, err := New(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create store: %v", err)
	}

	cleanup := func() { store = nil; os.RemoveAll(tmpDir) }
	return store, cleanup
}

func TestNewDisabled(t *testing.T) {
	cfg := Config{Enabled: false}
	store, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store != nil {
		t.Error("expected nil store when disabled")
	}
}

func TestBeginSession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	session, err := store.BeginSession("/tmp/test", []string{"PATH=/usr/bin", "HOME=/home/user"})
	if err != nil {
		t.Fatalf("begin session: %v", err)
	}

	if session.ID == "" {
		t.Fatal("session ID is empty")
	}
	if session.CWD != "/tmp/test" {
		t.Errorf("cwd = %s, want /tmp/test", session.CWD)
	}
	if len(session.Env) != 2 {
		t.Errorf("env count = %d, want 2", len(session.Env))
	}
	if session.Env["PATH"] != "/usr/bin" {
		t.Errorf("PATH env = %s, want /usr/bin", session.Env["PATH"])
	}
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if len(id1) != 8 {
		t.Errorf("ID length = %d, want 8", len(id1))
	}
	if id1 == id2 {
		// Very unlikely but possible with tiny IDs
		t.Skip("collision (extremely unlikely)")
	}
}

func TestRecordCommand(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	session, err := store.BeginSession("/tmp/test", nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	cmd := CommandEntry{
		Command:     "git status",
		Output:      "M file.go",
		TokensIn:    100,
		TokensOut:   10,
		SavedTokens: 90,
	}

	err = store.RecordCommand(session.ID, cmd)
	if err != nil {
		t.Fatalf("record command: %v", err)
	}

	// Verify by loading
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions))
	}

	if len(sessions[0].Commands) != 1 {
		t.Errorf("commands = %d, want 1", len(sessions[0].Commands))
	}

	if sessions[0].Commands[0].Command != "git status" {
		t.Errorf("command = %s, want git status", sessions[0].Commands[0].Command)
	}
}

func TestCheckRecovery(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// No session yet
	info, err := store.CheckRecovery()
	if err != nil {
		t.Fatalf("check recovery: %v", err)
	}
	if info.CanResume {
		t.Error("expected CanResume = false with no sessions")
	}

	// Create a session with commands
	session, err := store.BeginSession("/tmp/test", nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	for i := 0; i < 3; i++ {
		store.RecordCommand(session.ID, CommandEntry{
			Command:   "git log -n 1",
			Output:    "abc1234",
			TokensIn:  100,
			TokensOut: 10,
		})
	}

	// Check recovery
	info, err = store.CheckRecovery()
	if err != nil {
		t.Fatalf("check recovery: %v", err)
	}
	if !info.CanResume {
		t.Error("expected CanResume = true")
	}
	if info.Session == nil {
		t.Fatal("expected session in recovery info")
	}
	if len(info.InterruptedCommands) != 3 {
		t.Errorf("interrupted commands = %d, want 3", len(info.InterruptedCommands))
	}
	if info.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestListSessions(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session, _ := store.BeginSession("/tmp/test", nil)
		store.RecordCommand(session.ID, CommandEntry{
			Command: "test cmd",
		})
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("sessions = %d, want 3", len(sessions))
	}
}

func TestCloseSession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	session, _ := store.BeginSession("/tmp/test", nil)
	sessionID := session.ID

	// Verify session exists
	sessions, _ := store.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session before close")
	}

	// Close session
	err := store.CloseSession(sessionID)
	if err != nil {
		t.Fatalf("close session: %v", err)
	}

	// Verify it's archived (not in active list)
	sessions, _ = store.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("sessions after close = %d, want 0", len(sessions))
	}

	// Verify archive file exists
	archivePath := filepath.Join(store.baseDir, "archive", sessionID+".json")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("archive file should exist after close")
	}
}

func TestPruneOldSessions(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		session, _ := store.BeginSession("/tmp/test", nil)
		store.RecordCommand(session.ID, CommandEntry{Command: "test"})
	}

	// Prune to max 2
	removed, err := store.PruneOldSessions(2)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 3 {
		t.Errorf("removed = %d, want 3", removed)
	}

	sessions, _ := store.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("remaining sessions = %d, want 2", len(sessions))
	}
}

func TestCreateCheckpoint(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	session, _ := store.BeginSession("/tmp/test", nil)

	// Record some commands
	for i := 0; i < 3; i++ {
		store.RecordCommand(session.ID, CommandEntry{
			Command: "cmd " + string(rune('A'+i)),
		})
	}

	// Create checkpoint
	err := store.CreateCheckpoint(session.ID)
	if err != nil {
		t.Fatalf("checkpoint: %v", err)
	}

	// Verify checkpoint was saved at command count
	sessions, _ := store.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions))
	}

	if sessions[0].Checkpoint != 3 {
		t.Errorf("checkpoint = %d, want 3", sessions[0].Checkpoint)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("expected enabled by default")
	}
	if cfg.MaxSessions != 10 {
		t.Errorf("max sessions = %d, want 10", cfg.MaxSessions)
	}
	if cfg.CheckpointInterval == 0 {
		t.Error("expected non-zero checkpoint interval")
	}
	if cfg.BaseDir == "" {
		t.Error("expected non-empty base dir")
	}
}
