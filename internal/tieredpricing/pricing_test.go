package tieredpricing_test

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/tieredpricing"
)

func TestNewTieredPricing(t *testing.T) {
	p := tieredpricing.NewTieredPricing("gpt-4")
	if p == nil {
		t.Fatal("NewTieredPricing returned nil")
	}
}

func TestAddTier(t *testing.T) {
	p := tieredpricing.NewTieredPricing("gpt-4")
	p.AddTier(tieredpricing.Tier{
		Name:        "standard",
		MinTokens:   0,
		MaxTokens:   4096,
		InputPrice:  10.0,
		OutputPrice: 30.0,
	})
	tiers := p.GetTiers()
	if len(tiers) != 1 {
		t.Errorf("expected 1 tier, got %d", len(tiers))
	}
}

func TestCalculate(t *testing.T) {
	p := tieredpricing.NewTieredPricing("gpt-4")
	p.AddTier(tieredpricing.Tier{
		Name:        "base",
		MinTokens:   0,
		MaxTokens:   4096,
		InputPrice:  10.0,
		OutputPrice: 30.0,
	})
	cost := p.Calculate(500000, 500000)
	if cost < 0 {
		t.Errorf("Calculate returned negative cost: %f", cost)
	}
}

func TestCalculate_Zero(t *testing.T) {
	p := tieredpricing.NewTieredPricing("gpt-4")
	cost := p.Calculate(0, 0)
	if cost != 0 {
		t.Errorf("Calculate(0, 0) = %f, want 0", cost)
	}
}

func TestCalculate_MultipleTiers(t *testing.T) {
	p := tieredpricing.NewTieredPricing("gpt-4")
	p.AddTier(tieredpricing.Tier{
		Name:        "small",
		MinTokens:   0,
		MaxTokens:   4096,
		InputPrice:  5.0,
		OutputPrice: 15.0,
	})
	p.AddTier(tieredpricing.Tier{
		Name:        "large",
		MinTokens:   4097,
		MaxTokens:   8192,
		InputPrice:  10.0,
		OutputPrice: 30.0,
	})

	cost1 := p.Calculate(1000, 1000)
	cost2 := p.Calculate(5000, 5000)

	if cost1 < 0 || cost2 < 0 {
		t.Error("cost should not be negative")
	}
}

func TestTiersAreStored(t *testing.T) {
	p := tieredpricing.NewTieredPricing("test")
	if len(p.GetTiers()) != 0 {
		t.Error("new pricing should have 0 tiers")
	}
	p.AddTier(tieredpricing.Tier{Name: "a", MinTokens: 0, MaxTokens: 1000, InputPrice: 1})
	p.AddTier(tieredpricing.Tier{Name: "b", MinTokens: 1001, MaxTokens: 2000, InputPrice: 2})
	if len(p.GetTiers()) != 2 {
		t.Errorf("expected 2 tiers, got %d", len(p.GetTiers()))
	}
}
