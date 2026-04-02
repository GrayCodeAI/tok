package social

import "testing"

func TestSocialPlatform(t *testing.T) {
	sp := NewSocialPlatform()

	sp.RegisterUser(&UserProfile{
		ID:          "user1",
		DisplayName: "Alice",
	})

	sp.UpdateSavings("user1", 50000)
	profile := sp.GetUserProfile("user1")
	if profile == nil {
		t.Fatal("Expected profile")
	}
	if profile.TotalSavings != 50000 {
		t.Errorf("Expected 50000 savings, got %d", profile.TotalSavings)
	}

	leaderboard := sp.GetLeaderboard(10)
	if len(leaderboard) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(leaderboard))
	}
	if leaderboard[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", leaderboard[0].Rank)
	}
}

func TestBadgeAwarding(t *testing.T) {
	sp := NewSocialPlatform()
	sp.RegisterUser(&UserProfile{ID: "u1", DisplayName: "Test"})
	sp.AddBadge(&Badge{ID: "first_compression", Name: "First"})
	sp.AwardBadge("u1", "first_compression")

	profile := sp.GetUserProfile("u1")
	if len(profile.Badges) != 1 {
		t.Errorf("Expected 1 badge, got %d", len(profile.Badges))
	}
}
