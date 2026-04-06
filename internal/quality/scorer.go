// Package quality provides compression quality scoring and validation.
// This is a competitive feature vs LLMLingua and AutoCompressor that
// automatically measures compression quality beyond just token count.
package quality

import (
	"fmt"
	"math"
	"strings"
)

// QualityScore represents a comprehensive quality assessment of compression.
type QualityScore struct {
	Overall            float64 // 0-100, weighted average of all metrics
	SemanticPreserved  float64 // 0-100, how much meaning is retained
	StructureIntact    float64 // 0-100, code/document structure preservation
	ReadabilityScore   float64 // 0-100, human readability after compression
	InformationDensity float64 // 0-100, information per token
	CompressionRatio   float64 // 0-100, based on token reduction
	KeywordsPreserved  float64 // 0-100, important terms retained
	
	Details string // Human-readable explanation
	Grade   string // A+, A, B+, B, C+, C, D, F
}

// ScoreCompression evaluates the quality of compression.
func ScoreCompression(original, compressed string, originalTokens, compressedTokens int) QualityScore {
	score := QualityScore{}
	
	// 1. Compression Ratio (20% weight)
	if originalTokens > 0 {
		ratio := float64(compressedTokens) / float64(originalTokens)
		score.CompressionRatio = (1.0 - ratio) * 100
		if score.CompressionRatio > 100 {
			score.CompressionRatio = 100
		}
	}
	
	// 2. Keywords Preserved (25% weight)
	score.KeywordsPreserved = calculateKeywordPreservation(original, compressed)
	
	// 3. Structure Intact (20% weight)
	score.StructureIntact = calculateStructurePreservation(original, compressed)
	
	// 4. Readability (15% weight)
	score.ReadabilityScore = calculateReadability(compressed)
	
	// 5. Information Density (10% weight)
	score.InformationDensity = calculateInformationDensity(compressed, compressedTokens)
	
	// 6. Semantic Preserved (10% weight) - heuristic-based
	score.SemanticPreserved = calculateSemanticPreservation(original, compressed)
	
	// Calculate weighted overall score
	score.Overall = (
		score.CompressionRatio*0.20 +
		score.KeywordsPreserved*0.25 +
		score.StructureIntact*0.20 +
		score.ReadabilityScore*0.15 +
		score.InformationDensity*0.10 +
		score.SemanticPreserved*0.10)
	
	// Assign grade
	score.Grade = assignGrade(score.Overall)
	
	// Generate details
	score.Details = generateScoreDetails(score)
	
	return score
}

// calculateKeywordPreservation measures how many important words are kept.
func calculateKeywordPreservation(original, compressed string) float64 {
	keywords := extractKeywords(original)
	if len(keywords) == 0 {
		return 100.0
	}
	
	compressedLower := strings.ToLower(compressed)
	preserved := 0
	for _, kw := range keywords {
		if strings.Contains(compressedLower, strings.ToLower(kw)) {
			preserved++
		}
	}
	
	return (float64(preserved) / float64(len(keywords))) * 100
}

// extractKeywords identifies important words (nouns, technical terms, etc.)
func extractKeywords(text string) []string {
	// Simple heuristic: words longer than 5 chars, capitalized, or technical
	words := strings.Fields(text)
	keywords := make(map[string]bool)
	
	for _, word := range words {
		cleaned := strings.Trim(word, ".,;:!?()[]{}\"'")
		
		// Include if:
		// - Length > 5 chars
		// - Starts with capital
		// - Contains underscore (technical)
		// - All caps (acronym)
		if len(cleaned) > 5 || 
		   (len(cleaned) > 0 && cleaned[0] >= 'A' && cleaned[0] <= 'Z') ||
		   strings.Contains(cleaned, "_") ||
		   (len(cleaned) > 1 && strings.ToUpper(cleaned) == cleaned) {
			keywords[cleaned] = true
		}
	}
	
	result := make([]string, 0, len(keywords))
	for kw := range keywords {
		result = append(result, kw)
	}
	return result
}

// calculateStructurePreservation checks if structure markers are kept.
func calculateStructurePreservation(original, compressed string) float64 {
	structureMarkers := []string{
		"{", "}", "[", "]", "(", ")",
		"class ", "func ", "def ", "function",
		"error:", "warning:", "info:",
		"//", "/*", "*/", "#",
	}
	
	originalCount := 0
	compressedCount := 0
	
	for _, marker := range structureMarkers {
		originalCount += strings.Count(original, marker)
		compressedCount += strings.Count(compressed, marker)
	}
	
	if originalCount == 0 {
		return 100.0
	}
	
	preservation := (float64(compressedCount) / float64(originalCount)) * 100
	if preservation > 100 {
		preservation = 100
	}
	
	return preservation
}

// calculateReadability estimates how readable the compressed text is.
func calculateReadability(text string) float64 {
	if len(text) == 0 {
		return 0.0
	}
	
	lines := strings.Split(text, "\n")
	words := strings.Fields(text)
	
	// Factors that reduce readability:
	score := 100.0
	
	// Too short per line
	avgLineLen := float64(len(text)) / float64(len(lines))
	if avgLineLen < 10 {
		score -= 20
	}
	
	// Too few words
	if len(words) < 5 {
		score -= 20
	}
	
	// Check for complete words (not truncated)
	truncatedWords := 0
	for _, word := range words {
		if len(word) > 1 && strings.HasSuffix(word, "...") {
			truncatedWords++
		}
	}
	if len(words) > 0 {
		truncationRate := float64(truncatedWords) / float64(len(words))
		score -= truncationRate * 30
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateInformationDensity measures information per token.
func calculateInformationDensity(text string, tokens int) float64 {
	if tokens == 0 {
		return 0.0
	}
	
	// Measure unique meaningful words
	words := strings.Fields(text)
	uniqueWords := make(map[string]bool)
	meaningfulWords := 0
	
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true,
	}
	
	for _, word := range words {
		cleaned := strings.ToLower(strings.Trim(word, ".,;:!?()[]{}\"'"))
		if len(cleaned) > 2 && !stopWords[cleaned] {
			uniqueWords[cleaned] = true
			meaningfulWords++
		}
	}
	
	// Information density = unique meaningful words / total tokens
	density := (float64(len(uniqueWords)) / float64(tokens)) * 100
	
	// Normalize to 0-100 scale
	if density > 100 {
		density = 100
	}
	
	return density
}

// calculateSemanticPreservation estimates semantic similarity (heuristic).
func calculateSemanticPreservation(original, compressed string) float64 {
	// Simple heuristic: check if key phrases/sentences are preserved
	originalSentences := strings.Split(original, ".")
	_ = compressed // use compressed for future enhancement
	
	if len(originalSentences) == 0 {
		return 100.0
	}
	
	// Check how many original sentences have some representation in compressed
	preserved := 0
	for _, origSent := range originalSentences {
		origWords := strings.Fields(origSent)
		if len(origWords) < 3 {
			continue // Skip very short sentences
		}
		
		// Check if at least 30% of words appear in compressed text
		foundWords := 0
		for _, word := range origWords {
			if len(word) > 3 && strings.Contains(compressed, word) {
				foundWords++
			}
		}
		
		if float64(foundWords)/float64(len(origWords)) >= 0.3 {
			preserved++
		}
	}
	
	significantSentences := 0
	for _, sent := range originalSentences {
		if len(strings.Fields(sent)) >= 3 {
			significantSentences++
		}
	}
	
	if significantSentences == 0 {
		return 100.0
	}
	
	return (float64(preserved) / float64(significantSentences)) * 100
}

// assignGrade converts score to letter grade.
func assignGrade(score float64) string {
	switch {
	case score >= 97:
		return "A+"
	case score >= 93:
		return "A"
	case score >= 90:
		return "A-"
	case score >= 87:
		return "B+"
	case score >= 83:
		return "B"
	case score >= 80:
		return "B-"
	case score >= 77:
		return "C+"
	case score >= 73:
		return "C"
	case score >= 70:
		return "C-"
	case score >= 67:
		return "D+"
	case score >= 63:
		return "D"
	case score >= 60:
		return "D-"
	default:
		return "F"
	}
}

// generateScoreDetails creates human-readable explanation.
func generateScoreDetails(score QualityScore) string {
	var details strings.Builder
	
	details.WriteString(fmt.Sprintf("Overall Quality: %.1f%% (%s)\n", score.Overall, score.Grade))
	details.WriteString("\nBreakdown:\n")
	details.WriteString(fmt.Sprintf("  • Compression Ratio: %.1f%%\n", score.CompressionRatio))
	details.WriteString(fmt.Sprintf("  • Keywords Preserved: %.1f%%\n", score.KeywordsPreserved))
	details.WriteString(fmt.Sprintf("  • Structure Intact: %.1f%%\n", score.StructureIntact))
	details.WriteString(fmt.Sprintf("  • Readability: %.1f%%\n", score.ReadabilityScore))
	details.WriteString(fmt.Sprintf("  • Information Density: %.1f%%\n", score.InformationDensity))
	details.WriteString(fmt.Sprintf("  • Semantic Preserved: %.1f%%\n", score.SemanticPreserved))
	
	// Add recommendations
	details.WriteString("\nRecommendations:\n")
	if score.KeywordsPreserved < 70 {
		details.WriteString("  ⚠️  Consider using less aggressive compression to preserve key terms\n")
	}
	if score.ReadabilityScore < 70 {
		details.WriteString("  ⚠️  Output may be hard to read; try 'minimal' mode\n")
	}
	if score.StructureIntact < 70 {
		details.WriteString("  ⚠️  Code structure may be damaged; enable AST preservation\n")
	}
	if score.Overall >= 90 {
		details.WriteString("  ✅ Excellent compression quality!\n")
	} else if score.Overall >= 80 {
		details.WriteString("  ✅ Good compression quality\n")
	} else if score.Overall >= 70 {
		details.WriteString("  ⚠️  Acceptable but could be improved\n")
	} else {
		details.WriteString("  ❌ Quality needs improvement; try different settings\n")
	}
	
	return details.String()
}

// CompareCompressionMethods compares multiple compression approaches.
func CompareCompressionMethods(original string, originalTokens int, results map[string]struct {
	Compressed string
	Tokens     int
}) map[string]QualityScore {
	scores := make(map[string]QualityScore)
	
	for method, result := range results {
		scores[method] = ScoreCompression(original, result.Compressed, originalTokens, result.Tokens)
	}
	
	return scores
}

// RecommendBestMethod finds the best compression method based on quality.
func RecommendBestMethod(scores map[string]QualityScore) (string, QualityScore) {
	var bestMethod string
	var bestScore QualityScore
	bestOverall := -1.0
	
	for method, score := range scores {
		if score.Overall > bestOverall {
			bestOverall = score.Overall
			bestMethod = method
			bestScore = score
		}
	}
	
	return bestMethod, bestScore
}

// CalculateQualityTrend tracks quality over time.
func CalculateQualityTrend(historicalScores []QualityScore) (trend string, avgScore float64) {
	if len(historicalScores) == 0 {
		return "no data", 0.0
	}
	
	// Calculate average
	sum := 0.0
	for _, score := range historicalScores {
		sum += score.Overall
	}
	avgScore = sum / float64(len(historicalScores))
	
	// Determine trend
	if len(historicalScores) < 2 {
		return "stable", avgScore
	}
	
	// Compare first half vs second half
	midpoint := len(historicalScores) / 2
	firstHalfAvg := 0.0
	secondHalfAvg := 0.0
	
	for i := 0; i < midpoint; i++ {
		firstHalfAvg += historicalScores[i].Overall
	}
	firstHalfAvg /= float64(midpoint)
	
	for i := midpoint; i < len(historicalScores); i++ {
		secondHalfAvg += historicalScores[i].Overall
	}
	secondHalfAvg /= float64(len(historicalScores) - midpoint)
	
	diff := secondHalfAvg - firstHalfAvg
	
	switch {
	case math.Abs(diff) < 2:
		trend = "stable"
	case diff > 0:
		trend = "improving"
	default:
		trend = "declining"
	}
	
	return trend, avgScore
}
