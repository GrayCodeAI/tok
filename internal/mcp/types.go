// Package mcp implements the Model Context Protocol (MCP) server for tok.
// This allows AI assistants to use tok's filtering capabilities as MCP tools.
package mcp

import (
	"encoding/json"
)

// ProtocolVersion is the MCP protocol version we support
const ProtocolVersion = "2024-11-05"

// Message types
const (
	MessageTypeInitialize  = "initialize"
	MessageTypeInitialized = "initialized"
	MessageTypeToolsList   = "tools/list"
	MessageTypeToolsCall   = "tools/call"
	MessageTypePing        = "ping"
	MessageTypePong        = "pong"
	MessageTypeError       = "error"
)

// JSONRPCRequest is a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError is a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// InitializeParams contains initialization parameters
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

// InitializeResult is the response to initialize
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
}

// ClientCapabilities describes client capabilities
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability indicates client supports roots
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability indicates client supports sampling
type SamplingCapability struct{}

// ServerCapabilities describes server capabilities
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability describes tool support
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Implementation identifies the implementation
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an available tool
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolsListResult is the response to tools/list
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolsCallParams contains tool call parameters
type ToolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolsCallResult is the response to tools/call
type ToolsCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents a content item in a response
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewTextContent creates text content
func NewTextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// ToolHandler is a function that handles a tool call
type ToolHandler func(arguments json.RawMessage) (*ToolsCallResult, error)
