// Package api provides public API types for TokMan services.
package api

import "time"

// CompressionRequest is the public API request for compression.
type CompressionRequest struct {
	Input       string `json:"input"`
	Mode        string `json:"mode,omitempty"` // none, minimal, aggressive
	QueryIntent string `json:"query_intent,omitempty"`
	Budget      int    `json:"budget,omitempty"`
	Preset      string `json:"preset,omitempty"` // fast, balanced, full
}

// CompressionResponse is the public API response for compression.
type CompressionResponse struct {
	Output           string   `json:"output"`
	OriginalTokens   int      `json:"original_tokens"`
	CompressedTokens int      `json:"compressed_tokens"`
	SavingsPercent   float64  `json:"savings_percent"`
	LayersApplied    []string `json:"layers_applied,omitempty"`
}

// ExecuteRequest is the public API request for command execution.
type ExecuteRequest struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	FilterMode string            `json:"filter_mode,omitempty"`
	Timeout    time.Duration     `json:"timeout,omitempty"`
}

// ExecuteResponse is the public API response for command execution.
type ExecuteResponse struct {
	Stdout         string        `json:"stdout"`
	Stderr         string        `json:"stderr,omitempty"`
	ExitCode       int           `json:"exit_code"`
	TokensSaved    int           `json:"tokens_saved"`
	SavingsPercent float64       `json:"savings_percent"`
	Duration       time.Duration `json:"duration"`
}

// MetricsResponse is the public API response for analytics.
type MetricsResponse struct {
	TotalCommands    int64   `json:"total_commands"`
	TotalTokensSaved int64   `json:"total_tokens_saved"`
	AverageSavings   float64 `json:"average_savings"`
	P99LatencyMs     float64 `json:"p99_latency_ms"`
}

// AgentInfo is the public API type for agent information.
type AgentInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	HookPath    string `json:"hook_path,omitempty"`
}

// HealthResponse is the public API response for health checks.
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Uptime   time.Duration     `json:"uptime"`
	Services map[string]string `json:"services"`
}

// ErrorResponse is the public API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
