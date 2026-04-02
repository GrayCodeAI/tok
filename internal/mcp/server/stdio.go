// Package server provides stdio transport for MCP.
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/mcp"
)

// StdioServer provides stdio transport for MCP.
type StdioServer struct {
	registry  *mcp.ToolRegistry
	cache     *mcp.HashCache
	mu        sync.RWMutex
	startedAt time.Time
	running   bool
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// I/O streams (configurable for testing)
	in  io.Reader
	out io.Writer
	err io.Writer
}

// NewStdioServer creates a new stdio MCP server.
func NewStdioServer(registry *mcp.ToolRegistry, cache *mcp.HashCache) *StdioServer {
	return &StdioServer{
		registry:  registry,
		cache:     cache,
		startedAt: time.Now(),
		in:        os.Stdin,
		out:       os.Stdout,
		err:       os.Stderr,
	}
}

// NewStdioServerWithStreams creates a server with custom I/O streams (for testing).
func NewStdioServerWithStreams(registry *mcp.ToolRegistry, cache *mcp.HashCache, in io.Reader, out, err io.Writer) *StdioServer {
	return &StdioServer{
		registry:  registry,
		cache:     cache,
		startedAt: time.Now(),
		in:        in,
		out:       out,
		err:       err,
	}
}

// Start begins processing messages from stdin.
func (s *StdioServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

// Stop gracefully stops the server.
func (s *StdioServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns true if the server is running.
func (s *StdioServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *StdioServer) run(ctx context.Context) {
	defer s.wg.Done()

	scanner := bufio.NewScanner(s.in)
	// Increase buffer size for large messages
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(s.err, "scanner error: %v\n", err)
			}
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		s.wg.Add(1)
		go func(line string) {
			defer s.wg.Done()
			s.handleMessage(ctx, line)
		}(line)
	}
}

func (s *StdioServer) handleMessage(ctx context.Context, line string) {
	// Check for batch requests
	if len(line) > 0 && line[0] == '[' {
		s.handleBatchMessage(ctx, line)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		s.writeError(nil, -32700, "Parse error", err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		s.writeError(req.ID, -32600, "Invalid Request", "jsonrpc must be 2.0")
		return
	}

	result, err := s.dispatchMethod(ctx, req.Method, req.Params)
	if err != nil {
		s.writeError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	s.writeResponse(req.ID, result)
}

func (s *StdioServer) handleBatchMessage(ctx context.Context, line string) {
	var requests []JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &requests); err != nil {
		s.writeError(nil, -32700, "Parse error", err.Error())
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

	s.writeResponses(responses)
}

func (s *StdioServer) dispatchMethod(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(ctx, params)
	case "initialized":
		// Notification, no response needed
		return nil, nil
	case "tools/list":
		return s.handleToolsList(ctx, params)
	case "tools/call":
		return s.handleToolCall(ctx, params)
	case "cache/stats":
		return s.cache.Stats(), nil
	case "cache/invalidate":
		return s.handleCacheInvalidate(params)
	case "$/cancelRequest":
		// Cancellation notification, no response
		return nil, nil
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (s *StdioServer) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var initParams struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Roots struct {
				ListChanged bool `json:"listChanged"`
			} `json:"roots,omitempty"`
			Sampling struct{} `json:"sampling,omitempty"`
		} `json:"capabilities"`
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	log.Printf("Client connected: %s v%s (protocol: %s)",
		initParams.ClientInfo.Name,
		initParams.ClientInfo.Version,
		initParams.ProtocolVersion,
	)

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

func (s *StdioServer) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	type ToolsListResult struct {
		Tools []mcp.Tool `json:"tools"`
	}
	return ToolsListResult{Tools: s.registry.ListTools()}, nil
}

func (s *StdioServer) handleToolCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
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

	start := time.Now()
	result, err := handler(ctx, callParams.Arguments)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}

	// Log tool execution for analytics
	log.Printf("Tool %s executed in %v", callParams.Name, duration)

	type ToolCallResult struct {
		Content []ToolContent `json:"content"`
		IsError bool          `json:"isError,omitempty"`
	}

	content := fmt.Sprintf("%v", result)
	return ToolCallResult{
		Content: []ToolContent{
			{Type: "text", Text: content},
		},
	}, nil
}

func (s *StdioServer) handleCacheInvalidate(params json.RawMessage) (interface{}, error) {
	var invalidateParams struct {
		Pattern string `json:"pattern,omitempty"`
		All     bool   `json:"all,omitempty"`
	}
	if err := json.Unmarshal(params, &invalidateParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if invalidateParams.All {
		s.cache.Clear()
		return map[string]int{"deleted": 0}, nil
	}

	if invalidateParams.Pattern != "" {
		deleted := s.cache.InvalidateByPattern(invalidateParams.Pattern)
		return map[string]int{"deleted": deleted}, nil
	}

	return map[string]int{"deleted": 0}, nil
}

func (s *StdioServer) writeResponse(id interface{}, result interface{}) {
	// Notifications have no id
	if id == nil {
		return
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.out, string(data))
}

func (s *StdioServer) writeError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal error: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.out, string(respData))
}

func (s *StdioServer) writeResponses(responses []JSONRPCResponse) {
	data, err := json.Marshal(responses)
	if err != nil {
		log.Printf("failed to marshal batch response: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.out, string(data))
}

// LogError writes an error message to the error stream.
func (s *StdioServer) LogError(format string, args ...interface{}) {
	fmt.Fprintf(s.err, "[tokman-mcp] "+format+"\n", args...)
}

// LogInfo writes an info message to the error stream.
func (s *StdioServer) LogInfo(format string, args ...interface{}) {
	fmt.Fprintf(s.err, "[tokman-mcp] "+format+"\n", args...)
}
