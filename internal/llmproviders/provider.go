// Package llmproviders provides LLM provider integration for TokMan agents
package llmproviders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider defines an LLM provider interface
type Provider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	Embed(ctx context.Context, text string) ([]float64, error)
}

// CompletionRequest holds a completion request
type CompletionRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
	MaxTokens   int
	TopP        float64
	Stop        []string
	Tools       []ToolDefinition
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ToolDefinition represents a tool definition for function calling
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition represents a function definition
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CompletionResponse holds a completion response
type CompletionResponse struct {
	ID      string
	Model   string
	Content string
	Usage   UsageInfo
}

// UsageInfo holds token usage information
type UsageInfo struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content string
	Done    bool
}

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &CompletionResponse{
		ID:      result.ID,
		Model:   result.Model,
		Content: result.Choices[0].Message.Content,
		Usage: UsageInfo{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	ch <- StreamChunk{Content: "streaming...", Done: true}
	close(ch)
	return ch, nil
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	payload := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai embed failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	messages := make([]map[string]interface{}, 0)
	systemPrompt := ""

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"system":      systemPrompt,
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	content := ""
	for _, c := range result.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &CompletionResponse{
		ID:      result.ID,
		Model:   result.Model,
		Content: content,
		Usage: UsageInfo{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}, nil
}

func (p *AnthropicProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	ch <- StreamChunk{Content: "anthropic streaming...", Done: true}
	close(ch)
	return ch, nil
}

func (p *AnthropicProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("anthropic does not support embeddings")
}

// OllamaProvider implements the Provider interface for local Ollama
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: "http://localhost:11434",
		model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	prompt := ""
	for _, msg := range req.Messages {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"prompt":      prompt,
		"temperature": req.Temperature,
		"stream":      false,
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Model    string `json:"model"`
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return &CompletionResponse{
		Model:   result.Model,
		Content: result.Response,
	}, nil
}

func (p *OllamaProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	ch <- StreamChunk{Content: "ollama streaming...", Done: true}
	close(ch)
	return ch, nil
}

func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": text,
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/embeddings", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama embed failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return result.Embedding, nil
}

// ProviderFactory creates providers by type
type ProviderFactory struct {
	config map[string]ProviderConfig
}

// ProviderConfig holds provider configuration
type ProviderConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// NewProviderFactory creates a new factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		config: make(map[string]ProviderConfig),
	}
}

// Register registers a provider configuration
func (pf *ProviderFactory) Register(name string, config ProviderConfig) {
	pf.config[name] = config
}

// Create creates a provider by name
func (pf *ProviderFactory) Create(name string) (Provider, error) {
	config, ok := pf.config[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	switch name {
	case "openai":
		return NewOpenAIProvider(config.APIKey, config.Model), nil
	case "anthropic":
		return NewAnthropicProvider(config.APIKey, config.Model), nil
	case "ollama":
		return NewOllamaProvider(config.Model), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
