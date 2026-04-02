package social

import "time"

type UserProfile struct {
	ID           string    `json:"id"`
	GitHubLogin  string    `json:"github_login"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	TotalSavings int       `json:"total_savings"`
	Tier         string    `json:"tier"`
	Badges       []string  `json:"badges"`
	CreatedAt    time.Time `json:"created_at"`
}

type LeaderboardEntry struct {
	Rank         int    `json:"rank"`
	UserID       string `json:"user_id"`
	DisplayName  string `json:"display_name"`
	TotalSavings int    `json:"total_savings"`
	Tier         string `json:"tier"`
}

type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Rarity      string `json:"rarity"`
}

type SocialPlatform struct {
	users       map[string]*UserProfile
	leaderboard []LeaderboardEntry
	badges      map[string]*Badge
}

func NewSocialPlatform() *SocialPlatform {
	return &SocialPlatform{
		users:  make(map[string]*UserProfile),
		badges: make(map[string]*Badge),
	}
}

func (sp *SocialPlatform) RegisterUser(profile *UserProfile) {
	sp.users[profile.ID] = profile
}

func (sp *SocialPlatform) UpdateSavings(userID string, savings int) {
	if user, ok := sp.users[userID]; ok {
		user.TotalSavings += savings
		sp.recalculateTier(user)
	}
}

func (sp *SocialPlatform) recalculateTier(user *UserProfile) {
	switch {
	case user.TotalSavings > 1000000:
		user.Tier = "diamond"
	case user.TotalSavings > 500000:
		user.Tier = "platinum"
	case user.TotalSavings > 100000:
		user.Tier = "gold"
	case user.TotalSavings > 50000:
		user.Tier = "silver"
	case user.TotalSavings > 10000:
		user.Tier = "bronze"
	default:
		user.Tier = "starter"
	}
}

func (sp *SocialPlatform) GetLeaderboard(limit int) []LeaderboardEntry {
	var entries []LeaderboardEntry
	for _, user := range sp.users {
		entries = append(entries, LeaderboardEntry{
			UserID:       user.ID,
			DisplayName:  user.DisplayName,
			TotalSavings: user.TotalSavings,
			Tier:         user.Tier,
		})
	}

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].TotalSavings > entries[i].TotalSavings {
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

func (sp *SocialPlatform) AddBadge(badge *Badge) {
	sp.badges[badge.ID] = badge
}

func (sp *SocialPlatform) AwardBadge(userID, badgeID string) {
	if user, ok := sp.users[userID]; ok {
		if _, exists := sp.badges[badgeID]; exists {
			user.Badges = append(user.Badges, badgeID)
		}
	}
}

func (sp *SocialPlatform) GetUserProfile(userID string) *UserProfile {
	return sp.users[userID]
}
