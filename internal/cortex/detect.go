// Package cortex provides content-aware gate system (Claw Compactor's Cortex detection).
package cortex

import (
	"bufio"
	"math"
	"path/filepath"
	"regexp"
	"strings"
)

// ContentType represents the detected content type.
type ContentType int

const (
	Unknown ContentType = iota
	SourceCode
	BuildLog
	TestOutput
	StructuredData
	NaturalLanguage
	BinaryData
)

func (c ContentType) String() string {
	switch c {
	case SourceCode:
		return "source_code"
	case BuildLog:
		return "build_log"
	case TestOutput:
		return "test_output"
	case StructuredData:
		return "structured_data"
	case NaturalLanguage:
		return "natural_language"
	case BinaryData:
		return "binary_data"
	default:
		return "unknown"
	}
}

// Language represents the detected programming language.
type Language int

const (
	LangUnknown Language = iota
	LangGo
	LangRust
	LangPython
	LangJavaScript
	LangTypeScript
	LangJava
	LangC
	LangCpp
	LangRuby
	LangShell
	LangSQL
	LangMarkdown
	LangJSON
	LangYAML
	LangXML
)

func (l Language) String() string {
	switch l {
	case LangGo:
		return "go"
	case LangRust:
		return "rust"
	case LangPython:
		return "python"
	case LangJavaScript:
		return "javascript"
	case LangTypeScript:
		return "typescript"
	case LangJava:
		return "java"
	case LangC:
		return "c"
	case LangCpp:
		return "cpp"
	case LangRuby:
		return "ruby"
	case LangShell:
		return "shell"
	case LangSQL:
		return "sql"
	case LangMarkdown:
		return "markdown"
	case LangJSON:
		return "json"
	case LangYAML:
		return "yaml"
	case LangXML:
		return "xml"
	default:
		return "unknown"
	}
}

// DetectionResult contains content detection results.
type DetectionResult struct {
	ContentType ContentType
	Language    Language
	Confidence  float64
	Features    map[string]bool
	Stats       ContentStats
}

// ContentStats provides statistics about the content.
type ContentStats struct {
	TotalLines    int
	TotalChars    int
	NonEmptyLines int
	CommentLines  int
	CodeLines     int
	BlankLines    int
	AvgLineLength float64
	HasAnsiCodes  bool
	HasUnicode    bool
	Entropy       float64
}

// Detector provides content type detection.
type Detector struct {
	patterns     map[ContentType][]*regexp.Regexp
	langPatterns map[Language][]*regexp.Regexp
}

// NewDetector creates a new content detector.
func NewDetector() *Detector {
	d := &Detector{
		patterns:     make(map[ContentType][]*regexp.Regexp),
		langPatterns: make(map[Language][]*regexp.Regexp),
	}
	d.initPatterns()
	return d
}

// Detect analyzes content and returns detection results.
func (d *Detector) Detect(content string) DetectionResult {
	result := DetectionResult{
		ContentType: Unknown,
		Language:    LangUnknown,
		Confidence:  0.0,
		Features:    make(map[string]bool),
		Stats:       d.analyzeStats(content),
	}

	// Detect content type
	result.ContentType = d.detectContentType(content, result.Stats)

	// Detect language if source code
	if result.ContentType == SourceCode || result.ContentType == StructuredData {
		result.Language = d.detectLanguage(content)
	}

	// Calculate confidence
	result.Confidence = d.calculateConfidence(content, result)

	return result
}

// DetectFile analyzes a file and returns detection results.
func (d *Detector) DetectFile(path string, content string) DetectionResult {
	result := d.Detect(content)

	// Use file extension to improve language detection
	if result.Language == LangUnknown {
		result.Language = detectLanguageFromExt(path)
	}

	return result
}

func (d *Detector) initPatterns() {
	// Build log patterns
	d.patterns[BuildLog] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*(?:error|warning|info|debug)\s*[:\[]`),
		regexp.MustCompile(`(?i)(?:building|compiling|linking|generating)`),
		regexp.MustCompile(`(?i)(?:\d+\s*errors?|\d+\s*warnings?)`),
		regexp.MustCompile(`(?i)(?:\[.*\].*(?:info|error|warn|debug))`),
		regexp.MustCompile(`(?i)\[(?:ERROR|WARN|INFO|DEBUG)\]`),
	}

	// Test output patterns
	d.patterns[TestOutput] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*(?:pass|fail|skip|ok|run)\s+\w+`),
		regexp.MustCompile(`(?i)(?:===\s*(?:RUN|PASS|FAIL)|---\s*(?:PASS|FAIL|SKIP))`),
		regexp.MustCompile(`(?i)(?:test\s+passed|test\s+failed|tests?\s+complete)`),
		regexp.MustCompile(`(?i)(?:coverage[:\s]*\d+)`),
	}

	// Structured data patterns
	d.patterns[StructuredData] = []*regexp.Regexp{
		regexp.MustCompile(`^\s*[\{\[]`),
		regexp.MustCompile(`(?i)(?:\"[^\"]+\":\s*(?:\"|\d+|\[|\{))`),
		regexp.MustCompile(`^\s*---\s*$`),
		regexp.MustCompile(`^\s*\w+:\s*\w+`),
	}

	// Natural language patterns
	d.patterns[NaturalLanguage] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(the|a|an|is|are|was|were|be|been|being|have|has|had|do|does|did|will|would|could|should)\b`),
		regexp.MustCompile(`[.!?]\s+[A-Z]`),
	}

	// Language patterns
	d.langPatterns[LangGo] = []*regexp.Regexp{
		regexp.MustCompile(`\bpackage\s+\w+`),
		regexp.MustCompile(`\bfunc\s+\w+\s*\(`),
		regexp.MustCompile(`\bimport\s+\(`),
	}

	d.langPatterns[LangRust] = []*regexp.Regexp{
		regexp.MustCompile(`\bfn\s+\w+\s*\(`),
		regexp.MustCompile(`\buse\s+\w+::`),
		regexp.MustCompile(`\blet\s+mut\s+`),
		regexp.MustCompile(`\bimpl\s+\w+`),
	}

	d.langPatterns[LangPython] = []*regexp.Regexp{
		regexp.MustCompile(`\bdef\s+\w+\s*\(`),
		regexp.MustCompile(`\bimport\s+\w+`),
		regexp.MustCompile(`\bfrom\s+\w+\s+import`),
		regexp.MustCompile(`\bclass\s+\w+\s*[\(:]`),
	}

	d.langPatterns[LangJavaScript] = []*regexp.Regexp{
		regexp.MustCompile(`\bconst\s+\w+\s*=`),
		regexp.MustCompile(`\blet\s+\w+\s*=`),
		regexp.MustCompile(`\bvar\s+\w+\s*=`),
		regexp.MustCompile(`\bfunction\s+\w*\s*\(`),
		regexp.MustCompile(`=>\s*\{`),
	}

	d.langPatterns[LangTypeScript] = []*regexp.Regexp{
		regexp.MustCompile(`:\s*(?:string|number|boolean|any|void|interface|type)\b`),
		regexp.MustCompile(`\binterface\s+\w+\s*\{`),
		regexp.MustCompile(`\btype\s+\w+\s*=`),
	}
}

func (d *Detector) detectContentType(content string, stats ContentStats) ContentType {
	scores := make(map[ContentType]int)

	// Check patterns
	for ct, patterns := range d.patterns {
		for _, re := range patterns {
			if re.MatchString(content) {
				scores[ct] += 10
			}
		}
	}

	// Additional heuristics
	if stats.HasAnsiCodes {
		scores[BuildLog] += 5
		scores[TestOutput] += 5
	}

	if stats.CommentLines > 0 && stats.CodeLines > 0 {
		scores[SourceCode] += 15
	}

	// Find highest score
	var bestType ContentType
	bestScore := 0
	for ct, score := range scores {
		if score > bestScore {
			bestScore = score
			bestType = ct
		}
	}

	// If no strong match, check for source code characteristics
	if bestScore < 20 {
		if looksLikeSourceCode(content) {
			return SourceCode
		}
		if stats.TotalLines > 5 && bestScore == 0 {
			return NaturalLanguage
		}
	}

	return bestType
}

func (d *Detector) detectLanguage(content string) Language {
	// Check for JSON first (simple heuristic)
	trimmed := strings.TrimSpace(content)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		// Additional check for JSON structure
		if strings.Contains(content, "\"") && strings.Contains(content, ":") {
			return LangJSON
		}
	}

	scores := make(map[Language]int)

	for lang, patterns := range d.langPatterns {
		for _, re := range patterns {
			matches := re.FindAllString(content, -1)
			scores[lang] += len(matches) * 10
		}
	}

	// Find highest score
	var bestLang Language
	bestScore := 0
	for lang, score := range scores {
		if score > bestScore {
			bestScore = score
			bestLang = lang
		}
	}

	if bestScore >= 20 {
		return bestLang
	}

	return LangUnknown
}

func (d *Detector) analyzeStats(content string) ContentStats {
	stats := ContentStats{
		TotalChars: len(content),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var totalLineLength int
	for scanner.Scan() {
		line := scanner.Text()
		stats.TotalLines++

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			stats.BlankLines++
		} else {
			stats.NonEmptyLines++
			totalLineLength += len(line)

			if looksLikeComment(trimmed) {
				stats.CommentLines++
			} else {
				stats.CodeLines++
			}
		}

		if !stats.HasAnsiCodes && containsAnsi(line) {
			stats.HasAnsiCodes = true
		}
		if !stats.HasUnicode && containsUnicode(line) {
			stats.HasUnicode = true
		}
	}

	if stats.NonEmptyLines > 0 {
		stats.AvgLineLength = float64(totalLineLength) / float64(stats.NonEmptyLines)
	}

	stats.Entropy = calculateEntropy(content)

	return stats
}

func (d *Detector) calculateConfidence(content string, result DetectionResult) float64 {
	confidence := 0.5

	// Higher confidence for more content
	if result.Stats.TotalLines > 100 {
		confidence += 0.1
	}

	// Higher confidence with clear indicators
	if result.ContentType != Unknown {
		confidence += 0.2
	}

	if result.Language != LangUnknown {
		confidence += 0.2
	}

	// Lower confidence for mixed content
	if result.Stats.Entropy > 5.0 {
		confidence -= 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

func looksLikeSourceCode(content string) bool {
	indicators := []string{
		"func ", "def ", "class ", "import ", "package ", "#include",
		"public ", "private ", "var ", "let ", "const ", "function",
	}

	contentLower := strings.ToLower(content)
	count := 0
	for _, ind := range indicators {
		if strings.Contains(contentLower, ind) {
			count++
		}
	}
	return count >= 2
}

func looksLikeComment(line string) bool {
	return strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "/*") ||
		strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "<!--")
}

func containsAnsi(s string) bool {
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

func containsUnicode(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[byte]int)
	for i := 0; i < len(s); i++ {
		freq[s[i]]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

func detectLanguageFromExt(path string) Language {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return LangGo
	case ".rs":
		return LangRust
	case ".py":
		return LangPython
	case ".js":
		return LangJavaScript
	case ".ts", ".tsx":
		return LangTypeScript
	case ".java":
		return LangJava
	case ".c":
		return LangC
	case ".cpp", ".cc", ".hpp":
		return LangCpp
	case ".rb":
		return LangRuby
	case ".sh", ".bash":
		return LangShell
	case ".sql":
		return LangSQL
	case ".md":
		return LangMarkdown
	case ".json":
		return LangJSON
	case ".yaml", ".yml":
		return LangYAML
	case ".xml":
		return LangXML
	default:
		return LangUnknown
	}
}
