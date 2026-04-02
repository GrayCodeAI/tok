package contributiongraph

import "testing"

func TestKardashevRank(t *testing.T) {
	tests := []struct {
		savings int64
		tier    Tier
	}{
		{0, TierStarter},
		{5000, TierStarter},
		{10000, TierBronze},
		{100000, TierSilver},
		{1000000, TierGold},
		{10000000, TierPlatinum},
		{100000000, TierDiamond},
		{1000000000, TierLegend},
	}

	for _, tt := range tests {
		rank := NewKardashevRank(tt.savings)
		if rank.Tier != tt.tier {
			t.Errorf("KardashevRank(%d) = %s, want %s", tt.savings, rank.Tier, tt.tier)
		}
	}
}

func TestContributionGraph(t *testing.T) {
	cg := NewContributionGraph()
	cg.AddDay("2026-01-01", 100)
	cg.AddDay("2026-01-02", 500)
	cg.AddDay("2026-01-03", 1000)

	rendered := cg.Render(20, 7)
	if rendered == "" {
		t.Error("Expected non-empty render")
	}

	svg := cg.SVGBadge()
	if svg == "" {
		t.Error("Expected non-empty SVG")
	}

	stats := cg.Stats()
	if stats["total_days"].(int) != 3 {
		t.Errorf("Expected 3 days, got %v", stats["total_days"])
	}
}

func TestBadgeEngine(t *testing.T) {
	e := NewBadgeEngine()

	badges := e.List()
	if len(badges) < 5 {
		t.Errorf("Expected at least 5 badges, got %d", len(badges))
	}

	earned := e.CheckEarned(1000)
	if len(earned) == 0 {
		t.Error("Expected at least 1 badge earned with 1000 savings")
	}

	badge := e.Get("first_compression")
	if badge == nil {
		t.Error("Expected first_compression badge")
	}
}
