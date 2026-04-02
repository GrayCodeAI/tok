package historybloat

import (
	"strings"
)

type HistoryBloatReport struct {
	TotalTokens          int     `json:"total_tokens"`
	HistoryTokens        int     `json:"history_tokens"`
	HistoryPercentage    float64 `json:"history_percentage"`
	IsBloated            bool    `json:"is_bloated"`
	RedundantEntries     int     `json:"redundant_entries"`
	CompressionPotential int     `json:"compression_potential_tokens"`
	Recommendation       string  `json:"recommendation"`
}

type HistoryAnalyzer struct {
	threshold float64
}

func NewHistoryAnalyzer() *HistoryAnalyzer {
	return &HistoryAnalyzer{threshold: 60.0}
}

func NewHistoryAnalyzerWithThreshold(threshold float64) *HistoryAnalyzer {
	return &HistoryAnalyzer{threshold: threshold}
}

func (a *HistoryAnalyzer) Analyze(input string) *HistoryBloatReport {
	lines := strings.Split(input, "\n")
	totalTokens := len(input) / 4

	var historyLines []string
	inHistory := false
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lower, "<history>") || strings.Contains(lower, "previous") || strings.Contains(lower, "conversation") {
			inHistory = true
		}
		if strings.Contains(lower, "</history>") {
			inHistory = false
			continue
		}
		if inHistory || strings.HasPrefix(lower, "user:") || strings.HasPrefix(lower, "assistant:") || strings.HasPrefix(lower, "human:") || strings.HasPrefix(lower, "ai:") {
			historyLines = append(historyLines, line)
		}
	}

	historyContent := strings.Join(historyLines, "\n")
	historyTokens := len(historyContent) / 4

	var historyPct float64
	if totalTokens > 0 {
		historyPct = float64(historyTokens) / float64(totalTokens) * 100
	}

	redundant := a.countRedundant(historyLines)
	compressionPotential := redundant * 3

	isBloated := historyPct > a.threshold
	recommendation := "History is within acceptable limits"
	if isBloated {
		recommendation = "History exceeds " + string(rune(int(a.threshold))) + "% threshold. Consider compacting or summarizing conversation history."
	}

	return &HistoryBloatReport{
		TotalTokens:          totalTokens,
		HistoryTokens:        historyTokens,
		HistoryPercentage:    historyPct,
		IsBloated:            isBloated,
		RedundantEntries:     redundant,
		CompressionPotential: compressionPotential,
		Recommendation:       recommendation,
	}
}

func (a *HistoryAnalyzer) countRedundant(lines []string) int {
	seen := make(map[string]bool)
	redundant := 0
	for _, line := range lines {
		normalized := strings.TrimSpace(strings.ToLower(line))
		if len(normalized) < 30 {
			continue
		}
		if seen[normalized] {
			redundant++
		} else {
			seen[normalized] = true
		}
	}
	return redundant
}
