// Package llmproviders provides LLM provider integration for TokMan agents
package llmproviders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/GrayCodeAI/tokman/internal/circuitbreaker"
)

// MaxRetries is the default maximum number of retries for LLM requests.
const MaxRetries = 3

// DefaultRetryBackoff is the base delay between retries.
const DefaultRetryBackoff = 1 * time.Second

func newJSONRequest(ctx context.Context, method, url string, payload interface{}) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// Provider defines an LLM provider interface
type Provider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	Embed(ctx context.Context, text string) ([]float64, error)
}

// ProviderWithBreaker wraps a Provider with circuit breaker protection and retry logic.
type ProviderWithBreaker struct {
	provider Provider
	breaker  *circuitbreaker.Breaker
	maxRetry int
	backoff  time.Duration
}

// WrapProvider creates a ProviderWithBreaker that adds circuit breaker
// protection and exponential backoff retry to any Provider.
func WrapProvider(p Provider, breaker *circuitbreaker.Breaker) *ProviderWithBreaker {
	return &ProviderWithBreaker{
		provider: p,
		breaker:  breaker,
		maxRetry: MaxRetries,
		backoff:  DefaultRetryBackoff,
	}
}

// Name returns the wrapped provider name.
func (pw *ProviderWithBreaker) Name() string {
	return pw.provider.Name()
}

// Complete executes a completion request with circuit breaker protection and retry.
func (pw *ProviderWithBreaker) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= pw.maxRetry; attempt++ {
		if err := pw.breaker.Allow(); err != nil {
			return nil, fmt.Errorf("circuit breaker open for %s: %w", pw.provider.Name(), err)
		}

		resp, err := pw.provider.Complete(ctx, req)
		if err == nil {
			pw.breaker.RecordSuccess()
			return resp, nil
		}

		// Don't retry on client errors (4xx) - these are not transient
		if isClientError(err) {
			pw.breaker.RecordFailure()
			return nil, err
		}

		lastErr = err
		pw.breaker.RecordFailure()

		if attempt < pw.maxRetry {
			delay := pw.backoff * time.Duration(math.Pow(2, float64(attempt)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Retry after backoff
			}
		}
	}

	return nil, fmt.Errorf("all %d retries exhausted for %s: %w", pw.maxRetry, pw.provider.Name(), lastErr)
}

// StreamComplete proxies to the wrapped provider (no retry for streaming).
func (pw *ProviderWithBreaker) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	if err := pw.breaker.Allow(); err != nil {
		return nil, fmt.Errorf("circuit breaker open for %s: %w", pw.provider.Name(), err)
	}
	return pw.provider.StreamComplete(ctx, req)
}

// Embed proxies to the wrapped provider with circuit breaker protection.
func (pw *ProviderWithBreaker) Embed(ctx context.Context, text string) ([]float64, error) {
	if err := pw.breaker.Allow(); err != nil {
		return nil, fmt.Errorf("circuit breaker open for %s: %w", pw.provider.Name(), err)
	}

	result, err := pw.provider.Embed(ctx, text)
	if err != nil {
		pw.breaker.RecordFailure()
		return nil, err
	}

	pw.breaker.RecordSuccess()
	return result, nil
}

// isClientError checks if an error is a 4xx HTTP error (not retryable).
func isClientError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Check for common 4xx status codes in error messages
	for _, code := range []string{"400", "401", "403", "404", "422"} {
		if containsStr(msg, fmt.Sprintf("error %s", code)) {
			return true
		}
	}
	return false
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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

	httpReq, err := newJSONRequest(ctx, "POST", p.baseURL+"/chat/completions", payload)
	if err != nil {
		return nil, err
	}
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
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai response missing choices")
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

	httpReq, err := newJSONRequest(ctx, "POST", p.baseURL+"/embeddings", payload)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai embed failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embed error %d: %s", resp.StatusCode, string(body))
	}

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

	httpReq, err := newJSONRequest(ctx, "POST", "https://api.anthropic.com/v1/messages", payload)
	if err != nil {
		return nil, err
	}
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

	httpReq, err := newJSONRequest(ctx, "POST", p.baseURL+"/api/generate", payload)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

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

	httpReq, err := newJSONRequest(ctx, "POST", p.baseURL+"/api/embeddings", payload)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama embed failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed error %d: %s", resp.StatusCode, string(body))
	}

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

// Create creates a provider by name wrapped with circuit breaker protection.
func (pf *ProviderFactory) Create(name string) (Provider, error) {
	cfg, ok := pf.config[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	var base Provider
	switch name {
	case "openai":
		base = NewOpenAIProvider(cfg.APIKey, cfg.Model)
	case "anthropic":
		base = NewAnthropicProvider(cfg.APIKey, cfg.Model)
	case "ollama":
		base = NewOllamaProvider(cfg.Model)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}

	// Wrap with circuit breaker for production resilience
	breaker := circuitbreaker.New(circuitbreaker.DefaultConfig())
	return WrapProvider(base, breaker), nil
}
