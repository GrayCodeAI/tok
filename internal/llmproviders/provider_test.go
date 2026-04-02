// Package llmproviders provides LLM provider integration for TokMan agents
package llmproviders

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewOpenAIProvider(t *testing.T) {
	p := NewOpenAIProvider("test-key", "gpt-4")
	if p == nil {
		t.Fatal("expected provider")
	}
	if p.Name() != "openai" {
		t.Errorf("expected 'openai', got %s", p.Name())
	}
}

func TestNewAnthropicProvider(t *testing.T) {
	p := NewAnthropicProvider("test-key", "claude-3")
	if p == nil {
		t.Fatal("expected provider")
	}
	if p.Name() != "anthropic" {
		t.Errorf("expected 'anthropic', got %s", p.Name())
	}
}

func TestNewOllamaProvider(t *testing.T) {
	p := NewOllamaProvider("llama2")
	if p == nil {
		t.Fatal("expected provider")
	}
	if p.Name() != "ollama" {
		t.Errorf("expected 'ollama', got %s", p.Name())
	}
}

func TestProviderFactory(t *testing.T) {
	factory := NewProviderFactory()

	factory.Register("openai", ProviderConfig{
		APIKey: "test-key",
		Model:  "gpt-4",
	})

	factory.Register("anthropic", ProviderConfig{
		APIKey: "test-key",
		Model:  "claude-3",
	})

	factory.Register("ollama", ProviderConfig{
		Model: "llama2",
	})

	// Create OpenAI provider
	p1, err := factory.Create("openai")
	if err != nil {
		t.Fatalf("failed to create openai: %v", err)
	}
	if p1.Name() != "openai" {
		t.Errorf("expected 'openai', got %s", p1.Name())
	}

	// Create Anthropic provider
	p2, err := factory.Create("anthropic")
	if err != nil {
		t.Fatalf("failed to create anthropic: %v", err)
	}
	if p2.Name() != "anthropic" {
		t.Errorf("expected 'anthropic', got %s", p2.Name())
	}

	// Create Ollama provider
	p3, err := factory.Create("ollama")
	if err != nil {
		t.Fatalf("failed to create ollama: %v", err)
	}
	if p3.Name() != "ollama" {
		t.Errorf("expected 'ollama', got %s", p3.Name())
	}

	// Unknown provider
	_, err = factory.Create("unknown")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestCompletionRequest(t *testing.T) {
	req := CompletionRequest{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   1000,
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
	}

	if req.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %s", req.Model)
	}

	if len(req.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(req.Messages))
	}
}

func TestMessageJSON(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
		Name:    "test",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Role != "user" || decoded.Content != "Hello" {
		t.Errorf("expected role 'user' and content 'Hello', got %s/%s", decoded.Role, decoded.Content)
	}
}

func TestToolDefinition(t *testing.T) {
	tool := ToolDefinition{
		Type: "function",
		Function: FunctionDefinition{
			Name:        "search",
			Description: "Search the web",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
			},
		},
	}

	if tool.Type != "function" {
		t.Errorf("expected type 'function', got %s", tool.Type)
	}

	if tool.Function.Name != "search" {
		t.Errorf("expected function name 'search', got %s", tool.Function.Name)
	}
}

func TestStreamChunk(t *testing.T) {
	chunk := StreamChunk{
		Content: "Hello",
		Done:    false,
	}

	if chunk.Content != "Hello" {
		t.Errorf("expected content 'Hello', got %s", chunk.Content)
	}

	if chunk.Done {
		t.Error("expected done to be false")
	}
}

func TestUsageInfo(t *testing.T) {
	usage := UsageInfo{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.TotalTokens != usage.PromptTokens+usage.CompletionTokens {
		t.Errorf("expected total %d, got %d", usage.PromptTokens+usage.CompletionTokens, usage.TotalTokens)
	}
}

func TestCompletionResponse(t *testing.T) {
	resp := CompletionResponse{
		ID:      "resp-123",
		Model:   "gpt-4",
		Content: "Hello! How can I help you?",
		Usage: UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	if resp.ID != "resp-123" {
		t.Errorf("expected ID 'resp-123', got %s", resp.ID)
	}

	if resp.Usage.TotalTokens != 30 {
		t.Errorf("expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestOllamaComplete(t *testing.T) {
	// Test that Ollama provider can be created and has correct structure
	p := NewOllamaProvider("llama2")
	if p.baseURL != "http://localhost:11434" {
		t.Errorf("expected baseURL 'http://localhost:11434', got %s", p.baseURL)
	}
	if p.model != "llama2" {
		t.Errorf("expected model 'llama2', got %s", p.model)
	}
}

func TestAnthropicEmbedNotSupported(t *testing.T) {
	p := NewAnthropicProvider("test-key", "claude-3")
	_, err := p.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error for anthropic embeddings")
	}
}

func TestProviderConfig(t *testing.T) {
	config := ProviderConfig{
		APIKey:  "test-key",
		Model:   "gpt-4",
		BaseURL: "https://custom.api.com",
	}

	if config.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got %s", config.APIKey)
	}

	if config.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got %s", config.Model)
	}
}

func TestProviderFactoryRegister(t *testing.T) {
	factory := NewProviderFactory()

	factory.Register("custom", ProviderConfig{
		APIKey: "test",
		Model:  "custom-model",
	})

	if len(factory.config) != 1 {
		t.Errorf("expected 1 registered provider, got %d", len(factory.config))
	}

	_, err := factory.Create("custom")
	if err == nil {
		t.Error("expected error for unsupported provider type")
	}
}

func BenchmarkOpenAIProviderCreate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewOpenAIProvider("test-key", "gpt-4")
	}
}

func BenchmarkProviderFactoryCreate(b *testing.B) {
	factory := NewProviderFactory()
	factory.Register("openai", ProviderConfig{APIKey: "test", Model: "gpt-4"})
	factory.Register("ollama", ProviderConfig{Model: "llama2"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory.Create("ollama")
	}
}

// Test helper for HTTP mocking
type mockTransport struct {
	Response *http.Response
	Err      error
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	return t.Response, nil
}

func TestOpenAIProviderWithMock(t *testing.T) {
	// Create a mock response
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(`{
			"id": "test-id",
			"model": "gpt-4",
			"choices": [{"message": {"content": "Hello"}}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
		}`)),
	}

	client := &http.Client{
		Transport: &mockTransport{Response: mockResp},
		Timeout:   5 * time.Second,
	}

	p := &OpenAIProvider{
		apiKey:  "test",
		baseURL: "https://api.test.com",
		model:   "gpt-4",
		client:  client,
	}

	req := CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	resp, err := p.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello" {
		t.Errorf("expected content 'Hello', got %s", resp.Content)
	}

	if resp.Usage.TotalTokens != 30 {
		t.Errorf("expected 30 tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestOpenAIProviderErrorHandling(t *testing.T) {
	client := &http.Client{
		Transport: &mockTransport{
			Response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "Invalid API key"}`)),
			},
		},
		Timeout: 5 * time.Second,
	}

	p := &OpenAIProvider{
		apiKey:  "invalid",
		baseURL: "https://api.test.com",
		model:   "gpt-4",
		client:  client,
	}

	req := CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	_, err := p.Complete(context.Background(), req)
	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}
