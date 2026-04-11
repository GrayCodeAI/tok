package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Mode represents the filtering mode.
type Mode string

const (
	ModeNone       Mode = "none"
	ModeMinimal    Mode = "minimal"
	ModeAggressive Mode = "aggressive"
)

var allModes = []Mode{ModeNone, ModeMinimal, ModeAggressive}

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

// ModeNone = raw passthrough
func (e *Engine) ProcessWithLang(input string, lang string) (string, int) {
	// Language-specific processing can be added here
	return e.Process(input)
}

// DetectLanguageFromInput detects language from input content.
// Delegates to DetectLanguage and wraps the result as a Language.
func DetectLanguageFromInput(input string) Language {
	return Language(DetectLanguage(input))
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

// DetectLanguage attempts to detect the programming language from output
// using weighted scoring across multiple indicators.
func DetectLanguage(output string) string {
	scores := map[string]int{
		"go": 0, "python": 0, "rust": 0, "javascript": 0,
		"typescript": 0, "java": 0, "c": 0, "cpp": 0,
		"ruby": 0, "sql": 0, "shell": 0,
	}

	// Go indicators (high weight since "func" is distinctive)
	if strings.Contains(output, "func ") {
		scores["go"] = 10
	}
	if strings.Contains(output, "package ") {
		scores["go"] += 5
	}
	if strings.Contains(output, "import (") || strings.Contains(output, "fmt.") || strings.Contains(output, " := ") {
		scores["go"] += 5
	}

	// Rust indicators
	if strings.Contains(output, "fn ") || strings.Contains(output, "pub fn") {
		scores["rust"] += 10
	}
	if strings.Contains(output, "impl ") || strings.Contains(output, "trait ") || strings.Contains(output, "let mut") {
		scores["rust"] += 5
	}
	if strings.Contains(output, "&str") || strings.Contains(output, "Vec<") || strings.Contains(output, "Option<") {
		scores["rust"] += 10
	}

	// Python indicators
	if strings.Contains(output, "def ") {
		if strings.Contains(output, "self,") || strings.Contains(output, "self):") {
			scores["python"] += 10
		} else {
			scores["python"] += 5
		}
	}
	if strings.Contains(output, "import ") {
		if strings.Contains(output, "from ") {
			scores["python"] += 3
		}
	}
	// Penalize Python if there are curly braces (not Python style)
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		scores["python"] -= 5
	}

	// SQL indicators - even a single SQL keyword in a command-like context counts
	// Requires uppercase keywords to avoid false positives with English text
	sqlKeywords := 0
	for _, kw := range []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "JOIN", "GROUP BY", "ORDER BY"} {
		if strings.Contains(output, kw) {
			sqlKeywords++
		}
	}
	if sqlKeywords >= 1 {
		scores["sql"] = sqlKeywords * 15
	}

	// JavaScript/TypeScript
	if strings.Contains(output, "function ") || strings.Contains(output, "const ") || strings.Contains(output, "let ") {
		scores["javascript"] += 5
		if strings.Contains(output, "=>") {
			scores["javascript"] += 3
		}
	}
	// TypeScript type annotations (includes function return types like ": void", ": string", ": number")
	if strings.Contains(output, ": string") || strings.Contains(output, ": number") ||
		strings.Contains(output, ": boolean") || strings.Contains(output, ": void") ||
		strings.Contains(output, ": any") || strings.Contains(output, ": unknown") ||
		strings.Contains(output, "interface ") || strings.Contains(output, "type ") ||
		strings.Contains(output, "enum ") || strings.Contains(output, "namespace ") {
		scores["typescript"] += 15
	}

	// Java indicators
	if strings.Contains(output, "public class ") || strings.Contains(output, "private ") || strings.Contains(output, "protected ") {
		scores["java"] += 5
	}
	if strings.Contains(output, "System.out.") || strings.Contains(output, "public static void main") {
		scores["java"] += 10
	}

	// C/C++ indicators
	if strings.Contains(output, "#include") {
		scores["c"] += 5
		scores["cpp"] += 5
	}
	if strings.Contains(output, "std::") || strings.Contains(output, "cout") || strings.Contains(output, "cin") {
		scores["cpp"] += 15
	}
	if strings.Contains(output, "printf(") || strings.Contains(output, "malloc(") {
		scores["c"] += 10
	}

	// Ruby indicators
	if strings.Contains(output, "puts ") || strings.Contains(output, "require '") || strings.Contains(output, "end\n") {
		scores["ruby"] += 5
	}
	if strings.Contains(output, "def ") && !strings.Contains(output, "self:") {
		// Ruby uses "def" but without "self:" which Python uses
		scores["ruby"] += 3
	}

	// Shell indicators
	if strings.Contains(output, "chmod") || strings.Contains(output, "chown") || strings.Contains(output, "sudo ") {
		scores["shell"] += 3
	}

	// Find highest score
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

// estimateTokens is an alias for EstimateTokens (backward compatibility for internal use).
func estimateTokens(text string) int {
	return EstimateTokens(text)
}
