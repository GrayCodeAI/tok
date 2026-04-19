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
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/lakshmanpatel/tok/internal/compression"
	"github.com/lakshmanpatel/tok/internal/config"
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

	// Call session_start hooks with timeout to prevent goroutine leaks
	hookCtx, hookCancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer hookCancel()
		sm.executeHooks(hookCtx, HookSessionStart, session, nil)
	}()

	slog.Info("Created new session", "id", session.ID[:8], "agent", agent)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*Session, error) {
	sm.mu.RLock()
	if session, ok := sm.sessions[id]; ok {
		sm.mu.RUnlock()
		return session, nil
	}
	sm.mu.RUnlock()

	// Try to load from database
	session, err := sm.loadSession(id)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()
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

// buildContextSummary creates an optimized session context summary
func (sm *SessionManager) buildContextSummary(session *Session, maxTokens int) string {
	var sb strings.Builder

	// Add session header
	sb.WriteString("# Session Context\n")
	sb.WriteString(fmt.Sprintf("Agent: %s\n", session.Agent))
	sb.WriteString(fmt.Sprintf("Project: %s\n", session.ProjectPath))
	sb.WriteString(fmt.Sprintf("Turns: %d | Tokens: %d\n\n", session.Metadata.TotalTurns, session.Metadata.TotalTokens))

	// Add state information
	if session.State.Focus != "" {
		sb.WriteString(fmt.Sprintf("**Current Focus:** %s\n\n", session.State.Focus))
	}
	if session.State.NextAction != "" {
		sb.WriteString(fmt.Sprintf("**Next Action:** %s\n\n", session.State.NextAction))
	}

	// Add recent context blocks (most recent first, within token budget)
	tokenCount := 0
	sb.WriteString("## Recent Activity\n\n")

	for i := len(session.ContextBlocks) - 1; i >= 0; i-- {
		block := session.ContextBlocks[i]
		if tokenCount+block.Tokens > maxTokens {
			break
		}

		sb.WriteString(fmt.Sprintf("**%s** (%s)\n%s\n\n",
			block.Type,
			block.Timestamp.Format("15:04:05"),
			block.Content))

		tokenCount += block.Tokens
	}

	return sb.String()
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
	contextBlocks, err := json.Marshal(session.ContextBlocks)
	if err != nil {
		return fmt.Errorf("marshal context blocks: %w", err)
	}
	state, err := json.Marshal(session.State)
	if err != nil {
		return fmt.Errorf("marshal session state: %w", err)
	}
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("marshal session metadata: %w", err)
	}

	_, err = sm.db.Exec(`
		INSERT OR REPLACE INTO sessions 
		(id, agent, project_path, started_at, last_activity, context_blocks, state, metadata, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.Agent, session.ProjectPath, session.StartedAt,
		session.LastActivity, contextBlocks, state, metadata, session.ExpiresAt, session.IsActive)

	return err
}

// loadSession loads a session from the database
func (sm *SessionManager) loadSession(id string) (*Session, error) {
	var session Session
	var contextBlocks, state, metadata []byte

	err := sm.db.QueryRow(`
		SELECT id, agent, project_path, started_at, last_activity, context_blocks, state, metadata, expires_at, is_active
		FROM sessions WHERE id = ?
	`, id).Scan(&session.ID, &session.Agent, &session.ProjectPath, &session.StartedAt,
		&session.LastActivity, &contextBlocks, &state, &metadata, &session.ExpiresAt, &session.IsActive)

	if err != nil {
		return nil, err
	}

	if len(contextBlocks) > 0 {
		if err := json.Unmarshal(contextBlocks, &session.ContextBlocks); err != nil {
			return nil, fmt.Errorf("unmarshal context blocks: %w", err)
		}
	}
	if len(state) > 0 {
		if err := json.Unmarshal(state, &session.State); err != nil {
			return nil, fmt.Errorf("unmarshal session state: %w", err)
		}
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &session.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal session metadata: %w", err)
		}
	}

	return &session, nil
}

// CleanupExpired removes expired sessions
func (sm *SessionManager) CleanupExpired() error {
	_, err := sm.db.Exec("DELETE FROM sessions WHERE expires_at IS NOT NULL AND expires_at < ?", time.Now())
	return err
}

// GetSummary returns persisted session-store metrics for diagnostics and dashboards.
func (sm *SessionManager) GetSummary() (*SessionStoreSummary, error) {
	summary := &SessionStoreSummary{}
	var lastActivityEpoch sql.NullInt64

	err := sm.db.QueryRow(`
		SELECT
			COUNT(*) AS total_sessions,
			COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS active_sessions,
			CAST(strftime('%s', MAX(last_activity)) AS INTEGER) AS last_activity
		FROM sessions
	`).Scan(&summary.TotalSessions, &summary.ActiveSessions, &lastActivityEpoch)
	if err != nil {
		return nil, fmt.Errorf("query session summary: %w", err)
	}
	if lastActivityEpoch.Valid && lastActivityEpoch.Int64 > 0 {
		ts := time.Unix(lastActivityEpoch.Int64, 0)
		summary.LastActivity = &ts
	}

	if err := sm.db.QueryRow(`SELECT COUNT(*) FROM session_snapshots`).Scan(&summary.SnapshotCount); err != nil {
		return nil, fmt.Errorf("query session snapshots: %w", err)
	}

	var agent sql.NullString
	var count sql.NullInt64
	err = sm.db.QueryRow(`
		SELECT agent, COUNT(*) AS cnt
		FROM sessions
		WHERE TRIM(COALESCE(agent, '')) <> ''
		GROUP BY agent
		ORDER BY cnt DESC, agent ASC
		LIMIT 1
	`).Scan(&agent, &count)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query top session agent: %w", err)
	}
	if agent.Valid {
		summary.TopAgent = agent.String
	}
	if count.Valid {
		summary.TopAgentCount = count.Int64
	}

	sm.mu.RLock()
	summary.ActiveSessionID = sm.activeSession
	sm.mu.RUnlock()

	return summary, nil
}

// ListSessionOverviews returns recent persisted sessions with snapshot counts and totals.
func (sm *SessionManager) ListSessionOverviews(opts SessionListOptions) (*SessionOverviewList, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where, args := buildSessionFilters(opts)

	countQuery := `SELECT COUNT(*) FROM sessions WHERE ` + where
	var total int64
	if err := sm.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}

	query := `
		SELECT
			s.id,
			s.agent,
			s.project_path,
			s.started_at,
			s.last_activity,
			s.is_active,
			s.context_blocks,
			s.metadata,
			COALESCE(ss.snapshot_count, 0) AS snapshot_count,
			ss.last_snapshot_at
		FROM sessions s
		LEFT JOIN (
			SELECT
				session_id,
				COUNT(*) AS snapshot_count,
				MAX(created_at) AS last_snapshot_at
			FROM session_snapshots
			GROUP BY session_id
		) ss ON ss.session_id = s.id
		WHERE ` + where + `
		ORDER BY s.last_activity DESC, s.started_at DESC
		LIMIT ? OFFSET ?`

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := sm.db.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query session overviews: %w", err)
	}
	defer rows.Close()

	result := &SessionOverviewList{}
	for rows.Next() {
		var item SessionOverview
		var contextBlocksRaw, metadataRaw []byte
		var lastSnapshot sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Agent,
			&item.ProjectPath,
			&item.StartedAt,
			&item.LastActivity,
			&item.IsActive,
			&contextBlocksRaw,
			&metadataRaw,
			&item.SnapshotCount,
			&lastSnapshot,
		); err != nil {
			return nil, fmt.Errorf("scan session overview: %w", err)
		}
		if lastSnapshot.Valid {
			ts, err := parseSessionTimestamp(lastSnapshot.String)
			if err != nil {
				return nil, fmt.Errorf("parse session overview snapshot time: %w", err)
			}
			item.LastSnapshotAt = &ts
		}

		if len(contextBlocksRaw) > 0 {
			var blocks []ContextBlock
			if err := json.Unmarshal(contextBlocksRaw, &blocks); err != nil {
				return nil, fmt.Errorf("unmarshal session context blocks: %w", err)
			}
			item.ContextBlockCount = len(blocks)
		}
		if len(metadataRaw) > 0 {
			var metadata SessionMetadata
			if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
				return nil, fmt.Errorf("unmarshal session metadata: %w", err)
			}
			item.TotalTurns = metadata.TotalTurns
			item.TotalTokens = metadata.TotalTokens
			item.CompressionRatio = metadata.CompressionRatio
		}

		result.Sessions = append(result.Sessions, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session overviews: %w", err)
	}

	result.Total = total
	result.HasMore = int64(offset+len(result.Sessions)) < total
	return result, nil
}

// ListSnapshotSummaries returns snapshot history aggregated by session.
func (sm *SessionManager) ListSnapshotSummaries(limit int) ([]SessionSnapshotSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT
			s.id,
			COALESCE(s.agent, ''),
			COALESCE(s.project_path, ''),
			COUNT(ss.id) AS snapshot_count,
			MAX(ss.created_at) AS last_snapshot_at,
			COALESCE(MAX(ss.token_count), 0) AS latest_token_count
		FROM sessions s
		INNER JOIN session_snapshots ss ON ss.session_id = s.id
		GROUP BY s.id, s.agent, s.project_path
		ORDER BY last_snapshot_at DESC, snapshot_count DESC
		LIMIT ?`

	rows, err := sm.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("query snapshot summaries: %w", err)
	}
	defer rows.Close()

	var summaries []SessionSnapshotSummary
	for rows.Next() {
		var item SessionSnapshotSummary
		var lastSnapshot sql.NullString
		if err := rows.Scan(
			&item.SessionID,
			&item.Agent,
			&item.ProjectPath,
			&item.SnapshotCount,
			&lastSnapshot,
			&item.LatestTokenCount,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot summary: %w", err)
		}
		if lastSnapshot.Valid {
			ts, err := parseSessionTimestamp(lastSnapshot.String)
			if err != nil {
				return nil, fmt.Errorf("parse snapshot summary time: %w", err)
			}
			item.LastSnapshotAt = &ts
		}
		summaries = append(summaries, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snapshot summaries: %w", err)
	}
	return summaries, nil
}

// GetActiveContextMetrics returns context metrics for the active session.
func (sm *SessionManager) GetActiveContextMetrics() (*ActiveSessionContextMetrics, error) {
	active := sm.GetActiveSession()
	if active == nil {
		return nil, nil
	}

	metrics := &ActiveSessionContextMetrics{
		SessionID:         active.ID,
		Agent:             active.Agent,
		ProjectPath:       active.ProjectPath,
		Focus:             active.State.Focus,
		NextAction:        active.State.NextAction,
		TotalTurns:        active.Metadata.TotalTurns,
		TotalTokens:       active.Metadata.TotalTokens,
		CompressionRatio:  active.Metadata.CompressionRatio,
		ContextBlockCount: len(active.ContextBlocks),
		BlockTypeCounts:   make(map[string]int),
	}
	lastActivity := active.LastActivity
	metrics.LastActivity = &lastActivity

	for _, block := range active.ContextBlocks {
		metrics.BlockTypeCounts[string(block.Type)]++
	}

	return metrics, nil
}

// GetAnalyticsSnapshot returns the canonical session analytics payload for dashboards/TUIs.
func (sm *SessionManager) GetAnalyticsSnapshot(opts SessionListOptions, snapshotLimit int) (*SessionAnalyticsSnapshot, error) {
	summary, err := sm.GetSummary()
	if err != nil {
		return nil, err
	}
	recent, err := sm.ListSessionOverviews(opts)
	if err != nil {
		return nil, err
	}
	snapshots, err := sm.ListSnapshotSummaries(snapshotLimit)
	if err != nil {
		return nil, err
	}
	active, err := sm.GetActiveContextMetrics()
	if err != nil {
		return nil, err
	}

	return &SessionAnalyticsSnapshot{
		StoreSummary:    *summary,
		RecentSessions:  recent.Sessions,
		SnapshotHistory: snapshots,
		ActiveContext:   active,
	}, nil
}

// Close closes the session manager
func (sm *SessionManager) Close() error {
	return sm.db.Close()
}

func buildSessionFilters(opts SessionListOptions) (string, []any) {
	filters := []string{"1=1"}
	args := make([]any, 0, 4)

	if value := strings.TrimSpace(opts.Agent); value != "" {
		filters = append(filters, "agent = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(opts.ProjectPath); value != "" {
		filters = append(filters, "project_path = ?")
		args = append(args, value)
	}
	if opts.ActiveOnly {
		filters = append(filters, "is_active = 1")
	}

	return strings.Join(filters, " AND "), args
}

func parseSessionTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format %q", value)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
