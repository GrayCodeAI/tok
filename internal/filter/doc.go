// Package filter provides the core token compression pipeline for tok.
//
// The filter package implements a multi-layer compression architecture
// inspired by 120+ research papers from top institutions. It processes
// CLI output through a series of filter layers to reduce token usage
// in LLM interactions while preserving semantic meaning.
//
// # Pipeline Architecture
//
// The main entry point is PipelineCoordinator, which orchestrates up to
// 50+ compression layers organized into logical groups:
//
//   - Core Layers (1-9): Entropy, Perplexity, AST preservation, etc.
//   - Semantic Layers (11-20): Compaction, H2O, Attention Sink, etc.
//   - Research Layers (21-49): Advanced techniques like DiffAdapt, EPiC, etc.
//
// # Usage
//
// Create a coordinator with configuration and process text:
//
//	pipeline := filter.NewPipelineCoordinator(&config)
//	output, stats := pipeline.Process(inputText)
//	fmt.Printf("Saved %d tokens (%.1f%%)\n", stats.TotalSaved, stats.ReductionPercent)
//
// # Filter Engine
//
// For lightweight post-processing, the Engine type provides a simpler
// filter chain for tasks like ANSI stripping and comment removal.
package filter
