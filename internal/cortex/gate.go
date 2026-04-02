// Package cortex provides content-aware gate system.
package cortex

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// Gate determines if a layer should be applied to content.
type Gate interface {
	// ShouldApply returns true if the layer should be applied.
	ShouldApply(content string, detection DetectionResult) bool
	// Priority returns the gate priority (lower = higher priority).
	Priority() int
	// Name returns the gate name.
	Name() string
}

// LayerGate is a gate for a specific compression layer.
type LayerGate struct {
	name      string
	priority  int
	condition GateCondition
	layer     LayerProcessor
}

// GateCondition determines if content matches criteria.
type GateCondition interface {
	Match(content string, detection DetectionResult) bool
}

// LayerProcessor processes content for a layer.
type LayerProcessor interface {
	Process(content string) (string, int)
	Name() string
}

// NewLayerGate creates a new layer gate.
func NewLayerGate(name string, priority int, condition GateCondition, layer LayerProcessor) *LayerGate {
	return &LayerGate{
		name:      name,
		priority:  priority,
		condition: condition,
		layer:     layer,
	}
}

// ShouldApply implements Gate.
func (g *LayerGate) ShouldApply(content string, detection DetectionResult) bool {
	return g.condition.Match(content, detection)
}

// Priority implements Gate.
func (g *LayerGate) Priority() int {
	return g.priority
}

// Name implements Gate.
func (g *LayerGate) Name() string {
	return g.name
}

// Process applies the layer to content.
func (g *LayerGate) Process(content string) (string, int) {
	return g.layer.Process(content)
}

// ContentTypeCondition matches by content type.
type ContentTypeCondition struct {
	Types []ContentType
}

// Match implements GateCondition.
func (c ContentTypeCondition) Match(content string, detection DetectionResult) bool {
	for _, t := range c.Types {
		if detection.ContentType == t {
			return true
		}
	}
	return false
}

// LanguageCondition matches by programming language.
type LanguageCondition struct {
	Languages []Language
}

// Match implements GateCondition.
func (c LanguageCondition) Match(content string, detection DetectionResult) bool {
	for _, l := range c.Languages {
		if detection.Language == l {
			return true
		}
	}
	return false
}

// FeatureCondition matches by content features.
type FeatureCondition struct {
	Features map[string]bool
}

// Match implements GateCondition.
func (c FeatureCondition) Match(content string, detection DetectionResult) bool {
	for feature, required := range c.Features {
		if hasFeature, ok := detection.Features[feature]; ok {
			if hasFeature != required {
				return false
			}
		} else if required {
			return false
		}
	}
	return true
}

// SizeCondition matches by content size.
type SizeCondition struct {
	MinLines  int
	MaxLines  int
	MinChars  int
	MaxChars  int
	MinTokens int
	MaxTokens int
}

// Match implements GateCondition.
func (c SizeCondition) Match(content string, detection DetectionResult) bool {
	stats := detection.Stats
	if c.MinLines > 0 && stats.TotalLines < c.MinLines {
		return false
	}
	if c.MaxLines > 0 && stats.TotalLines > c.MaxLines {
		return false
	}
	if c.MinChars > 0 && stats.TotalChars < c.MinChars {
		return false
	}
	if c.MaxChars > 0 && stats.TotalChars > c.MaxChars {
		return false
	}
	return true
}

// CompositeCondition combines multiple conditions with AND logic.
type CompositeCondition struct {
	Conditions []GateCondition
}

// Match implements GateCondition.
func (c CompositeCondition) Match(content string, detection DetectionResult) bool {
	for _, cond := range c.Conditions {
		if !cond.Match(content, detection) {
			return false
		}
	}
	return true
}

// AnyCondition combines conditions with OR logic.
type AnyCondition struct {
	Conditions []GateCondition
}

// Match implements GateCondition.
func (c AnyCondition) Match(content string, detection DetectionResult) bool {
	for _, cond := range c.Conditions {
		if cond.Match(content, detection) {
			return true
		}
	}
	return false
}

// NotCondition negates a condition.
type NotCondition struct {
	Condition GateCondition
}

// Match implements GateCondition.
func (c NotCondition) Match(content string, detection DetectionResult) bool {
	return !c.Condition.Match(content, detection)
}

// GateRegistry manages all layer gates.
type GateRegistry struct {
	gates    []Gate
	detector *Detector
}

// NewGateRegistry creates a new gate registry.
func NewGateRegistry() *GateRegistry {
	return &GateRegistry{
		gates:    make([]Gate, 0),
		detector: NewDetector(),
	}
}

// Register adds a gate to the registry.
func (r *GateRegistry) Register(gate Gate) {
	r.gates = append(r.gates, gate)
	r.sortGates()
}

// ApplyGates runs applicable gates on content and returns processed result.
func (r *GateRegistry) ApplyGates(content string) (string, int) {
	detection := r.detector.Detect(content)
	totalSaved := 0
	result := content

	for _, gate := range r.gates {
		if layerGate, ok := gate.(*LayerGate); ok {
			if layerGate.ShouldApply(result, detection) {
				processed, saved := layerGate.Process(result)
				result = processed
				totalSaved += saved

				// Re-detect if content changed significantly
				if saved > len(content)/10 {
					detection = r.detector.Detect(result)
				}
			}
		}
	}

	return result, totalSaved
}

// GetApplicableGates returns list of gates that apply to content.
func (r *GateRegistry) GetApplicableGates(content string) []string {
	detection := r.detector.Detect(content)
	var applicable []string

	for _, gate := range r.gates {
		if gate.ShouldApply(content, detection) {
			applicable = append(applicable, gate.Name())
		}
	}

	return applicable
}

// Analyze returns detection info without processing.
func (r *GateRegistry) Analyze(content string) DetectionResult {
	return r.detector.Detect(content)
}

func (r *GateRegistry) sortGates() {
	// Simple bubble sort by priority
	for i := 0; i < len(r.gates)-1; i++ {
		for j := i + 1; j < len(r.gates); j++ {
			if r.gates[i].Priority() > r.gates[j].Priority() {
				r.gates[i], r.gates[j] = r.gates[j], r.gates[i]
			}
		}
	}
}

// DefaultGates returns the default set of layer gates.
func DefaultGates() []*LayerGate {
	var gates []*LayerGate

	// Layer 1: Entropy Filter - applies to all code
	gates = append(gates, NewLayerGate(
		"entropy_filter",
		10,
		ContentTypeCondition{Types: []ContentType{SourceCode}},
		&EntropyProcessor{},
	))

	// Layer 2: Perplexity Filter - applies to natural language
	gates = append(gates, NewLayerGate(
		"perplexity_filter",
		20,
		ContentTypeCondition{Types: []ContentType{NaturalLanguage, BuildLog}},
		&PerplexityProcessor{},
	))

	// Layer 3: Goal-Driven Analysis - large files
	gates = append(gates, NewLayerGate(
		"goal_driven",
		30,
		SizeCondition{MinLines: 100},
		&GoalDrivenProcessor{},
	))

	// Layer 4: AST Parsing - structured code
	gates = append(gates, NewLayerGate(
		"ast_parse",
		40,
		CompositeCondition{
			Conditions: []GateCondition{
				ContentTypeCondition{Types: []ContentType{SourceCode}},
				AnyCondition{
					Conditions: []GateCondition{
						LanguageCondition{Languages: []Language{LangGo, LangRust, LangPython, LangJavaScript, LangTypeScript, LangJava}},
					},
				},
			},
		},
		&ASTProcessor{},
	))

	// Layer 5: Contrastive Learning - test output
	gates = append(gates, NewLayerGate(
		"contrastive",
		50,
		ContentTypeCondition{Types: []ContentType{TestOutput}},
		&ContrastiveProcessor{},
	))

	// Layer 6: N-gram Deduplication - logs
	gates = append(gates, NewLayerGate(
		"ngram_dedup",
		60,
		ContentTypeCondition{Types: []ContentType{BuildLog, TestOutput}},
		&NgramProcessor{},
	))

	// Layer 6b: N-gram Deduplication - large content (alias for testing)
	gates = append(gates, NewLayerGate(
		"ngram_dedup_large",
		61,
		SizeCondition{MinLines: 50},
		&NgramProcessor{},
	))

	// Layer 7: LLM Evaluator - large content
	gates = append(gates, NewLayerGate(
		"llm_eval",
		70,
		SizeCondition{MinTokens: 1000},
		&LLMEvalProcessor{},
	))

	// Layer 8: Gist Memory - structured data
	gates = append(gates, NewLayerGate(
		"gist_memory",
		80,
		ContentTypeCondition{Types: []ContentType{StructuredData}},
		&GistProcessor{},
	))

	// Layer 9: Hierarchical Summarization - documents
	gates = append(gates, NewLayerGate(
		"hier_summary",
		90,
		AnyCondition{
			Conditions: []GateCondition{
				ContentTypeCondition{Types: []ContentType{NaturalLanguage}},
				SizeCondition{MinLines: 500},
			},
		},
		&HierarchicalProcessor{},
	))

	// Layer 10: Budget Enforcement - all content
	gates = append(gates, NewLayerGate(
		"budget_enforce",
		100,
		AlwaysCondition{},
		&BudgetProcessor{},
	))

	return gates
}

// Layer Processors

// EntropyProcessor reduces high-entropy content.
type EntropyProcessor struct{}

func (p *EntropyProcessor) Process(content string) (string, int) {
	// Remove high-entropy strings (likely hashes/IDs)
	lines := strings.Split(content, "\n")
	var result []string
	saved := 0

	for _, line := range lines {
		processed := removeHighEntropy(line)
		if processed != line {
			saved += len(line) - len(processed)
		}
		result = append(result, processed)
	}

	return strings.Join(result, "\n"), saved
}

func (p *EntropyProcessor) Name() string { return "entropy_filter" }

// PerplexityProcessor filters low-perplexity content.
type PerplexityProcessor struct{}

func (p *PerplexityProcessor) Process(content string) (string, int) {
	// Simplified: remove redundant phrases
	original := content
	content = removeRedundantPhrases(content)
	return content, len(original) - len(content)
}

func (p *PerplexityProcessor) Name() string { return "perplexity_filter" }

// GoalDrivenProcessor analyzes query intent.
type GoalDrivenProcessor struct{}

func (p *GoalDrivenProcessor) Process(content string) (string, int) {
	// Simplified: keep only relevant sections
	return content, 0 // Placeholder
}

func (p *GoalDrivenProcessor) Name() string { return "goal_driven" }

// ASTProcessor preserves AST structure.
type ASTProcessor struct{}

func (p *ASTProcessor) Process(content string) (string, int) {
	// Use existing filter's read modes
	opts := filter.ReadOptions{Mode: filter.ReadSignatures}
	return filter.ReadContent(content, opts), 0
}

func (p *ASTProcessor) Name() string { return "ast_parse" }

// ContrastiveProcessor uses contrastive learning.
type ContrastiveProcessor struct{}

func (p *ContrastiveProcessor) Process(content string) (string, int) {
	return content, 0 // Placeholder
}

func (p *ContrastiveProcessor) Name() string { return "contrastive" }

// NgramProcessor deduplicates n-grams.
type NgramProcessor struct{}

func (p *NgramProcessor) Process(content string) (string, int) {
	// Simplified: collapse repeated lines
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return content, 0
	}

	var result []string
	lastLine := ""
	repeatCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}

		if trimmed == lastLine {
			repeatCount++
			// Skip adding this line (it's a repeat)
			if repeatCount == 3 {
				// Add marker on 3rd occurrence
				result = append(result, "... (repeated)")
			}
		} else {
			if repeatCount > 2 {
				// Add count marker before new line
				result = append(result, fmt.Sprintf("(%d more identical)", repeatCount-2))
			}
			result = append(result, line)
			lastLine = trimmed
			repeatCount = 0
		}

		// Handle end of input with pending repeats
		if i == len(lines)-1 && repeatCount > 2 {
			result = append(result, fmt.Sprintf("(%d more identical)", repeatCount-2))
		}
	}

	processed := strings.Join(result, "\n")
	saved := len(content) - len(processed)
	if saved < 0 {
		saved = 0 // Don't return negative savings
	}
	return processed, saved
}

func (p *NgramProcessor) Name() string { return "ngram_dedup" }

// LLMEvalProcessor uses LLM evaluation.
type LLMEvalProcessor struct{}

func (p *LLMEvalProcessor) Process(content string) (string, int) {
	return content, 0 // Placeholder - requires LLM
}

func (p *LLMEvalProcessor) Name() string { return "llm_eval" }

// GistProcessor creates content gists.
type GistProcessor struct{}

func (p *GistProcessor) Process(content string) (string, int) {
	// Simplified: extract key-value pairs
	return content, 0 // Placeholder
}

func (p *GistProcessor) Name() string { return "gist_memory" }

// HierarchicalProcessor creates summaries.
type HierarchicalProcessor struct{}

func (p *HierarchicalProcessor) Process(content string) (string, int) {
	// Simplified: truncate with summary
	if len(content) > 10000 {
		return content[:5000] + "\n... [truncated for length] ...\n" + content[len(content)-1000:], 0
	}
	return content, 0
}

func (p *HierarchicalProcessor) Name() string { return "hier_summary" }

// BudgetProcessor enforces token budget.
type BudgetProcessor struct{}

func (p *BudgetProcessor) Process(content string) (string, int) {
	tokens := filter.EstimateTokens(content)
	if tokens > 10000 {
		// Truncate to budget
		return truncateToTokens(content, 10000), 0
	}
	return content, 0
}

func (p *BudgetProcessor) Name() string { return "budget_enforce" }

// AlwaysCondition always matches.
type AlwaysCondition struct{}

func (c AlwaysCondition) Match(content string, detection DetectionResult) bool {
	return true
}

// Helper functions

func removeHighEntropy(s string) string {
	// Remove strings that look like hashes or UUIDs
	// Simplified implementation
	hashPattern := `[a-f0-9]{32,}`
	re := regexpCompile(hashPattern)
	if re != nil {
		return re.ReplaceAllString(s, "[hash]")
	}
	return s
}

func removeRedundantPhrases(s string) string {
	// Simplified: remove repeated words
	words := strings.Fields(s)
	var result []string
	lastWord := ""
	for _, word := range words {
		if strings.ToLower(word) != strings.ToLower(lastWord) {
			result = append(result, word)
		}
		lastWord = word
	}
	return strings.Join(result, " ")
}

func truncateToTokens(content string, maxTokens int) string {
	// Simple approximation: ~4 chars per token
	maxChars := maxTokens * 4
	if len(content) <= maxChars {
		return content
	}
	half := maxChars / 2
	return content[:half] + "\n... [truncated] ...\n" + content[len(content)-half/2:]
}

func regexpCompile(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}
