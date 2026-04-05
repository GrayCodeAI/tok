// Package engine provides the lightweight filter engine for quick output
// post-processing, distinct from the full 26+ layer pipeline.
//
// The engine handles simple formatting tasks:
//   - ANSI code stripping
//   - Comment removal
//   - Import condensing
//   - Semantic pruning
//   - Position-aware reordering
//   - Hierarchical summarization
//   - Query-aware filtering
//
// Usage:
//
//	engine := filter.NewEngine(filter.ModeMinimal)
//	output, saved := engine.Process(input)
//
// For full compression, use filter.PipelineCoordinator instead.
package engine
