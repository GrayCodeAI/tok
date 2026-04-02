// Package gateway provides multi-provider LLM gateway.
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider interface for LLM providers.
type Provider interface {
	Name() string
	Compress(ctx context.Context, req *CompressionRequest) (*CompressionResponse, error)
	HealthCheck() error
}

// CompressionRequest represents a compression request.
type CompressionRequest struct {
	Content     string
	Mode        string
	MaxTokens   int
	ContextType string
}

// CompressionResponse represents a compression response.
type CompressionResponse struct {
	Compressed   string
	TokensSaved  int
	ProviderUsed string
	LatencyMs    int64
}

// AnthropicProvider implements Anthropic API.
type AnthropicProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// NewAnthropicProvider creates Anthropic provider.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "claude-3-5-sonnet-20241022",
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Compress(ctx context.Context, req *CompressionRequest) (*CompressionResponse, error) {
	// Anthropic-specific compression using their API
	payload := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("Summarize this content concisely: %s", req.Content),
			},
		},
	}

	data, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error: %s", body)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	compressed := req.Content
	if len(result.Content) > 0 {
		compressed = result.Content[0].Text
	}

	return &CompressionResponse{
		Compressed:   compressed,
		TokensSaved:  len(req.Content) - len(compressed),
		ProviderUsed: p.Name(),
		LatencyMs:    latency,
	}, nil
}

func (p *AnthropicProvider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

// OpenAIProvider implements OpenAI API.
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// NewOpenAIProvider creates OpenAI provider.
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gpt-4-turbo",
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Compress(ctx context.Context, req *CompressionRequest) (*CompressionResponse, error) {
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a text compression assistant. Summarize the input concisely.",
			},
			{
				"role":    "user",
				"content": req.Content,
			},
		},
		"max_tokens": 1024,
	}

	data, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai API error: %s", body)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	compressed := req.Content
	if len(result.Choices) > 0 {
		compressed = result.Choices[0].Message.Content
	}

	return &CompressionResponse{
		Compressed:   compressed,
		TokensSaved:  len(req.Content) - len(compressed),
		ProviderUsed: p.Name(),
		LatencyMs:    latency,
	}, nil
}

func (p *OpenAIProvider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

// Gateway manages multiple providers with fallback.
type Gateway struct {
	providers []Provider
	quota     *QuotaManager
}

// QuotaManager manages usage quotas.
type QuotaManager struct {
	limits  map[string]int64
	used    map[string]int64
	resetAt map[string]time.Time
}

// NewGateway creates a new gateway.
func NewGateway() *Gateway {
	return &Gateway{
		providers: make([]Provider, 0),
		quota: &QuotaManager{
			limits:  make(map[string]int64),
			used:    make(map[string]int64),
			resetAt: make(map[string]time.Time),
		},
	}
}

// AddProvider adds a provider to the gateway.
func (g *Gateway) AddProvider(p Provider) {
	g.providers = append(g.providers, p)
}

// SetQuota sets the quota for a provider.
func (g *Gateway) SetQuota(provider string, limit int64) {
	g.quota.limits[provider] = limit
	g.quota.used[provider] = 0
	g.quota.resetAt[provider] = time.Now().Add(24 * time.Hour)
}

// Compress compresses content with fallback.
func (g *Gateway) Compress(ctx context.Context, req *CompressionRequest) (*CompressionResponse, error) {
	var lastErr error

	for _, provider := range g.providers {
		// Check quota
		if g.quota.limits[provider.Name()] > 0 {
			if g.quota.used[provider.Name()] >= g.quota.limits[provider.Name()] {
				continue // Skip if quota exceeded
			}
		}

		resp, err := provider.Compress(ctx, req)
		if err != nil {
			lastErr = err
			continue
		}

		// Record usage
		g.quota.used[provider.Name()] += int64(len(req.Content))

		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed: %w", lastErr)
	}

	return nil, fmt.Errorf("no providers available")
}

// HealthChecks runs health checks on all providers.
func (g *Gateway) HealthChecks() map[string]error {
	results := make(map[string]error)
	for _, p := range g.providers {
		results[p.Name()] = p.HealthCheck()
	}
	return results
}
