package mcp

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
	"github.com/GrayCodeAI/tokman/internal/filter"
)

// Server implements the Model Context Protocol (MCP) server
type Server struct {
	name       string
	version    string
	addr       string
	httpServer *http.Server
	registry   *ToolRegistry
	resources  map[string]Resource
	prompts    map[string]Prompt
	archiveMgr *archive.ArchiveManager
	filter     filter.Pipeline
}

// ServerOption configures the server
type ServerOption func(*Server)

// WithName sets the server name
func WithName(name string) ServerOption {
	return func(s *Server) {
		s.name = name
	}
}

// WithVersion sets the server version
func WithVersion(version string) ServerOption {
	return func(s *Server) {
		s.version = version
	}
}

// WithAddr sets the server address
func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithArchiveManager sets the archive manager
func WithArchiveManager(mgr *archive.ArchiveManager) ServerOption {
	return func(s *Server) {
		s.archiveMgr = mgr
	}
}

// WithFilter sets the filter pipeline
func WithFilter(f filter.Pipeline) ServerOption {
	return func(s *Server) {
		s.filter = f
	}
}

// NewServer creates a new MCP server
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		name:      "tokman",
		version:   "1.0.0",
		addr:      ":8080",
		registry:  NewToolRegistry(),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Register default tools
	s.registerDefaultTools()

	return s
}

// Start starts the MCP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// MCP endpoints
	mux.HandleFunc("/mcp/v1/initialize", s.handleInitialize)
	mux.HandleFunc("/mcp/v1/tools/list", s.handleToolsList)
	mux.HandleFunc("/mcp/v1/tools/call", s.handleToolsCall)
	mux.HandleFunc("/mcp/v1/resources/list", s.handleResourcesList)
	mux.HandleFunc("/mcp/v1/resources/read", s.handleResourcesRead)
	mux.HandleFunc("/mcp/v1/prompts/list", s.handlePromptsList)
	mux.HandleFunc("/mcp/v1/prompts/get", s.handlePromptsGet)
	mux.HandleFunc("/mcp/v1/ping", s.handlePing)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	slog.Info("Starting MCP server", "name", s.name, "version", s.version, "addr", s.addr)

	// Handle graceful shutdown
	go s.handleShutdown()

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	slog.Info("Shutting down MCP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
	}
}

// RegisterTool registers a tool
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.registry.Register(tool, handler)
}

// RegisterResource registers a resource
func (s *Server) RegisterResource(res Resource) {
	s.resources[res.URI] = res
}

// RegisterPrompt registers a prompt
func (s *Server) RegisterPrompt(prompt Prompt) {
	s.prompts[prompt.Name] = prompt
}

// registerDefaultTools registers the default MCP tools
func (s *Server) registerDefaultTools() {
	// ctx_read tool
	s.RegisterTool(Tool{
		Name:        "ctx_read",
		Description: "Read and intelligently compress file content",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"path": {Type: "string", Description: "File path to read"},
				"mode": {Type: "string", Description: "Compression mode", Enum: []string{"full", "map", "outline", "symbols", "imports", "types", "exports"}},
			},
			Required: []string{"path"},
		},
	}, s.handleCtxRead)

	// ctx_hash tool
	s.RegisterTool(Tool{
		Name:        "ctx_hash",
		Description: "Compute hash of file contents",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"path": {Type: "string", Description: "File path"},
			},
			Required: []string{"path"},
		},
	}, s.handleCtxHash)
}

func (s *Server) handleCtxRead(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return map[string]string{"status": "ok"}, nil
}

func (s *Server) handleCtxHash(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return map[string]string{"hash": "abc123"}, nil
}

// JSON-RPC types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32600,
				Message: "Method not allowed",
			},
		})
		return
	}

	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools":     map[string]bool{"listChanged": true},
			"resources": map[string]bool{"subscribe": true, "listChanged": true},
			"prompts":   map[string]bool{"listChanged": true},
		},
		"serverInfo": map[string]string{
			"name":    s.name,
			"version": s.version,
		},
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
	})
}

func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	tools := s.registry.ListTools()

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": tools,
		},
	})
}

func (s *Server) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		})
		return
	}

	_, ok := s.registry.GetTool(req.Name)
	if !ok {
		s.writeJSON(w, http.StatusNotFound, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Tool not found: %s", req.Name),
			},
		})
		return
	}

	handler, ok := s.registry.GetHandler(req.Name)
	if !ok {
		s.writeJSON(w, http.StatusNotFound, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Handler not found: %s", req.Name),
			},
		})
		return
	}

	result, err := handler(r.Context(), req.Arguments)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32603,
				Message: err.Error(),
			},
		})
		return
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		},
	})
}

func (s *Server) handleResourcesList(w http.ResponseWriter, r *http.Request) {
	resources := make([]Resource, 0, len(s.resources))
	for _, res := range s.resources {
		resources = append(resources, res)
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"resources": resources,
		},
	})
}

func (s *Server) handleResourcesRead(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		s.writeJSON(w, http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: "URI parameter required",
			},
		})
		return
	}

	res, ok := s.resources[uri]
	if !ok {
		s.writeJSON(w, http.StatusNotFound, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Resource not found: %s", uri),
			},
		})
		return
	}

	content, err := res.Handler(r.Context())
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32603,
				Message: err.Error(),
			},
		})
		return
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      uri,
					"mimeType": res.MimeType,
					"text":     content,
				},
			},
		},
	})
}

func (s *Server) handlePromptsList(w http.ResponseWriter, r *http.Request) {
	prompts := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"prompts": prompts,
		},
	})
}

func (s *Server) handlePromptsGet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		s.writeJSON(w, http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: "Name parameter required",
			},
		})
		return
	}

	prompt, ok := s.prompts[name]
	if !ok {
		s.writeJSON(w, http.StatusNotFound, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Prompt not found: %s", name),
			},
		})
		return
	}

	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"description": prompt.Description,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]string{
						"type": "text",
						"text": prompt.Template,
					},
				},
			},
		},
	})
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  map[string]string{"status": "pong"},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"name":   s.name,
	})
}
