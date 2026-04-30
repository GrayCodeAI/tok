package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// FilterParams for tok_filter tool
type FilterParams struct {
	Text   string `json:"text"`
	Mode   string `json:"mode,omitempty"`
	Budget int    `json:"budget,omitempty"`
	Query  string `json:"query,omitempty"`
}

// CompressFileParams for tok_compress_file tool
type CompressFileParams struct {
	Path     string `json:"path"`
	Mode     string `json:"mode,omitempty"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// AnalyzeParams for tok_analyze_output tool
type AnalyzeParams struct {
	Text string `json:"text"`
}

// FilterResult for filter tool response
type FilterResult struct {
	FilteredText   string   `json:"filtered_text"`
	OriginalTokens int      `json:"original_tokens"`
	FilteredTokens int      `json:"filtered_tokens"`
	TokensSaved    int      `json:"tokens_saved"`
	SavingsPercent float64  `json:"savings_percent"`
	LayersApplied  []string `json:"layers_applied"`
}

// handleFilter processes the tok_filter tool
func (s *Server) handleFilter(arguments json.RawMessage) (*ToolsCallResult, error) {
	var params FilterParams
	if err := json.Unmarshal(arguments, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if params.Text == "" {
		return nil, fmt.Errorf("text parameter is required")
	}

	// Configure pipeline
	mode := filter.ModeMinimal
	switch params.Mode {
	case "aggressive":
		mode = filter.ModeAggressive
	}

	cfg := filter.PipelineConfig{
		Mode:        mode,
		Budget:      params.Budget,
		QueryIntent: params.Query,
	}

	// Use existing pipeline or create new one
	pipeline := s.pipeline
	if pipeline == nil {
		pipeline = filter.NewPipelineCoordinator(&cfg)
	}

	// Process text
	filtered, stats, err := pipeline.Process(params.Text)
	if err != nil {
		return nil, err
	}

	// Build result
	layersApplied := make([]string, 0, len(stats.LayerStats))
	for name := range stats.LayerStats {
		layersApplied = append(layersApplied, name)
	}

	result := FilterResult{
		FilteredText:   filtered,
		OriginalTokens: stats.OriginalTokens,
		FilteredTokens: stats.FinalTokens,
		TokensSaved:    stats.TotalSaved,
		SavingsPercent: stats.ReductionPercent,
		LayersApplied:  layersApplied,
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ToolsCallResult{
		Content: []Content{
			NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleCompressFile processes the tok_compress_file tool
func (s *Server) handleCompressFile(arguments json.RawMessage) (*ToolsCallResult, error) {
	var params CompressFileParams
	if err := json.Unmarshal(arguments, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if params.Path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	// Read file
	content, err := os.ReadFile(params.Path)
	if err != nil {
		return &ToolsCallResult{
			Content: []Content{
				NewTextContent(fmt.Sprintf("Error reading file: %v", err)),
			},
			IsError: true,
		}, nil
	}

	text := string(content)
	originalLines := strings.Split(text, "\n")

	// Configure pipeline
	mode := filter.ModeMinimal
	switch params.Mode {
	case "aggressive":
		mode = filter.ModeAggressive
	}

	cfg := filter.PipelineConfig{Mode: mode}
	pipeline := s.pipeline
	if pipeline == nil {
		pipeline = filter.NewPipelineCoordinator(&cfg)
	}

	// Process
	filtered, stats, err := pipeline.Process(text)
	if err != nil {
		return nil, err
	}

	// Apply max_lines limit if specified
	if params.MaxLines > 0 {
		lines := strings.Split(filtered, "\n")
		if len(lines) > params.MaxLines {
			lines = lines[:params.MaxLines]
			lines = append(lines, fmt.Sprintf("... (%d more lines)",
				len(strings.Split(filtered, "\n"))-params.MaxLines))
			filtered = strings.Join(lines, "\n")
		}
	}

	// Build response
	response := fmt.Sprintf(`File: %s
Original: %d lines, ~%d tokens
Filtered: ~%d lines, ~%d tokens
Savings: %d tokens (%.1f%%)

--- Filtered Content ---
%s`,
		params.Path,
		len(originalLines),
		stats.OriginalTokens,
		len(strings.Split(filtered, "\n")),
		stats.FinalTokens,
		stats.TotalSaved,
		stats.ReductionPercent,
		filtered,
	)

	return &ToolsCallResult{
		Content: []Content{
			NewTextContent(response),
		},
	}, nil
}

// handleAnalyzeOutput analyzes without filtering
func (s *Server) handleAnalyzeOutput(arguments json.RawMessage) (*ToolsCallResult, error) {
	var params AnalyzeParams
	if err := json.Unmarshal(arguments, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if params.Text == "" {
		return nil, fmt.Errorf("text parameter is required")
	}

	// Analyze structure
	lines := strings.Split(params.Text, "\n")
	tokens := len(params.Text) / 4 // Rough estimate

	// Detect patterns
	patterns := detectPatterns(params.Text)

	// Estimate compression potential
	compression := estimateCompression(params.Text)

	analysis := fmt.Sprintf(`Analysis Report:
================
Total Lines: %d
Estimated Tokens: %d

Detected Patterns:
%s

Compression Potential:
- Minimal mode: ~%.0f%% reduction
- Balanced mode: ~%.0f%% reduction  
- Aggressive mode: ~%.0f%% reduction

Recommendations:
%s`,
		len(lines),
		tokens,
		strings.Join(patterns, "\n"),
		compression[filter.ModeMinimal]*100,
		compression[filter.ModeMinimal]*100,
		compression[filter.ModeAggressive]*100,
		getRecommendations(patterns),
	)

	return &ToolsCallResult{
		Content: []Content{
			NewTextContent(analysis),
		},
	}, nil
}

// handleGetStats returns usage statistics
func (s *Server) handleGetStats(arguments json.RawMessage) (*ToolsCallResult, error) {
	// Get tracking data if available
	tracker := tracking.GetGlobalTracker()

	var sessionStats string
	if tracker != nil {
		summary, err := tracker.GetSavings("")
		if err == nil && summary != nil {
			sessionStats = fmt.Sprintf(`Current Session:
- Commands processed: %d
- Total tokens saved: %d
- Average compression: %.1f%%
`,
				summary.TotalCommands,
				summary.TotalSaved,
				summary.ReductionPct)
		}
	}

	if sessionStats == "" {
		sessionStats = `Current Session:
- Commands processed: 0 (tracking not enabled)
- Total tokens saved: 0
- Average compression: N/A

For detailed stats, enable tracking in config:
[tracking]
enabled = true`
	}

	stats := fmt.Sprintf(`tok Statistics
=================

%s

Available Tools:
- tok_filter: Filter arbitrary text
- tok_compress_file: Compress file content
- tok_analyze_output: Analyze without filtering
- tok_get_stats: This tool
- tok_explain_layers: Explain compression layers
`, sessionStats)

	return &ToolsCallResult{
		Content: []Content{
			NewTextContent(stats),
		},
	}, nil
}

// handleExplainLayers explains the compression pipeline
func (s *Server) handleExplainLayers(arguments json.RawMessage) (*ToolsCallResult, error) {
	explanation := `tok Compression Pipeline (20 Layers)
========================================

LAYER 1: Entropy Filtering
Removes low-information tokens based on Shannon entropy.
Good for: Removing repetitive log lines, boilerplate.

LAYER 2: Perplexity Pruning  
Iterative token removal using perplexity scoring (LLMLingua).
Good for: Natural language text.

LAYER 3: Goal-Driven Selection
CRF-style line scoring based on query intent (SWE-Pruner).
Good for: Task-specific filtering.

LAYER 4: AST Preservation
Syntax-aware compression preserving code structure (LongCodeZip).
Good for: Source code files.

LAYER 5: Contrastive Ranking
Question-relevance scoring (LongLLMLingua).
Good for: Documentation, queries.

LAYER 6: N-gram Abbreviation
Lossless pattern compression (CompactPrompt).
Good for: Repetitive patterns.

LAYER 7: Evaluator Heads
Early-layer attention simulation (EHPC).
Good for: Long context windows.

LAYER 8: Gist Compression
Virtual token embedding (Stanford/Berkeley).
Good for: Semantic compression.

LAYER 9: Hierarchical Summary
Recursive summarization (AutoCompressor).
Good for: Multi-level summaries.

LAYER 10: Budget Enforcement
Strict token limits.
Good for: Hard constraints.

LAYER 11: Compaction
Semantic compression for conversations (MemGPT).
Good for: Chat logs, threads.

LAYER 12: Attribution Filter
78% pruning via attribution (ProCut).
Good for: Large outputs.

LAYER 13: H2O Filter
Heavy-Hitter Oracle (NeurIPS 2023).
Good for: 30x+ compression.

LAYER 14: Attention Sink
StreamingLLM stability.
Good for: Infinite context.

LAYER 15: Meta-Token
27% lossless compression.
Good for: Structured data.

LAYER 16: Semantic Chunk
Context-aware boundaries (ChunkKV).
Good for: Long documents.

LAYER 17: Semantic Cache
KVReviver + semantic reuse.
Good for: Repeated patterns.

LAYER 18: Lazy Pruner
2.34x speedup (LazyLLM).
Good for: Performance.

LAYER 19: Semantic Anchor
Attention gradient detection.
Good for: Context preservation.

LAYER 20: Agent Memory
Knowledge graph extraction.
Good for: Multi-turn conversations.

MODES:
- minimal: Light filtering, preserves most content (default)
- aggressive: Heavy filtering, maximum compression

PRESETS:
- fast: Fewer layers, quick results
- balanced: Default mix (minimal mode)
- full: All 20 layers enabled
`

	return &ToolsCallResult{
		Content: []Content{
			NewTextContent(explanation),
		},
	}, nil
}

// detectPatterns detects content patterns
func detectPatterns(text string) []string {
	patterns := []string{}

	if strings.Contains(text, "error") || strings.Contains(text, "Error") {
		patterns = append(patterns, "- Error messages detected")
	}
	if strings.Contains(text, "import ") || strings.Contains(text, "from ") {
		patterns = append(patterns, "- Import statements (Python/JS)")
	}
	if strings.Contains(text, "func ") || strings.Contains(text, "def ") {
		patterns = append(patterns, "- Function definitions")
	}
	if strings.Contains(text, "class ") {
		patterns = append(patterns, "- Class definitions")
	}
	if strings.Contains(text, "// ") || strings.Contains(text, "# ") || strings.Contains(text, "/*") {
		patterns = append(patterns, "- Comments present")
	}
	if strings.Contains(text, "http") || strings.Contains(text, "HTTP") {
		patterns = append(patterns, "- URLs/HTTP references")
	}
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		patterns = append(patterns, "- Structured data (JSON-like)")
	}
	if len(patterns) == 0 {
		patterns = append(patterns, "- Generic text content")
	}

	return patterns
}

// estimateCompression estimates compression ratios
func estimateCompression(text string) map[filter.Mode]float64 {
	lines := len(strings.Split(text, "\n"))

	// Heuristic estimates
	switch {
	case lines < 50:
		return map[filter.Mode]float64{
			filter.ModeMinimal:    0.15,
			filter.ModeAggressive: 0.30,
		}
	case lines < 200:
		return map[filter.Mode]float64{
			filter.ModeMinimal:    0.30,
			filter.ModeAggressive: 0.60,
		}
	case lines < 1000:
		return map[filter.Mode]float64{
			filter.ModeMinimal:    0.45,
			filter.ModeAggressive: 0.75,
		}
	default:
		return map[filter.Mode]float64{
			filter.ModeMinimal:    0.55,
			filter.ModeAggressive: 0.90,
		}
	}
}

// getRecommendations returns filter recommendations
func getRecommendations(patterns []string) string {
	recs := []string{}

	for _, p := range patterns {
		switch {
		case strings.Contains(p, "Error"):
			recs = append(recs, "- Use 'err' mode to focus on error messages")
		case strings.Contains(p, "Import"):
			recs = append(recs, "- Use 'minimal' mode to preserve dependencies")
		case strings.Contains(p, "Function"):
			recs = append(recs, "- AST layer will preserve function signatures")
		}
	}

	if len(recs) == 0 {
		return "- Use 'minimal' mode for general purpose filtering"
	}

	return strings.Join(recs, "\n")
}
