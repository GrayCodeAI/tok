// Package server provides HTTP and stdio transports for MCP.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/mcp"
)

// HTTPServer provides HTTP transport for MCP.
type HTTPServer struct {
	registry   *mcp.ToolRegistry
	cache      *mcp.HashCache
	mu         sync.RWMutex
	startedAt  time.Time
	tokenSaved int64
	port       int
	server     *http.Server
	listener   net.Listener
}

// NewHTTPServer creates a new HTTP MCP server.
func NewHTTPServer(registry *mcp.ToolRegistry, cache *mcp.HashCache, port int) *HTTPServer {
	return &HTTPServer{
		registry:  registry,
		cache:     cache,
		port:      port,
		startedAt: time.Now(),
	}
}

// Start starts the HTTP server.
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// JSON-RPC endpoint
	mux.HandleFunc("/mcp", s.handleJSONRPC)

	// Legacy endpoints for compatibility
	mux.HandleFunc("/v1/tools/list", s.handleToolsList)
	mux.HandleFunc("/v1/tools/call", s.handleToolCall)

	// Health and status
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the server.
func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// Port returns the actual port (may differ from requested if 0).
func (s *HTTPServer) Port() int {
	return s.port
}

// Addr returns the server address.
func (s *HTTPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return fmt.Sprintf(":%d", s.port)
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *HTTPServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, nil, -32600, "Invalid Request", "POST required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		s.writeError(w, nil, -32600, "Invalid Request", "Content-Type must be application/json")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10*1024*1024))
	if err != nil {
		s.writeError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	// Check for batch requests
	if len(body) > 0 && body[0] == '[' {
		s.handleBatchRequest(r.Context(), w, body)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		s.writeError(w, req.ID, -32600, "Invalid Request", "jsonrpc must be 2.0")
		return
	}

	result, err := s.dispatchMethod(r.Context(), req.Method, req.Params)
	if err != nil {
		s.writeError(w, req.ID, -32603, "Internal error", err.Error())
		return
	}

	s.writeResponse(w, req.ID, result)
}

func (s *HTTPServer) handleBatchRequest(ctx context.Context, w http.ResponseWriter, body []byte) {
	var requests []JSONRPCRequest
	if err := json.Unmarshal(body, &requests); err != nil {
		s.writeError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	responses := make([]JSONRPCResponse, 0, len(requests))
	for _, req := range requests {
		if req.JSONRPC != "2.0" {
			responses = append(responses, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32600,
					Message: "Invalid Request",
					Data:    "jsonrpc must be 2.0",
				},
			})
			continue
		}

		result, err := s.dispatchMethod(ctx, req.Method, req.Params)
		if err != nil {
			responses = append(responses, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32603,
					Message: "Internal error",
					Data:    err.Error(),
				},
			})
		} else {
			responses = append(responses, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  result,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		log.Printf("HTTP server batch encode error: %v", err)
	}
}

func (s *HTTPServer) dispatchMethod(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(ctx, params)
	case "tools/list":
		return s.handleToolsListRPC(ctx, params)
	case "tools/call":
		return s.handleToolCallRPC(ctx, params)
	case "cache/stats":
		return s.handleCacheStats(ctx, params)
	case "cache/invalidate":
		return s.handleCacheInvalidate(ctx, params)
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (s *HTTPServer) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return mcp.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsCapability{
				ListChanged: true,
			},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    "tokman-mcp",
			Version: "1.0.0",
		},
	}, nil
}

func (s *HTTPServer) handleToolsListRPC(ctx context.Context, params json.RawMessage) (interface{}, error) {
	type ToolsListResult struct {
		Tools []mcp.Tool `json:"tools"`
	}
	return ToolsListResult{Tools: s.registry.ListTools()}, nil
}

func (s *HTTPServer) handleToolCallRPC(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	handler, ok := s.registry.GetHandler(callParams.Name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", callParams.Name)
	}

	result, err := handler(ctx, callParams.Arguments)
	if err != nil {
		return nil, err
	}

	type ToolCallResult struct {
		Content []ToolContent `json:"content"`
	}
	return ToolCallResult{
		Content: []ToolContent{
			{Type: "text", Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

// ToolContent represents content in a tool result.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (s *HTTPServer) handleCacheStats(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return s.cache.Stats(), nil
}

func (s *HTTPServer) handleCacheInvalidate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var invalidateParams struct {
		Pattern string `json:"pattern,omitempty"`
		All     bool   `json:"all,omitempty"`
	}
	if err := json.Unmarshal(params, &invalidateParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if invalidateParams.All {
		s.cache.Clear()
		return map[string]int{"deleted": s.cache.Len()}, nil
	}

	if invalidateParams.Pattern != "" {
		deleted := s.cache.InvalidateByPattern(invalidateParams.Pattern)
		return map[string]int{"deleted": deleted}, nil
	}

	return map[string]int{"deleted": 0}, nil
}

func (s *HTTPServer) writeResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}); err != nil {
		log.Printf("HTTP server response encode error: %v", err)
	}
}

func (s *HTTPServer) writeError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}); err != nil {
		log.Printf("HTTP server error encode error: %v", err)
	}
}

func (s *HTTPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": s.registry.ListTools(),
	}); err != nil {
		log.Printf("HTTP server tools list encode error: %v", err)
	}
}

func (s *HTTPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name   string                 `json:"name"`
		Params map[string]interface{} `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	handler, ok := s.registry.GetHandler(req.Name)
	if !ok {
		http.Error(w, "tool not found", http.StatusNotFound)
		return
	}

	result, err := handler(r.Context(), req.Params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	}); err != nil {
		log.Printf("HTTP server tool call encode error: %v", err)
	}
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"uptime":  time.Since(s.startedAt).String(),
	}); err != nil {
		log.Printf("HTTP server health encode error: %v", err)
	}
}

func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	tokenSaved := s.tokenSaved
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mcp.Status{
		Version:          "1.0.0",
		Uptime:           time.Since(s.startedAt),
		CacheStats:       s.cache.Stats(),
		SessionStart:     s.startedAt,
		TotalTokensSaved: tokenSaved,
	}); err != nil {
		log.Printf("HTTP server status encode error: %v", err)
	}
}
