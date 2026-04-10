package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/GrayCodeAI/tokman/internal/compression"
	"github.com/GrayCodeAI/tokman/internal/config"
)

// SessionManager manages sessions and their state
type SessionManager struct {
	db            *sql.DB
	sessions      map[string]*Session
	hooks         map[HookType][]Hook
	activeSession string
	compressor    *compression.BrotliCompressor
	mu            sync.RWMutex
}

// HookType represents the type of hook
type HookType string

const (
	HookSessionStart HookType = "session_start"
	HookPreToolUse   HookType = "pre_tool_use"
	HookPostToolUse  HookType = "post_tool_use"
	HookPreCompact   HookType = "pre_compact"
)

// Hook is a function that gets called at specific points
type Hook func(ctx context.Context, session *Session, data interface{}) error

// NewSessionManager creates a new session manager
func NewSessionManager() (*SessionManager, error) {
	dataDir := config.DataPath()
	dbPath := filepath.Join(dataDir, "sessions.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}

	sm := &SessionManager{
		db:         db,
		sessions:   make(map[string]*Session),
		hooks:      make(map[HookType][]Hook),
		compressor: compression.NewBrotliCompressor(),
	}

	if err := sm.initializeSchema(); err != nil {
		return nil, err
	}

	return sm, nil
}

// initializeSchema creates the database tables
func (sm *SessionManager) initializeSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    agent TEXT,
    project_path TEXT,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME DEFAULT CURRENT_TIMESTAMP,
    context_blocks TEXT, -- JSON array
    state TEXT, -- JSON object
    metadata TEXT, -- JSON object
    expires_at DATETIME,
    is_active BOOLEAN DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_sessions_agent ON sessions(agent);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_path);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_sessions_activity ON sessions(last_activity);

CREATE TABLE IF NOT EXISTS session_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    content TEXT,
    token_count INTEGER,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snapshots_session ON session_snapshots(session_id);
`
	_, err := sm.db.Exec(schema)
	return err
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(agent, projectPath string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:            generateSessionID(),
		Agent:         agent,
		ProjectPath:   projectPath,
		StartedAt:     time.Now(),
		LastActivity:  time.Now(),
		ContextBlocks: []ContextBlock{},
		State: SessionState{
			Variables: make(map[string]interface{}),
		},
		Metadata: SessionMetadata{
			TotalTurns:  0,
			TotalTokens: 0,
		},
		IsActive: true,
	}

	// Persist to database
	if err := sm.persistSession(session); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	sm.sessions[session.ID] = session
	sm.activeSession = session.ID

	// Call session_start hooks
	go sm.executeHooks(context.Background(), HookSessionStart, session, nil)

	slog.Info("Created new session", "id", session.ID[:8], "agent", agent)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if session, ok := sm.sessions[id]; ok {
		return session, nil
	}

	// Try to load from database
	session, err := sm.loadSession(id)
	if err != nil {
		return nil, err
	}

	sm.sessions[id] = session
	return session, nil
}

// GetActiveSession returns the currently active session
func (sm *SessionManager) GetActiveSession() *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.activeSession == "" {
		return nil
	}

	return sm.sessions[sm.activeSession]
}

// SetActiveSession sets the active session
func (sm *SessionManager) SetActiveSession(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.sessions[id]; !ok {
		// Try to load from database
		session, err := sm.loadSession(id)
		if err != nil {
			return err
		}
		sm.sessions[id] = session
	}

	sm.activeSession = id
	return nil
}

// AddContextBlock adds a context block to the active session
func (sm *SessionManager) AddContextBlock(blockType ContextBlockType, content string, tokens int) error {
	session := sm.GetActiveSession()
	if session == nil {
		return fmt.Errorf("no active session")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	block := ContextBlock{
		Type:      blockType,
		Content:   content,
		Timestamp: time.Now(),
		Tokens:    tokens,
	}

	session.ContextBlocks = append(session.ContextBlocks, block)
	session.LastActivity = time.Now()
	session.Metadata.TotalTurns++
	session.Metadata.TotalTokens += tokens

	// Persist update
	return sm.persistSession(session)
}

// PreCompact executes PreCompact hooks and returns optimized context
func (sm *SessionManager) PreCompact(ctx context.Context, maxTokens int) (string, error) {
	session := sm.GetActiveSession()
	if session == nil {
		return "", fmt.Errorf("no active session")
	}

	// Build context summary
	summary := sm.buildContextSummary(session, maxTokens)

	// Execute PreCompact hooks
	hookData := map[string]interface{}{
		"summary":   summary,
		"maxTokens": maxTokens,
		"session":   session,
	}

	if err := sm.executeHooks(ctx, HookPreCompact, session, hookData); err != nil {
		slog.Error("PreCompact hook failed", "error", err)
	}

	return summary, nil
}

// buildContextSummary creates an optimized summary of session context
func (sm *SessionManager) buildContextSummary(session *Session, maxTokens int) string {
	var summary string

	// Add session header
	summary += fmt.Sprintf("# Session Context\n")
	summary += fmt.Sprintf("Agent: %s\n", session.Agent)
	summary += fmt.Sprintf("Project: %s\n", session.ProjectPath)
	summary += fmt.Sprintf("Turns: %d | Tokens: %d\n\n", session.Metadata.TotalTurns, session.Metadata.TotalTokens)

	// Add state information
	if session.State.Focus != "" {
		summary += fmt.Sprintf("**Current Focus:** %s\n\n", session.State.Focus)
	}
	if session.State.NextAction != "" {
		summary += fmt.Sprintf("**Next Action:** %s\n\n", session.State.NextAction)
	}

	// Add recent context blocks (most recent first, within token budget)
	tokenCount := 0
	summary += "## Recent Activity\n\n"

	for i := len(session.ContextBlocks) - 1; i >= 0; i-- {
		block := session.ContextBlocks[i]
		if tokenCount+block.Tokens > maxTokens {
			break
		}

		summary += fmt.Sprintf("**%s** (%s)\n%s\n\n",
			block.Type,
			block.Timestamp.Format("15:04:05"),
			block.Content)

		tokenCount += block.Tokens
	}

	return summary
}

// CreateSnapshot creates a snapshot of the current session
func (sm *SessionManager) CreateSnapshot(sessionID string) (*SessionSnapshot, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Serialize session state
	content, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	// Compress
	compressed, err := sm.compressor.CompressWithMetadata(content)
	if err != nil {
		return nil, fmt.Errorf("failed to compress session: %w", err)
	}

	// Save to database
	result, err := sm.db.Exec(
		"INSERT INTO session_snapshots (session_id, content, token_count) VALUES (?, ?, ?)",
		sessionID, compressed.Data, session.Metadata.TotalTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	snapshotID, _ := result.LastInsertId()

	snapshot := &SessionSnapshot{
		ID:         snapshotID,
		SessionID:  sessionID,
		CreatedAt:  time.Now(),
		Content:    string(content),
		TokenCount: session.Metadata.TotalTokens,
	}

	slog.Info("Created session snapshot", "session", sessionID[:8], "snapshot", snapshotID)
	return snapshot, nil
}

// RestoreSnapshot restores a session from a snapshot
func (sm *SessionManager) RestoreSnapshot(snapshotID int64) (*Session, error) {
	var sessionID string
	var content []byte

	err := sm.db.QueryRow(
		"SELECT session_id, content FROM session_snapshots WHERE id = ?",
		snapshotID).Scan(&sessionID, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot: %w", err)
	}

	// Decompress
	decompressed, err := sm.compressor.Decompress(content)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress snapshot: %w", err)
	}

	// Unmarshal
	var session Session
	if err := json.Unmarshal(decompressed, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	session.LastActivity = time.Now()

	// Save as new session
	session.ID = generateSessionID()
	if err := sm.persistSession(&session); err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.sessions[session.ID] = &session
	sm.mu.Unlock()

	slog.Info("Restored session from snapshot", "original", sessionID[:8], "new", session.ID[:8])
	return &session, nil
}

// RegisterHook registers a hook for a specific event
func (sm *SessionManager) RegisterHook(hookType HookType, hook Hook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.hooks[hookType] = append(sm.hooks[hookType], hook)
}

// executeHooks executes all registered hooks for an event
func (sm *SessionManager) executeHooks(ctx context.Context, hookType HookType, session *Session, data interface{}) error {
	sm.mu.RLock()
	hooks := sm.hooks[hookType]
	sm.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx, session, data); err != nil {
			slog.Error("Hook execution failed", "type", hookType, "error", err)
			// Continue with other hooks
		}
	}

	return nil
}

// persistSession saves a session to the database
func (sm *SessionManager) persistSession(session *Session) error {
	contextBlocks, _ := json.Marshal(session.ContextBlocks)
	state, _ := json.Marshal(session.State)
	metadata, _ := json.Marshal(session.Metadata)

	_, err := sm.db.Exec(`
		INSERT OR REPLACE INTO sessions 
		(id, agent, project_path, started_at, last_activity, context_blocks, state, metadata, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.Agent, session.ProjectPath, session.StartedAt,
		session.LastActivity, contextBlocks, state, metadata, session.IsActive)

	return err
}

// loadSession loads a session from the database
func (sm *SessionManager) loadSession(id string) (*Session, error) {
	var session Session
	var contextBlocks, state, metadata []byte

	err := sm.db.QueryRow(`
		SELECT id, agent, project_path, started_at, last_activity, context_blocks, state, metadata, is_active
		FROM sessions WHERE id = ?
	`, id).Scan(&session.ID, &session.Agent, &session.ProjectPath, &session.StartedAt,
		&session.LastActivity, &contextBlocks, &state, &metadata, &session.IsActive)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(contextBlocks, &session.ContextBlocks)
	json.Unmarshal(state, &session.State)
	json.Unmarshal(metadata, &session.Metadata)

	return &session, nil
}

// CleanupExpired removes expired sessions
func (sm *SessionManager) CleanupExpired() error {
	_, err := sm.db.Exec("DELETE FROM sessions WHERE expires_at IS NOT NULL AND expires_at < ?", time.Now())
	return err
}

// Close closes the session manager
func (sm *SessionManager) Close() error {
	return sm.db.Close()
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
