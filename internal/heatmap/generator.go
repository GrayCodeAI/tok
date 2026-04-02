package heatmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HeatmapGenerator struct {
	analyzer *SectionAnalyzer
}

func NewHeatmapGenerator() *HeatmapGenerator {
	return &HeatmapGenerator{
		analyzer: NewSectionAnalyzer(),
	}
}

func (g *HeatmapGenerator) Generate(input string) *HeatmapData {
	sections := g.analyzer.Analyze(input)
	totalTokens := 0
	for _, s := range sections {
		totalTokens += s.TokenCount
	}

	return &HeatmapData{
		TotalTokens: totalTokens,
		Sections:    sections,
		Timestamp:   0,
	}
}

func (g *HeatmapGenerator) GenerateWithWaste(input string, wasteScore float64) *HeatmapData {
	data := g.Generate(input)
	data.WasteScore = wasteScore
	return data
}

func (g *HeatmapGenerator) ToJSON(data *HeatmapData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (g *HeatmapGenerator) ToCSV(data *HeatmapData) string {
	var sb strings.Builder
	sb.WriteString("type,token_count,percentage,start_line,end_line\n")
	for _, s := range data.Sections {
		sb.WriteString(fmt.Sprintf("%s,%d,%.2f,%d,%d\n",
			s.Type, s.TokenCount, s.Percentage, s.StartLine, s.EndLine))
	}
	return sb.String()
}

func (g *HeatmapGenerator) Summary(data *HeatmapData) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total Tokens: %d\n", data.TotalTokens))
	sb.WriteString(fmt.Sprintf("Sections: %d\n", len(data.Sections)))
	if data.WasteScore > 0 {
		sb.WriteString(fmt.Sprintf("Waste Score: %.1f%%\n", data.WasteScore))
	}
	sb.WriteString("\nBreakdown:\n")
	for _, s := range data.Sections {
		bar := strings.Repeat("█", int(s.Percentage/2))
		sb.WriteString(fmt.Sprintf("  %-10s %6d tokens (%5.1f%%) %s\n",
			s.Type, s.TokenCount, s.Percentage, bar))
	}
	return sb.String()
}
