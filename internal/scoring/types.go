package scoring

import (
	"fmt"
	"time"
)

// UserPreferences stores user scoring preferences
type UserPreferences struct {
	KeywordWeights map[string]float64
	TierThresholds map[SignalTier]float64
}

// NewUserPreferences creates default user preferences
func NewUserPreferences() *UserPreferences {
	return &UserPreferences{
		KeywordWeights: make(map[string]float64),
		TierThresholds: map[SignalTier]float64{
			TierCritical:   0.85,
			TierImportant:  0.65,
			TierNiceToHave: 0.45,
			TierNoise:      0.0,
		},
	}
}

// ScoringOptions provides options for scoring content
type ScoringOptions struct {
	Query        string
	Timestamp    *time.Time
	DocumentFreq map[string]int
}

// ScoredContent represents scored content with statistics
type ScoredContent struct {
	Lines      []*ScoredLine
	TotalLines int
	AvgScore   float64
	MaxScore   float64
	MinScore   float64
	TierCounts map[SignalTier]int
}

// ScoredLine represents a single scored line
type ScoredLine struct {
	LineNumber int
	Content    string
	Score      float64
	Tier       SignalTier
}

// Summary returns a human-readable summary
func (sc *ScoredContent) Summary() string {
	return fmt.Sprintf("Scored %d lines (avg: %.2f, max: %.2f)",
		len(sc.Lines), sc.AvgScore, sc.MaxScore)
}
