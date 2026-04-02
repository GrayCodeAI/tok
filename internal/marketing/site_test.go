package marketing

import "testing"

func TestMarketingSite(t *testing.T) {
	site := NewMarketingSite()

	pricing := site.GetPricing()
	if len(pricing) != 3 {
		t.Errorf("Expected 3 pricing tiers, got %d", len(pricing))
	}

	comparisons := site.GetComparisons()
	if len(comparisons) < 3 {
		t.Errorf("Expected at least 3 comparisons, got %d", len(comparisons))
	}
}

func TestNewsletter(t *testing.T) {
	n := NewNewsletter()

	n.Subscribe("test@example.com")
	if n.Count() != 1 {
		t.Errorf("Expected 1 subscriber, got %d", n.Count())
	}

	n.Unsubscribe("test@example.com")
	if n.Count() != 0 {
		t.Errorf("Expected 0 subscribers, got %d", n.Count())
	}
}

func TestReferralProgram(t *testing.T) {
	rp := NewReferralProgram()

	rp.Refer("user1", "user2")
	if rp.GetRewards("user1") != 10 {
		t.Errorf("Expected 10 rewards, got %d", rp.GetRewards("user1"))
	}
}
