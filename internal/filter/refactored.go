package filter

// CoreFilters manages layers 1-9
type CoreFilters struct {
	entropy      *EntropyFilter
	perplexity   *PerplexityFilter
	goalDriven   *GoalDrivenFilter
	ast          *ASTPreserveFilter
	contrastive  *ContrastiveFilter
	ngram        *NgramAbbreviator
	evaluator    *EvaluatorHeadsFilter
	gist         *GistFilter
	hierarchical *HierarchicalSummaryFilter
}

// NewCoreFilters creates core filter set
func NewCoreFilters(cfg PipelineConfig) *CoreFilters {
	c := &CoreFilters{}
	
	if cfg.EnableEntropy {
		c.entropy = NewEntropyFilter()
	}
	if cfg.EnablePerplexity {
		c.perplexity = NewPerplexityFilter()
	}
	if cfg.EnableGoalDriven && cfg.QueryIntent != "" {
		c.goalDriven = NewGoalDrivenFilter(cfg.QueryIntent)
	}
	if cfg.EnableAST {
		c.ast = NewASTPreserveFilter()
	}
	if cfg.EnableContrastive && cfg.QueryIntent != "" {
		c.contrastive = NewContrastiveFilter(cfg.QueryIntent)
	}
	if cfg.NgramEnabled {
		c.ngram = NewNgramAbbreviator()
	}
	if cfg.EnableEvaluator {
		c.evaluator = NewEvaluatorHeadsFilter()
	}
	if cfg.EnableGist {
		c.gist = NewGistFilter()
	}
	if cfg.EnableHierarchical {
		c.hierarchical = NewHierarchicalSummaryFilter()
	}
	
	return c
}

// Apply applies all core filters
func (c *CoreFilters) Apply(input string, mode Mode, stats *PipelineStats) string {
	output := input
	
	if c.entropy != nil {
		output = applyFilter(c.entropy, output, mode, stats)
	}
	if c.perplexity != nil {
		output = applyFilter(c.perplexity, output, mode, stats)
	}
	if c.goalDriven != nil {
		output = applyFilter(c.goalDriven, output, mode, stats)
	}
	if c.ast != nil {
		output = applyFilter(c.ast, output, mode, stats)
	}
	if c.contrastive != nil {
		output = applyFilter(c.contrastive, output, mode, stats)
	}
	if c.ngram != nil {
		output = applyFilter(c.ngram, output, mode, stats)
	}
	if c.evaluator != nil {
		output = applyFilter(c.evaluator, output, mode, stats)
	}
	if c.gist != nil {
		output = applyFilter(c.gist, output, mode, stats)
	}
	if c.hierarchical != nil {
		output = applyFilter(c.hierarchical, output, mode, stats)
	}
	
	return output
}

// SemanticFilters manages layers 11-20
type SemanticFilters struct {
	compaction    *CompactionLayer
	attribution   *AttributionFilter
	h2o           *H2OFilter
	attentionSink *AttentionSinkFilter
	metaToken     *MetaTokenFilter
	semanticChunk *SemanticChunkFilter
	sketchStore   *SketchStoreFilter
	lazyPruner    *LazyPrunerFilter
	semanticAnchor *SemanticAnchorFilter
	agentMemory   *AgentMemoryFilter
}

// NewSemanticFilters creates semantic filter set
func NewSemanticFilters(cfg PipelineConfig) *SemanticFilters {
	s := &SemanticFilters{}
	
	if cfg.EnableCompaction {
		s.compaction = NewCompactionLayer(DefaultCompactionConfig())
	}
	if cfg.EnableAttribution {
		s.attribution = NewAttributionFilter(DefaultAttributionConfig())
	}
	if cfg.EnableH2O {
		s.h2o = NewH2OFilter(DefaultH2OConfig())
	}
	if cfg.EnableAttentionSink {
		s.attentionSink = NewAttentionSinkFilter(DefaultSinkConfig())
	}
	if cfg.EnableMetaToken {
		s.metaToken = NewMetaTokenFilter(DefaultMetaTokenConfig())
	}
	if cfg.EnableSemanticChunk {
		s.semanticChunk = NewSemanticChunkFilter(DefaultSemanticChunkConfig())
	}
	if cfg.EnableSketchStore {
		s.sketchStore = NewSketchStoreFilter(DefaultSketchStoreConfig())
	}
	if cfg.EnableLazyPruner {
		s.lazyPruner = NewLazyPrunerFilter(DefaultLazyPrunerConfig())
	}
	if cfg.EnableSemanticAnchor {
		s.semanticAnchor = NewSemanticAnchorFilter(DefaultSemanticAnchorConfig())
	}
	if cfg.EnableAgentMemory {
		s.agentMemory = NewAgentMemoryFilter(DefaultAgentMemoryConfig())
	}
	
	return s
}

// Apply applies all semantic filters
func (s *SemanticFilters) Apply(input string, mode Mode, stats *PipelineStats) string {
	output := input
	
	if s.compaction != nil {
		output = applyFilter(s.compaction, output, mode, stats)
	}
	if s.attribution != nil {
		output = applyFilter(s.attribution, output, mode, stats)
	}
	if s.h2o != nil {
		output = applyFilter(s.h2o, output, mode, stats)
	}
	if s.attentionSink != nil {
		output = applyFilter(s.attentionSink, output, mode, stats)
	}
	if s.metaToken != nil {
		output = applyFilter(s.metaToken, output, mode, stats)
	}
	if s.semanticChunk != nil {
		output = applyFilter(s.semanticChunk, output, mode, stats)
	}
	if s.sketchStore != nil {
		output = applyFilter(s.sketchStore, output, mode, stats)
	}
	if s.lazyPruner != nil {
		output = applyFilter(s.lazyPruner, output, mode, stats)
	}
	if s.semanticAnchor != nil {
		output = applyFilter(s.semanticAnchor, output, mode, stats)
	}
	if s.agentMemory != nil {
		output = applyFilter(s.agentMemory, output, mode, stats)
	}
	
	return output
}

// applyFilter is a helper to apply a filter and record stats
func applyFilter(f Filter, input string, mode Mode, stats *PipelineStats) string {
	output, saved := f.Apply(input, mode)
	if stats != nil {
		stats.LayerStats[f.Name()] = LayerStat{
			TokensSaved: saved,
		}
	}
	return output
}

// RefactoredCoordinator is the new simplified coordinator
type RefactoredCoordinator struct {
	config   PipelineConfig
	core     *CoreFilters
	semantic *SemanticFilters
	budget   *BudgetEnforcer
	cache    *LayerCache
}

// NewRefactoredCoordinator creates a refactored coordinator
func NewRefactoredCoordinator(cfg PipelineConfig) *RefactoredCoordinator {
	return &RefactoredCoordinator{
		config:   cfg,
		core:     NewCoreFilters(cfg),
		semantic: NewSemanticFilters(cfg),
		budget:   NewBudgetEnforcer(),
		cache:    GetGlobalLayerCache(),
	}
}

// Process processes input through the pipeline
func (r *RefactoredCoordinator) Process(input string) (string, *PipelineStats) {
	stats := &PipelineStats{
		OriginalTokens: EstimateTokens(input),
		LayerStats:     make(map[string]LayerStat),
	}
	
	output := input
	
	// Core filters (1-9)
	output = r.core.Apply(output, r.config.Mode, stats)
	
	// Budget enforcement (10)
	if r.config.Budget > 0 {
		r.budget.SetBudget(r.config.Budget)
		output = applyFilter(r.budget, output, r.config.Mode, stats)
	}
	
	// Semantic filters (11-20)
	output = r.semantic.Apply(output, r.config.Mode, stats)
	
	stats.FinalTokens = EstimateTokens(output)
	stats.TotalSaved = stats.OriginalTokens - stats.FinalTokens
	
	return output, stats
}
