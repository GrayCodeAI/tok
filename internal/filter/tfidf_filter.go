package filter

import (
	"math"
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// TFIDFFilter implements DSPC-style coarse-grained TF-IDF filtering.
// Research Source: "DSPC: Dual-Stage Progressive Compression" (Sep 2025)
// Key Innovation: Training-free coarse-to-fine compression using TF-IDF for
// sentence filtering + attention contribution for token pruning.
// Results: Beats LongLLMLingua by 7.76% using only 3x fewer tokens.
//
// This is a NEW pre-filter layer that runs before expensive layers (L2-L5).
// It scores sentences by TF-IDF and removes low-information sentences early,
// reducing the token budget for subsequent processing.
type TFIDFFilter struct {
	config TFIDFConfig
}

// TFIDFConfig holds configuration for TF-IDF filtering
type TFIDFConfig struct {
	// Enabled controls whether the filter is active
	Enabled bool

	// MinSentences is the minimum number of sentences required to apply filtering
	MinSentences int

	// Threshold for sentence importance (0.0-1.0)
	// Sentences below this TF-IDF score are pruned
	Threshold float64

	// MinContentLength is the minimum character length to apply filtering
	MinContentLength int
}

// DefaultTFIDFConfig returns default configuration
func DefaultTFIDFConfig() TFIDFConfig {
	return TFIDFConfig{
		Enabled:          true,
		MinSentences:     5,
		Threshold:        0.15,
		MinContentLength: 200,
	}
}


// NewTFIDFFilterWithConfig creates a filter with custom config
func NewTFIDFFilterWithConfig(cfg TFIDFConfig) *TFIDFFilter {
	return &TFIDFFilter{config: cfg}
}

// Name returns the filter name
func (f *TFIDFFilter) Name() string {
	return "tfidf"
}

// Apply applies TF-IDF based coarse filtering
func (f *TFIDFFilter) Apply(input string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Split into sentences/lines
	sentences := f.splitSentences(input)
	if len(sentences) < f.config.MinSentences {
		return input, 0
	}

	// Compute TF-IDF scores for each sentence
	scores := f.computeTFIDF(sentences)

	// Determine threshold based on mode
	threshold := f.config.Threshold
	if mode == ModeAggressive {
		threshold += 0.1
	}

	// Keep sentences above threshold
	var kept []string
	for i, sent := range sentences {
		if scores[i] >= threshold || f.isStructural(sent) {
			kept = append(kept, sent)
		}
	}

	// Safety: keep at least 30% of sentences
	minKeep := int(math.Ceil(float64(len(sentences)) * 0.3))
	if len(kept) < minKeep {
		// Add back highest-scoring removed sentences
		type scoredSent struct {
			content string
			score   float64
		}
		var removed []scoredSent
		keptSet := make(map[string]bool)
		for _, k := range kept {
			keptSet[k] = true
		}
		for i, sent := range sentences {
			if !keptSet[sent] {
				removed = append(removed, scoredSent{sent, scores[i]})
			}
		}
		// Sort by score descending (simple selection)
		for i := 0; i < len(removed) && len(kept) < minKeep; i++ {
			bestIdx := i
			for j := i + 1; j < len(removed); j++ {
				if removed[j].score > removed[bestIdx].score {
					bestIdx = j
				}
			}
			removed[i], removed[bestIdx] = removed[bestIdx], removed[i]
			kept = append(kept, removed[i].content)
		}
	}

	// Reconstruct preserving original order
	output := f.reconstructInOrder(input, kept)

	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 3 {
		return input, 0
	}

	return output, saved
}

// splitSentences splits content into sentences/lines
func (f *TFIDFFilter) splitSentences(input string) []string {
	// First try splitting by newlines
	lines := strings.Split(input, "\n")
	var sentences []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}

	// If too few lines, try sentence-level splitting
	if len(sentences) < f.config.MinSentences {
		sentenceRe := regexp.MustCompile(`[^.!?\n]+[.!?]`)
		matches := sentenceRe.FindAllString(input, -1)
		if len(matches) >= f.config.MinSentences {
			return matches
		}
	}

	return sentences
}

// computeTFIDF computes TF-IDF scores for each sentence
func (f *TFIDFFilter) computeTFIDF(sentences []string) []float64 {
	n := len(sentences)
	scores := make([]float64, n)

	// Build term frequency per sentence and document frequency
	sentenceTFs := make([]map[string]int, n)
	df := make(map[string]int)

	for i, sent := range sentences {
		words := tokenizeWords(sent)
		tf := make(map[string]int)
		seen := make(map[string]bool)
		for _, w := range words {
			w = strings.ToLower(w)
			tf[w]++
			if !seen[w] {
				df[w]++
				seen[w] = true
			}
		}
		sentenceTFs[i] = tf
	}

	// Compute TF-IDF score for each sentence
	for i, tf := range sentenceTFs {
		var score float64
		totalTerms := 0
		for _, count := range tf {
			totalTerms += count
		}

		for term, count := range tf {
			// TF component (normalized)
			tfVal := float64(count) / float64(totalTerms)
			// IDF component
			idf := math.Log(float64(n) / float64(df[term]+1))
			score += tfVal * idf
		}

		// Normalize by sentence length to avoid bias toward long sentences
		if totalTerms > 0 {
			score /= math.Sqrt(float64(totalTerms))
		}

		scores[i] = score
	}

	// Normalize scores to [0, 1]
	maxScore := 0.0
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}
	if maxScore > 0 {
		for i := range scores {
			scores[i] /= maxScore
		}
	}

	return scores
}

// isStructural checks if a sentence should always be preserved
func (f *TFIDFFilter) isStructural(sent string) bool {
	structural := []string{
		"func ", "function ", "def ", "class ", "struct ",
		"import ", "package ", "use ", "require",
		"```", "ERROR", "FAIL", "PASS",
	}
	for _, s := range structural {
		if strings.Contains(sent, s) {
			return true
		}
	}
	return false
}

// reconstructInOrder rebuilds content preserving original sentence order
func (f *TFIDFFilter) reconstructInOrder(original string, kept []string) string {
	keptSet := make(map[string]int) // sentence -> priority
	for i, k := range kept {
		keptSet[k] = i
	}

	var result strings.Builder
	lines := strings.Split(original, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString("\n")
			continue
		}
		if _, ok := keptSet[trimmed]; ok {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// tokenizeWords splits text into words (simple whitespace + punctuation)
func tokenizeWords(text string) []string {
	return regexp.MustCompile(`\w+`).FindAllString(text, -1)
}
