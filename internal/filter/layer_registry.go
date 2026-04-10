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
		{ID: "1_entropy", Name: "Entropy Filtering", Tier: LayerTierStable, Status: "implemented", PaperRef: "Selective Context (2023)"},
		{ID: "2_perplexity", Name: "Perplexity Pruning", Tier: LayerTierStable, Status: "implemented", PaperRef: "LLMLingua (2023)"},
		{ID: "3_goal_driven", Name: "Goal Driven Selection", Tier: LayerTierStable, Status: "implemented", PaperRef: "SWE-Pruner (2026)"},
		{ID: "4_ast_preserve", Name: "AST Preservation", Tier: LayerTierStable, Status: "implemented", PaperRef: "LongCodeZip"},
		{ID: "5_contrastive", Name: "Contrastive Ranking", Tier: LayerTierStable, Status: "implemented", PaperRef: "LongLLMLingua"},
		{ID: "6_ngram", Name: "N-gram Abbreviation", Tier: LayerTierStable, Status: "implemented", PaperRef: "CompactPrompt"},
		{ID: "7_evaluator", Name: "Evaluator Heads", Tier: LayerTierStable, Status: "implemented", PaperRef: "EHPC"},
		{ID: "8_gist", Name: "Gist Compression", Tier: LayerTierStable, Status: "implemented", PaperRef: "Gist Tokens"},
		{ID: "9_hierarchical", Name: "Hierarchical Summary", Tier: LayerTierStable, Status: "implemented", PaperRef: "AutoCompressor"},
		{ID: "10_budget", Name: "Budget Enforcement", Tier: LayerTierStable, Status: "implemented", PaperRef: "Practical budget control"},
		{ID: "11_compaction", Name: "Compaction", Tier: LayerTierStable, Status: "implemented", PaperRef: "MemGPT"},
		{ID: "12_attribution", Name: "Attribution Filter", Tier: LayerTierStable, Status: "implemented", PaperRef: "ProCut"},
		{ID: "13_h2o", Name: "H2O Filter", Tier: LayerTierStable, Status: "implemented", PaperRef: "H2O"},
		{ID: "14_attention_sink", Name: "Attention Sink", Tier: LayerTierStable, Status: "implemented", PaperRef: "StreamingLLM"},
		{ID: "15_meta_token", Name: "Meta Token", Tier: LayerTierExperimental, Status: "implemented", PaperRef: "Meta Token (2025)"},
		{ID: "16_semantic_chunk", Name: "Semantic Chunk", Tier: LayerTierStable, Status: "implemented", PaperRef: "ChunkKV-like"},
		{ID: "17_sketch_store", Name: "Sketch Store", Tier: LayerTierStable, Status: "implemented", PaperRef: "KVReviver"},
		{ID: "18_lazy_pruner", Name: "Lazy Pruner", Tier: LayerTierStable, Status: "implemented", PaperRef: "LazyLLM"},
		{ID: "19_semantic_anchor", Name: "Semantic Anchor", Tier: LayerTierStable, Status: "implemented", PaperRef: "SAC"},
		{ID: "20_agent_memory", Name: "Agent Memory", Tier: LayerTierStable, Status: "implemented", PaperRef: "Focus-inspired"},
	}
	return implemented
}
