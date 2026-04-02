package tieredpricing

type Tier struct {
	Name        string  `json:"name"`
	MinTokens   int64   `json:"min_tokens"`
	MaxTokens   int64   `json:"max_tokens"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
}

type TieredPricing struct {
	model string
	tiers []Tier
}

func NewTieredPricing(model string) *TieredPricing {
	return &TieredPricing{model: model}
}

func (p *TieredPricing) AddTier(tier Tier) {
	p.tiers = append(p.tiers, tier)
}

func (p *TieredPricing) Calculate(inputTokens, outputTokens int64) float64 {
	var cost float64
	for _, tier := range p.tiers {
		if inputTokens >= tier.MinTokens && (tier.MaxTokens == 0 || inputTokens <= tier.MaxTokens) {
			cost += float64(inputTokens) / 1000000 * tier.InputPrice
		}
		if outputTokens >= tier.MinTokens && (tier.MaxTokens == 0 || outputTokens <= tier.MaxTokens) {
			cost += float64(outputTokens) / 1000000 * tier.OutputPrice
		}
	}
	return cost
}

func (p *TieredPricing) GetTiers() []Tier {
	return p.tiers
}
