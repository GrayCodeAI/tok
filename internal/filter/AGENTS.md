# AGENTS.md -- `internal/filter/` Package Guide

> **Purpose:** This document catalogs every non-test `.go` file in the `internal/filter`
> package. The package implements a layered compression pipeline (31-stage core with
> experimental extension toward 50+) that reduces LLM
> context while preserving semantic meaning. All files live in a single Go package
> (`package filter`) because they share core types (`PipelineState`, `Mode`,
> `ContentType`, etc.) and tightly-coupled interfaces.

> **Subdirectory convention:** this directory currently keeps all active layer
> implementations in the parent package for tight coordination.

---

## 1. Core Pipeline

These files orchestrate the entire compression system. Every other file in the package
plugs into this core.

| File | Description |
|------|-------------|
| `pipeline_process.go` | Main pipeline execution engine that applies ordered compression layers and produces the final output. |
| `pipeline_types.go` | Pipeline config types, `PipelineCoordinator` struct, layer config structs, and `PipelineStats`. |
| `pipeline_init.go` | Pipeline initialization: `NewPipelineCoordinator()`, filter instantiation, layer building. |
| `pipeline_early_exit.go` | Stage gates and early-exit logic (`shouldSkip*` methods). |
| `pipeline_runtime.go` | Runtime pipeline execution helpers. |
| `pipeline_stats.go` | Pipeline statistics collection and reporting. |
| `pipeline_stats_safe.go` | Thread-safe pipeline stats accessors. |
| `manager.go` | High-level lifecycle manager -- loads/saves pipeline configurations, handles SHA-256 cache keys, and coordinates preset selection. |
| `content_route_strategy.go` | Content router (formerly `adaptive_router.go`) that detects content type (JSON, code, logs, etc.) and selects the optimal compression strategy. |
| `core_filters.go` | `CoreFilters` -- manages layers 1-9 pipeline (formerly `refactored.go`). |
| `filter.go` | Defines the core `Mode` type (compress/reversible/aggressive), `Filter` interface, and the top-level `Apply()` entry point. |
| `presets.go` | Defines `PipelinePreset` constants (fast/balanced/full) and the layer sets each preset activates. |
| `tier_config.go` | Tier-based auto-configuration for layer enablement. |
| `parallel.go` | `ParallelProcessor` and `ParallelCompressor` for batch compression across multiple cores. |
| `coordinator_pool.go` | `CoordinatorPool` -- reusable `PipelineCoordinator` pool via `sync.Pool` (formerly `pool.go`). |
| `cached_pipeline.go` | `CachedPipeline` -- fingerprint-based caching of pipeline results. |
| `layer_registry.go` | `LayerRegistry` -- registry of available compression layers. |
| `layer_gate.go` | `LayerGate` -- controls which layers can run (experimental/stable/alpha gates). |
| `layer_cache.go` | `LayerCache` -- per-layer result caching. |

---

## 2. Compression Layers (Foundational L1--L20)

Each file implements one or more foundational layers. Most are
backed by published research (cited in file comments).

| File | Layer(s) | Description |
|------|----------|-------------|
| `entropy.go` | L1 | Entropy-based token pruning using Shannon entropy with SIMD-accelerated scoring; drops low-information tokens. |
| `perplexity.go` | L2 | LLMLingua-style perplexity-based iterative pruning (Microsoft/Tsinghua, 2023); ranks tokens by perplexity and removes the least surprising. |
| `goal_driven.go` | L3 | SWE-Pruner style goal-driven compression; uses task context to prioritize relevant code constructs. |
| `ast_preserve.go` | L4 | LongCodeZip-style AST-aware compression (NUS, 2025); preserves function signatures while compressing bodies. |
| `contrastive.go` | L5 | LongLLMLingua contrastive perplexity; question-aware compression via contrastive scoring. |
| `ngram.go` | L6 | N-gram abbreviation filter; compresses by abbreviating common multi-word patterns using SIMD-accelerated matching. |
| `evaluator_heads.go` | L7 | EHPC-style evaluator-heads compression (Tsinghua/Huawei, 2025); identifies important tokens via attention head simulation. |
| `gist.go` | L8 | Gisting compression (Stanford/Berkeley, 2023); compresses prompts into virtual "gist tokens" representing condensed meaning. |
| `hierarchical.go` | L9 *(also `hierarchical_summary.go`)* | Multi-level hierarchical summarization; AutoCompressor-style recursive summarization. |
| `budget.go` | L10 | Budget enforcer that scores output segments and keeps only the most important ones to hit a strict token limit. |
| `compaction.go` | L11 | Full compactor engine with SHA-256 chunk deduplication, cross-reference resolution, delta encoding, and merge logic. |
| `attribution.go` | L12 | Token attribution filter; scores tokens by their contribution to downstream predictions and drops low-attribution tokens. |
| `h2o.go` | L13 | Heavy-Hitter Oracle (H2O) compression; heap-based approach to preserve "heavy-hitter" tokens. |
| `attention_sink.go` | L14 | StreamingLLM-style attention-sink preservation; keeps attention-sink tokens at sequence boundaries. |
| `meta_token.go` | L15 | Meta-token compression; replaces repeated sub-sequences with short virtual tokens and a lookup table. |
| `semantic_chunk.go` | L16 | ChunkKV-style semantic chunk compression; groups tokens into semantic chunks and prunes low-relevance chunks. |
| `sketch_store.go` | L17 | Sketch-based storage; maintains lightweight hash-based sketches for fast similarity comparison and cross-reference resolution. |
| `lazy_pruner.go` | L18 | Budget-aware dynamic pruning (LazyLLM style); defers pruning decisions until budget is exhausted. |
| `semantic_anchor.go` | L19 | Semantic-Anchor Compression (SAC, 2024); identifies anchor points and compresses relative to them. |
| `agent_memory.go` | L20 | Agent memory mode (Focus-inspired); maintains compressed working memory for multi-turn agent sessions. |
| `semantic.go` | -- | General-purpose semantic filter; prunes low-information segments using statistical analysis and unicode-level heuristics. |

---

## 3. Unified Research Layers (Consolidated L21-L45 into L14-L16)

| File | Layer | Merges | Description |
|------|-------|--------|-------------|
| `edge_case_unified.go` | L14 | L21 MarginalInfoGain, L22 NearDedup, L23 CoTCompress, L24 CodingAgentContext, L25 PerceptionCompress | Unified edge case handling |
| `reasoning_unified.go` | L15 | L26 LightThinker, L27 ThinkSwitcher, L28 GMSA, L29 CARL, L30 SlimInfer | Unified reasoning trace compression |
| `advanced_research_unified.go` | L16 | L31-L45 (DiffAdapt, EPiC, SSDP, AgentOCR, etc.) | Unified advanced research optimizations |

---

## 4. Extended Research & Utility Layers

| File | Description |
|------|-------------|
| `adaptive_context_optimize.go` | ACON: adaptive context optimization (ICLR 2026) |
| `critical_action_retain.go` | CARL: retain critical tool-call entries, drop non-critical |
| `group_merge_semantic_align.go` | GMSA: group merging & semantic alignment |
| `multi_agent_debate_collapse.go` | S2MAD: collapse agreement phrases in multi-agent debate |
| `tree_search_diverge_prune.go` | SSDP: prune redundant tree-of-thought branches |
| `difficulty_adaptive_compress.go` | DiffAdapt: difficulty-adaptive compression ratio |
| `causal_edge_preserve.go` | EPiC: preserve causal edge lines in reasoning |
| `token_dense_dialect.go` | TDD: replace common terms with Unicode symbol shorthand |
| `columnar_json_encode.go` | TOON: columnar encoding for homogeneous JSON arrays |
| `path_shim_injector.go` | PATH shim injector for auto-filtering subprocesses |
| `image_compress.go` | PhotonFilter: detect & compress base64-encoded images |
| `compression_stage_map.go` | FusionStageMap: maps Claw-style stages to Tok layer IDs |
| `coverage_check.go` | Coverage test runner script |
| `reasoning_step_compress.go` | LightThinker: compress reasoning output per step |
| `memory_augment_compress.go` | LightMem: replace repeated facts with short references |
| `orphan_line_drop.go` | SlimInfer: drop isolated lines not referenced by others |
| `path_alias_compress.go` | PathShorten: alias repeated long paths with short tokens |
| `json_stream_sampler.go` | JSONSampler: down-sample dense JSON line streams |
| `structural_collapse.go` | StructuralCollapse: compact repetitive structural boilerplate |
| `chain_of_thought_compress.go` | CoTCompress: compress chain-of-thought reasoning traces |
| `near_duplicate_collapse.go` | NearDedup: SimHash-based near-duplicate line collapse |
| `conversation_turn_dedup.go` | ConversationDedup: deduplicate across conversation turns |
| `coding_agent_context.go` | CodingAgentContext: prune coding agent tool outputs |
| `perceptual_redundancy_drop.go` | PerceptionCompress: remove perceptually redundant lines |
| `reasoning_route_compress.go` | ThinkSwitcher: route to fast/light/heavy compression |
| `agent_density_compress.go` | AgentOCR: measure density per agent turn, collapse low-density |
| `agent_history_compress.go` | AgentOCRHistory: compact older conversation turns |
| `role_budget_compress.go` | RoleBudget: allocate budget by multi-agent role |
| `graph_reasoning_compress.go` | GraphCoT: keep high-centrality reasoning lines |
| `latent_collab_collapse.go` | LatentCollab: collapse semantically equivalent multi-agent turns |
| `kv_cache_stabilize.go` | QuantumLock: stabilize system prompts for KV-cache alignment |
| `small_model_bridge.go` | SmallKV: reconstruct bridges for broken patterns |
| `signature_truncate.go` | SmartTruncate: truncate preserving function signatures |
| `log_fold_compress.go` | LogCrunch: fold repetitive logs, preserve errors |
| `diff_fold_compress.go` | DiffCrunch: compact large diffs |
| `auto_content_compress.go` | ContextCrunch: auto-detecting log/diff compression |
| `search_result_dedup.go` | SearchCrunch: deduplicate search results |
| `error_pattern_learner.go` | EngramLearner: learn compression failure patterns |
| `progressive_summarize.go` | TieredSummary: L0/L1/L2 progressive summarization |
| `content_profile_detect.go` | ContentProfile: auto-detect compression profile |
| `code_comment_strip.go` | CommentFilter: strip comments per language |
| `panic_safe_wrapper.go` | SafeFilter: nil-safe panic-recovery wrapper |
| `threshold_feedback_learn.go` | FeedbackLoop: learn thresholds from compression quality |
| `cross_layer_feedback.go` | InterLayerFeedback: cross-layer compression feedback |
| `kv_cache_aligner.go` | KVCacheAligner: KV-cache prompt alignment |
| `signature_patterns.go` | Precompiled regex patterns for signatures/imports |
| `file_read_modes.go` | ReadMode: file reading strategy definitions |
| `adaptive_learning.go` | AdaptiveLearning: merged EngramLearner + TieredSummary |
| `crunch_bench.go` | CrunchBench: comprehensive compression benchmarking |
| `swe_adaptive_loop.go` | SWEAdaptiveLoop: iterative prune loop inspired by SWE-Pruner |

---

## 5. Adaptive Selection

These files dynamically adjust which compression layers run and how aggressively they
compress, based on content characteristics.

| File | Description |
|------|-------------|
| `adaptive.go` | Adaptive layer selector; uses heuristic content-type analysis to dynamically enable/disable layers and tune thresholds per input. Contains `DensityAdaptiveAllocator` (DAST-style). |

---

## 6. Utilities

Shared helpers, data structures, and low-level primitives used across the package.

| File | Description |
|------|-------------|
| `utils.go` | Core tokenizer and string normalization helpers (`cleanWord`, `tokenizeRe`); used by nearly every filter. |
| `constants.go` | Package-wide constant definitions. |
| `cache_lru_compat.go` | `LRUCache` backward-compatibility shim that aliases to `cache.LRUCache`. |
| `bytepool.go` | `BytePool` and `FastStringBuilder` -- reusable byte buffer pools to reduce GC pressure. |
| `ansi.go` | ANSI escape-sequence stripper; uses SIMD-optimized byte scanning for 10--40x speedup over regex. |
| `noise.go` | Progress-bar and noise detector; identifies and removes transient CLI output. |
| `dedup.go` | Line-level deduplication filter; removes duplicate lines common in logs and test output. |
| `equivalence.go` | Semantic equivalence checker; verifies that compressed output preserves critical information. |
| `fingerprint.go` | Content-hash (SHA-256) based cache key generation. |
| `doc.go` | Package documentation. |
| `optimizations_benchmark_test.go` | Benchmarks for SIMD-optimized filter operations. |
| `pipeline_bench_test.go` | Pipeline-level benchmarks. |
| `pipeline_runtime_test.go` | Pipeline runtime tests. |

---

## 7. Code-Aware Processing

Filters that understand programming-language structure (imports, braces, comments).

| File | Description |
|------|-------------|
| `ast_preserve.go` | *(Also in L4.)* AST-aware filter that parses function signatures, class declarations, and control-flow boundaries. |
| `brace_depth.go` | `BodyFilter` that strips function bodies based on brace depth; preserves signatures while removing implementation detail. |
| `signature_patterns.go` | Precompiled regex patterns for function/type signature detection and import matching across ~20 languages. |
| `code_comment_strip.go` | Language-specific comment pattern registry and stripper; defines line-comment, block-comment, and doc-comment syntax. |
| `import.go` | Import statement condenser; collapses verbose import blocks into compact representations per language. |

---

## 8. Quality / Analysis

Metrics and analysis tools that measure or preserve compression quality.

| File | Description |
|------|-------------|
| `quality.go` | `QualityMetrics` -- measures information preservation ratio, key-term retention, and structural integrity after compression. |
| `quality_guardrail.go` | `QualityGuardrail` -- optional output quality check that prevents over-aggressive compression. |
| `attribution.go` | *(Also in L12.)* Per-token attribution scoring that measures each token's contribution to downstream predictions. |

---

## 9. Parallel & Batch Processing

| File | Description |
|------|-------------|
| `parallel.go` | `ParallelProcessor`, `ParallelCompressor`, and `ParallelProcessResult` for batch compression across multiple CPU cores. |

---

## 10. Streaming

| File | Description |
|------|-------------|
| `streaming.go` | `StreamingProcessor` for real-time compression of streaming content; designed for chat agents and long-running sessions. |

---

## 11. Reversible Compression

| File | Description |
|------|-------------|
| `reversible.go` | Reversible compression with on-disk full-content storage and SHA-256 integrity verification. |

---

## 12. Position-Aware Compression

| File | Description |
|------|-------------|
| `position_aware.go` | Reorders output segments to counteract the "lost in the middle" phenomenon; puts important content at sequence ends. |

---

## 13. Question-Aware Compression

| File | Description |
|------|-------------|
| `query_aware.go` | LongLLMLingua-style question-aware recovery; preserves query-relevant subsequences by scoring token--question similarity. |

---

## 14. LLM-Aware Compression

| File | Description |
|------|-------------|
| `llm_aware.go` | LLM-aware filter using a local LLM for high-quality summarization when heuristic compression is insufficient. |
| `llm_compress.go` | LLM-driven compression via external process invocation for context-aware summarization with JSON I/O. |

---

## 15. Session Management

| File | Description |
|------|-------------|
| `session.go` | Session manager with SHA-256-keyed state persistence; stores compressed context snapshots on disk for continuity across agent turns. |

---

## 16. Log Aggregation

| File | Description |
|------|-------------|
| `log_fold_compress.go` | `LogCrunch` -- folds repetitive INFO/DEBUG logs while preserving warnings, errors, and stack traces. |
| `diff_fold_compress.go` | `DiffCrunch` -- compacts large diffs by pruning repetitive unchanged context lines. |
| `search_result_dedup.go` | `SearchCrunch` -- deduplicates repeated search result lines and keeps top unique hits. |

---

## 17. Multi-File Handling

| File | Description |
|------|-------------|
| `multi_file.go` | Multi-file filter with cross-file relationship detection; sorts, deduplicates, and creates consolidated views of related file outputs. |

---

## 18. TOML Filter Configuration

| File | Description |
|------|-------------|
| `pipeline_toml.go` | Integration of TOML-based declarative filter definitions into the pipeline. |
