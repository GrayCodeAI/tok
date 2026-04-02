package responsecompress

import (
	"strings"
)

type ResponseCompressor struct {
	stripWhitespace bool
	stripFiller     bool
	maxTokens       int
}

func NewResponseCompressor() *ResponseCompressor {
	return &ResponseCompressor{
		stripWhitespace: true,
		stripFiller:     true,
		maxTokens:       0,
	}
}

func (rc *ResponseCompressor) Compress(response string) string {
	output := response

	if rc.stripWhitespace {
		output = stripWhitespace(output)
	}

	if rc.stripFiller {
		output = stripFillerWords(output)
	}

	if rc.maxTokens > 0 {
		output = truncateToTokens(output, rc.maxTokens)
	}

	return output
}

func (rc *ResponseCompressor) SetMaxTokens(n int) {
	rc.maxTokens = n
}

func (rc *ResponseCompressor) SetStripWhitespace(v bool) {
	rc.stripWhitespace = v
}

func (rc *ResponseCompressor) SetStripFiller(v bool) {
	rc.stripFiller = v
}

func stripWhitespace(input string) string {
	lines := strings.Split(input, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, strings.TrimRight(line, " \t"))
	}
	output := strings.Join(result, "\n")

	for strings.Contains(output, "\n\n\n") {
		output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(output)
}

func stripFillerWords(input string) string {
	fillers := []string{
		"Here is the ", "Here's the ", "The following ",
		"Note that ", "Keep in mind ", "As mentioned ",
		"As you can see ", "Of course ", "Certainly ",
	}
	output := input
	for _, filler := range fillers {
		output = strings.ReplaceAll(output, filler, "")
	}
	return output
}

func truncateToTokens(input string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(input) <= maxChars {
		return input
	}
	return input[:maxChars]
}

type ResponseMetrics struct {
	OriginalTokens   int     `json:"original_tokens"`
	CompressedTokens int     `json:"compressed_tokens"`
	Savings          int     `json:"savings"`
	SavingsPct       float64 `json:"savings_pct"`
}

func CalculateResponseMetrics(original, compressed string) *ResponseMetrics {
	origTokens := len(original) / 4
	compTokens := len(compressed) / 4
	savings := origTokens - compTokens
	pct := 0.0
	if origTokens > 0 {
		pct = float64(savings) / float64(origTokens) * 100
	}
	return &ResponseMetrics{
		OriginalTokens:   origTokens,
		CompressedTokens: compTokens,
		Savings:          savings,
		SavingsPct:       pct,
	}
}
