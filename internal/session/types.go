package session

import "time"

// Session represents an active interaction session
type Session struct {
	ID            string          `json:"id"`
	Agent         string          `json:"agent"`
	ProjectPath   string          `json:"project_path"`
	StartedAt     time.Time       `json:"started_at"`
	LastActivity  time.Time       `json:"last_activity"`
	ContextBlocks []ContextBlock  `json:"context_blocks"`
	State         SessionState    `json:"state"`
	Metadata      SessionMetadata `json:"metadata"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
	IsActive      bool            `json:"is_active"`
}

// SessionState represents the current state of a session
type SessionState struct {
	Variables  map[string]interface{} `json:"variables"`
	Focus      string                 `json:"focus,omitempty"`
	NextAction string                 `json:"next_action,omitempty"`
}

// SessionMetadata contains session statistics
type SessionMetadata struct {
	TotalTurns       int     `json:"total_turns"`
	TotalTokens      int     `json:"total_tokens"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// ContextBlockType represents the type of context block
type ContextBlockType string

const (
	BlockTypeUserQuery  ContextBlockType = "user_query"
	BlockTypeToolResult ContextBlockType = "tool_result"
	BlockTypeSummary    ContextBlockType = "summary"
	BlockTypeSystem     ContextBlockType = "system"
	BlockTypeError      ContextBlockType = "error"
)

// ContextBlock represents a piece of context in a session
type ContextBlock struct {
	Type      ContextBlockType `json:"type"`
	Content   string           `json:"content"`
	Timestamp time.Time        `json:"timestamp"`
	Tokens    int              `json:"tokens"`
}

// SessionSnapshot represents a saved session state
type SessionSnapshot struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	CreatedAt  time.Time `json:"created_at"`
	Content    string    `json:"content"`
	TokenCount int       `json:"token_count"`
}

// SessionListOptions provides filtering for listing sessions
type SessionListOptions struct {
	Agent       string
	ProjectPath string
	ActiveOnly  bool
	Limit       int
	Offset      int
}

// SessionListResult contains the result of a list operation
type SessionListResult struct {
	Sessions []Session `json:"sessions"`
	Total    int64     `json:"total"`
	HasMore  bool      `json:"has_more"`
}

// PreCompactOptions contains options for PreCompact operation
type PreCompactOptions struct {
	MaxTokens       int
	PreserveRecent  int // Number of recent turns to preserve
	IncludeState    bool
	IncludeMetadata bool
}

// PreCompactResult contains the result of PreCompact operation
type PreCompactResult struct {
	Summary     string `json:"summary"`
	TokensUsed  int    `json:"tokens_used"`
	BlocksKept  int    `json:"blocks_kept"`
	BlocksTotal int    `json:"blocks_total"`
}
