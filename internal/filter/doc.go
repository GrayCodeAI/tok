// Package filter implements the 31-layer token compression pipeline.
//
// # Organization
//
// This package contains 85 source files organized by concern:
//
// Core Pipeline (pipeline_*.go, manager.go)
//
//	Pipeline coordinator, config types, layer initialization, and execution.
//
// Filter Interface (filter.go)
//
//	Shared types: Mode, Filter, Engine, Language detection.
//
// Layer 1-10: Core Compression (entropy.go, perplexity.go, goal_driven.go, etc.)
//
//	Research-backed token reduction from 120+ papers.
//
// Layer 11-20: Semantic Filters (compaction.go, attribution.go, h2o.go, etc.)
//
//	Context-aware compression for conversations and complex output.
//
// Layer 21-27: Research Filters (swezze.go, mixed_dim.go, beaver.go, etc.)
//
//	Latest research from arXiv 2025-2026.
//
// Adaptive Layers (adaptive.go, density_adaptive.go, dynamic_ratio.go, etc.)
//
//	Self-tuning compression based on content characteristics.
//
// Utility Filters (ansi.go, noise.go, dedup.go, brace_depth.go, etc.)
//
//	Formatting, deduplication, and structural filters.
//
// Infrastructure (lru_cache.go, semantic_cache.go, streaming.go, session.go, etc.)
//
//	Caching, streaming, session management for pipeline support.
//
// # Usage
//
// For full 31-layer pipeline:
//
//	cfg := filter.PipelineConfig{EnableEntropy: true, ...}
//	pipeline := filter.NewPipelineCoordinator(cfg)
//	output, stats := pipeline.Process(input)
//
// For lightweight filtering (ANSI, comments, imports):
//
//	engine := filter.NewEngine(filter.ModeMinimal)
//	output, saved := engine.Process(input)
package filter
