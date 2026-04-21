package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Mode represents the filtering mode.
type Mode string

const (
	ModeNone       Mode = "none"
	ModeMinimal    Mode = "minimal"
	ModeAggressive Mode = "aggressive"
)

// Language represents a programming language for filtering
type Language string

const (
	LangRust       Language = "rust"
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangGo         Language = "go"
	LangC          Language = "c"
	LangCpp        Language = "cpp"
	LangJava       Language = "java"
	LangRuby       Language = "ruby"
	LangShell      Language = "sh"
	LangSQL        Language = "sql"
	LangUnknown    Language = "unknown"
)

// Filter defines the interface for output filters.
type Filter interface {
	// Name returns the filter name.
	Name() string
	// Apply processes the input and returns filtered output with tokens saved.
	Apply(input string, mode Mode) (output string, tokensSaved int)
}

// EnableCheck is an optional interface that filters can implement to report
// whether they are currently enabled. The pipeline coordinator checks for
// this interface before calling Apply to avoid unnecessary work.
type EnableCheck interface {
	IsEnabled() bool
}

// ApplicabilityCheck is an optional interface that filters can implement to
// report whether they should run for a given input. The coordinator calls
// this before Apply to implement stage gates (skip cheap before expensive).
type ApplicabilityCheck interface {
	IsApplicable(input string) bool
}

// Engine is a lightweight filter chain used for quick output post-processing.
// Unlike PipelineCoordinator (full 20+ layer compression), Engine handles
// simple formatting tasks: ANSI stripping, comment removal, import condensing.
type Engine struct {
	filters        []Filter
	mode           Mode
	queryIntent    string // Query intent for query-aware compression
	promptTemplate string // Prompt template name for LLM summarization
}

// EngineConfig holds configuration for the filter engine
type EngineConfig struct {
	Mode             Mode
	QueryIntent      string
	LLMEnabled       bool
	MultiFileEnabled bool
	PromptTemplate   string // Template name for LLM summarization
}

// NewEngine creates a new filter engine with all registered filters.
func NewEngine(mode Mode) *Engine {
	return NewEngineWithQuery(mode, "")
}

// NewEngineWithQuery creates a new filter engine with query-aware compression.
func NewEngineWithQuery(mode Mode, queryIntent string) *Engine {
	return NewEngineWithConfig(EngineConfig{
		Mode:        mode,
		QueryIntent: queryIntent,
	})
}

// NewEngineWithConfig creates a filter engine with full configuration options.
func NewEngineWithConfig(cfg EngineConfig) *Engine {
	filters := []Filter{
		NewANSIFilter(),
		newCommentFilter(),
		NewImportFilter(),
	}

	// Add multi-file filter early if enabled (for cross-file optimization)
	if cfg.MultiFileEnabled {
		filters = append(filters, NewMultiFileFilter(MultiFileConfig{
			PreserveBoundaries: true,
		}))
	}

	// Add research-based semantic filters
	filters = append(filters,
		NewSemanticFilter(),      // Semantic pruning - research-based
		NewPositionAwareFilter(), // Position-bias optimization - reorders for LLM recall
		NewHierarchicalFilter(),  // Multi-level summarization for large outputs
	)

	// Add query-aware filter if intent is provided
	if cfg.QueryIntent != "" {
		filters = append(filters, NewQueryAwareFilter(cfg.QueryIntent))
	}

	if cfg.Mode == ModeAggressive {
		filters = append(filters, NewBodyFilter())
	}

	return &Engine{
		filters:        filters,
		mode:           cfg.Mode,
		queryIntent:    cfg.QueryIntent,
		promptTemplate: cfg.PromptTemplate,
	}
}

// Process applies all filters to the input.
func (e *Engine) Process(input string) (string, int) {
	output := input
	totalSaved := 0

	for _, filter := range e.filters {
		// Skip body filter in minimal mode
		if e.mode == ModeMinimal && filter.Name() == "body" {
			continue
		}

		filtered, saved := filter.Apply(output, e.mode)
		output = filtered
		totalSaved += saved
	}

	return output, totalSaved
}

// SetMode changes the filter mode.
func (e *Engine) SetMode(mode Mode) {
	e.mode = mode
}

// EstimateTokens provides a heuristic token count.
// Delegates to core.EstimateTokens for single source of truth (T22).
func EstimateTokens(text string) int {
	return core.EstimateTokens(text)
}

// IsCode checks if the output looks like source code.
func IsCode(output string) bool {
	codeIndicators := []string{
		"func ", "function ", "def ", "class ", "struct ",
		"import ", "package ", "use ", "require(",
		"pub fn", "pub struct", "pub async",
		"//", "/*", "#!", "package main",
	}

	for _, indicator := range codeIndicators {
		if strings.Contains(output, indicator) {
			return true
		}
	}

	return false
}

// languageIndicator defines a pattern and score for language detection
type languageIndicator struct {
	patterns []string
	score    int
}

// languageRules contains detection rules for each language
var languageRules = map[string][]languageIndicator{
	"go": {
		{[]string{"func "}, 10},
		{[]string{"package "}, 5},
		{[]string{"import (", "fmt.", " := "}, 5},
	},
	"rust": {
		{[]string{"fn ", "pub fn"}, 10},
		{[]string{"impl ", "trait ", "let mut"}, 5},
		{[]string{"&str", "Vec<", "Option<"}, 10},
	},
	"python": {
		{[]string{"self,", "self):"}, 10},
		{[]string{"from "}, 3},
	},
	"java": {
		{[]string{"public class ", "private ", "protected "}, 5},
		{[]string{"System.out.", "public static void main"}, 10},
	},
	"cpp": {
		{[]string{"std::", "cout", "cin"}, 15},
	},
	"c": {
		{[]string{"printf(", "malloc("}, 10},
	},
	"ruby": {
		{[]string{"puts ", "require '", "end\n"}, 5},
	},
	"shell": {
		{[]string{"chmod", "chown", "sudo "}, 3},
	},
}

// sqlKeywords for SQL detection
var sqlKeywords = []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "JOIN", "GROUP BY", "ORDER BY"}

// typescriptIndicators for TypeScript detection
var typescriptIndicators = []string{
	": string", ": number", ": boolean", ": void",
	": any", ": unknown", "interface ", "type ", "enum ", "namespace ",
}

// DetectLanguage attempts to detect the programming language from output
// using weighted scoring across multiple indicators.
func DetectLanguage(output string) string {
	scores := make(map[string]int)

	// Apply rule-based scoring
	for lang, rules := range languageRules {
		for _, rule := range rules {
			for _, pattern := range rule.patterns {
				if strings.Contains(output, pattern) {
					scores[lang] += rule.score
					break
				}
			}
		}
	}

	// Python-specific detection with curly brace penalty
	detectPython(output, scores)

	// SQL detection
	detectSQL(output, scores)

	// JavaScript/TypeScript detection
	detectJavaScriptFamily(output, scores)

	// C/C++ shared indicators
	if strings.Contains(output, "#include") {
		scores["c"] += 5
		scores["cpp"] += 5
	}

	// Ruby-specific detection
	detectRuby(output, scores)

	return selectBestLanguage(scores)
}

func detectPython(output string, scores map[string]int) {
	if strings.Contains(output, "def ") {
		scores["python"] += 5
	}
	if strings.Contains(output, "import ") {
		scores["python"] += 2
	}
	// Penalize Python if there are curly braces
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		scores["python"] -= 5
	}
}

func detectSQL(output string, scores map[string]int) {
	keywordCount := 0
	for _, kw := range sqlKeywords {
		if strings.Contains(output, kw) {
			keywordCount++
		}
	}
	if keywordCount > 0 {
		scores["sql"] = keywordCount * 15
	}
}

func detectJavaScriptFamily(output string, scores map[string]int) {
	if strings.Contains(output, "function ") || strings.Contains(output, "const ") || strings.Contains(output, "let ") {
		scores["javascript"] += 5
		if strings.Contains(output, "=>") {
			scores["javascript"] += 3
		}
	}

	for _, indicator := range typescriptIndicators {
		if strings.Contains(output, indicator) {
			scores["typescript"] += 15
			break
		}
	}
}

func detectRuby(output string, scores map[string]int) {
	if strings.Contains(output, "def ") && !strings.Contains(output, "self:") {
		scores["ruby"] += 3
	}
}

func selectBestLanguage(scores map[string]int) string {
	bestLang := "unknown"
	bestScore := 0
	for lang, score := range scores {
		if score > bestScore {
			bestScore = score
			bestLang = lang
		}
	}
	return bestLang
}

// estimateTokens is an alias for EstimateTokens (used internally by filter layers).
func estimateTokens(text string) int {
	return core.EstimateTokens(text)
}

// DetectLanguageFromInput detects language from input content.
// Delegates to DetectLanguage and wraps the result as a Language.
func DetectLanguageFromInput(input string) Language {
	return Language(DetectLanguage(input))
}
