package filter

import (
	"fmt"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// TieredSummaryFilter implements L0/L1/L2 progressive summarization.
// L0: Surface summary (keywords, topics)
// L1: Structural summary (sections, relationships)
// L2: Deep summary (full semantic compression)
type TieredSummaryFilter struct {
	enabled bool
}

// SummaryTier defines the level of summarization.
type SummaryTier int

const (
	TierL0 SummaryTier = iota // Surface: keywords, entities
	TierL1                    // Structural: sections, outlines
	TierL2                    // Deep: full semantic compression
)

// TieredResult holds summaries at all three levels.
type TieredResult struct {
	L0 *L0Summary
	L1 *L1Summary
	L2 *L2Summary
}

// L0Summary is the surface-level summary.
type L0Summary struct {
	Keywords   []string `json:"keywords"`
	Entities   []string `json:"entities"`
	Topics     []string `json:"topics"`
	TokenCount int      `json:"token_count"`
}

// L1Summary is the structural summary.
type L1Summary struct {
	Title      string    `json:"title"`
	Sections   []Section `json:"sections"`
	Outline    string    `json:"outline"`
	TokenCount int       `json:"token_count"`
}

// Section represents a document section.
type Section struct {
	Heading string `json:"heading"`
	Summary string `json:"summary"`
	Level   int    `json:"level"`
}

// L2Summary is the deep semantic summary.
type L2Summary struct {
	Summary      string   `json:"summary"`
	KeyPoints    []string `json:"key_points"`
	Implications []string `json:"implications,omitempty"`
	TokenCount   int      `json:"token_count"`
}

// NewTieredSummaryFilter creates a new tiered summary filter.
func NewTieredSummaryFilter() *TieredSummaryFilter {
	return &TieredSummaryFilter{enabled: true}
}

// Name returns the filter name.
func (tsf *TieredSummaryFilter) Name() string { return "tiered_summary" }

// Apply generates tiered summaries from input.
func (tsf *TieredSummaryFilter) Apply(input string, mode Mode) (string, int) {
	if !tsf.enabled || mode == ModeNone || len(input) < 500 {
		return input, 0
	}

	result := tsf.GenerateTiers(input)

	// Select tier based on mode and content size
	tokens := core.EstimateTokens(input)
	var output string
	var tier SummaryTier

	switch {
	case tokens > 10000 || mode == ModeAggressive:
		// Use L2 deep summary for very large content
		output = tsf.formatL2(result.L2)
		tier = TierL2
	case tokens > 2000:
		// Use L1 structural summary for medium content
		output = tsf.formatL1(result.L1)
		tier = TierL1
	default:
		// Use L0 surface summary for smaller content
		output = tsf.formatL0(result.L0)
		tier = TierL0
	}

	saved := tokens - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}

	// Add tier marker
	output = fmt.Sprintf("[tier:%s]\n%s", tierName(tier), output)

	return output, saved
}

// GenerateTiers creates all three summary levels.
func (tsf *TieredSummaryFilter) GenerateTiers(input string) *TieredResult {
	return &TieredResult{
		L0: tsf.generateL0(input),
		L1: tsf.generateL1(input),
		L2: tsf.generateL2(input),
	}
}

// generateL0 creates surface-level summary.
func (tsf *TieredSummaryFilter) generateL0(input string) *L0Summary {
	// Extract keywords (heuristic: frequently occurring words)
	keywords := extractKeywords(input, 10)

	// Extract entities (heuristic: capitalized phrases)
	entities := extractEntities(input)

	// Infer topics from keywords
	topics := inferTopics(keywords)

	summary := &L0Summary{
		Keywords: keywords,
		Entities: entities,
		Topics:   topics,
	}
	summary.TokenCount = core.EstimateTokens(strings.Join(keywords, " ") + " " + strings.Join(entities, " "))
	return summary
}

// generateL1 creates structural summary.
func (tsf *TieredSummaryFilter) generateL1(input string) *L1Summary {
	lines := strings.Split(input, "\n")
	var sections []Section
	var currentSection *Section

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect section headers (markdown style)
		if strings.HasPrefix(trimmed, "#") {
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}
			level := 0
			for i, c := range trimmed {
				if c == '#' {
					level = i + 1
				} else {
					break
				}
			}
			currentSection = &Section{
				Heading: strings.TrimSpace(trimmed[level:]),
				Level:   level,
			}
		} else if currentSection != nil {
			currentSection.Summary += trimmed + " "
		}
	}

	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	// Truncate section summaries
	for i := range sections {
		sections[i].Summary = truncateSentence(sections[i].Summary, 150)
	}

	// Build outline
	outline := buildOutline(sections)

	// Extract title
	title := ""
	if len(sections) > 0 {
		title = sections[0].Heading
	}

	summary := &L1Summary{
		Title:    title,
		Sections: sections,
		Outline:  outline,
	}
	summary.TokenCount = core.EstimateTokens(outline)
	return summary
}

// generateL2 creates deep semantic summary.
func (tsf *TieredSummaryFilter) generateL2(input string) *L2Summary {
	// Extract key points (heuristic: first sentence of paragraphs)
	paragraphs := strings.Split(input, "\n\n")
	var keyPoints []string

	for _, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		// Take first sentence of significant paragraphs
		sentences := splitSentences(trimmed)
		if len(sentences) > 0 && len(trimmed) > 50 {
			keyPoints = append(keyPoints, sentences[0])
		}
		if len(keyPoints) >= 5 {
			break
		}
	}

	// Generate summary from key points
	summary := strings.Join(keyPoints, " ")

	// Limit summary length
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}

	result := &L2Summary{
		Summary:   summary,
		KeyPoints: keyPoints,
	}
	result.TokenCount = core.EstimateTokens(summary)
	return result
}

// formatL0 formats L0 summary for output.
func (tsf *TieredSummaryFilter) formatL0(l0 *L0Summary) string {
	var parts []string
	if len(l0.Topics) > 0 {
		parts = append(parts, fmt.Sprintf("Topics: %s", strings.Join(l0.Topics, ", ")))
	}
	if len(l0.Keywords) > 0 {
		parts = append(parts, fmt.Sprintf("Keywords: %s", strings.Join(l0.Keywords, ", ")))
	}
	if len(l0.Entities) > 0 {
		parts = append(parts, fmt.Sprintf("Entities: %s", strings.Join(l0.Entities, ", ")))
	}
	return strings.Join(parts, "\n")
}

// formatL1 formats L1 summary for output.
func (tsf *TieredSummaryFilter) formatL1(l1 *L1Summary) string {
	var parts []string
	if l1.Title != "" {
		parts = append(parts, fmt.Sprintf("# %s", l1.Title))
	}
	parts = append(parts, "## Outline")
	parts = append(parts, l1.Outline)
	if len(l1.Sections) > 0 {
		parts = append(parts, "## Sections")
		for _, sec := range l1.Sections {
			indent := strings.Repeat("  ", sec.Level-1)
			parts = append(parts, fmt.Sprintf("%s- %s: %s", indent, sec.Heading, sec.Summary))
		}
	}
	return strings.Join(parts, "\n")
}

// formatL2 formats L2 summary for output.
func (tsf *TieredSummaryFilter) formatL2(l2 *L2Summary) string {
	var parts []string
	parts = append(parts, "## Summary")
	parts = append(parts, l2.Summary)
	if len(l2.KeyPoints) > 0 {
		parts = append(parts, "\n## Key Points")
		for i, point := range l2.KeyPoints {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, point))
		}
	}
	return strings.Join(parts, "\n")
}

// Helper functions

func tierName(tier SummaryTier) string {
	switch tier {
	case TierL0:
		return "L0"
	case TierL1:
		return "L1"
	case TierL2:
		return "L2"
	default:
		return "unknown"
	}
}

func extractKeywords(input string, count int) []string {
	// Simple heuristic: most frequent words
	words := strings.Fields(strings.ToLower(input))
	freq := make(map[string]int)
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}

	for _, w := range words {
		w = strings.TrimFunc(w, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		if len(w) > 3 && !stopwords[w] {
			freq[w]++
		}
	}

	// Get top keywords
	type wordFreq struct {
		word string
		freq int
	}
	var wf []wordFreq
	for w, f := range freq {
		wf = append(wf, wordFreq{w, f})
	}

	// Simple bubble sort for top N
	for i := 0; i < len(wf)-1 && i < count; i++ {
		for j := i + 1; j < len(wf); j++ {
			if wf[j].freq > wf[i].freq {
				wf[i], wf[j] = wf[j], wf[i]
			}
		}
	}

	var result []string
	for i := 0; i < len(wf) && i < count; i++ {
		result = append(result, wf[i].word)
	}
	return result
}

func extractEntities(input string) []string {
	// Simple heuristic: capitalized phrases
	words := strings.Fields(input)
	var entities []string
	var currentEntity []string

	for _, w := range words {
		if len(w) > 0 && w[0] >= 'A' && w[0] <= 'Z' {
			currentEntity = append(currentEntity, w)
		} else {
			if len(currentEntity) > 0 {
				entity := strings.Join(currentEntity, " ")
				if len(entity) > 3 {
					entities = append(entities, entity)
				}
				currentEntity = nil
			}
		}
	}

	if len(currentEntity) > 0 {
		entity := strings.Join(currentEntity, " ")
		if len(entity) > 3 {
			entities = append(entities, entity)
		}
	}

	return deduplicateStrings(entities)
}

func inferTopics(keywords []string) []string {
	// Simple topic inference from keywords
	topics := make(map[string]bool)

	codeIndicators := []string{"function", "class", "method", "variable", "import", "package"}
	docIndicators := []string{"document", "section", "chapter", "introduction", "conclusion"}
	dataIndicators := []string{"data", "json", "database", "query", "table", "column"}

	for _, kw := range keywords {
		for _, ind := range codeIndicators {
			if kw == ind {
				topics["code"] = true
			}
		}
		for _, ind := range docIndicators {
			if kw == ind {
				topics["documentation"] = true
			}
		}
		for _, ind := range dataIndicators {
			if kw == ind {
				topics["data"] = true
			}
		}
	}

	var result []string
	for t := range topics {
		result = append(result, t)
	}
	return result
}

func buildOutline(sections []Section) string {
	var parts []string
	for _, sec := range sections {
		indent := strings.Repeat("  ", sec.Level-1)
		parts = append(parts, fmt.Sprintf("%s- %s", indent, sec.Heading))
	}
	return strings.Join(parts, "\n")
}

func splitSentences(text string) []string {
	// Simple sentence splitting
	var sentences []string
	var current strings.Builder

	for i, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			// Check if next char is space or end
			if i+1 >= len(text) || text[i+1] == ' ' || text[i+1] == '\n' {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}

	return sentences
}

func truncateSentence(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// Compile-time check
var _ Filter = (*TieredSummaryFilter)(nil)
