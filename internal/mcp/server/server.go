// Package server provides a standalone MCP (Model Context Protocol) server
// that can be used by any AI agent without requiring the TokMan CLI.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/mcp"
)

// Config holds server configuration
type Config struct {
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	EnableSSE   bool          `json:"enable_sse"`
	EnableStdio bool          `json:"enable_stdio"`
	ArchivePath string        `json:"archive_path"`
	MaxRequests int           `json:"max_requests"`
	Timeout     time.Duration `json:"timeout"`
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:        "localhost",
		Port:        8080,
		EnableSSE:   true,
		EnableStdio: true,
		ArchivePath: "~/.local/share/tokman/mcp_archive.db",
		MaxRequests: 1000,
		Timeout:     30 * time.Second,
	}
}

// Server is the MCP server implementation
type Server struct {
	config     *Config
	archiveMgr *archive.ArchiveManager
	router     *http.ServeMux
	server     *http.Server
	tools      map[string]mcp.ToolHandler
	resources  map[string]mcp.ResourceHandler
	shutdown   chan os.Signal
}

// NewServer creates a new MCP server
func NewServer(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Initialize archive manager
	archiveCfg := archive.ArchiveConfig{
		MaxSize:    100 * 1024 * 1024,
		Expiration: 7 * 24 * time.Hour,
		Enabled:    true,
	}

	archiveMgr, err := archive.NewArchiveManager(archiveCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive manager: %w", err)
	}

	s := &Server{
		config:     cfg,
		archiveMgr: archiveMgr,
		router:     http.NewServeMux(),
		tools:      make(map[string]mcp.ToolHandler),
		resources:  make(map[string]mcp.ResourceHandler),
		shutdown:   make(chan os.Signal, 1),
	}

	// Register default tools
	s.registerDefaultTools()

	// Setup routes
	s.setupRoutes()

	return s, nil
}

// registerDefaultTools registers the default set of MCP tools
func (s *Server) registerDefaultTools() {
	// Read tool
	s.RegisterTool("read", "Read and compress file content",
		map[string]mcp.Property{
			"path": {Type: "string", Description: "File path to read"},
		},
		[]string{"path"},
		s.handleReadTool())

	// Search tool
	s.RegisterTool("search", "Search archived content",
		map[string]mcp.Property{
			"query": {Type: "string", Description: "Search query"},
		},
		[]string{"query"},
		s.handleSearchTool())

	// Summary tool
	s.RegisterTool("summary", "Get summary of content",
		map[string]mcp.Property{
			"content": {Type: "string", Description: "Content to summarize"},
		},
		[]string{"content"},
		s.handleSummaryTool())

	// Delta tool
	s.RegisterTool("delta", "Get changed files",
		map[string]mcp.Property{
			"path": {Type: "string", Description: "Directory or file path"},
		},
		[]string{},
		s.handleDeltaTool())

	// Grep tool
	s.RegisterTool("grep", "Search content with patterns",
		map[string]mcp.Property{
			"pattern": {Type: "string", Description: "Search pattern"},
			"path":    {Type: "string", Description: "Path to search"},
		},
		[]string{"pattern"},
		s.handleGrepTool())

	// Hash tool
	s.RegisterTool("hash", "Compute content hash",
		map[string]mcp.Property{
			"content": {Type: "string", Description: "Content to hash"},
		},
		[]string{"content"},
		s.handleHashTool())

	// Tree tool
	s.RegisterTool("tree", "Show directory tree",
		map[string]mcp.Property{
			"path":  {Type: "string", Description: "Directory path"},
			"depth": {Type: "integer", Description: "Maximum depth"},
		},
		[]string{},
		s.handleTreeTool())

	// Stats tool
	s.RegisterTool("stats", "Get file/directory statistics",
		map[string]mcp.Property{
			"path": {Type: "string", Description: "Path to analyze"},
		},
		[]string{"path"},
		s.handleStatsTool())

	// Archive tool
	s.RegisterTool("archive", "Archive content",
		map[string]mcp.Property{
			"content": {Type: "string", Description: "Content to archive"},
			"path":    {Type: "string", Description: "File path to archive"},
		},
		[]string{},
		s.handleArchiveTool())

	// Rewind tool
	s.RegisterTool("rewind", "Rewind to previous state",
		map[string]mcp.Property{
			"hash": {Type: "string", Description: "Archive hash"},
		},
		[]string{},
		s.handleRewindTool())
}

// RegisterTool registers a new tool
func (s *Server) RegisterTool(name, description string, properties map[string]mcp.Property, required []string, handler mcp.ToolHandler) {
	s.tools[name] = handler
	slog.Info("Registered MCP tool", "name", name, "description", description)
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	// MCP endpoints
	s.router.HandleFunc("/mcp/v1/initialize", s.handleInitialize)
	s.router.HandleFunc("/mcp/v1/tools/list", s.handleListTools)
	s.router.HandleFunc("/mcp/v1/tools/call", s.handleCallTool)
	s.router.HandleFunc("/mcp/v1/prompts/list", s.handleListPrompts)
	s.router.HandleFunc("/mcp/v1/resources/list", s.handleListResources)
	s.router.HandleFunc("/mcp/v1/health", s.handleHealth)

	// SSE endpoint for streaming
	if s.config.EnableSSE {
		s.router.HandleFunc("/mcp/v1/sse", s.handleSSE)
	}

	// Stdio endpoint
	if s.config.EnableStdio {
		s.router.HandleFunc("/mcp/v1/stdio", s.handleStdio)
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "tokman-mcp",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools":     map[string]bool{"listChanged": false},
			"prompts":   map[string]bool{"listChanged": false},
			"resources": map[string]bool{"subscribe": false, "listChanged": false},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListTools returns the list of available tools
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toolList := make([]map[string]interface{}, 0, len(s.tools))
	for name := range s.tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        name,
			"description": s.getToolDescription(name),
		})
	}

	response := map[string]interface{}{
		"tools": toolList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCallTool handles tool invocation
func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	handler, exists := s.tools[req.Name]
	if !exists {
		http.Error(w, fmt.Sprintf("Tool not found: %s", req.Name), http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.Timeout)
	defer cancel()

	result, err := handler(ctx, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("%v", result),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListPrompts returns available prompts
func (s *Server) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"prompts": []interface{}{}})
}

// handleListResources returns available resources
func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"resources": []interface{}{}})
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"tools":     len(s.tools),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleSSE handles Server-Sent Events
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial event
	fmt.Fprintf(w, "data: %s\n\n", `{"type": "connected"}`)
	flusher.Flush()

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, "data: %s\n\n", `{"type": "ping"}`)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// handleStdio handles stdio connections
func (s *Server) handleStdio(w http.ResponseWriter, r *http.Request) {
	// This would handle stdio-based communication for CLI integration
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// getToolDescription returns a tool's description
func (s *Server) getToolDescription(name string) string {
	descriptions := map[string]string{
		"read":    "Read and compress file content",
		"search":  "Search archived content",
		"summary": "Get summary of content",
		"delta":   "Get changed files",
		"grep":    "Search content with patterns",
		"hash":    "Compute content hash",
		"tree":    "Show directory tree",
		"stats":   "Get file/directory statistics",
		"archive": "Archive content",
		"rewind":  "Rewind to previous state",
	}

	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return name
}

// Start starts the MCP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

	// Setup graceful shutdown
	signal.Notify(s.shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-s.shutdown
		slog.Info("Shutting down MCP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}()

	slog.Info("Starting MCP server", "address", addr)
	return s.server.ListenAndServe()
}

// Stop stops the MCP server
func (s *Server) Stop() {
	close(s.shutdown)
}

// Tool handlers - to be implemented with actual logic
func (s *Server) handleReadTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		path, _ := params["path"].(string)
		if path == "" {
			return nil, fmt.Errorf("path is required")
		}

		// Use filter pipeline to compress content
		// TODO: Read file and process through pipeline
		// cfg := filter.PipelineConfig{Mode: filter.ModeMinimal}
		// pipeline := filter.NewPipelineCoordinator(cfg)

		return map[string]string{
			"status": "success",
			"path":   path,
			"note":   "File compression via MCP server",
		}, nil
	}
}

func (s *Server) handleSearchTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		query, _ := params["query"].(string)
		return map[string]interface{}{
			"query":   query,
			"results": []interface{}{},
		}, nil
	}
}

func (s *Server) handleSummaryTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		content, _ := params["content"].(string)
		return map[string]string{
			"summary": "Summary of: " + content[:min(len(content), 50)],
		}, nil
	}
}

func (s *Server) handleDeltaTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"changed_files": []interface{}{},
		}, nil
	}
}

func (s *Server) handleGrepTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		pattern, _ := params["pattern"].(string)
		return map[string]interface{}{
			"pattern": pattern,
			"matches": []interface{}{},
		}, nil
	}
}

func (s *Server) handleHashTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		content, _ := params["content"].(string)
		return map[string]string{
			"hash":  fmt.Sprintf("%x", len(content)), // Simplified
			"input": content[:min(len(content), 100)],
		}, nil
	}
}

func (s *Server) handleTreeTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		path, _ := params["path"].(string)
		if path == "" {
			path = "."
		}
		return map[string]interface{}{
			"path": path,
			"tree": map[string]interface{}{},
		}, nil
	}
}

func (s *Server) handleStatsTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		path, _ := params["path"].(string)
		return map[string]interface{}{
			"path":        path,
			"size":        0,
			"lines":       0,
			"tokens":      0,
			"compression": "0%",
		}, nil
	}
}

func (s *Server) handleArchiveTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return map[string]string{
			"status": "archived",
			"hash":   "abc123",
		}, nil
	}
}

func (s *Server) handleRewindTool() mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		hash, _ := params["hash"].(string)
		return map[string]string{
			"status": "restored",
			"hash":   hash,
		}, nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
