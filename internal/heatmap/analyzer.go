package heatmap

import "strings"

type SectionType string

const (
	SectionSystem  SectionType = "system"
	SectionTools   SectionType = "tools"
	SectionContext SectionType = "context"
	SectionHistory SectionType = "history"
	SectionQuery   SectionType = "query"
	SectionOutput  SectionType = "output"
	SectionUnknown SectionType = "unknown"
)

type HeatmapSection struct {
	Type       SectionType `json:"type"`
	TokenCount int         `json:"token_count"`
	Percentage float64     `json:"percentage"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
	Content    string      `json:"content,omitempty"`
}

type HeatmapData struct {
	TotalTokens int              `json:"total_tokens"`
	Sections    []HeatmapSection `json:"sections"`
	WasteScore  float64          `json:"waste_score"`
	Timestamp   int64            `json:"timestamp"`
}

type SectionAnalyzer struct {
	systemMarkers  []string
	toolMarkers    []string
	contextMarkers []string
	historyMarkers []string
}

func NewSectionAnalyzer() *SectionAnalyzer {
	return &SectionAnalyzer{
		systemMarkers: []string{
			"<system>", "system prompt", "you are a", "your role",
			"your task", "instructions", "guidelines", "rules:",
			"system message", "assistant is", "helpful assistant",
		},
		toolMarkers: []string{
			"<tool>", "<function>", "tool call", "function call",
			"tool_use", "tool_call", "tool_result", "parameters", "arguments",
		},
		contextMarkers: []string{
			"<context>", "file:", "path:", "source code",
			"code snippet", "```", "import ", "package ",
		},
		historyMarkers: []string{
			"<history>", "previous", "earlier", "conversation",
			"user:", "assistant:", "human:", "ai:",
		},
	}
}

func (a *SectionAnalyzer) ClassifyLine(line string) SectionType {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return SectionUnknown
	}

	for _, marker := range a.systemMarkers {
		if strings.Contains(lower, marker) {
			return SectionSystem
		}
	}
	for _, marker := range a.toolMarkers {
		if strings.Contains(lower, marker) {
			return SectionTools
		}
	}
	for _, marker := range a.contextMarkers {
		if strings.Contains(lower, marker) {
			return SectionContext
		}
	}
	for _, marker := range a.historyMarkers {
		if strings.Contains(lower, marker) {
			return SectionHistory
		}
	}

	return SectionQuery
}

func (a *SectionAnalyzer) Analyze(input string) []HeatmapSection {
	lines := strings.Split(input, "\n")
	var sections []HeatmapSection
	var currentType SectionType = SectionUnknown
	var startLine int
	var contentLines []string

	flushSection := func(endLine int) {
		if currentType != SectionUnknown || len(contentLines) > 0 {
			content := strings.Join(contentLines, "\n")
			tokenCount := estimateTokens(content)
			sections = append(sections, HeatmapSection{
				Type:       currentType,
				TokenCount: tokenCount,
				StartLine:  startLine,
				EndLine:    endLine,
			})
		}
	}

	for i, line := range lines {
		lineType := a.ClassifyLine(line)
		if lineType != currentType && i > 0 {
			flushSection(i - 1)
			currentType = lineType
			startLine = i
			contentLines = []string{line}
		} else {
			currentType = lineType
			contentLines = append(contentLines, line)
		}
	}

	if len(contentLines) > 0 {
		flushSection(len(lines) - 1)
	}

	totalTokens := 0
	for _, s := range sections {
		totalTokens += s.TokenCount
	}

	if totalTokens > 0 {
		for i := range sections {
			sections[i].Percentage = float64(sections[i].TokenCount) / float64(totalTokens) * 100
		}
	}

	return sections
}

func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text) / 4
}
