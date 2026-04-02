// Package tools provides tests for MCP tool implementations.
package tools

import (
	"context"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/mcp"
)

func TestRegisterAllTools(t *testing.T) {
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)

	RegisterAllTools(registry, cache)

	tools := registry.ListTools()
	if len(tools) != 22 {
		t.Errorf("expected 22 tools, got %d", len(tools))
	}

	// Verify all 22 tools are present
	if len(tools) != 22 {
		t.Errorf("expected 22 tools, got %d", len(tools))
	}

	// Verify all expected tools are present
	expectedTools := []string{
		"ctx_read", "ctx_delta", "ctx_grep", "ctx_hash",
		"ctx_cache_info", "ctx_invalidate", "ctx_compact",
		"ctx_summary", "ctx_remember", "ctx_recall",
		"ctx_search_memory", "ctx_bundle", "ctx_bundle_changed",
		"ctx_bundle_summary", "ctx_exec", "ctx_tldr",
		"ctx_patterns", "ctx_modes", "ctx_mode",
		"ctx_status", "ctx_config", "ctx_mcp",
	}

	for _, name := range expectedTools {
		_, ok := registry.GetTool(name)
		if !ok {
			t.Errorf("expected tool %s not found", name)
		}
	}
}

func TestCtxReadHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)
	RegisterAllTools(registry, cache)

	handler, ok := registry.GetHandler("ctx_read")
	if !ok {
		t.Fatal("ctx_read handler not found")
	}

	// Test with invalid params
	_, err := handler(context.Background(), map[string]interface{}{})
	if err != mcp.ErrInvalidParams {
		t.Errorf("expected ErrInvalidParams, got %v", err)
	}
}

func TestCtxHashHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)
	RegisterAllTools(registry, cache)

	handler, ok := registry.GetHandler("ctx_hash")
	if !ok {
		t.Fatal("ctx_hash handler not found")
	}

	// Test with content
	result, err := handler(context.Background(), map[string]interface{}{
		"content": "hello world",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	if resultMap["full_hash"] == "" {
		t.Error("expected full_hash to be set")
	}
	if resultMap["short_hash"] == "" {
		t.Error("expected short_hash to be set")
	}
}

func TestCtxCacheInfoHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)
	RegisterAllTools(registry, cache)

	handler, ok := registry.GetHandler("ctx_cache_info")
	if !ok {
		t.Fatal("ctx_cache_info handler not found")
	}

	result, err := handler(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	if _, ok := resultMap["total_entries"]; !ok {
		t.Error("expected total_entries in result")
	}
}

func TestCtxStatusHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)
	RegisterAllTools(registry, cache)

	handler, ok := registry.GetHandler("ctx_status")
	if !ok {
		t.Fatal("ctx_status handler not found")
	}

	result, err := handler(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	if resultMap["tools_count"] != 22 {
		t.Errorf("expected 22 tools, got %v", resultMap["tools_count"])
	}
}
