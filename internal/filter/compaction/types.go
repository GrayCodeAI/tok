package compaction

import (
	"sync"
	"time"
	
	"github.com/GrayCodeAI/tokman/internal/llm"
)

// CompactionLayer provides semantic compression for chat/conversation content.
// Refactored from original 968-line compaction.go
// Paper: "MemGPT" — Packer et al., UC Berkeley, 2023
type CompactionLayer struct {
	config         CompactionConfig
	summarizer     *llm.Summarizer
	sessionTracker SessionTracker
	cache          map[string]*CompactionResult
	cacheMu        sync.RWMutex
}

// CompactionConfig holds configuration for the compaction layer
type CompactionConfig struct {
	Enabled              bool
	ThresholdLines       int
	ThresholdTokens      int
	PreserveRecentTurns  int
	MaxSummaryTokens     int
	ContentTypes         []string
	CacheEnabled         bool
	AutoDetect           bool
	StateSnapshotFormat  bool
	ExtractKeyValuePairs bool
	MaxContextEntries    int
}

// CompactionResult represents a compaction result
type CompactionResult struct {
	Snapshot         *StateSnapshot
	OriginalTokens   int
	FinalTokens      int
	SavedTokens      int
	CompressionRatio float64
	Cached           bool
	Timestamp        time.Time
}

// StateSnapshot represents semantic compaction output
type StateSnapshot struct {
	SessionHistory SessionHistory  `json:"session_history"`
	CurrentState   CurrentState    `json:"current_state"`
	Context        SnapshotContext `json:"context"`
	PendingPlan    []Milestone     `json:"pending_plan"`
}

// SessionHistory tracks what happened in the session
type SessionHistory struct {
	UserQueries []string `json:"user_queries"`
	ActivityLog []string `json:"activity_log"`
	FilesRead   []string `json:"files_read,omitempty"`
	FilesEdited []string `json:"files_edited,omitempty"`
	CommandsRun []string `json:"commands_run,omitempty"`
	Decisions   []string `json:"decisions,omitempty"`
}

// CurrentState tracks what's currently active
type CurrentState struct {
	Focus     string `json:"focus"`
	NextAction string `json:"next_action"`
}

// SnapshotContext tracks what to remember
type SnapshotContext struct {
	Critical []string `json:"critical"`
	Working  []string `json:"working"`
}

// Milestone represents a future milestone
type Milestone struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

// SessionTracker interface for tracking sessions
type SessionTracker interface {
	Track(query string)
	GetRecent(n int) []string
}

// DefaultCompactionConfig returns default configuration
func DefaultCompactionConfig() CompactionConfig {
	return CompactionConfig{
		Enabled:              false,
		ThresholdLines:       100,
		ThresholdTokens:      2000,
		PreserveRecentTurns:  5,
		MaxSummaryTokens:     500,
		ContentTypes:         []string{"chat", "conversation", "session"},
		CacheEnabled:         true,
		AutoDetect:           true,
		StateSnapshotFormat:  true,
		ExtractKeyValuePairs: true,
		MaxContextEntries:    20,
	}
}
