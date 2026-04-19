package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/lakshmanpatel/tok/internal/filter"
)

// Server implements an MCP server
type Server struct {
	name        string
	version     string
	tools       map[string]*registeredTool
	handlers    map[string]ToolHandler
	pipeline    *filter.PipelineCoordinator
	mu          sync.RWMutex
	initialized bool
}

// registeredTool combines tool definition with handler
type registeredTool struct {
	tool    Tool
	handler ToolHandler
}

// NewServer creates a new MCP server
func NewServer(name, version string, pipeline *filter.PipelineCoordinator) *Server {
	s := &Server{
		name:     name,
		version:  version,
		tools:    make(map[string]*registeredTool),
		handlers: make(map[string]ToolHandler),
		pipeline: pipeline,
	}

	// Register built-in tools
	s.registerBuiltinTools()

	return s
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}

	s.tools[tool.Name] = &registeredTool{
		tool:    tool,
		handler: handler,
	}

	return nil
}

// registerBuiltinTools registers tok's built-in tools
func (s *Server) registerBuiltinTools() {
	// Tool: tok_filter
	s.RegisterTool(Tool{
		Name:        "tok_filter",
		Description: "Filter and compress text using tok's compression pipeline. Reduces token usage by 60-90%.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"text": {
					"type": "string",
					"description": "Text content to filter"
				},
				"mode": {
					"type": "string",
					"enum": ["minimal", "aggressive"],
					"default": "minimal",
					"description": "Compression level: minimal preserves more, aggressive compresses more"
				},
				"budget": {
					"type": "integer",
					"description": "Maximum tokens allowed in output (0 = unlimited)"
				},
				"query": {
					"type": "string",
					"description": "Query intent for goal-driven filtering (debug/review/deploy)"
				}
			},
			"required": ["text"]
		}`),
	}, s.handleFilter)

	// Tool: tok_compress_file
	s.RegisterTool(Tool{
		Name:        "tok_compress_file",
		Description: "Read and compress a file. Returns filtered content with token savings report.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "Path to file to compress"
				},
				"mode": {
					"type": "string",
					"enum": ["minimal", "aggressive"],
					"default": "minimal",
					"description": "Compression level: minimal preserves more, aggressive compresses more"
				},
				"max_lines": {
					"type": "integer",
					"description": "Maximum lines to return (0 = unlimited)"
				}
			},
			"required": ["path"]
		}`),
	}, s.handleCompressFile)

	// Tool: tok_analyze_output
	s.RegisterTool(Tool{
		Name:        "tok_analyze_output",
		Description: "Analyze command output structure without filtering. Returns metrics on what would be filtered.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"text": {
					"type": "string",
					"description": "Command output to analyze"
				}
			},
			"required": ["text"]
		}`),
	}, s.handleAnalyzeOutput)

	// Tool: tok_get_stats
	s.RegisterTool(Tool{
		Name:        "tok_get_stats",
		Description: "Get tok usage statistics including total tokens saved and compression ratio.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, s.handleGetStats)

	// Tool: tok_explain_layers
	s.RegisterTool(Tool{
		Name:        "tok_explain_layers",
		Description: "Explain what each compression layer does and which are enabled.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, s.handleExplainLayers)
}

// RunStdio runs the server with stdio transport
func (s *Server) RunStdio() error {
	return s.runTransport(os.Stdin, os.Stdout)
}

// runTransport runs the server with arbitrary input/output
func (s *Server) runTransport(input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	encoder := json.NewEncoder(output)

	for {
		// Read line (JSON-RPC message)
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		// Parse request
		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(encoder, nil, ErrorCodeParseError, "Parse error", err.Error())
			continue
		}

		// Handle request
		resp := s.handleRequest(&req)
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("encode error: %w", err)
		}
	}
}

// handleRequest processes a single JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case MessageTypeInitialize:
		result, err := s.handleInitialize(req.Params)
		if err != nil {
			resp.Error = &JSONRPCError{
				Code:    ErrorCodeInvalidParams,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}

	case MessageTypeToolsList:
		resp.Result = s.handleToolsList()

	case MessageTypeToolsCall:
		result, err := s.handleToolsCall(req.Params)
		if err != nil {
			resp.Error = &JSONRPCError{
				Code:    ErrorCodeInternalError,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}

	case MessageTypePing:
		resp.Result = struct{}{}

	default:
		resp.Error = &JSONRPCError{
			Code:    ErrorCodeMethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return resp
}

// handleInitialize processes the initialize request
func (s *Server) handleInitialize(params json.RawMessage) (*InitializeResult, error) {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: Implementation{
			Name:    s.name,
			Version: s.version,
		},
	}, nil
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList() *ToolsListResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, rt := range s.tools {
		tools = append(tools, rt.tool)
	}

	return &ToolsListResult{Tools: tools}
}

// handleToolsCall routes tool calls to handlers
func (s *Server) handleToolsCall(params json.RawMessage) (*ToolsCallResult, error) {
	var callParams ToolsCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	s.mu.RLock()
	rt, exists := s.tools[callParams.Name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", callParams.Name)
	}

	return rt.handler(callParams.Arguments)
}

// sendError sends an error response
func (s *Server) sendError(encoder *json.Encoder, id interface{}, code int, message string, data interface{}) {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	encoder.Encode(resp)
}
