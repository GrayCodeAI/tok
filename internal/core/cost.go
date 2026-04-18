package core

import (
	"strings"
	"sync"
)

// ModelPricing holds per-token pricing for a model.
type ModelPricing struct {
	Model            string
	InputPerMillion  float64 // Cost per 1M input tokens
	OutputPerMillion float64 // Cost per 1M output tokens
}

// CommonModelPricing provides pricing for popular models (updated April 2026).
// Use RegisterModelPricing to add or override entries at runtime.
var CommonModelPricing = map[string]ModelPricing{
	// OpenAI
	"gpt-4o": {
		Model:            "gpt-4o",
		InputPerMillion:  2.50,
		OutputPerMillion: 10.00,
	},
	"gpt-4o-mini": {
		Model:            "gpt-4o-mini",
		InputPerMillion:  0.15,
		OutputPerMillion: 0.60,
	},
	"gpt-4.1": {
		Model:            "gpt-4.1",
		InputPerMillion:  2.00,
		OutputPerMillion: 8.00,
	},
	"gpt-4.1-mini": {
		Model:            "gpt-4.1-mini",
		InputPerMillion:  0.40,
		OutputPerMillion: 1.60,
	},
	"gpt-4.1-nano": {
		Model:            "gpt-4.1-nano",
		InputPerMillion:  0.10,
		OutputPerMillion: 0.40,
	},
	// Anthropic
	"claude-3.5-sonnet": {
		Model:            "claude-3.5-sonnet",
		InputPerMillion:  3.00,
		OutputPerMillion: 15.00,
	},
	"claude-3-haiku": {
		Model:            "claude-3-haiku",
		InputPerMillion:  0.25,
		OutputPerMillion: 1.25,
	},
	"claude-4-sonnet": {
		Model:            "claude-4-sonnet",
		InputPerMillion:  3.00,
		OutputPerMillion: 15.00,
	},
	"claude-4-opus": {
		Model:            "claude-4-opus",
		InputPerMillion:  15.00,
		OutputPerMillion: 75.00,
	},
}

var pricingMu sync.RWMutex

func normalizeModelPricingKey(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

// RegisterModelPricing adds or overwrites pricing for a model at runtime.
// This allows users and plugins to keep pricing current without code changes.
func RegisterModelPricing(model string, input, output float64) {
	model = normalizeModelPricingKey(model)
	pricingMu.Lock()
	CommonModelPricing[model] = ModelPricing{
		Model:            model,
		InputPerMillion:  input,
		OutputPerMillion: output,
	}
	pricingMu.Unlock()
}

// GetModelPricing returns the pricing for a model, or the default if unknown.
func GetModelPricing(model string) ModelPricing {
	model = normalizeModelPricingKey(model)
	pricingMu.RLock()
	defer pricingMu.RUnlock()
	if p, ok := CommonModelPricing[model]; ok {
		return p
	}
	return CommonModelPricing["gpt-4o-mini"]
}

// HasModelPricing reports whether a model has an explicit pricing entry.
func HasModelPricing(model string) bool {
	model = normalizeModelPricingKey(model)
	pricingMu.RLock()
	defer pricingMu.RUnlock()
	_, ok := CommonModelPricing[model]
	return ok
}

// CalculateSavings computes dollar savings from token reduction.
func CalculateSavings(tokensSaved int, model string) float64 {
	pricing := GetModelPricing(model)
	// Assume all saved tokens would have been input tokens
	return float64(tokensSaved) / 1_000_000 * pricing.InputPerMillion
}
