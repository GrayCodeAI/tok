// Package integration provides end-to-end integration tests.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/cortex"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/mcp"
	"github.com/GrayCodeAI/tokman/internal/mcp/tools"
)

// TestFullPipeline tests the complete TokMan pipeline.
func TestFullPipeline(t *testing.T) {
	// Set up MCP server components
	registry := mcp.NewToolRegistry()
	cache := mcp.NewHashCache(100, 100*1024*1024)
	tools.RegisterAllTools(registry, cache)

	// Test 1: Hash computation
	t.Run("hash_computation", func(t *testing.T) {
		handler, ok := registry.GetHandler("ctx_hash")
		if !ok {
			t.Fatal("ctx_hash not found")
		}

		result, err := handler(context.Background(), map[string]interface{}{
			"content": "hello world",
		})
		if err != nil {
			t.Fatalf("handler failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result is not a map")
		}

		fullHash, ok := resultMap["full_hash"].(string)
		if !ok || len(fullHash) != 64 {
			t.Errorf("expected 64 char hex hash, got %v", fullHash)
		}
	})

	// Test 2: Compression
	t.Run("compression", func(t *testing.T) {
		handler, ok := registry.GetHandler("ctx_compact")
		if !ok {
			t.Fatal("ctx_compact not found")
		}

		content := "line1\nline2\nline3\n"
		result, err := handler(context.Background(), map[string]interface{}{
			"content": content,
			"mode":    "minimal",
		})
		if err != nil {
			t.Fatalf("handler failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result is not a map")
		}

		if _, ok := resultMap["original_tokens"]; !ok {
			t.Error("expected original_tokens in result")
		}
		if _, ok := resultMap["final_tokens"]; !ok {
			t.Error("expected final_tokens in result")
		}
	})

	// Test 3: Cortex gate application
	t.Run("cortex_gates", func(t *testing.T) {
		gateRegistry := cortex.NewGateRegistry()
		gates := cortex.DefaultGates()
		for _, gate := range gates {
			gateRegistry.Register(gate)
		}

		// Test with Go code
		goCode := `package main
func main() {
	println("hello")
}`

		detection := gateRegistry.Analyze(goCode)
		if detection.ContentType != cortex.SourceCode {
			t.Errorf("expected SourceCode, got %v", detection.ContentType)
		}
		if detection.Language != cortex.LangGo {
			t.Errorf("expected LangGo, got %v", detection.Language)
		}
	})

	// Test 4: Filter engine
	t.Run("filter_engine", func(t *testing.T) {
		engine := filter.NewEngine(filter.ModeMinimal)

		logContent := `[INFO] Starting build
[ERROR] Compilation failed
[ERROR] Compilation failed
[INFO] Build complete`

		processed, saved := engine.Process(logContent)
		if len(processed) == 0 {
			t.Error("processed content is empty")
		}
		t.Logf("Saved %d bytes", saved)
	})

	// Test 5: Memory operations
	t.Run("memory_operations", func(t *testing.T) {
		rememberHandler, _ := registry.GetHandler("ctx_remember")
		recallHandler, _ := registry.GetHandler("ctx_recall")
		searchHandler, _ := registry.GetHandler("ctx_search_memory")

		// Store memory
		_, err := rememberHandler(context.Background(), map[string]interface{}{
			"key":   "test_key",
			"value": "test_value",
			"tags":  []interface{}{"test", "integration"},
		})
		if err != nil {
			t.Fatalf("remember failed: %v", err)
		}

		// Recall memory
		result, err := recallHandler(context.Background(), map[string]interface{}{
			"key": "test_key",
		})
		if err != nil {
			t.Fatalf("recall failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["value"] != "test_value" {
			t.Errorf("expected test_value, got %v", resultMap["value"])
		}

		// Search memory
		searchResult, err := searchHandler(context.Background(), map[string]interface{}{
			"query": "test",
		})
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}

		searchMap := searchResult.(map[string]interface{})
		if count, ok := searchMap["count"].(int); !ok || count == 0 {
			t.Error("expected at least one search result")
		}
	})
}

// TestPerformance verifies system performance characteristics.
func TestPerformance(t *testing.T) {
	t.Run("large_file_processing", func(t *testing.T) {
		// Generate 10KB of log content
		var content string
		for i := 0; i < 500; i++ {
			content += "[INFO] This is a log message with some data that might be repeated\n"
		}

		engine := filter.NewEngine(filter.ModeAggressive)

		start := time.Now()
		processed, _ := engine.Process(content)
		duration := time.Since(start)

		t.Logf("Processed %d bytes in %v", len(content), duration)
		t.Logf("Output: %d bytes", len(processed))

		// Should process in under 2s (accommodates race detector overhead)
		if duration > 2*time.Second {
			t.Errorf("processing took too long: %v", duration)
		}
	})
}
