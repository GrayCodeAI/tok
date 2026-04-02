package agent

import (
	"context"
	"fmt"
)

type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []Message, opts ChatOptions) (*Response, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ChatOptions struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Tools       []Tool
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"`
}

type AgentManager struct {
	providers map[string]Provider
	active    string
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		providers: make(map[string]Provider),
	}
}

func (am *AgentManager) RegisterProvider(name string, provider Provider) {
	am.providers[name] = provider
}

func (am *AgentManager) SetActiveProvider(name string) error {
	if _, ok := am.providers[name]; !ok {
		return fmt.Errorf("provider not found: %s", name)
	}
	am.active = name
	return nil
}

func (am *AgentManager) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*Response, error) {
	provider, ok := am.providers[am.active]
	if !ok {
		return nil, fmt.Errorf("no active provider")
	}
	return provider.Chat(ctx, messages, opts)
}

func (am *AgentManager) ListProviders() []string {
	var names []string
	for name := range am.providers {
		names = append(names, name)
	}
	return names
}

type RetryConfig struct {
	MaxRetries  int
	BackoffBase float64
	Retryable   []int
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		BackoffBase: 2.0,
		Retryable:   []int{429, 500, 502, 503, 504},
	}
}

type StuckLoopDetector struct {
	calls     map[string]int
	threshold int
}

func NewStuckLoopDetector(threshold int) *StuckLoopDetector {
	return &StuckLoopDetector{
		calls:     make(map[string]int),
		threshold: threshold,
	}
}

func (d *StuckLoopDetector) Record(toolCall string) bool {
	d.calls[toolCall]++
	return d.calls[toolCall] >= d.threshold
}

func (d *StuckLoopDetector) Reset() {
	d.calls = make(map[string]int)
}

type ContextTracker struct {
	MaxTokens      int
	CurrentTokens  int
	CompactionFunc func() string
}

func (ct *ContextTracker) AddTokens(n int) bool {
	ct.CurrentTokens += n
	return ct.CurrentTokens <= ct.MaxTokens
}

func (ct *ContextTracker) NeedsCompaction() bool {
	return ct.CurrentTokens > ct.MaxTokens
}
