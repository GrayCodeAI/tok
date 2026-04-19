package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lakshmanpatel/tok/internal/filter"
)

func TestServer_Initialize(t *testing.T) {
	server := NewServer("test", "1.0.0", nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MessageTypeInitialize,
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}`),
	}

	resp := server.handleRequest(&req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*InitializeResult)
	if !ok {
		t.Fatalf("expected InitializeResult, got %T", resp.Result)
	}

	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("expected protocol version %s, got %s", ProtocolVersion, result.ProtocolVersion)
	}

	if result.ServerInfo.Name != "test" {
		t.Errorf("expected server name 'test', got %s", result.ServerInfo.Name)
	}
}

func TestServer_ToolsList(t *testing.T) {
	server := NewServer("test", "1.0.0", nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MessageTypeToolsList,
	}

	resp := server.handleRequest(&req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*ToolsListResult)
	if !ok {
		t.Fatalf("expected ToolsListResult, got %T", resp.Result)
	}

	// Should have 5 built-in tools
	if len(result.Tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(result.Tools))
	}

	// Check for expected tools
	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"tok_filter",
		"tok_compress_file",
		"tok_analyze_output",
		"tok_get_stats",
		"tok_explain_layers",
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("expected tool %q not found", name)
		}
	}
}

func TestServer_ToolsCall_Filter(t *testing.T) {
	cfg := filter.PipelineConfig{Mode: filter.ModeMinimal}
	pipeline := filter.NewPipelineCoordinator(cfg)
	server := NewServer("test", "1.0.0", pipeline)

	// First initialize
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  MessageTypeInitialize,
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}`),
	}
	server.handleRequest(&initReq)

	// Now call the filter tool
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  MessageTypeToolsCall,
		Params:  json.RawMessage(`{"name":"tok_filter","arguments":{"text":"Hello world test","mode":"minimal"}}`),
	}

	resp := server.handleRequest(&req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*ToolsCallResult)
	if !ok {
		t.Fatalf("expected ToolsCallResult, got %T", resp.Result)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	// Content should be JSON with filter results
	if result.Content[0].Type != "text" {
		t.Errorf("expected content type 'text', got %s", result.Content[0].Type)
	}
}

func TestServer_RunStdio(t *testing.T) {
	server := NewServer("test", "1.0.0", nil)

	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}` + "\n")
	output := &bytes.Buffer{}

	// Run in goroutine since RunStdio blocks
	done := make(chan error, 1)
	go func() {
		done <- server.runTransport(input, output)
	}()

	// Wait for processing (EOF will end it)
	<-done

	// Check output contains valid JSON response
	outputStr := output.String()
	if !strings.Contains(outputStr, "result") {
		t.Error("expected result in output")
	}
}

func TestServer_UnknownMethod(t *testing.T) {
	server := NewServer("test", "1.0.0", nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	resp := server.handleRequest(&req)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}

	if resp.Error.Code != ErrorCodeMethodNotFound {
		t.Errorf("expected error code %d, got %d", ErrorCodeMethodNotFound, resp.Error.Code)
	}
}

func TestNewTextContent(t *testing.T) {
	content := NewTextContent("test message")

	if content.Type != "text" {
		t.Errorf("expected type 'text', got %s", content.Type)
	}

	if content.Text != "test message" {
		t.Errorf("expected text 'test message', got %s", content.Text)
	}
}
