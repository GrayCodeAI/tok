// Package mcphost provides MCP (Model Context Protocol) host management for TokMan
package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Host manages MCP server connections
type Host struct {
	id           string
	name         string
	config       HostConfig
	servers      map[string]*Server
	sessions     map[string]*Session
	mu           sync.RWMutex
	eventHandler EventHandler
}

// HostConfig holds host configuration
type HostConfig struct {
	Name              string
	Version           string
	MaxConnections    int
	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration
	AutoReconnect     bool
	Capabilities      HostCapabilities
}

// HostCapabilities defines what the host supports
type HostCapabilities struct {
	Tools     bool
	Resources bool
	Prompts   bool
	Logging   bool
}

// Server represents a connected MCP server
type Server struct {
	ID           string
	Name         string
	Version      string
	Transport    Transport
	Status       ServerStatus
	Capabilities ServerCapabilities
	Tools        []Tool
	Resources    []Resource
	Prompts      []Prompt
	LastPing     time.Time
	ConnectTime  time.Time
	mu           sync.RWMutex
}

// ServerStatus represents server connection status
type ServerStatus string

const (
	ServerStatusDisconnected ServerStatus = "disconnected"
	ServerStatusConnecting   ServerStatus = "connecting"
	ServerStatusConnected    ServerStatus = "connected"
	ServerStatusError        ServerStatus = "error"
)

// ServerCapabilities defines what a server supports
type ServerCapabilities struct {
	Tools     *ToolsCapability
	Resources *ResourcesCapability
	Prompts   *PromptsCapability
	Logging   *LoggingCapability
}

// ToolsCapability defines tool support
type ToolsCapability struct {
	ListChanged bool
}

// ResourcesCapability defines resource support
type ResourcesCapability struct {
	Subscribe   bool
	ListChanged bool
}

// PromptsCapability defines prompt support
type PromptsCapability struct {
	ListChanged bool
}

// LoggingCapability defines logging support
type LoggingCapability struct {
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// Resource represents an MCP resource
type Resource struct {
	URI         string
	Name        string
	Description string
	MimeType    string
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string
	Description string
	Arguments   []PromptArgument
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string
	Description string
	Required    bool
}

// Transport defines the interface for MCP transports
type Transport interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Send(message Message) error
	Receive() (Message, error)
	IsConnected() bool
}

// Message represents an MCP protocol message
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Session represents a client session
type Session struct {
	ID           string
	ClientInfo   ClientInfo
	ServerID     string
	Status       SessionStatus
	CreatedAt    time.Time
	LastActivity time.Time
	mu           sync.RWMutex
}

// SessionStatus represents session status
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusInactive SessionStatus = "inactive"
	SessionStatusClosed   SessionStatus = "closed"
)

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// EventHandler handles host events
type EventHandler func(event HostEvent)

// HostEvent represents a host event
type HostEvent struct {
	Type      EventType
	Timestamp time.Time
	ServerID  string
	SessionID string
	Message   string
	Data      interface{}
}

// EventType represents event types
type EventType string

const (
	EventServerConnected    EventType = "server_connected"
	EventServerDisconnected EventType = "server_disconnected"
	EventServerError        EventType = "server_error"
	EventSessionCreated     EventType = "session_created"
	EventSessionClosed      EventType = "session_closed"
	EventToolCalled         EventType = "tool_called"
	EventResourceAccessed   EventType = "resource_accessed"
)

// NewHost creates a new MCP host
func NewHost(config HostConfig) *Host {
	return &Host{
		id:       generateHostID(),
		name:     config.Name,
		config:   config,
		servers:  make(map[string]*Server),
		sessions: make(map[string]*Session),
	}
}

// ID returns the host ID
func (h *Host) ID() string {
	return h.id
}

// Name returns the host name
func (h *Host) Name() string {
	return h.name
}

// SetEventHandler sets the event handler
func (h *Host) SetEventHandler(handler EventHandler) {
	h.eventHandler = handler
}

func (h *Host) emit(event HostEvent) {
	if h.eventHandler != nil {
		h.eventHandler(event)
	}
}

// RegisterServer registers an MCP server
func (h *Host) RegisterServer(id string, name string, transport Transport) (*Server, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.servers[id]; exists {
		return nil, fmt.Errorf("server %s already registered", id)
	}

	server := &Server{
		ID:        id,
		Name:      name,
		Transport: transport,
		Status:    ServerStatusDisconnected,
		Tools:     make([]Tool, 0),
		Resources: make([]Resource, 0),
		Prompts:   make([]Prompt, 0),
	}

	h.servers[id] = server

	return server, nil
}

// ConnectServer connects to an MCP server
func (h *Host) ConnectServer(ctx context.Context, serverID string) error {
	h.mu.RLock()
	server, exists := h.servers[serverID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	server.mu.Lock()
	server.Status = ServerStatusConnecting
	server.mu.Unlock()

	// Connect transport
	ctx, cancel := context.WithTimeout(ctx, h.config.ConnectionTimeout)
	defer cancel()

	if err := server.Transport.Connect(ctx); err != nil {
		server.mu.Lock()
		server.Status = ServerStatusError
		server.mu.Unlock()

		h.emit(HostEvent{
			Type:      EventServerError,
			Timestamp: time.Now(),
			ServerID:  serverID,
			Message:   err.Error(),
		})

		return err
	}

	// Perform initialization handshake
	if err := h.initializeServer(ctx, server); err != nil {
		server.Transport.Disconnect()
		server.mu.Lock()
		server.Status = ServerStatusError
		server.mu.Unlock()
		return err
	}

	server.mu.Lock()
	server.Status = ServerStatusConnected
	server.ConnectTime = time.Now()
	server.mu.Unlock()

	h.emit(HostEvent{
		Type:      EventServerConnected,
		Timestamp: time.Now(),
		ServerID:  serverID,
		Message:   fmt.Sprintf("Connected to server %s", server.Name),
	})

	return nil
}

func (h *Host) initializeServer(ctx context.Context, server *Server) error {
	// Send initialize request
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    h.config.Capabilities,
		"clientInfo": map[string]string{
			"name":    h.config.Name,
			"version": h.config.Version,
		},
	}

	paramsJSON, _ := json.Marshal(initParams)
	initRequest := Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  paramsJSON,
	}

	if err := server.Transport.Send(initRequest); err != nil {
		return err
	}

	// Wait for response
	response, err := server.Transport.Receive()
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("initialization failed: %s", response.Error.Message)
	}

	// Parse server capabilities
	var result struct {
		ProtocolVersion string             `json:"protocolVersion"`
		Capabilities    ServerCapabilities `json:"capabilities"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}

	if err := json.Unmarshal(response.Result, &result); err != nil {
		return err
	}

	server.mu.Lock()
	server.Capabilities = result.Capabilities
	server.Version = result.ServerInfo.Version
	server.mu.Unlock()

	// Send initialized notification
	notification := Message{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	return server.Transport.Send(notification)
}

// DisconnectServer disconnects from a server
func (h *Host) DisconnectServer(serverID string) error {
	h.mu.RLock()
	server, exists := h.servers[serverID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	server.mu.Lock()
	if server.Status != ServerStatusConnected {
		server.mu.Unlock()
		return nil
	}
	server.mu.Unlock()

	// Send close notification
	closeMsg := Message{
		JSONRPC: "2.0",
		Method:  "notifications/closed",
	}
	_ = server.Transport.Send(closeMsg)

	// Disconnect transport
	err := server.Transport.Disconnect()

	server.mu.Lock()
	server.Status = ServerStatusDisconnected
	server.mu.Unlock()

	h.emit(HostEvent{
		Type:      EventServerDisconnected,
		Timestamp: time.Now(),
		ServerID:  serverID,
		Message:   fmt.Sprintf("Disconnected from server %s", server.Name),
	})

	return err
}

// GetServer returns a server by ID
func (h *Host) GetServer(id string) (*Server, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	server, exists := h.servers[id]
	if !exists {
		return nil, fmt.Errorf("server %s not found", id)
	}

	return server, nil
}

// ListServers returns all registered servers
func (h *Host) ListServers() []*Server {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*Server, 0, len(h.servers))
	for _, server := range h.servers {
		result = append(result, server)
	}

	return result
}

// CreateSession creates a new client session
func (h *Host) CreateSession(clientInfo ClientInfo, serverID string) (*Session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.servers[serverID]; !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	session := &Session{
		ID:           generateSessionID(),
		ClientInfo:   clientInfo,
		ServerID:     serverID,
		Status:       SessionStatusActive,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	h.sessions[session.ID] = session

	h.emit(HostEvent{
		Type:      EventSessionCreated,
		Timestamp: time.Now(),
		SessionID: session.ID,
		Message:   fmt.Sprintf("Session created for %s", clientInfo.Name),
	})

	return session, nil
}

// CloseSession closes a session
func (h *Host) CloseSession(sessionID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.mu.Lock()
	session.Status = SessionStatusClosed
	session.mu.Unlock()

	delete(h.sessions, sessionID)

	h.emit(HostEvent{
		Type:      EventSessionClosed,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Message:   "Session closed",
	})

	return nil
}

// ListTools returns all tools from a server
func (h *Host) ListTools(serverID string) ([]Tool, error) {
	server, err := h.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	return server.Tools, nil
}

// CallTool invokes a tool on a server
func (h *Host) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (json.RawMessage, error) {
	server, err := h.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	server.mu.RLock()
	if server.Status != ServerStatusConnected {
		server.mu.RUnlock()
		return nil, fmt.Errorf("server not connected")
	}
	server.mu.RUnlock()

	// Build tool call request
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}
	paramsJSON, _ := json.Marshal(params)

	request := Message{
		JSONRPC: "2.0",
		ID:      generateRequestID(),
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	// Send request
	if err := server.Transport.Send(request); err != nil {
		return nil, err
	}

	// Wait for response with timeout
	ctx, cancel := context.WithTimeout(ctx, h.config.RequestTimeout)
	defer cancel()

	done := make(chan Message, 1)
	go func() {
		response, _ := server.Transport.Receive()
		done <- response
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("tool call timeout")
	case response := <-done:
		if response.Error != nil {
			return nil, fmt.Errorf("tool call failed: %s", response.Error.Message)
		}

		h.emit(HostEvent{
			Type:      EventToolCalled,
			Timestamp: time.Now(),
			ServerID:  serverID,
			Message:   fmt.Sprintf("Tool %s called", toolName),
			Data:      arguments,
		})

		return response.Result, nil
	}
}

// ReadResource reads a resource from a server
func (h *Host) ReadResource(ctx context.Context, serverID string, uri string) (json.RawMessage, error) {
	server, err := h.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	server.mu.RLock()
	if server.Status != ServerStatusConnected {
		server.mu.RUnlock()
		return nil, fmt.Errorf("server not connected")
	}
	server.mu.RUnlock()

	params := map[string]string{"uri": uri}
	paramsJSON, _ := json.Marshal(params)

	request := Message{
		JSONRPC: "2.0",
		ID:      generateRequestID(),
		Method:  "resources/read",
		Params:  paramsJSON,
	}

	if err := server.Transport.Send(request); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, h.config.RequestTimeout)
	defer cancel()

	done := make(chan Message, 1)
	go func() {
		response, _ := server.Transport.Receive()
		done <- response
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("resource read timeout")
	case response := <-done:
		if response.Error != nil {
			return nil, fmt.Errorf("resource read failed: %s", response.Error.Message)
		}

		h.emit(HostEvent{
			Type:      EventResourceAccessed,
			Timestamp: time.Now(),
			ServerID:  serverID,
			Message:   fmt.Sprintf("Resource %s accessed", uri),
		})

		return response.Result, nil
	}
}

// Shutdown gracefully shuts down the host
func (h *Host) Shutdown() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Disconnect all servers
	for _, server := range h.servers {
		if server.Status == ServerStatusConnected {
			_ = server.Transport.Disconnect()
		}
	}

	// Close all sessions
	for _, session := range h.sessions {
		session.Status = SessionStatusClosed
	}

	return nil
}

func generateHostID() string {
	return fmt.Sprintf("host-%d", time.Now().Unix())
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

var requestCounter int

func generateRequestID() int {
	requestCounter++
	return requestCounter
}
