// Package layers contains all compression layer implementations.
//
// Each layer implements the filter.Filter interface with Apply() method.
// Layers are organized by research lineage and execution order.
//
// Execution Order (pipeline):
//
//	Layer  1: Entropy Filtering (Selective Context, Mila 2023)
//	Layer  2: Perplexity Pruning (LLMLingua, Microsoft 2023)
//	Layer  3: Goal-Driven Selection (SWE-Pruner, Shanghai Jiao Tong 2025)
//	Layer  4: AST Preservation (LongCodeZip, NUS 2025)
//	Layer  5: Contrastive Ranking (LongLLMLingua, Microsoft 2024)
//	Layer  6: N-gram Abbreviation (CompactPrompt, 2025)
//	Layer  7: Evaluator Heads (EHPC, Tsinghua/Huawei 2025)
//	Layer  8: Gist Compression (Stanford/Berkeley, 2023)
//	Layer  9: Hierarchical Summary (AutoCompressor, Princeton/MIT 2023)
//	Layer 10: Budget Enforcement
//	Layer 11: Compaction (Semantic compression)
//	Layer 12: Attribution Filter (ProCut, LinkedIn 2025)
//	Layer 13: H2O Filter (Heavy-Hitter Oracle, NeurIPS 2023)
//	Layer 14: Attention Sink (StreamingLLM, 2023)
//	Layer 15: Meta-Token (arXiv:2506.00307, 2025)
//	Layer 16: Semantic Chunk (ChunkKV-style)
//	Layer 17: Sketch Store (KVReviver, Dec 2025)
//	Layer 18: Lazy Pruner (LazyLLM, July 2024)
//	Layer 19: Semantic Anchor (Attention Gradient Detection)
//	Layer 20: Agent Memory (Focus-inspired)
//	Layer 21-27: 2026 Research (SWEzze, MixedDimKV, BEAVER, PoC, TokenQuant, TokenRetention, ACON)
//
// This package re-exports layer constructors from the parent filter package
// for cleaner import paths. The actual implementations remain in the parent
// package to avoid circular dependencies during the transition period.
package layers
