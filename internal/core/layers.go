// Package core provides compression layer implementations.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Layer represents a compression layer.
type Layer interface {
	Name() string
	Apply(content string) (string, int)
	ShouldApply(contentType string) bool
}

// LayerRegistry manages compression layers.
type LayerRegistry struct {
	mu     sync.RWMutex
	layers []Layer
}

// NewLayerRegistry creates a new layer registry.
func NewLayerRegistry() *LayerRegistry {
	return &LayerRegistry{
		layers: make([]Layer, 0),
	}
}

// Register registers a layer.
func (r *LayerRegistry) Register(layer Layer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.layers = append(r.layers, layer)
}

// Apply applies all registered layers.
func (r *LayerRegistry) Apply(content string, contentType string) (string, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalSaved := 0
	result := content

	for _, layer := range r.layers {
		if layer.ShouldApply(contentType) {
			newResult, saved := layer.Apply(result)
			result = newResult
			totalSaved += saved
		}
	}

	return result, totalSaved
}

// GetLayers returns all registered layers.
func (r *LayerRegistry) GetLayers() []Layer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Layer, len(r.layers))
	copy(result, r.layers)
	return result
}

// Layer 5: Contrastive Learning - Identify semantically similar content
type ContrastiveLayer struct {
	embeddings map[string][]float64
	threshold  float64
}

// NewContrastiveLayer creates a new contrastive learning layer.
func NewContrastiveLayer() *ContrastiveLayer {
	return &ContrastiveLayer{
		embeddings: make(map[string][]float64),
		threshold:  0.85, // Similarity threshold
	}
}

func (l *ContrastiveLayer) Name() string { return "contrastive" }

func (l *ContrastiveLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *ContrastiveLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	uniqueLines := make([]string, 0, len(lines))
	seenHashes := make(map[string]bool)
	saved := 0

	for _, line := range lines {
		if len(line) < 10 {
			uniqueLines = append(uniqueLines, line)
			continue
		}

		hash := l.hashLine(line)
		if seenHashes[hash] {
			saved += len(line)
			continue
		}

		// Check similarity with existing lines
		isSimilar := false
		for existingHash := range seenHashes {
			if l.similarity(hash, existingHash) > l.threshold {
				isSimilar = true
				saved += len(line)
				break
			}
		}

		if !isSimilar {
			seenHashes[hash] = true
			uniqueLines = append(uniqueLines, line)
		}
	}

	return strings.Join(uniqueLines, "\n"), saved
}

func (l *ContrastiveLayer) hashLine(line string) string {
	h := sha256.Sum256([]byte(line))
	return hex.EncodeToString(h[:8])
}

func (l *ContrastiveLayer) similarity(a, b string) float64 {
	// Simplified similarity based on character overlap
	if a == b {
		return 1.0
	}
	return 0.0
}

// Layer 6: N-gram Deduplication
type NgramDeduplicationLayer struct {
	n         int
	threshold float64
}

// NewNgramDeduplicationLayer creates a new n-gram deduplication layer.
func NewNgramDeduplicationLayer() *NgramDeduplicationLayer {
	return &NgramDeduplicationLayer{
		n:         5,
		threshold: 0.8,
	}
}

func (l *NgramDeduplicationLayer) Name() string { return "ngram_dedup" }

func (l *NgramDeduplicationLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *NgramDeduplicationLayer) Apply(content string) (string, int) {
	paragraphs := strings.Split(content, "\n\n")
	uniqueParagraphs := make([]string, 0, len(paragraphs))
	ngramSets := make([]map[string]bool, 0)
	saved := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if len(para) == 0 {
			continue
		}

		ngrams := l.getNgrams(para)
		isDuplicate := false

		for _, existingSet := range ngramSets {
			if l.jaccardSimilarity(ngrams, existingSet) > l.threshold {
				isDuplicate = true
				saved += len(para)
				break
			}
		}

		if !isDuplicate {
			uniqueParagraphs = append(uniqueParagraphs, para)
			ngramSets = append(ngramSets, ngrams)
		}
	}

	return strings.Join(uniqueParagraphs, "\n\n"), saved
}

func (l *NgramDeduplicationLayer) getNgrams(s string) map[string]bool {
	tokens := strings.Fields(s)
	ngrams := make(map[string]bool)

	if len(tokens) < l.n {
		ngrams[strings.Join(tokens, " ")] = true
		return ngrams
	}

	for i := 0; i <= len(tokens)-l.n; i++ {
		ngram := strings.Join(tokens[i:i+l.n], " ")
		ngrams[ngram] = true
	}

	return ngrams
}

func (l *NgramDeduplicationLayer) jaccardSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// Layer 7: Code Fold Detection
type CodeFoldLayer struct {
	minFoldSize int
}

// NewCodeFoldLayer creates a new code folding layer.
func NewCodeFoldLayer() *CodeFoldLayer {
	return &CodeFoldLayer{minFoldSize: 10}
}

func (l *CodeFoldLayer) Name() string { return "code_fold" }

func (l *CodeFoldLayer) ShouldApply(contentType string) bool {
	return strings.Contains(contentType, "code") ||
		strings.Contains(contentType, "go") ||
		strings.Contains(contentType, "javascript")
}

func (l *CodeFoldLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	inFold := false
	foldStart := 0
	saved := 0

	for i, line := range lines {
		indent := l.countIndent(line)

		if !inFold && l.isFoldStart(line) {
			inFold = true
			foldStart = i
			result = append(result, line)
		} else if inFold {
			if i-foldStart > l.minFoldSize && indent <= l.countIndent(lines[foldStart]) {
				// End of fold
				inFold = false
				saved += l.sumLength(lines[foldStart+1 : i])
				result = append(result, "    // ... (folded)")
				result = append(result, line)
			} else if i == len(lines)-1 {
				// End of file, keep fold
				result = append(result, lines[foldStart+1:]...)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n"), saved
}

func (l *CodeFoldLayer) countIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' || c == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

func (l *CodeFoldLayer) isFoldStart(line string) bool {
	// Detect function definitions, class definitions, etc.
	patterns := []string{
		"^\\s*func\\s+",
		"^\\s*class\\s+",
		"^\\s*if\\s+.*{$",
		"^\\s*for\\s+.*{$",
		"^\\s*while\\s+.*{$",
		"^\\s*switch\\s+.*{$",
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, line)
		if matched {
			return strings.HasSuffix(line, "{")
		}
	}
	return false
}

func (l *CodeFoldLayer) sumLength(lines []string) int {
	sum := 0
	for _, line := range lines {
		sum += len(line) + 1 // +1 for newline
	}
	return sum
}

// Layer 8: Import/Dependency Collapse
type ImportCollapseLayer struct{}

// NewImportCollapseLayer creates a new import collapse layer.
func NewImportCollapseLayer() *ImportCollapseLayer {
	return &ImportCollapseLayer{}
}

func (l *ImportCollapseLayer) Name() string { return "import_collapse" }

func (l *ImportCollapseLayer) ShouldApply(contentType string) bool {
	return strings.Contains(contentType, "go") ||
		strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "typescript")
}

func (l *ImportCollapseLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	inImportBlock := false
	imports := []string{}
	saved := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Go imports
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			collapsed := l.collapseGoImports(imports)
			result = append(result, collapsed)
			saved += l.sumLength(imports)
			imports = nil
			continue
		}
		if inImportBlock {
			imports = append(imports, line)
			continue
		}

		// Single line imports
		if strings.HasPrefix(trimmed, "import ") && !strings.Contains(trimmed, "(") {
			imports = append(imports, line)
			if len(imports) >= 3 {
				collapsed := l.collapseGoImports(imports)
				result = append(result, collapsed)
				saved += l.sumLength(imports) - len(collapsed)
				imports = nil
			}
			continue
		}

		result = append(result, line)
	}

	// Handle remaining imports
	if len(imports) > 0 {
		result = append(result, imports...)
	}

	return strings.Join(result, "\n"), saved
}

func (l *ImportCollapseLayer) collapseGoImports(imports []string) string {
	if len(imports) == 0 {
		return ""
	}
	if len(imports) == 1 {
		return strings.TrimSpace(imports[0])
	}

	// Extract import paths
	paths := make([]string, 0, len(imports))
	for _, imp := range imports {
		// Remove whitespace and quotes
		cleaned := strings.TrimSpace(imp)
		cleaned = strings.Trim(cleaned, `"`)
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			paths = append(paths, cleaned)
		}
	}

	// Group by prefix
	groups := make(map[string][]string)
	for _, path := range paths {
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			prefix := parts[0]
			groups[prefix] = append(groups[prefix], path)
		}
	}

	// Build collapsed form
	var result []string
	result = append(result, "import (")

	// Sort keys for consistent output
	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, prefix := range keys {
		if len(groups[prefix]) > 2 {
			result = append(result, "\t// "+prefix+"/... ("+
				string(rune(len(groups[prefix])+'0'))+" imports)")
		} else {
			for _, path := range groups[prefix] {
				result = append(result, "\t\""+path+"\"")
			}
		}
	}

	result = append(result, ")")
	return strings.Join(result, "\n")
}

func (l *ImportCollapseLayer) sumLength(lines []string) int {
	sum := 0
	for _, line := range lines {
		sum += len(line) + 1
	}
	return sum
}

// Layer 9: Comment Removal Heuristics
type CommentRemovalLayer struct {
	preserveDoc     bool
	preserveTODO    bool
	minContentRatio float64
}

// NewCommentRemovalLayer creates a new comment removal layer.
func NewCommentRemovalLayer() *CommentRemovalLayer {
	return &CommentRemovalLayer{
		preserveDoc:     true,
		preserveTODO:    true,
		minContentRatio: 0.3,
	}
}

func (l *CommentRemovalLayer) Name() string { return "comment_removal" }

func (l *CommentRemovalLayer) ShouldApply(contentType string) bool {
	return strings.Contains(contentType, "code")
}

func (l *CommentRemovalLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	saved := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if line is a comment
		isComment := strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*")

		if !isComment {
			result = append(result, line)
			continue
		}

		// Preserve important comments
		if l.preserveDoc && strings.HasPrefix(trimmed, "// ") &&
			(len(trimmed) < 4 || trimmed[3] >= 'A' && trimmed[3] <= 'Z') {
			result = append(result, line)
			continue
		}

		if l.preserveTODO && (strings.Contains(trimmed, "TODO") ||
			strings.Contains(trimmed, "FIXME") ||
			strings.Contains(trimmed, "NOTE") ||
			strings.Contains(trimmed, "WARNING")) {
			result = append(result, line)
			continue
		}

		// Remove the comment
		saved += len(line) + 1
	}

	return strings.Join(result, "\n"), saved
}

// Layer 10: Budget Enforcement
type BudgetLayer struct {
	maxTokens int
}

// NewBudgetLayer creates a new budget enforcement layer.
func NewBudgetLayer(maxTokens int) *BudgetLayer {
	return &BudgetLayer{maxTokens: maxTokens}
}

func (l *BudgetLayer) Name() string { return "budget" }

func (l *BudgetLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *BudgetLayer) Apply(content string) (string, int) {
	// Estimate tokens (rough approximation)
	tokens := len(content) / 4
	if tokens <= l.maxTokens {
		return content, 0
	}

	// Apply progressive truncation
	lines := strings.Split(content, "\n")
	var result []string
	currentTokens := 0
	saved := 0

	// Keep important lines first (definitions, errors)
	priorityLines := make([]string, 0)
	otherLines := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if l.isHighPriority(trimmed) {
			priorityLines = append(priorityLines, line)
		} else {
			otherLines = append(otherLines, line)
		}
	}

	// Add priority lines first
	for _, line := range priorityLines {
		lineTokens := len(line) / 4
		if currentTokens+lineTokens > l.maxTokens {
			saved += len(line) + 1
			continue
		}
		result = append(result, line)
		currentTokens += lineTokens
	}

	// Fill with other lines
	for _, line := range otherLines {
		lineTokens := len(line) / 4
		if currentTokens+lineTokens > l.maxTokens {
			saved += len(line) + 1
			continue
		}
		result = append(result, line)
		currentTokens += lineTokens
	}

	return strings.Join(result, "\n"), saved
}

func (l *BudgetLayer) isHighPriority(line string) bool {
	patterns := []string{
		"^func ",
		"^class ",
		"^struct ",
		"^interface ",
		"error",
		"Error",
		"^// ",
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, line)
		if matched {
			return true
		}
	}
	return false
}

// RegisterDefaultLayers registers all default layers.
func RegisterDefaultLayers(registry *LayerRegistry) {
	registry.Register(NewContrastiveLayer())
	registry.Register(NewNgramDeduplicationLayer())
	registry.Register(NewCodeFoldLayer())
	registry.Register(NewImportCollapseLayer())
	registry.Register(NewCommentRemovalLayer())
	registry.Register(NewBudgetLayer(4000))
}
