// Package core provides final compression layers (21-31).
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Layer 21: Semantic Similarity Filter
type SemanticSimilarityLayer struct {
	embeddings map[string][]float64
	threshold  float64
	mu         sync.RWMutex
}

// NewSemanticSimilarityLayer creates a semantic similarity layer.
func NewSemanticSimilarityLayer() *SemanticSimilarityLayer {
	return &SemanticSimilarityLayer{
		embeddings: make(map[string][]float64),
		threshold:  0.85,
	}
}

func (l *SemanticSimilarityLayer) Name() string { return "semantic_similarity" }

func (l *SemanticSimilarityLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *SemanticSimilarityLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var unique []string
	saved := 0

	for _, line := range lines {
		if len(line) < 10 {
			unique = append(unique, line)
			continue
		}

		emb := l.embed(line)
		isDuplicate := false

		l.mu.RLock()
		for _, existing := range l.embeddings {
			if cosineSimilarity(emb, existing) > l.threshold {
				isDuplicate = true
				break
			}
		}
		l.mu.RUnlock()

		if !isDuplicate {
			l.mu.Lock()
			l.embeddings[line] = emb
			l.mu.Unlock()
			unique = append(unique, line)
		} else {
			saved += len(line)
		}
	}

	return strings.Join(unique, "\n"), saved
}

func (l *SemanticSimilarityLayer) embed(text string) []float64 {
	// Simple character frequency embedding
	vec := make([]float64, 256)
	for i := range text {
		if i < len(text) {
			vec[text[i]]++
		}
	}
	// Normalize
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Layer 22: Code Fold Detection
type CodeFoldLayer22 struct {
	minFoldLines int
}

func NewCodeFoldLayer22() *CodeFoldLayer22 {
	return &CodeFoldLayer22{minFoldLines: 5}
}

func (l *CodeFoldLayer22) Name() string { return "code_fold_22" }

func (l *CodeFoldLayer22) ShouldApply(contentType string) bool {
	return strings.Contains(contentType, "code")
}

func (l *CodeFoldLayer22) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	i := 0
	saved := 0

	for i < len(lines) {
		line := lines[i]
		if l.isFoldStart(line) {
			foldStart := i
			nesting := 1
			i++

			for i < len(lines) && nesting > 0 {
				if l.isFoldStart(lines[i]) {
					nesting++
				} else if l.isFoldEnd(lines[i]) {
					nesting--
				}
				i++
			}

			foldEnd := i - 1
			foldLines := foldEnd - foldStart

			if foldLines >= l.minFoldLines {
				result = append(result, lines[foldStart])
				result = append(result, fmt.Sprintf("    // ... %d lines folded ...", foldLines-2))
				result = append(result, lines[foldEnd])
				saved += foldLines * len("    ")
			} else {
				for j := foldStart; j <= foldEnd; j++ {
					result = append(result, lines[j])
				}
			}
		} else {
			result = append(result, line)
			i++
		}
	}

	return strings.Join(result, "\n"), saved
}

func (l *CodeFoldLayer22) isFoldStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasSuffix(trimmed, "{") && !strings.HasPrefix(trimmed, "//")
}

func (l *CodeFoldLayer22) isFoldEnd(line string) bool {
	return strings.TrimSpace(line) == "}"
}

// Layer 23: Import/Dependency Collapse
type ImportCollapseLayer23 struct{}

func NewImportCollapseLayer23() *ImportCollapseLayer23 {
	return &ImportCollapseLayer23{}
}

func (l *ImportCollapseLayer23) Name() string { return "import_collapse_23" }

func (l *ImportCollapseLayer23) ShouldApply(contentType string) bool {
	return true
}

func (l *ImportCollapseLayer23) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	inImport := false
	var imports []string
	saved := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Go/Python/Java style imports
		if strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "from ") ||
			strings.HasPrefix(trimmed, "using ") ||
			strings.HasPrefix(trimmed, "require(") ||
			strings.HasPrefix(trimmed, "include ") {
			if !inImport {
				inImport = true
			}
			imports = append(imports, line)
			continue
		}

		if inImport && len(imports) > 0 {
			// Collapse accumulated imports
			if len(imports) > 3 {
				result = append(result, fmt.Sprintf("// %d imports collapsed", len(imports)))
				saved += len(imports) * 30
			} else {
				result = append(result, imports...)
			}
			imports = nil
			inImport = false
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n"), saved
}

// Layer 24: Comment Removal Heuristics
type CommentRemovalLayer24 struct {
	preserveDoc bool
}

func NewCommentRemovalLayer24() *CommentRemovalLayer24 {
	return &CommentRemovalLayer24{preserveDoc: true}
}

func (l *CommentRemovalLayer24) Name() string { return "comment_removal_24" }

func (l *CommentRemovalLayer24) ShouldApply(contentType string) bool {
	return strings.Contains(contentType, "code")
}

func (l *CommentRemovalLayer24) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	var result []string
	saved := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for different comment styles
		isComment := strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "--") ||
			strings.HasPrefix(trimmed, "/*")

		if !isComment {
			result = append(result, line)
			continue
		}

		// Preserve documentation comments
		if l.preserveDoc {
			if strings.HasPrefix(trimmed, "///") ||
				strings.HasPrefix(trimmed, "//go:") ||
				strings.HasPrefix(trimmed, "/**") ||
				strings.Contains(trimmed, "TODO") ||
				strings.Contains(trimmed, "FIXME") ||
				strings.Contains(trimmed, "NOTE") {
				result = append(result, line)
				continue
			}
		}

		saved += len(line) + 1
	}

	return strings.Join(result, "\n"), saved
}

// Layer 25: Whitespace Normalization
type WhitespaceLayer struct{}

func NewWhitespaceLayer() *WhitespaceLayer {
	return &WhitespaceLayer{}
}

func (l *WhitespaceLayer) Name() string { return "whitespace" }

func (l *WhitespaceLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *WhitespaceLayer) Apply(content string) (string, int) {
	originalLen := len(content)

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Convert tabs to spaces
	content = strings.ReplaceAll(content, "\t", "    ")

	// Remove trailing whitespace
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Remove consecutive blank lines
	var result []string
	prevBlank := false
	for _, line := range lines {
		isBlank := len(strings.TrimSpace(line)) == 0
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}

	content = strings.Join(result, "\n")
	saved := originalLen - len(content)
	return content, saved
}

// Layer 26: String Interning
type StringInternLayer struct {
	strings map[string]string
	mu      sync.RWMutex
}

func NewStringInternLayer() *StringInternLayer {
	return &StringInternLayer{
		strings: make(map[string]string),
	}
}

func (l *StringInternLayer) Name() string { return "string_intern" }

func (l *StringInternLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *StringInternLayer) Apply(content string) (string, int) {
	// Find string literals
	re := regexp.MustCompile(`"([^"]{10,})"`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	saved := 0

	// Replace from end to start to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		if len(m) < 4 {
			continue
		}
		str := content[m[2]:m[3]]

		l.mu.RLock()
		interned, ok := l.strings[str]
		l.mu.RUnlock()

		if !ok {
			l.mu.Lock()
			l.strings[str] = str
			l.mu.Unlock()
			interned = str
		} else if interned != str {
			// String exists, replace with reference
			ref := fmt.Sprintf("[str:%x]", hashString(str)[:8])
			content = content[:m[0]] + ref + content[m[1]:]
			saved += len(str) - len(ref)
		}
	}

	return content, saved
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Layer 27: Number Precision Reduction
type PrecisionLayer struct {
	maxDecimals int
}

func NewPrecisionLayer() *PrecisionLayer {
	return &PrecisionLayer{maxDecimals: 4}
}

func (l *PrecisionLayer) Name() string { return "precision" }

func (l *PrecisionLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *PrecisionLayer) Apply(content string) (string, int) {
	// Reduce precision of floating point numbers
	re := regexp.MustCompile(`\b\d+\.\d{5,}\b`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		var f float64
		fmt.Sscanf(match, "%f", &f)
		format := fmt.Sprintf("%%.%df", l.maxDecimals)
		return fmt.Sprintf(format, f)
	})
	return content, 0
}

// Layer 28: URL Shortening
type URLLayer struct {
	urls map[string]string
	seq  int
	mu   sync.Mutex
}

func NewURLLayer() *URLLayer {
	return &URLLayer{
		urls: make(map[string]string),
	}
}

func (l *URLLayer) Name() string { return "url_shorten" }

func (l *URLLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *URLLayer) Apply(content string) (string, int) {
	re := regexp.MustCompile(`https?://[^\s<>"{}|\^` + "`" + `\[\]]+`)
	matches := re.FindAllStringIndex(content, -1)
	saved := 0

	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		url := content[m[0]:m[1]]

		l.mu.Lock()
		l.seq++
		ref := fmt.Sprintf("[url:%d]", l.seq)
		l.urls[ref] = url
		l.mu.Unlock()

		content = content[:m[0]] + ref + content[m[1]:]
		saved += len(url) - len(ref)
	}

	return content, saved
}

// Layer 29: UUID/Hash Truncation
type TruncateLayer struct{}

func NewTruncateLayer() *TruncateLayer {
	return &TruncateLayer{}
}

func (l *TruncateLayer) Name() string { return "truncate_ids" }

func (l *TruncateLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *TruncateLayer) Apply(content string) (string, int) {
	// UUID pattern
	uuidRe := regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	content = uuidRe.ReplaceAllString(content, "[UUID]")

	// SHA-256 hashes
	hashRe := regexp.MustCompile(`\b[0-9a-f]{64}\b`)
	content = hashRe.ReplaceAllString(content, "[hash]")

	// Git commits
	commitRe := regexp.MustCompile(`\b[0-9a-f]{40}\b`)
	content = commitRe.ReplaceAllString(content, "[commit]")

	return content, 0
}

// Layer 30: Repetition Detection
type RepetitionLayer struct {
	threshold int
}

func NewRepetitionLayer() *RepetitionLayer {
	return &RepetitionLayer{threshold: 3}
}

func (l *RepetitionLayer) Name() string { return "repetition" }

func (l *RepetitionLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *RepetitionLayer) Apply(content string) (string, int) {
	lines := strings.Split(content, "\n")
	type run struct {
		line  string
		start int
		count int
	}

	var runs []run
	current := run{start: 0}

	for i, line := range lines {
		if i > 0 && line == lines[i-1] {
			current.count++
		} else {
			if current.count >= l.threshold {
				runs = append(runs, current)
			}
			current = run{line: line, start: i, count: 1}
		}
	}

	if current.count >= l.threshold {
		runs = append(runs, current)
	}

	// Build result
	var result []string
	lastEnd := 0
	saved := 0

	for _, r := range runs {
		// Add lines before run
		result = append(result, lines[lastEnd:r.start]...)
		// Add marker
		result = append(result, fmt.Sprintf("// repeated %d times:", r.count))
		result = append(result, r.line)
		saved += (r.count - 1) * len(r.line)
		lastEnd = r.start + r.count
	}

	result = append(result, lines[lastEnd:]...)
	return strings.Join(result, "\n"), saved
}

// Layer 31: Final Compression Pass
type FinalPassLayer struct{}

func NewFinalPassLayer() *FinalPassLayer {
	return &FinalPassLayer{}
}

func (l *FinalPassLayer) Name() string { return "final_pass" }

func (l *FinalPassLayer) ShouldApply(contentType string) bool {
	return true
}

func (l *FinalPassLayer) Apply(content string) (string, int) {
	// Final cleanup and optimization
	originalLen := len(content)

	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Ensure single newline at end
	content = strings.TrimRight(content, "\n") + "\n"

	// Compress consecutive spaces
	re := regexp.MustCompile(` +`)
	content = re.ReplaceAllString(content, " ")

	// Sort imports if present (Go/Python style)
	content = l.sortImports(content)

	saved := originalLen - len(content)
	return content, saved
}

func (l *FinalPassLayer) sortImports(content string) string {
	lines := strings.Split(content, "\n")

	var imports []string
	var other []string
	inImport := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import") || strings.HasPrefix(trimmed, "from") {
			inImport = true
			imports = append(imports, line)
		} else if inImport && (len(trimmed) == 0 || !strings.HasPrefix(trimmed, ".")) {
			inImport = false
			other = append(other, line)
		} else if inImport {
			imports = append(imports, line)
		} else {
			other = append(other, line)
		}
	}

	if len(imports) > 0 {
		sort.Strings(imports)
		var result []string
		result = append(result, other[:max(0, len(other)-len(lines)+len(imports))]...)
		result = append(result, imports...)
		result = append(result, other[len(other)-len(lines)+len(imports):]...)
		return strings.Join(result, "\n")
	}

	return content
}

// RegisterFinalLayers registers layers 21-31.
func RegisterFinalLayers(registry *LayerRegistry) {
	registry.Register(NewSemanticSimilarityLayer())
	registry.Register(NewCodeFoldLayer22())
	registry.Register(NewImportCollapseLayer23())
	registry.Register(NewCommentRemovalLayer24())
	registry.Register(NewWhitespaceLayer())
	registry.Register(NewStringInternLayer())
	registry.Register(NewPrecisionLayer())
	registry.Register(NewURLLayer())
	registry.Register(NewTruncateLayer())
	registry.Register(NewRepetitionLayer())
	registry.Register(NewFinalPassLayer())
}
