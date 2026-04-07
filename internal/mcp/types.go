// Package mcp implements the Model Context Protocol server for TokMan.
// Provides 24 MCP tools for intelligent context management.
package mcp

import (
	"context"
	"fmt"
	"time"
)

// ContentType represents the detected content type.
type ContentType string

const (
	ContentTypeCode     ContentType = "code"
	ContentTypeLog      ContentType = "log"
	ContentTypeJSON     ContentType = "json"
	ContentTypeDiff     ContentType = "diff"
	ContentTypeHTML     ContentType = "html"
	ContentTypeText     ContentType = "text"
	ContentTypeMarkdown ContentType = "markdown"
	ContentTypeUnknown  ContentType = "unknown"
)

// Language represents the detected programming language.
type Language string

const (
	LangGo         Language = "go"
	LangRust       Language = "rust"
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangJava       Language = "java"
	LangC          Language = "c"
	LangCpp        Language = "cpp"
	LangRuby       Language = "ruby"
	LangShell      Language = "shell"
	LangUnknown    Language = "unknown"
)

// ContextMode represents the read mode for ctx_read.
type ContextMode string

const (
	ModeFull    ContextMode = "full"
	ModeMap     ContextMode = "map"
	ModeOutline ContextMode = "outline"
	ModeSymbols ContextMode = "symbols"
	ModeImports ContextMode = "imports"
	ModeTypes   ContextMode = "types"
	ModeExports ContextMode = "exports"
)

// Request represents an MCP tool call request.
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// Response represents an MCP tool call response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an MCP error.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes per MCP spec.
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// Tool represents an MCP tool.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema represents the JSON schema for tool input.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property represents a JSON schema property.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ServerCapabilities represents MCP server capabilities.
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools"`
}

// ToolsCapability represents tools support.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// InitializeResult represents the initialize response.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerInfo represents server information.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// FusionContext carries context through the MCP pipeline.
// Immutable - each operation returns a new context.
type FusionContext struct {
	Content     string                 `json:"content"`
	ContentType ContentType            `json:"content_type"`
	Language    Language               `json:"language"`
	FilePath    string                 `json:"file_path"`
	Hash        string                 `json:"hash"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewFusionContext creates a new fusion context.
func NewFusionContext(content, filePath string) *FusionContext {
	return &FusionContext{
		Content:   content,
		FilePath:  filePath,
		Hash:      ComputeHashShort(content),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// ToolHandler is the function signature for tool implementations.
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// ToolRegistry manages available tools.
type ToolRegistry struct {
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// Register registers a tool and its handler.
func (r *ToolRegistry) Register(tool Tool, handler ToolHandler) {
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// GetTool returns a tool by name.
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetHandler returns a handler by name.
func (r *ToolRegistry) GetHandler(name string) (ToolHandler, bool) {
	handler, ok := r.handlers[name]
	return handler, ok
}

// ListTools returns all registered tools.
func (r *ToolRegistry) ListTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// CacheEntry represents a cached file entry.
type CacheEntry struct {
	Hash      string    `json:"hash"`
	Content   string    `json:"content"`
	FilePath  string    `json:"file_path"`
	Timestamp time.Time `json:"timestamp"`
	Accessed  time.Time `json:"accessed"`
	HitCount  int       `json:"hit_count"`
}

// IsExpired checks if the cache entry is expired (older than 24 hours).
func (e *CacheEntry) IsExpired() bool {
	return time.Since(e.Timestamp) > 24*time.Hour
}

// MemoryEntry represents a persistent memory entry.
type MemoryEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Bundle represents a collection of files.
type Bundle struct {
	ID        string                 `json:"id"`
	Files     []BundleFile           `json:"files"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// BundleFile represents a file in a bundle.
type BundleFile struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	Content  string `json:"content,omitempty"`
	Size     int    `json:"size"`
	Language string `json:"language"`
}

// Resource represents an MCP resource.
type Resource struct {
	URI         string          `json:"uri"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	MimeType    string          `json:"mimeType,omitempty"`
	Handler     ResourceHandler `json:"-"`
}

// ResourceHandler handles resource requests.
type ResourceHandler func(ctx context.Context) (string, error)

// Prompt represents an MCP prompt.
type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template"`
}

// Stats represents cache statistics.
type Stats struct {
	TotalEntries int64     `json:"total_entries"`
	TotalSize    int64     `json:"total_size_bytes"`
	HitRate      float64   `json:"hit_rate"`
	HitCount     int64     `json:"hit_count"`
	MissCount    int64     `json:"miss_count"`
	OldestEntry  time.Time `json:"oldest_entry"`
	NewestEntry  time.Time `json:"newest_entry"`
}

// Status represents overall MCP server status.
type Status struct {
	Version          string        `json:"version"`
	Uptime           time.Duration `json:"uptime"`
	CacheStats       Stats         `json:"cache_stats"`
	ActiveMode       ContextMode   `json:"active_mode"`
	SessionStart     time.Time     `json:"session_start"`
	TotalCommands    int64         `json:"total_commands"`
	TotalTokensSaved int64         `json:"total_tokens_saved"`
}

// Common parameter structures.

// ReadParams for ctx_read tool.
type ReadParams struct {
	Path string      `json:"path"`
	Mode ContextMode `json:"mode,omitempty"`
}

// DeltaParams for ctx_delta tool.
type DeltaParams struct {
	Path     string `json:"path"`
	BaseHash string `json:"base_hash,omitempty"`
}

// GrepParams for ctx_grep tool.
type GrepParams struct {
	Path    string `json:"path"`
	Pattern string `json:"pattern"`
	Context int    `json:"context,omitempty"`
}

// RememberParams for ctx_remember tool.
type RememberParams struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Tags  []string `json:"tags,omitempty"`
}

// BundleParams for ctx_bundle tool.
type BundleParams struct {
	Paths    []string `json:"paths"`
	Compress bool     `json:"compress,omitempty"`
}

// ExecParams for ctx_exec tool.
type ExecParams struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// Error messages.
var (
	ErrToolNotFound   = fmt.Errorf("tool not found")
	ErrInvalidParams  = fmt.Errorf("invalid parameters")
	ErrFileNotFound   = fmt.Errorf("file not found")
	ErrCacheMiss      = fmt.Errorf("cache miss")
	ErrMemoryNotFound = fmt.Errorf("memory entry not found")
	ErrAlreadyExists  = fmt.Errorf("already exists")
	ErrInternal       = fmt.Errorf("internal error")
)
