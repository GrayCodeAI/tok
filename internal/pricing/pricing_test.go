package pricing

import "testing"

func TestPricingCache(t *testing.T) {
	cache := NewPricingCache(0)
	cache.LoadDefaults()

	p, ok := cache.Get("gpt-4o")
	if !ok {
		t.Error("Expected gpt-4o pricing")
	}
	if p.InputCostPer1M != 2.5 {
		t.Errorf("Expected 2.5 input cost, got %f", p.InputCostPer1M)
	}

	p, ok = cache.Get("claude-3-5-sonnet")
	if !ok {
		t.Error("Expected claude-3-5-sonnet pricing")
	}
	if p.InputCostPer1M != 3 {
		t.Errorf("Expected 3 input cost, got %f", p.InputCostPer1M)
	}
}

func TestCalculateCost(t *testing.T) {
	cache := NewPricingCache(0)
	cache.LoadDefaults()

	cost := cache.CalculateCost("gpt-4o", 1000000, 500000)
	if cost == 0 {
		t.Error("Expected non-zero cost")
	}
	if cost < 2.5 || cost > 7.5 {
		t.Errorf("Expected cost between 2.5 and 7.5, got %f", cost)
	}
}

func TestCheapestForCapability(t *testing.T) {
	cache := NewPricingCache(0)
	cache.LoadDefaults()

	models := []string{"gpt-4o", "gpt-4o-mini", "claude-3-haiku"}
	cheapest, cost := cache.CheapestForCapability(models, 100000, 50000)
	if cheapest == "" {
		t.Error("Expected cheapest model")
	}
	if cost == 0 {
		t.Error("Expected non-zero cost")
	}
}

func TestPriceComparison(t *testing.T) {
	cache := NewPricingCache(0)
	cache.LoadDefaults()

	output := cache.PriceComparison([]string{"gpt-4o", "gpt-4o-mini"}, 100000, 50000)
	if output == "" {
		t.Error("Expected non-empty comparison")
	}
}

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"openai/gpt-4o", "gpt-4o"},
		{"anthropic/claude-3-haiku", "claude-3-haiku"},
		{"GPT-4O", "gpt-4o"},
		{"gpt-4o", "gpt-4o"},
	}

	for _, tt := range tests {
		got := normalizeModelName(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeModelName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
