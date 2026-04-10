package filter

// LayerTier describes maturity/activation tier for a layer.
type LayerTier string

const (
	LayerTierStable       LayerTier = "stable"
	LayerTierExperimental LayerTier = "experimental"
	LayerTierRecovery     LayerTier = "recovery"
	LayerTierPlanned      LayerTier = "planned"
)

// LayerMeta documents a layer and its research provenance.
type LayerMeta struct {
	ID       string
	Name     string
	Tier     LayerTier
	Status   string // implemented | planned
	PaperRef string
}

// LayerRegistry stores metadata for all known layers.
type LayerRegistry struct {
	layers map[string]LayerMeta
}

func NewLayerRegistry() *LayerRegistry {
	r := &LayerRegistry{layers: make(map[string]LayerMeta, 64)}
	for _, m := range defaultLayerMeta() {
		r.layers[m.ID] = m
	}
	return r
}

func (r *LayerRegistry) Get(id string) (LayerMeta, bool) {
	m, ok := r.layers[id]
	return m, ok
}

func defaultLayerMeta() []LayerMeta {
	implemented := []LayerMeta{
		{ID: "pre_extractive", Name: "Extractive Prefilter", Tier: LayerTierStable, Status: "implemented", PaperRef: "Selective Context (2023)"},
		{ID: "pre_tfidf", Name: "TF-IDF Prefilter", Tier: LayerTierStable, Status: "implemented", PaperRef: "Classic IR"},
		{ID: "1_entropy", Name: "Entropy Filtering", Tier: LayerTierStable, Status: "implemented", PaperRef: "Selective Context (2023)"},
		{ID: "2_perplexity", Name: "Perplexity Pruning", Tier: LayerTierStable, Status: "implemented", PaperRef: "LLMLingua (2023)"},
		{ID: "3_goal_driven", Name: "Goal Driven Selection", Tier: LayerTierStable, Status: "implemented", PaperRef: "SWE-Pruner (2026)"},
		{ID: "4_ast_preserve", Name: "AST Preservation", Tier: LayerTierStable, Status: "implemented", PaperRef: "LongCodeZip"},
		{ID: "5_contrastive", Name: "Contrastive Ranking", Tier: LayerTierStable, Status: "implemented", PaperRef: "LongLLMLingua"},
		{ID: "6_ngram", Name: "N-gram Abbreviation", Tier: LayerTierStable, Status: "implemented", PaperRef: "CompactPrompt"},
		{ID: "7_evaluator", Name: "Evaluator Heads", Tier: LayerTierStable, Status: "implemented", PaperRef: "EHPC"},
		{ID: "8_gist", Name: "Gist Compression", Tier: LayerTierStable, Status: "implemented", PaperRef: "Gist Tokens"},
		{ID: "9_hierarchical", Name: "Hierarchical Summary", Tier: LayerTierStable, Status: "implemented", PaperRef: "AutoCompressor"},
		{ID: "10_neural", Name: "Neural Compression", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "LLM Summarization"},
		{ID: "11_compaction", Name: "Compaction", Tier: LayerTierStable, Status: "implemented", PaperRef: "MemGPT"},
		{ID: "12_attribution", Name: "Attribution Filter", Tier: LayerTierStable, Status: "implemented", PaperRef: "ProCut"},
		{ID: "13_h2o", Name: "H2O Filter", Tier: LayerTierStable, Status: "implemented", PaperRef: "H2O"},
		{ID: "14_attention_sink", Name: "Attention Sink", Tier: LayerTierStable, Status: "implemented", PaperRef: "StreamingLLM"},
		{ID: "15_meta_token", Name: "Meta Token", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Meta Token (2025)"},
		{ID: "16_semantic_chunk", Name: "Semantic Chunk", Tier: LayerTierStable, Status: "implemented", PaperRef: "ChunkKV-like"},
		{ID: "17_sketch_store", Name: "Sketch Store", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "KVReviver"},
		{ID: "18_lazy_pruner", Name: "Lazy Pruner", Tier: LayerTierStable, Status: "implemented", PaperRef: "LazyLLM"},
		{ID: "19_semantic_anchor", Name: "Semantic Anchor", Tier: LayerTierStable, Status: "implemented", PaperRef: "SAC"},
		{ID: "20_agent_memory", Name: "Agent Memory", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Focus-inspired"},
		{ID: "20_symbolic_compress", Name: "Symbolic Compress", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "MetaGlyph"},
		{ID: "21_phrase_grouping", Name: "Phrase Grouping", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "CompactPrompt"},
		{ID: "22_numerical_quant", Name: "Numerical Quantization", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "CompactPrompt"},
		{ID: "23_dynamic_ratio", Name: "Dynamic Ratio", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "PruneSID"},
		{ID: "24_hypernym", Name: "Hypernym Compression", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Hypernym Prompt Compression"},
		{ID: "25_semantic_cache", Name: "Semantic Cache", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "SemantiCache"},
		{ID: "26_scope", Name: "Scope Filter", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "SCOPE"},
		{ID: "27_kvzip", Name: "KVzip", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "KVzip"},
		{ID: "28_question_aware", Name: "Question Aware", Tier: LayerTierRecovery, Status: "implemented", PaperRef: "LongLLMLingua"},
		{ID: "29_density_adaptive", Name: "Density Adaptive", Tier: LayerTierRecovery, Status: "implemented", PaperRef: "DAST"},
	}

	planned := []LayerMeta{
		{ID: "30_salience_graph", Name: "Salience Graph", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "GraphRAG-style ranking"},
		{ID: "31_trace_preserve", Name: "Trace Preserve", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Failure trace retention"},
		{ID: "32_ast_diff_focus", Name: "AST Diff Focus", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Syntax-guided diff compression"},
		{ID: "33_unit_test_focus", Name: "Unit Test Focus", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Test-aware pruning"},
		{ID: "34_symbol_table", Name: "Symbol Table Keep", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Code map compression"},
		{ID: "35_path_anchor", Name: "Path Anchor", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "File-path anchor heuristics"},
		{ID: "36_stacktrace_focus", Name: "Stacktrace Focus", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Runtime failure compression"},
		{ID: "37_exit_signal_keep", Name: "Exit Signal Keep", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Outcome-preserving compression"},
		{ID: "38_semantic_dedup", Name: "Semantic Dedup", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Semantic redundancy removal"},
		{ID: "39_recall_booster", Name: "Recall Booster", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Recall-constrained pruning"},
		{ID: "40_log_cluster", Name: "Log Cluster", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Log summarization clustering"},
		{ID: "41_error_window", Name: "Error Window", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Error-centric windowing"},
		{ID: "42_dependency_focus", Name: "Dependency Focus", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Dependency-aware context"},
		{ID: "43_symbolic_patch", Name: "Symbolic Patch", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Patch tokenization"},
		{ID: "44_runtime_anchor", Name: "Runtime Anchor", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Session continuity anchors"},
		{ID: "45_multiturn_merge", Name: "Multi-turn Merge", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Conversation compression"},
		{ID: "46_context_cache", Name: "Context Cache", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Cache-aware context reuse"},
		{ID: "47_confidence_gate", Name: "Confidence Gate", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Uncertainty-guided pruning"},
		{ID: "48_loss_guard", Name: "Loss Guard", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Faithfulness constraints"},
		{ID: "49_repair_pass", Name: "Repair Pass", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Context repair"},
	}

	return append(implemented, planned...)
}
