package wrapped

import (
	"fmt"
	"strings"
)

type WrappedStats struct {
	TotalTokens       int     `json:"total_tokens"`
	TotalSaved        int     `json:"total_saved"`
	SavingsPercentage float64 `json:"savings_percentage"`
	TopModel          string  `json:"top_model"`
	TopCommand        string  `json:"top_command"`
	TotalCommands     int     `json:"total_commands"`
	TotalSessions     int     `json:"total_sessions"`
	AvgCompression    float64 `json:"avg_compression"`
	BestDay           string  `json:"best_day"`
	BestDaySavings    int     `json:"best_day_savings"`
}

type WrappedGenerator struct{}

func NewWrappedGenerator() *WrappedGenerator {
	return &WrappedGenerator{}
}

func (g *WrappedGenerator) Generate(stats *WrappedStats) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════╗\n")
	sb.WriteString("║        TOKMAN WRAPPED 2026           ║\n")
	sb.WriteString("╠══════════════════════════════════════╣\n")
	sb.WriteString("║                                      ║\n")
	sb.WriteString(fmt.Sprintf("║  Total Tokens Processed: %8d  ║\n", stats.TotalTokens))
	sb.WriteString(fmt.Sprintf("║  Tokens Saved:           %8d  ║\n", stats.TotalSaved))
	sb.WriteString(fmt.Sprintf("║  Savings:                %7.1f%%  ║\n", stats.SavingsPercentage))
	sb.WriteString("║                                      ║\n")
	sb.WriteString(fmt.Sprintf("║  Top Model:    %-20s  ║\n", stats.TopModel))
	sb.WriteString(fmt.Sprintf("║  Top Command:  %-20s  ║\n", stats.TopCommand))
	sb.WriteString(fmt.Sprintf("║  Commands:     %-20d  ║\n", stats.TotalCommands))
	sb.WriteString(fmt.Sprintf("║  Sessions:     %-20d  ║\n", stats.TotalSessions))
	sb.WriteString("║                                      ║\n")
	sb.WriteString(fmt.Sprintf("║  Best Day: %s (%d saved)  ║\n", stats.BestDay, stats.BestDaySavings))
	sb.WriteString("║                                      ║\n")
	sb.WriteString("╚══════════════════════════════════════╝\n")

	return sb.String()
}

func (g *WrappedGenerator) GenerateSVG(stats *WrappedStats) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300">
<rect width="400" height="300" fill="#1a1a2e"/>
<text x="200" y="40" text-anchor="middle" fill="#e94560" font-size="24" font-weight="bold">TokMan Wrapped 2026</text>
<text x="50" y="90" fill="#eee" font-size="16">Total Tokens: %d</text>
<text x="50" y="120" fill="#eee" font-size="16">Saved: %d (%.1f%%)</text>
<text x="50" y="160" fill="#eee" font-size="16">Top Model: %s</text>
<text x="50" y="190" fill="#eee" font-size="16">Commands: %d</text>
<text x="50" y="220" fill="#eee" font-size="16">Sessions: %d</text>
<text x="50" y="260" fill="#e94560" font-size="14">#TokManWrapped</text>
</svg>`, stats.TotalTokens, stats.TotalSaved, stats.SavingsPercentage, stats.TopModel, stats.TotalCommands, stats.TotalSessions)
}
