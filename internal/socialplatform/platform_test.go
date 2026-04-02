package socialplatform

import "testing"

func TestSocialPlatformManager(t *testing.T) {
	m := NewSocialPlatformManager()

	m.RegisterUser(&UserProfile{ID: "u1", Username: "alice", DisplayName: "Alice"})
	m.RegisterUser(&UserProfile{ID: "u2", Username: "bob", DisplayName: "Bob"})

	m.UpdateSavings("u1", 50000)
	m.UpdateSavings("u2", 150000)

	leaderboard := m.GetLeaderboard(10)
	if len(leaderboard) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(leaderboard))
	}
	if leaderboard[0].Username != "bob" {
		t.Errorf("Expected bob at rank 1, got %s", leaderboard[0].Username)
	}
	if leaderboard[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", leaderboard[0].Rank)
	}
}

func TestSocialPlatformBadges(t *testing.T) {
	m := NewSocialPlatformManager()
	m.RegisterUser(&UserProfile{ID: "u1", Username: "alice"})
	m.AwardBadge("u1", "first_compression")

	stats := m.GetStats()
	if stats.TotalUsers != 1 {
		t.Errorf("Expected 1 user, got %d", stats.TotalUsers)
	}
}

func TestSocialPlatformExport(t *testing.T) {
	m := NewSocialPlatformManager()
	m.RegisterUser(&UserProfile{ID: "u1", Username: "test"})

	export, err := m.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}
	if len(export) == 0 {
		t.Error("Expected non-empty export")
	}
}
