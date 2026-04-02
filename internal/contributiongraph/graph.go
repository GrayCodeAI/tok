package contributiongraph

import (
	"fmt"
	"strings"
)

type Tier string

const (
	TierStarter  Tier = "starter"
	TierBronze   Tier = "bronze"
	TierSilver   Tier = "silver"
	TierGold     Tier = "gold"
	TierPlatinum Tier = "platinum"
	TierDiamond  Tier = "diamond"
	TierLegend   Tier = "legend"
)

type KardashevRank struct {
	Tier      Tier    `json:"tier"`
	Level     int     `json:"level"`
	Savings   int64   `json:"savings"`
	NextTier  Tier    `json:"next_tier"`
	NextLevel int64   `json:"next_level"`
	Progress  float64 `json:"progress"`
}

func NewKardashevRank(savings int64) *KardashevRank {
	switch {
	case savings >= 1000000000:
		return &KardashevRank{Tier: TierLegend, Level: 7, Savings: savings, Progress: 100}
	case savings >= 100000000:
		return &KardashevRank{Tier: TierDiamond, Level: 6, Savings: savings, NextTier: TierLegend, NextLevel: 1000000000, Progress: float64(savings) / 1000000000 * 100}
	case savings >= 10000000:
		return &KardashevRank{Tier: TierPlatinum, Level: 5, Savings: savings, NextTier: TierDiamond, NextLevel: 100000000, Progress: float64(savings) / 100000000 * 100}
	case savings >= 1000000:
		return &KardashevRank{Tier: TierGold, Level: 4, Savings: savings, NextTier: TierPlatinum, NextLevel: 10000000, Progress: float64(savings) / 10000000 * 100}
	case savings >= 100000:
		return &KardashevRank{Tier: TierSilver, Level: 3, Savings: savings, NextTier: TierGold, NextLevel: 1000000, Progress: float64(savings) / 1000000 * 100}
	case savings >= 10000:
		return &KardashevRank{Tier: TierBronze, Level: 2, Savings: savings, NextTier: TierSilver, NextLevel: 100000, Progress: float64(savings) / 100000 * 100}
	default:
		return &KardashevRank{Tier: TierStarter, Level: 1, Savings: savings, NextTier: TierBronze, NextLevel: 10000, Progress: float64(savings) / 10000 * 100}
	}
}

type ContributionDay struct {
	Date    string `json:"date"`
	Savings int    `json:"savings"`
	Level   int    `json:"level"`
}

type ContributionGraph struct {
	days []ContributionDay
}

func NewContributionGraph() *ContributionGraph {
	return &ContributionGraph{}
}

func (cg *ContributionGraph) AddDay(date string, savings int) {
	level := 0
	switch {
	case savings >= 1000:
		level = 4
	case savings >= 500:
		level = 3
	case savings >= 100:
		level = 2
	case savings >= 10:
		level = 1
	}
	cg.days = append(cg.days, ContributionDay{Date: date, Savings: savings, Level: level})
}

func (cg *ContributionGraph) Render(width, height int) string {
	if len(cg.days) == 0 {
		return "No contributions yet.\n"
	}

	var sb strings.Builder
	sb.WriteString("Contribution Graph (3D isometric view)\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	chars := []string{" ", "░", "▒", "▓", "█"}
	for row := 0; row < 7 && row < height; row++ {
		for col := 0; col < width/2 && col < len(cg.days)/7; col++ {
			idx := col*7 + row
			if idx < len(cg.days) {
				sb.WriteString(chars[cg.days[idx].Level])
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("─", width) + "\n")
	return sb.String()
}

func (cg *ContributionGraph) SVGBadge() string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="30">
<rect width="200" height="30" rx="5" fill="#1a1a2e"/>
<text x="100" y="20" text-anchor="middle" fill="#e94560" font-size="12" font-weight="bold">%d days active</text>
</svg>`, len(cg.days))
}

func (cg *ContributionGraph) Stats() map[string]interface{} {
	totalSavings := 0
	for _, d := range cg.days {
		totalSavings += d.Savings
	}
	return map[string]interface{}{
		"total_days":    len(cg.days),
		"total_savings": totalSavings,
	}
}

type BadgeEngine struct {
	badges map[string]*Badge
}

type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Rarity      string `json:"rarity"`
	Requirement string `json:"requirement"`
}

func NewBadgeEngine() *BadgeEngine {
	engine := &BadgeEngine{
		badges: make(map[string]*Badge),
	}
	engine.registerBuiltInBadges()
	return engine
}

func (e *BadgeEngine) registerBuiltInBadges() {
	badges := []Badge{
		{ID: "first_compression", Name: "First Compression", Description: "Run your first compression", Icon: "⚡", Rarity: "common", Requirement: "0 savings"},
		{ID: "token_saver", Name: "Token Saver", Description: "Save 1000 tokens", Icon: "💾", Rarity: "uncommon", Requirement: "1K savings"},
		{ID: "compression_master", Name: "Compression Master", Description: "Save 100K tokens", Icon: "🏆", Rarity: "rare", Requirement: "100K savings"},
		{ID: "token_titan", Name: "Token Titan", Description: "Save 1M tokens", Icon: "🌟", Rarity: "epic", Requirement: "1M savings"},
		{ID: "legend", Name: "Legend", Description: "Save 100M tokens", Icon: "👑", Rarity: "legendary", Requirement: "100M savings"},
		{ID: "early_adopter", Name: "Early Adopter", Description: "Used TokMan in first week", Icon: "🚀", Rarity: "rare", Requirement: "Early usage"},
		{ID: "multi_provider", Name: "Multi-Provider", Description: "Used 5+ providers", Icon: "🔗", Rarity: "uncommon", Requirement: "5 providers"},
		{ID: "filter_crafter", Name: "Filter Crafter", Description: "Created 10 custom filters", Icon: "🔧", Rarity: "rare", Requirement: "10 filters"},
	}
	for i := range badges {
		e.badges[badges[i].ID] = &badges[i]
	}
}

func (e *BadgeEngine) Get(id string) *Badge {
	return e.badges[id]
}

func (e *BadgeEngine) List() []*Badge {
	var result []*Badge
	for _, b := range e.badges {
		result = append(result, b)
	}
	return result
}

func (e *BadgeEngine) CheckEarned(savings int64) []*Badge {
	var earned []*Badge
	for _, b := range e.badges {
		switch b.ID {
		case "first_compression":
			if savings > 0 {
				earned = append(earned, b)
			}
		case "token_saver":
			if savings >= 1000 {
				earned = append(earned, b)
			}
		case "compression_master":
			if savings >= 100000 {
				earned = append(earned, b)
			}
		case "token_titan":
			if savings >= 1000000 {
				earned = append(earned, b)
			}
		case "legend":
			if savings >= 100000000 {
				earned = append(earned, b)
			}
		}
	}
	return earned
}
