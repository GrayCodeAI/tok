package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ModelPricing struct {
	Model           string  `json:"model"`
	InputCostPer1M  float64 `json:"input_cost_per_1m"`
	OutputCostPer1M float64 `json:"output_cost_per_1m"`
	Provider        string  `json:"provider"`
	Available       bool    `json:"available"`
}

type PricingSource string

const (
	SourceLiteLLM    PricingSource = "litellm"
	SourceOpenRouter PricingSource = "openrouter"
	SourceManual     PricingSource = "manual"
)

type PricingCache struct {
	prices  map[string]ModelPricing
	updated time.Time
	ttl     time.Duration
}

func NewPricingCache(ttl time.Duration) *PricingCache {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &PricingCache{
		prices: make(map[string]ModelPricing),
		ttl:    ttl,
	}
}

func (c *PricingCache) Get(model string) (ModelPricing, bool) {
	p, ok := c.prices[normalizeModelName(model)]
	return p, ok
}

func (c *PricingCache) Set(model string, pricing ModelPricing) {
	c.prices[normalizeModelName(model)] = pricing
}

func (c *PricingCache) IsExpired() bool {
	return time.Since(c.updated) > c.ttl
}

func (c *PricingCache) LoadDefaults() {
	defaults := []ModelPricing{
		{Model: "gpt-4o", InputCostPer1M: 2.5, OutputCostPer1M: 10, Provider: "openai", Available: true},
		{Model: "gpt-4o-mini", InputCostPer1M: 0.15, OutputCostPer1M: 0.6, Provider: "openai", Available: true},
		{Model: "gpt-4-turbo", InputCostPer1M: 10, OutputCostPer1M: 30, Provider: "openai", Available: true},
		{Model: "gpt-3.5-turbo", InputCostPer1M: 0.5, OutputCostPer1M: 1.5, Provider: "openai", Available: true},
		{Model: "claude-3-5-sonnet", InputCostPer1M: 3, OutputCostPer1M: 15, Provider: "anthropic", Available: true},
		{Model: "claude-3-sonnet", InputCostPer1M: 3, OutputCostPer1M: 15, Provider: "anthropic", Available: true},
		{Model: "claude-3-haiku", InputCostPer1M: 0.25, OutputCostPer1M: 1.25, Provider: "anthropic", Available: true},
		{Model: "claude-3-opus", InputCostPer1M: 15, OutputCostPer1M: 75, Provider: "anthropic", Available: true},
		{Model: "gemini-1.5-pro", InputCostPer1M: 1.25, OutputCostPer1M: 5, Provider: "google", Available: true},
		{Model: "gemini-1.5-flash", InputCostPer1M: 0.075, OutputCostPer1M: 0.3, Provider: "google", Available: true},
		{Model: "gemini-2.0-flash", InputCostPer1M: 0.1, OutputCostPer1M: 0.4, Provider: "google", Available: true},
		{Model: "groq/mixtral-8x7b", InputCostPer1M: 0.27, OutputCostPer1M: 0.27, Provider: "groq", Available: true},
		{Model: "groq/llama3-70b", InputCostPer1M: 0.59, OutputCostPer1M: 0.79, Provider: "groq", Available: true},
		{Model: "deepseek/deepseek-chat", InputCostPer1M: 0.14, OutputCostPer1M: 0.28, Provider: "deepseek", Available: true},
		{Model: "deepseek/deepseek-coder", InputCostPer1M: 0.14, OutputCostPer1M: 0.28, Provider: "deepseek", Available: true},
	}
	for _, p := range defaults {
		c.Set(p.Model, p)
	}
	c.updated = time.Now()
}

func (c *PricingCache) FetchFromLiteLLM() error {
	resp, err := http.Get("https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var models map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&models); err != nil {
		return err
	}

	for model, data := range models {
		if m, ok := data.(map[string]interface{}); ok {
			pricing := ModelPricing{
				Model:     model,
				Available: true,
			}
			if cost, ok := m["input_cost_per_token"].(float64); ok {
				pricing.InputCostPer1M = cost * 1000000
			}
			if cost, ok := m["output_cost_per_token"].(float64); ok {
				pricing.OutputCostPer1M = cost * 1000000
			}
			c.Set(model, pricing)
		}
	}
	c.updated = time.Now()
	return nil
}

func (c *PricingCache) List() []ModelPricing {
	var prices []ModelPricing
	for _, p := range c.prices {
		prices = append(prices, p)
	}
	return prices
}

func (c *PricingCache) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, ok := c.Get(model)
	if !ok {
		return 0
	}
	cost := (float64(inputTokens)/1000000)*pricing.InputCostPer1M +
		(float64(outputTokens)/1000000)*pricing.OutputCostPer1M
	return cost
}

func (c *PricingCache) CheapestForCapability(models []string, inputTokens, outputTokens int) (string, float64) {
	cheapest := ""
	cheapestCost := float64(0)

	for _, model := range models {
		if pricing, ok := c.Get(model); ok && pricing.Available {
			cost := c.CalculateCost(model, inputTokens, outputTokens)
			if cheapest == "" || cost < cheapestCost {
				cheapest = model
				cheapestCost = cost
			}
		}
	}
	return cheapest, cheapestCost
}

func (c *PricingCache) PriceComparison(models []string, inputTokens, outputTokens int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Cost for %d input + %d output tokens:\n", inputTokens, outputTokens))

	var entries []struct {
		model string
		cost  float64
	}
	for _, m := range models {
		cost := c.CalculateCost(m, inputTokens, outputTokens)
		entries = append(entries, struct {
			model string
			cost  float64
		}{m, cost})
	}

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].cost < entries[i].cost {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  %-30s $%.4f\n", e.model, e.cost))
	}
	return sb.String()
}

func normalizeModelName(model string) string {
	model = strings.ToLower(model)
	model = strings.TrimPrefix(model, "openai/")
	model = strings.TrimPrefix(model, "anthropic/")
	model = strings.TrimPrefix(model, "google/")
	return model
}
