package socialplatform

import (
	"encoding/json"
	"time"
)

type UserProfile struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	TotalSaved  int64     `json:"total_saved"`
	Tier        string    `json:"tier"`
	Badges      []string  `json:"badges"`
	JoinedAt    time.Time `json:"joined_at"`
}

type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	TotalSaved  int64  `json:"total_saved"`
	Tier        string `json:"tier"`
}

type SocialStats struct {
	TotalUsers int     `json:"total_users"`
	TotalSaved int64   `json:"total_saved"`
	AvgSavings float64 `json:"avg_savings"`
	TopUser    string  `json:"top_user"`
	TopSavings int64   `json:"top_savings"`
}

type SocialPlatformManager struct {
	users       map[string]*UserProfile
	leaderboard []LeaderboardEntry
}

func NewSocialPlatformManager() *SocialPlatformManager {
	return &SocialPlatformManager{
		users: make(map[string]*UserProfile),
	}
}

func (m *SocialPlatformManager) RegisterUser(profile *UserProfile) {
	m.users[profile.ID] = profile
	m.recalculateTiers()
}

func (m *SocialPlatformManager) UpdateSavings(userID string, saved int64) {
	if user, ok := m.users[userID]; ok {
		user.TotalSaved += saved
		m.recalculateTier(user)
	}
}

func (m *SocialPlatformManager) recalculateTiers() {
	for _, user := range m.users {
		m.recalculateTier(user)
	}
}

func (m *SocialPlatformManager) recalculateTier(user *UserProfile) {
	switch {
	case user.TotalSaved >= 1000000000:
		user.Tier = "legend"
	case user.TotalSaved >= 100000000:
		user.Tier = "diamond"
	case user.TotalSaved >= 10000000:
		user.Tier = "platinum"
	case user.TotalSaved >= 1000000:
		user.Tier = "gold"
	case user.TotalSaved >= 100000:
		user.Tier = "silver"
	case user.TotalSaved >= 10000:
		user.Tier = "bronze"
	default:
		user.Tier = "starter"
	}
}

func (m *SocialPlatformManager) GetLeaderboard(limit int) []LeaderboardEntry {
	var entries []LeaderboardEntry
	for _, user := range m.users {
		entries = append(entries, LeaderboardEntry{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			TotalSaved:  user.TotalSaved,
			Tier:        user.Tier,
		})
	}

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].TotalSaved > entries[i].TotalSaved {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	if limit > len(entries) {
		limit = len(entries)
	}
	for i := 0; i < limit; i++ {
		entries[i].Rank = i + 1
	}
	return entries[:limit]
}

func (m *SocialPlatformManager) AwardBadge(userID, badge string) {
	if user, ok := m.users[userID]; ok {
		user.Badges = append(user.Badges, badge)
	}
}

func (m *SocialPlatformManager) GetStats() *SocialStats {
	var totalSaved int64
	var topUser string
	var topSavings int64
	for _, user := range m.users {
		totalSaved += user.TotalSaved
		if user.TotalSaved > topSavings {
			topSavings = user.TotalSaved
			topUser = user.Username
		}
	}
	avg := 0.0
	if len(m.users) > 0 {
		avg = float64(totalSaved) / float64(len(m.users))
	}
	return &SocialStats{
		TotalUsers: len(m.users),
		TotalSaved: totalSaved,
		AvgSavings: avg,
		TopUser:    topUser,
		TopSavings: topSavings,
	}
}

func (m *SocialPlatformManager) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(m.GetStats(), "", "  ")
}
