package filter

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/GrayCodeAI/tokman/internal/simd"
)

// Pre-compiled high importance keywords for fast scoring
var highKeywords = []string{
	"error", "failed", "failure", "fatal", "panic",
	"exception", "critical", "bug", "security",
	"diff --git", "deleted", "added", "modified",
}

// Pre-compiled medium importance keywords
var mediumKeywords = []string{
	"warning", "deprecated", "todo", "fixme",
	"test", "assert", "expect", "verify",
	"function", "class", "struct", "interface",
}

// Selective cache for repeated content (Phase 6)
// Only caches content that appears multiple times
type selectiveCache struct {
	mu          sync.RWMutex
	frequencies map[string]int // Track frequency of content hashes
	cache       map[string]cachedResult
	maxEntries  int
	minHits     int // Minimum hits before caching
}

type cachedResult struct {
	output      string
	tokensSaved int
}

var globalSelectiveCache = &selectiveCache{
	frequencies: make(map[string]int),
	cache:       make(map[string]cachedResult),
	maxEntries:  1000,
	minHits:     2, // Cache after 2nd occurrence
}

// makeCacheKey generates a cache key from content using first 100 chars + length.
// This provides fast lookup without computing full hashes.
func makeCacheKey(content string) string {
	return fmt.Sprintf("%d:%s", len(content), content[:min(100, len(content))])
}

// checkCache returns cached result if available, otherwise tracks frequency
func (sc *selectiveCache) checkCache(content string) (string, int, bool) {
	key := makeCacheKey(content)

	sc.mu.RLock()
	if cached, ok := sc.cache[key]; ok {
		sc.mu.RUnlock()
		return cached.output, cached.tokensSaved, true
	}
	sc.mu.RUnlock()

	// Track frequency
	sc.mu.Lock()
	sc.frequencies[key]++
	sc.mu.Unlock()

	return "", 0, false
}

// shouldCache returns true if content should be cached (high frequency)
func (sc *selectiveCache) shouldCache(content string) bool {
	key := makeCacheKey(content)

	sc.mu.RLock()
	freq := sc.frequencies[key]
	sc.mu.RUnlock()

	return freq >= sc.minHits
}

// storeCache caches a result if the content is frequent
func (sc *selectiveCache) storeCache(content string, output string, tokensSaved int) {
	if !sc.shouldCache(content) {
		return
	}

	key := makeCacheKey(content)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Evict old entries if full
	if len(sc.cache) >= sc.maxEntries {
		count := 0
		for k := range sc.cache {
			delete(sc.cache, k)
			count++
			if count >= sc.maxEntries/2 {
				break
			}
		}
	}

	sc.cache[key] = cachedResult{output: output, tokensSaved: tokensSaved}
}

// HierarchicalFilter implements multi-level summarization for large outputs.
// Based on "Hierarchical Context Compression" research - creates a tree-like
// structure where each level provides progressively more detail.
//
// For outputs exceeding a threshold (default 10K lines), this filter:
// 1. Segments the output into logical sections
// 2. Generates summaries at multiple abstraction levels
// 3. Preserves the most important sections verbatim
// 4. Compresses mid-importance sections into summaries
// 5. Drops low-importance sections entirely
type HierarchicalFilter struct {
	// Threshold for triggering hierarchical compression (in lines)
	lineThreshold int
	// Maximum depth of summarization hierarchy
	maxDepth int
	// Whether to use semantic scoring for section importance
	useSemanticScoring bool
	// Cached semantic filter (reused across calls)
	semanticFilter *SemanticFilter
}

// NewHierarchicalFilter creates a new hierarchical summarization filter.
func NewHierarchicalFilter() *HierarchicalFilter {
	return &HierarchicalFilter{
		lineThreshold:      500, // 500 lines = ~10K tokens
		maxDepth:           3,   // 3 levels: overview, summary, detail
		useSemanticScoring: true,
		semanticFilter:     NewSemanticFilter(), // Cache for reuse
	}
}

// Name returns the filter name.
func (f *HierarchicalFilter) Name() string {
	return "hierarchical"
}

// Apply applies hierarchical summarization to large outputs.
// Optimized: Early exit for small/medium inputs, streaming for very large inputs

func (f *HierarchicalFilter) Apply(input string, mode Mode) (string, int) {
	// Quick size check before splitting
	inputLen := len(input)

	// Don't process small outputs - early exit with minimal work
	if inputLen < f.lineThreshold*40 { // ~40 chars per line average
		return input, 0
	}

	if cached, saved, found := globalSelectiveCache.checkCache(input); found {
		return cached, saved
	}

	// Estimate line count without full split for large inputs
	estimatedLines := inputLen / 40
	if estimatedLines < f.lineThreshold {
		return input, 0
	}

	// For very large inputs (>100KB), use streaming to reduce memory pressure
	var output string
	var tokensSaved int

	if inputLen > 100000 {
		output, tokensSaved = f.applyStreaming(input, mode)
	} else {
		// Now split lines
		lines := strings.Split(input, "\n")
		lineCount := len(lines)

		// Don't process small outputs
		if lineCount < f.lineThreshold {
			return input, 0
		}

		// Segment into logical sections
		sections := f.segmentIntoSections(lines)
		if len(sections) == 0 {
			return input, 0
		}

		// Score each section
		scored := f.scoreSections(sections)

		// Build hierarchical output based on mode
		output = f.buildHierarchicalOutput(scored, mode, lineCount)

		tokensSaved = EstimateTokens(input) - EstimateTokens(output)
		if tokensSaved < 0 {
			tokensSaved = 0
		}
	}

	globalSelectiveCache.storeCache(input, output, tokensSaved)

	return output, tokensSaved
}

// applyStreaming processes very large inputs using chunked streaming
// Reduces memory pressure by processing in segments
func (f *HierarchicalFilter) applyStreaming(input string, mode Mode) (string, int) {
	inputLen := len(input)
	chunkSize := 50000 // 50KB chunks
	var chunks []string

	// Split into chunks at line boundaries
	for i := 0; i < inputLen; i += chunkSize {
		end := i + chunkSize
		if end > inputLen {
			end = inputLen
		}

		// Find nearest line boundary
		if end < inputLen {
			for end > i && input[end] != '\n' {
				end--
			}
			if end == i {
				end = i + chunkSize // No newline found, use original
			} else {
				end++ // Include the newline
			}
		}

		chunks = append(chunks, input[i:end])
	}

	// Process each chunk independently
	var results []string
	totalSaved := 0

	for _, chunk := range chunks {
		result, saved := f.processChunk(chunk, mode)
		results = append(results, result)
		totalSaved += saved
	}

	output := strings.Join(results, "\n")
	return output, totalSaved
}

// processChunk processes a single chunk of content
func (f *HierarchicalFilter) processChunk(chunk string, mode Mode) (string, int) {
	lines := strings.Split(chunk, "\n")
	lineCount := len(lines)

	if lineCount < f.lineThreshold/2 {
		// Very small chunk, just return summary
		if len(chunk) > 100 {
			return chunk[:100] + "...", EstimateTokens(chunk) - 25
		}
		return chunk, 0
	}

	sections := f.segmentIntoSections(lines)
	if len(sections) == 0 {
		return chunk, 0
	}

	scored := f.scoreSections(sections)
	output := f.buildHierarchicalOutput(scored, mode, lineCount)

	tokensSaved := EstimateTokens(chunk) - EstimateTokens(output)
	if tokensSaved < 0 {
		tokensSaved = 0
	}

	return output, tokensSaved
}

// section represents a logical section of the output
type section struct {
	content   string
	startLine int
	endLine   int
	level     int // 0 = top-level, 1 = nested, etc.
	score     float64
	summary   string
}

// segmentIntoSections divides output into logical sections
// Optimized: Uses sampling for large inputs to avoid O(n) line-by-line checks
func (f *HierarchicalFilter) segmentIntoSections(lines []string) []section {
	lineCount := len(lines)

	// For large inputs, use section shortcuts (sampling-based)
	// Reduces from O(n) to O(n/samplingRate) for boundary detection
	// Lowered threshold from 2000 to 500 for better P99
	if lineCount > 500 {
		return f.segmentWithSampling(lines)
	}

	return f.segmentFull(lines)
}

// segmentFull performs full line-by-line segmentation (original algorithm)
func (f *HierarchicalFilter) segmentFull(lines []string) []section {
	var sections []section
	var currentSection []string
	sectionStart := 0
	currentLevel := 0

	for i, line := range lines {
		level := f.detectSectionLevel(line)

		// Check for section boundary
		isBoundary := f.isSectionBoundary(line, i, lines)

		if isBoundary && len(currentSection) > 0 {
			// Save current section
			sections = append(sections, section{
				content:   strings.Join(currentSection, "\n"),
				startLine: sectionStart,
				endLine:   i - 1,
				level:     currentLevel,
			})
			currentSection = nil
			sectionStart = i
			currentLevel = level
		}

		currentSection = append(currentSection, line)
	}

	// Add final section
	if len(currentSection) > 0 {
		sections = append(sections, section{
			content:   strings.Join(currentSection, "\n"),
			startLine: sectionStart,
			endLine:   len(lines) - 1,
			level:     currentLevel,
		})
	}

	return sections
}

// segmentWithSampling uses sampling to find section boundaries faster
// Processes every Nth line in first pass, then refines boundaries
func (f *HierarchicalFilter) segmentWithSampling(lines []string) []section {
	lineCount := len(lines)

	// Sampling rate: check 1 in every 10 lines for large inputs
	samplingRate := 10
	if lineCount > 10000 {
		samplingRate = 20
	}

	// First pass: find likely boundary positions using sampling
	likelyBoundaries := make([]int, 0, lineCount/samplingRate+1)
	likelyBoundaries = append(likelyBoundaries, 0) // Always start at 0

	for i := 0; i < lineCount; i += samplingRate {
		if i > 0 && f.isQuickBoundary(lines[i]) {
			likelyBoundaries = append(likelyBoundaries, i)
		}
	}
	likelyBoundaries = append(likelyBoundaries, lineCount-1) // Always end at last line

	// Second pass: refine boundaries by checking nearby lines
	// Build sections from refined boundaries
	var sections []section

	for b := 0; b < len(likelyBoundaries)-1; b++ {
		start := likelyBoundaries[b]
		end := likelyBoundaries[b+1]

		// Refine: scan a small window around the boundary
		refinedStart := start
		if b > 0 {
			// Look for actual boundary within ±5 lines
			windowStart := max(0, start-5)
			windowEnd := min(lineCount-1, start+5)
			for i := windowStart; i <= windowEnd; i++ {
				if f.isQuickBoundary(lines[i]) {
					refinedStart = i
					break
				}
			}
		}

		// Build section content
		sectionEnd := min(end, lineCount-1)
		if sectionEnd > refinedStart {
			content := strings.Join(lines[refinedStart:sectionEnd+1], "\n")
			sections = append(sections, section{
				content:   content,
				startLine: refinedStart,
				endLine:   sectionEnd,
				level:     0, // Simplified level for sampled sections
			})
		}
	}

	return sections
}

// isQuickBoundary performs a fast boundary check without context
func (f *HierarchicalFilter) isQuickBoundary(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}

	// Fast first-character check
	firstChar := trimmed[0]

	// Visual dividers and headers
	if firstChar == '=' || firstChar == '-' || firstChar == '+' || firstChar == '#' {
		return true
	}

	// Quick prefix check for common patterns
	if firstChar == 'd' && len(trimmed) > 10 && strings.HasPrefix(trimmed, "diff --git") {
		return true
	}

	// SIMD check for remaining patterns
	return simd.ContainsAny(trimmed, []string{
		"test result:", "running ", "error[", "error:",
		"Compiling ", "Building ", "Finished ",
	})
}

// detectSectionLevel determines the nesting level of a section
func (f *HierarchicalFilter) detectSectionLevel(line string) int {
	trimmed := strings.TrimSpace(line)

	// Headers and dividers indicate top-level sections
	if strings.HasPrefix(trimmed, "===") ||
		strings.HasPrefix(trimmed, "##") {
		return 0
	}

	// Subsection markers ("---" is a subsection divider)
	if strings.HasPrefix(trimmed, "---") {
		return 1
	}

	return 2 // Default to detail level
}

// isSectionBoundary detects if a line starts a new section
// Optimized: Uses SIMD operations and pre-compiled patterns
func (f *HierarchicalFilter) isSectionBoundary(line string, idx int, lines []string) bool {
	trimmed := strings.TrimSpace(line)
	trimmedLen := len(trimmed)

	// Fast path: empty line check
	if trimmedLen == 0 {
		if idx > 0 && idx < len(lines)-1 {
			prevLen := len(lines[idx-1])
			nextLen := len(lines[idx+1])
			// Quick check without TrimSpace for performance
			if prevLen > 50 || nextLen > 50 {
				return true
			}
		}
		return false
	}

	// First character check for fast filtering
	firstChar := trimmed[0]

	// Visual dividers (starts with =, -, +, #)
	if firstChar == '=' && strings.HasPrefix(trimmed, "===") {
		return true
	}
	if firstChar == '-' && strings.HasPrefix(trimmed, "---") {
		return true
	}
	if firstChar == '+' && strings.HasPrefix(trimmed, "+++") {
		return true
	}

	// Markdown headers (starts with #)
	if firstChar == '#' {
		return true
	}

	// Fast path: check first word for common patterns
	// Avoid full Contains for better performance
	if firstChar == 'd' && strings.HasPrefix(trimmed, "diff --git") {
		return true
	}

	// Use SIMD-optimized contains for remaining patterns
	if simd.ContainsAny(trimmed, []string{
		"test result:", "running ", " tests",
		"Compiling ", "Building ", "Finished ",
		"error[", "error: ",
	}) {
		return true
	}

	return false
}

// scoreSections assigns importance scores to each section
// Optimized: Reuses cached semantic filter, uses SIMD for scoring
func (f *HierarchicalFilter) scoreSections(sections []section) []section {
	if !f.useSemanticScoring {
		// Uniform scoring without semantic analysis
		for i := range sections {
			sections[i].score = 0.5
		}
		return sections
	}

	// Use cached semantic filter (optimization: avoid re-allocation)
	sf := f.semanticFilter
	if sf == nil {
		sf = NewSemanticFilter()
		f.semanticFilter = sf
	}

	// Score sections with early exit for large inputs
	numSections := len(sections)
	for i := range sections {
		// For large numbers of sections, use fast scoring
		if numSections > 100 {
			sections[i].score = f.fastSectionScore(sections[i])
		} else {
			sections[i].score = f.calculateSectionScore(sections[i], sf)
		}

		// Only generate summaries for high-scoring sections
		if sections[i].score >= 0.4 {
			sections[i].summary = f.generateSectionSummary(sections[i])
		}
	}

	return sections
}

// fastSectionScore provides a fast approximation for large inputs
// Uses SIMD-only matching without semantic filter overhead
func (f *HierarchicalFilter) fastSectionScore(s section) float64 {
	content := s.content
	score := 0.0

	// SIMD-optimized keyword counting
	highMatches := 0
	for _, kw := range highKeywords {
		if simd.ContainsWord(content, kw) {
			highMatches++
		}
	}

	// Early exit if no matches
	if highMatches == 0 {
		// Quick check for file references
		if simd.ContainsAny(content, fileExtensions) {
			return 0.4
		}
		return 0.2
	}

	score += float64(highMatches) * 0.15

	// File references
	if simd.ContainsAny(content, fileExtensions) {
		score += 0.3
	}

	// Clamp
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateSectionScore computes importance for a section
// Optimized: Uses pre-compiled keywords and SIMD operations
func (f *HierarchicalFilter) calculateSectionScore(s section, sf *SemanticFilter) float64 {
	score := 0.0
	content := s.content

	// Use SIMD-optimized case-insensitive matching
	// Count matches using fast byte scanning
	highMatches := 0
	mediumMatches := 0

	// Single pass through content for all keywords (O(n) instead of O(n*k))
	for _, kw := range highKeywords {
		if simd.ContainsWord(content, kw) {
			highMatches++
		}
	}

	for _, kw := range mediumKeywords {
		if simd.ContainsWord(content, kw) {
			mediumMatches++
		}
	}

	score += float64(highMatches) * 0.2
	score += float64(mediumMatches) * 0.1

	// File references (very important for debugging)
	// Use SIMD-optimized byte scanning
	if simd.ContainsAny(content, []string{".go:", ".rs:", ".py:", ".js:", ".ts:"}) {
		score += 0.3
	}

	// Stack traces - use SIMD word matching
	if simd.ContainsAny(content, []string{"at ", "Traceback", "stack trace"}) {
		score += 0.4
	}

	// Length penalty (longer sections are less dense)
	lineCount := s.endLine - s.startLine + 1
	if lineCount > 100 {
		score *= 0.8
	}

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// generateSectionSummary creates a one-line summary of a section
// Optimized: Early exit and limited line scanning
func (f *HierarchicalFilter) generateSectionSummary(s section) string {
	// Fast path: use first non-empty line for large sections
	content := s.content
	if len(content) > 5000 {
		// For large sections, just use first meaningful line
		idx := 0
		for idx < len(content) && content[idx] == '\n' {
			idx++
		}
		if idx >= len(content) {
			return "[section]"
		}
		end := idx
		for end < len(content) && content[end] != '\n' {
			end++
		}
		line := strings.TrimSpace(content[idx:end])
		if len(line) > 80 {
			return line[:77] + "..."
		}
		if line == "" {
			return "[section]"
		}
		return line
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "[empty section]"
	}

	// Find the most representative line (limit to first 20 lines)
	var bestLine string
	bestScore := -1.0
	maxLines := 20
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	for i := 0; i < maxLines; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}

		// Score this line
		score := f.scoreLineForSummary(trimmed)
		if score > bestScore {
			bestScore = score
			bestLine = trimmed
		}
	}

	if bestLine == "" {
		return "[section]"
	}

	// Truncate if needed
	if len(bestLine) > 80 {
		return bestLine[:77] + "..."
	}

	return bestLine
}

// scoreLineForSummary rates how good a line is as a summary
func (f *HierarchicalFilter) scoreLineForSummary(line string) float64 {
	lower := strings.ToLower(line)
	score := 0.0

	// Prefer lines with key information
	if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		score += 0.5
	}
	if strings.Contains(lower, "test") || strings.Contains(lower, "pass") {
		score += 0.3
	}

	// Prefer shorter lines
	if len(line) < 60 {
		score += 0.2
	}

	// Avoid pure symbols or numbers
	hasLetters := false
	for _, r := range line {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetters = true
			break
		}
	}
	if !hasLetters {
		score -= 0.5
	}

	return score
}

// buildHierarchicalOutput constructs the compressed output
func (f *HierarchicalFilter) buildHierarchicalOutput(sections []section, mode Mode, totalLines int) string {
	var output []string

	// Add header showing compression stats
	output = append(output, f.formatHeader(totalLines, len(sections)))

	// Determine thresholds based on mode
	highThreshold, midThreshold := f.getThresholds(mode)

	for _, s := range sections {
		switch {
		case s.score >= highThreshold:
			// Keep full content
			output = append(output, s.content)

		case s.score >= midThreshold:
			// Keep summary with line range
			output = append(output, f.formatSummarySection(s))

		default:
			// Skip low-importance sections
			// Optionally: add a one-liner indicating skipped content
		}
	}

	return strings.Join(output, "\n")
}

// formatHeader creates the compression header
func (f *HierarchicalFilter) formatHeader(totalLines, sectionCount int) string {
	return "\n[Hierarchical Summary: " + strconv.Itoa(totalLines) + " lines → " + strconv.Itoa(sectionCount) + " sections]\n"
}

// formatSummarySection formats a section as a summary
func (f *HierarchicalFilter) formatSummarySection(s section) string {
	lineCount := s.endLine - s.startLine + 1
	return "\n├─ [L" + strconv.Itoa(s.startLine+1) + "-" + strconv.Itoa(s.endLine+1) + "] " + s.summary + " (" + strconv.Itoa(lineCount) + " lines, score: " + f.formatScore(s.score) + ")\n"
}

// formatScore formats a score to 2 decimal places
func (f *HierarchicalFilter) formatScore(score float64) string {
	intPart := int(score * 100)
	return strconv.Itoa(intPart/100) + "." + fmt.Sprintf("%02d", intPart%100)
}

// getThresholds returns score thresholds for different compression levels
func (f *HierarchicalFilter) getThresholds(mode Mode) (high, mid float64) {
	switch mode {
	case ModeAggressive:
		return 0.7, 0.4 // Only high-value content
	case ModeMinimal:
		return 0.5, 0.25
	default:
		return 0.6, 0.3
	}
}

// SetLineThreshold configures the line threshold for hierarchical compression
func (f *HierarchicalFilter) SetLineThreshold(threshold int) {
	f.lineThreshold = threshold
}

// SetMaxDepth configures the maximum summarization depth
func (f *HierarchicalFilter) SetMaxDepth(depth int) {
	f.maxDepth = depth
}
