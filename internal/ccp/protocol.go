package ccp

import (
	"database/sql"
	"time"
)

type MemoryType string

const (
	MemoryTask      MemoryType = "task"
	MemoryFinding   MemoryType = "finding"
	MemoryDecision  MemoryType = "decision"
	MemoryKnowledge MemoryType = "knowledge"
)

type MemoryEntry struct {
	ID        int64      `json:"id"`
	Type      MemoryType `json:"type"`
	Content   string     `json:"content"`
	SessionID string     `json:"session_id"`
	Project   string     `json:"project"`
	Relevance float64    `json:"relevance"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

type ContextContinuityProtocol struct {
	db  *sql.DB
	ttl time.Duration
}

func NewCCP(db *sql.DB) *ContextContinuityProtocol {
	return &ContextContinuityProtocol{
		db:  db,
		ttl: 7 * 24 * time.Hour,
	}
}

func (ccp *ContextContinuityProtocol) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS ccp_memory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		content TEXT NOT NULL,
		session_id TEXT NOT NULL,
		project TEXT NOT NULL,
		relevance REAL NOT NULL DEFAULT 0.5,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME
	);
	CREATE INDEX IF NOT EXISTS idx_ccp_type ON ccp_memory(type);
	CREATE INDEX IF NOT EXISTS idx_ccp_session ON ccp_memory(session_id);
	CREATE INDEX IF NOT EXISTS idx_ccp_project ON ccp_memory(project);
	`
	_, err := ccp.db.Exec(query)
	return err
}

func (ccp *ContextContinuityProtocol) Store(entry *MemoryEntry) error {
	entry.ExpiresAt = time.Now().Add(ccp.ttl)
	_, err := ccp.db.Exec(`
		INSERT INTO ccp_memory (type, content, session_id, project, relevance, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, entry.Type, entry.Content, entry.SessionID, entry.Project, entry.Relevance, entry.ExpiresAt)
	return err
}

func (ccp *ContextContinuityProtocol) Retrieve(project string, memType MemoryType, limit int) ([]MemoryEntry, error) {
	rows, err := ccp.db.Query(`
		SELECT id, type, content, session_id, project, relevance, created_at, expires_at
		FROM ccp_memory
		WHERE project = ? AND type = ? AND expires_at > datetime('now')
		ORDER BY relevance DESC, created_at DESC
		LIMIT ?
	`, project, memType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		err := rows.Scan(&e.ID, &e.Type, &e.Content, &e.SessionID, &e.Project, &e.Relevance, &e.CreatedAt, &e.ExpiresAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (ccp *ContextContinuityProtocol) Cleanup() error {
	_, err := ccp.db.Exec("DELETE FROM ccp_memory WHERE expires_at < datetime('now')")
	return err
}

type Scratchpad struct {
	agentID string
	content string
	shared  map[string]string
}

func NewScratchpad(agentID string) *Scratchpad {
	return &Scratchpad{
		agentID: agentID,
		shared:  make(map[string]string),
	}
}

func (s *Scratchpad) Write(key, value string) {
	s.shared[key] = value
}

func (s *Scratchpad) Read(key string) string {
	return s.shared[key]
}

func (s *Scratchpad) List() map[string]string {
	return s.shared
}
