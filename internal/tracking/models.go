package tracking

import "time"

// CommandRecord represents a single command execution record.
type CommandRecord struct {
	ID             int64     `json:"id"`
	Command        string    `json:"command"`
	OriginalOutput string    `json:"original_output,omitempty"`
	FilteredOutput string    `json:"filtered_output,omitempty"`
	OriginalTokens int       `json:"original_tokens"`
	FilteredTokens int       `json:"filtered_tokens"`
	SavedTokens    int       `json:"saved_tokens"`
	ProjectPath    string    `json:"project_path"`
	SessionID      string    `json:"session_id,omitempty"`
	ExecTimeMs     int64     `json:"exec_time_ms"`
	Timestamp      time.Time `json:"timestamp"`
	ParseSuccess   bool      `json:"parse_success"`
	// AI Agent attribution fields
	AgentName   string `json:"agent_name,omitempty"`   // e.g., "Claude Code", "OpenCode", "Cursor"
	ModelName   string `json:"model_name,omitempty"`   // e.g., "claude-3-opus", "gpt-4", "gemini-pro"
	Provider    string `json:"provider,omitempty"`     // e.g., "Anthropic", "OpenAI", "Google"
	ModelFamily string `json:"model_family,omitempty"` // e.g., "claude", "gpt", "gemini"
	// Smart context read metadata
	ContextKind         string `json:"context_kind,omitempty"`          // e.g., "read", "delta", "mcp"
	ContextMode         string `json:"context_mode,omitempty"`          // requested mode: auto, graph, delta, ...
	ContextResolvedMode string `json:"context_resolved_mode,omitempty"` // effective mode after auto-resolution
	ContextTarget       string `json:"context_target,omitempty"`        // file path or target identifier
	ContextRelatedFiles int    `json:"context_related_files,omitempty"` // number of related files included
	ContextBundle       bool   `json:"context_bundle,omitempty"`        // whether multiple files were delivered together
}

// SavingsSummary represents aggregated token savings.
type SavingsSummary struct {
	TotalCommands int     `json:"total_commands"`
	TotalSaved    int     `json:"total_saved"`
	TotalOriginal int     `json:"total_original"`
	TotalFiltered int     `json:"total_filtered"`
	ReductionPct  float64 `json:"reduction_percent"`
}

// CommandStats represents statistics for a specific command type.
type CommandStats struct {
	Command        string  `json:"command"`
	ExecutionCount int     `json:"execution_count"`
	TotalSaved     int     `json:"total_saved"`
	TotalOriginal  int     `json:"total_original"`
	ReductionPct   float64 `json:"reduction_percent"`
}

// SessionInfo represents information about a shell session.
type SessionInfo struct {
	SessionID   string    `json:"session_id"`
	StartedAt   time.Time `json:"started_at"`
	ProjectPath string    `json:"project_path"`
}

// ReportFilter represents filters for generating reports.
type ReportFilter struct {
	ProjectPath string     `json:"project_path,omitempty"`
	SessionID   string     `json:"session_id,omitempty"`
	Command     string     `json:"command,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

// Team represents an organization/workspace in the multi-tenant system.
type Team struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	Description        string    `json:"description,omitempty"`
	OwnerID            *int64    `json:"owner_id,omitempty"`
	MonthlyTokenBudget int       `json:"monthly_token_budget"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// User represents a team member with role-based access.
type User struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	TeamID    int64      `json:"team_id"`
	FullName  string     `json:"full_name,omitempty"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	Role      string     `json:"role"` // "admin", "editor", "viewer"
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// FilterMetric tracks individual filter effectiveness.
type FilterMetric struct {
	ID                 int64     `json:"id"`
	FilterName         string    `json:"filter_name"`
	TeamID             *int64    `json:"team_id,omitempty"`
	CommandID          *int64    `json:"command_id,omitempty"`
	TokensBefore       int       `json:"tokens_before"`
	TokensAfter        int       `json:"tokens_after"`
	TokensSaved        int       `json:"tokens_saved"`
	ProcessingTimeUs   int64     `json:"processing_time_us"`
	EffectivenessScore float64   `json:"effectiveness_score"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
}

// CostAggregation stores pre-aggregated cost metrics for fast queries.
type CostAggregation struct {
	ID                  int64     `json:"id"`
	TeamID              int64     `json:"team_id"`
	DateBucket          string    `json:"date_bucket"` // YYYY-MM-DD
	Period              string    `json:"period"`      // "daily", "weekly", "monthly"
	TotalCommands       int       `json:"total_commands"`
	TotalOriginalTokens int       `json:"total_original_tokens"`
	TotalFilteredTokens int       `json:"total_filtered_tokens"`
	TotalSavedTokens    int       `json:"total_saved_tokens"`
	EstimatedCostUSD    float64   `json:"estimated_cost_usd"`
	EstimatedSavingsUSD float64   `json:"estimated_savings_usd"`
	AvgReductionPercent float64   `json:"avg_reduction_percent"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// AuditLog represents an audit trail event.
type AuditLog struct {
	ID           int64     `json:"id"`
	TeamID       int64     `json:"team_id"`
	UserID       *int64    `json:"user_id,omitempty"`
	Action       string    `json:"action"` // "create", "update", "delete", "login", etc.
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	OldValue     string    `json:"old_value,omitempty"`
	NewValue     string    `json:"new_value,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FilterEffectiveness represents aggregated filter performance.
type FilterEffectiveness struct {
	FilterName          string     `json:"filter_name"`
	TeamID              *int64     `json:"team_id,omitempty"`
	UsageCount          int        `json:"usage_count"`
	AvgEffectiveness    float64    `json:"avg_effectiveness"`
	TotalSaved          int        `json:"total_saved"`
	AvgProcessingTimeUs float64    `json:"avg_processing_time_us"`
	FirstUsed           *time.Time `json:"first_used,omitempty"`
	LastUsed            *time.Time `json:"last_used,omitempty"`
}

// TrendData represents historical trend for dashboard charts.
type TrendData struct {
	TeamID                 int64   `json:"team_id"`
	DateBucket             string  `json:"date_bucket"`
	TotalCommands          int     `json:"total_commands"`
	TotalOriginalTokens    int     `json:"total_original_tokens"`
	TotalFilteredTokens    int     `json:"total_filtered_tokens"`
	TotalSavedTokens       int     `json:"total_saved_tokens"`
	EstimatedCostUSD       float64 `json:"estimated_cost_usd"`
	EstimatedSavingsUSD    float64 `json:"estimated_savings_usd"`
	AvgReductionPercent    float64 `json:"avg_reduction_percent"`
	ActualReductionPercent float64 `json:"actual_reduction_percent"`
}
