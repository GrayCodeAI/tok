// Package session provides session recovery functionality for crash resilience
package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/GrayCodeAI/tokman/internal/config"
)

// RecoveryManager manages session recovery and auto-save functionality
type RecoveryManager struct {
	dataDir          string
	autoSaveInterval time.Duration
	sessions         map[string]*RecoverableSession
	shutdown         chan struct{}
}

// RecoverableSession represents a session that can be recovered after crash
type RecoverableSession struct {
	ID           string                 `json:"id"`
	Agent        string                 `json:"agent"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	Commands     []RecoveredCommand     `json:"commands"`
	Context      map[string]interface{} `json:"context"`
	State        RecoveryState          `json:"recovery_state"`
	Checksum     string                 `json:"checksum"`
}

// RecoveredCommand represents a command that can be recovered
type RecoveredCommand struct {
	Command          string    `json:"command"`
	Timestamp        time.Time `json:"timestamp"`
	OriginalSize     int       `json:"original_size"`
	FilteredSize     int       `json:"filtered_size"`
	TokensSaved      int       `json:"tokens_saved"`
	WorkingDir       string    `json:"working_dir"`
	ExitCode         int       `json:"exit_code"`
	CompressedOutput string    `json:"compressed_output,omitempty"`
}

// RecoveryState represents the current recovery state of a session
type RecoveryState string

const (
	StateActive    RecoveryState = "active"
	StatePaused    RecoveryState = "paused"
	StateCrashed   RecoveryState = "crashed"
	StateRecovered RecoveryState = "recovered"
	StateClosed    RecoveryState = "closed"
)

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() (*RecoveryManager, error) {
	dataDir := filepath.Join(config.DataPath(), "recovery")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recovery directory: %w", err)
	}

	return &RecoveryManager{
		dataDir:          dataDir,
		autoSaveInterval: 30 * time.Second,
		sessions:         make(map[string]*RecoverableSession),
		shutdown:         make(chan struct{}),
	}, nil
}

// Start starts the auto-save goroutine
func (rm *RecoveryManager) Start() {
	go rm.autoSave()
	slog.Info("Session recovery manager started", "data_dir", rm.dataDir)
}

// Stop stops the auto-save goroutine
func (rm *RecoveryManager) Stop() {
	close(rm.shutdown)
	// Final save of all sessions
	rm.SaveAllSessions()
}

// autoSave periodically saves all active sessions
func (rm *RecoveryManager) autoSave() {
	ticker := time.NewTicker(rm.autoSaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.SaveAllSessions()
		case <-rm.shutdown:
			return
		}
	}
}

// CreateSession creates a new recoverable session
func (rm *RecoveryManager) CreateSession(id, agent string) *RecoverableSession {
	session := &RecoverableSession{
		ID:           id,
		Agent:        agent,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Commands:     make([]RecoveredCommand, 0),
		Context:      make(map[string]interface{}),
		State:        StateActive,
	}

	rm.sessions[id] = session
	rm.saveSession(session)

	slog.Info("Created recoverable session", "id", id, "agent", agent)
	return session
}

// GetSession retrieves a session by ID
func (rm *RecoveryManager) GetSession(id string) (*RecoverableSession, bool) {
	session, exists := rm.sessions[id]
	return session, exists
}

// AddCommand adds a command to a session
func (rm *RecoveryManager) AddCommand(sessionID string, cmd RecoveredCommand) error {
	session, exists := rm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Commands = append(session.Commands, cmd)
	session.LastActivity = time.Now()

	// Auto-save after each command
	go rm.saveSession(session)

	return nil
}

// UpdateContext updates session context
func (rm *RecoveryManager) UpdateContext(sessionID string, key string, value interface{}) error {
	session, exists := rm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Context[key] = value
	session.LastActivity = time.Now()

	return nil
}

// PauseSession pauses a session (preserves state but stops tracking)
func (rm *RecoveryManager) PauseSession(id string) error {
	session, exists := rm.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.State = StatePaused
	rm.saveSession(session)

	slog.Info("Session paused", "id", id)
	return nil
}

// ResumeSession resumes a paused session
func (rm *RecoveryManager) ResumeSession(id string) (*RecoverableSession, error) {
	session, exists := rm.sessions[id]
	if !exists {
		// Try to load from disk
		loaded, err := rm.loadSession(id)
		if err != nil {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		session = loaded
		rm.sessions[id] = session
	}

	session.State = StateActive
	session.LastActivity = time.Now()
	rm.saveSession(session)

	slog.Info("Session resumed", "id", id)
	return session, nil
}

// CloseSession closes a session
func (rm *RecoveryManager) CloseSession(id string) error {
	session, exists := rm.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.State = StateClosed
	rm.saveSession(session)

	// Remove from memory but keep on disk for history
	delete(rm.sessions, id)

	slog.Info("Session closed", "id", id)
	return nil
}

// ListRecoverableSessions lists all sessions that can be recovered
func (rm *RecoveryManager) ListRecoverableSessions() ([]*RecoverableSession, error) {
	var sessions []*RecoverableSession

	// List from memory
	for _, session := range rm.sessions {
		if session.State == StateActive || session.State == StateCrashed {
			sessions = append(sessions, session)
		}
	}

	// List from disk
	entries, err := os.ReadDir(rm.dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		id := entry.Name()
		if _, exists := rm.sessions[id]; exists {
			continue // Already in memory
		}

		session, err := rm.loadSession(id)
		if err != nil {
			slog.Warn("Failed to load session", "id", id, "error", err)
			continue
		}

		if session.State == StateActive || session.State == StateCrashed {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// RecoverSession recovers a crashed session
func (rm *RecoveryManager) RecoverSession(id string) (*RecoverableSession, error) {
	session, exists := rm.sessions[id]
	if !exists {
		// Try to load from disk
		loaded, err := rm.loadSession(id)
		if err != nil {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		session = loaded
	}

	if session.State != StateCrashed && session.State != StateActive {
		return nil, fmt.Errorf("session is not recoverable, state: %s", session.State)
	}

	session.State = StateRecovered
	session.LastActivity = time.Now()
	rm.sessions[id] = session
	rm.saveSession(session)

	slog.Info("Session recovered", "id", id, "commands", len(session.Commands))
	return session, nil
}

// CleanupOldSessions removes sessions older than the specified duration
func (rm *RecoveryManager) CleanupOldSessions(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	entries, err := os.ReadDir(rm.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			sessionID := entry.Name()
			path := filepath.Join(rm.dataDir, sessionID)

			if err := os.Remove(path); err != nil {
				slog.Warn("Failed to remove old session", "id", sessionID, "error", err)
			} else {
				slog.Info("Removed old session", "id", sessionID, "age", time.Since(info.ModTime()))
			}

			// Also remove from memory if present
			delete(rm.sessions, sessionID)
		}
	}

	return nil
}

// SaveAllSessions saves all active sessions to disk
func (rm *RecoveryManager) SaveAllSessions() {
	for _, session := range rm.sessions {
		if session.State == StateActive {
			rm.saveSession(session)
		}
	}
}

// saveSession saves a session to disk
func (rm *RecoveryManager) saveSession(session *RecoverableSession) error {
	path := filepath.Join(rm.dataDir, session.ID)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// loadSession loads a session from disk
func (rm *RecoveryManager) loadSession(id string) (*RecoverableSession, error) {
	path := filepath.Join(rm.dataDir, id)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session RecoverableSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetRecoveryStats returns recovery statistics
func (rm *RecoveryManager) GetRecoveryStats() RecoveryStats {
	stats := RecoveryStats{
		TotalSessions: len(rm.sessions),
	}

	for _, session := range rm.sessions {
		switch session.State {
		case StateActive:
			stats.ActiveSessions++
		case StatePaused:
			stats.PausedSessions++
		case StateCrashed:
			stats.CrashedSessions++
		}

		stats.TotalCommands += len(session.Commands)
	}

	return stats
}

// RecoveryStats holds recovery statistics
type RecoveryStats struct {
	TotalSessions   int `json:"total_sessions"`
	ActiveSessions  int `json:"active_sessions"`
	PausedSessions  int `json:"paused_sessions"`
	CrashedSessions int `json:"crashed_sessions"`
	TotalCommands   int `json:"total_commands"`
}

// GenerateRecoveryReport generates a recovery report
func (rm *RecoveryManager) GenerateRecoveryReport() string {
	stats := rm.GetRecoveryStats()

	report := fmt.Sprintf("Session Recovery Report\n")
	report += fmt.Sprintf("======================\n\n")
	report += fmt.Sprintf("Total Sessions: %d\n", stats.TotalSessions)
	report += fmt.Sprintf("Active: %d\n", stats.ActiveSessions)
	report += fmt.Sprintf("Paused: %d\n", stats.PausedSessions)
	report += fmt.Sprintf("Crashed: %d\n", stats.CrashedSessions)
	report += fmt.Sprintf("Total Commands: %d\n", stats.TotalCommands)

	return report
}

// CheckForCrashes checks for crashed sessions from previous runs
func (rm *RecoveryManager) CheckForCrashes() ([]*RecoverableSession, error) {
	entries, err := os.ReadDir(rm.dataDir)
	if err != nil {
		return nil, err
	}

	var crashed []*RecoverableSession

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		session, err := rm.loadSession(entry.Name())
		if err != nil {
			continue
		}

		// If session was active but not properly closed, mark as crashed
		if session.State == StateActive {
			session.State = StateCrashed
			rm.saveSession(session)
			crashed = append(crashed, session)
		}
	}

	return crashed, nil
}
