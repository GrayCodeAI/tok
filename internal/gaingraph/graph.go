package gaingraph

import (
	"fmt"
	"strings"
	"sync"
)

type DailyEntry struct {
	Date     string `json:"date"`
	Tokens   int64  `json:"tokens"`
	Saved    int64  `json:"saved"`
	Commands int    `json:"commands"`
}

type GainGraph struct {
	entries map[string]*DailyEntry
	mu      sync.RWMutex
}

func NewGainGraph() *GainGraph {
	return &GainGraph{
		entries: make(map[string]*DailyEntry),
	}
}

func (g *GainGraph) Record(date string, tokens, saved int64, commands int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if entry, ok := g.entries[date]; ok {
		entry.Tokens += tokens
		entry.Saved += saved
		entry.Commands += commands
	} else {
		g.entries[date] = &DailyEntry{Date: date, Tokens: tokens, Saved: saved, Commands: commands}
	}
}

func (g *GainGraph) GetDaily(days int) []*DailyEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []*DailyEntry
	for _, e := range g.entries {
		result = append(result, e)
	}
	if days > 0 && days < len(result) {
		result = result[:days]
	}
	return result
}

func (g *GainGraph) RenderGraph(width int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.entries) == 0 {
		return "No data\n"
	}

	chars := []string{" ", "░", "▒", "▓", "█"}
	var sb strings.Builder
	sb.WriteString("Token Savings Graph (30 days)\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	for i := 0; i < width/2 && i < len(g.entries); i++ {
		maxSaved := int64(1)
		for _, e := range g.entries {
			if e.Saved > maxSaved {
				maxSaved = e.Saved
			}
		}
		level := 0
		for _, e := range g.entries {
			if e.Saved > 0 {
				ratio := e.Saved * 4 / maxSaved
				level = int(ratio)
			}
		}
		if level > 4 {
			level = 4
		}
		sb.WriteString(chars[level])
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")

	totalSaved := int64(0)
	for _, e := range g.entries {
		totalSaved += e.Saved
	}
	sb.WriteString(fmt.Sprintf("Total: %d tokens saved\n", totalSaved))
	return sb.String()
}
