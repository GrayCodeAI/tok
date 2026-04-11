package filter

// SetTOMLFilter sets a TOML filter to be applied first in the pipeline.
// This is called from outside the filter package to avoid import cycles.
func (p *PipelineCoordinator) SetTOMLFilter(filter Filter, name string) {
	p.tomlFilterWrapper = filter
	p.tomlFilterName = name
}

// GetTOMLFilterName returns the name of the configured TOML filter.
func (p *PipelineCoordinator) GetTOMLFilterName() string {
	return p.tomlFilterName
}

// GetEntropyFilter returns the entropy filter
func (c *PipelineCoordinator) GetEntropyFilter() *EntropyFilter {
	return c.entropyFilter
}

// GetPerplexityFilter returns the perplexity filter
func (c *PipelineCoordinator) GetPerplexityFilter() *PerplexityFilter {
	return c.perplexityFilter
}

// GetGoalDrivenFilter returns the goal-driven filter
func (c *PipelineCoordinator) GetGoalDrivenFilter() *GoalDrivenFilter {
	return c.goalDrivenFilter
}

// GetASTPreserveFilter returns the AST preservation filter
func (c *PipelineCoordinator) GetASTPreserveFilter() *ASTPreserveFilter {
	return c.astPreserveFilter
}

// GetContrastiveFilter returns the contrastive filter
func (c *PipelineCoordinator) GetContrastiveFilter() *ContrastiveFilter {
	return c.contrastiveFilter
}

// GetNgramAbbreviator returns the N-gram abbreviator
func (c *PipelineCoordinator) GetNgramAbbreviator() *NgramAbbreviator {
	return c.ngramAbbreviator
}

// GetEvaluatorHeadsFilter returns the evaluator heads filter
func (c *PipelineCoordinator) GetEvaluatorHeadsFilter() *EvaluatorHeadsFilter {
	return c.evaluatorHeadsFilter
}

// GetGistFilter returns the gist filter
func (c *PipelineCoordinator) GetGistFilter() *GistFilter {
	return c.gistFilter
}

// GetHierarchicalSummaryFilter returns the hierarchical summary filter
func (c *PipelineCoordinator) GetHierarchicalSummaryFilter() *HierarchicalSummaryFilter {
	return c.hierarchicalSummaryFilter
}

// GetCompactionLayer returns the compaction layer
func (c *PipelineCoordinator) GetCompactionLayer() *CompactionLayer {
	return c.compactionLayer
}

// GetAttributionFilter returns the attribution filter
func (c *PipelineCoordinator) GetAttributionFilter() *AttributionFilter {
	return c.attributionFilter
}

// GetH2OFilter returns the H2O filter
func (c *PipelineCoordinator) GetH2OFilter() *H2OFilter {
	return c.h2oFilter
}

// GetAttentionSinkFilter returns the attention sink filter
func (c *PipelineCoordinator) GetAttentionSinkFilter() *AttentionSinkFilter {
	return c.attentionSinkFilter
}

// GetMetaTokenFilter returns the meta-token filter
func (c *PipelineCoordinator) GetMetaTokenFilter() *MetaTokenFilter {
	return c.metaTokenFilter
}

// GetSemanticChunkFilter returns the semantic chunk filter
func (c *PipelineCoordinator) GetSemanticChunkFilter() *SemanticChunkFilter {
	return c.semanticChunkFilter
}

// GetSketchStoreFilter returns the sketch store filter
func (c *PipelineCoordinator) GetSketchStoreFilter() *SketchStoreFilter {
	return c.sketchStoreFilter
}

// GetLazyPrunerFilter returns the lazy pruner filter
func (c *PipelineCoordinator) GetLazyPrunerFilter() *LazyPrunerFilter {
	return c.lazyPrunerFilter
}

// GetSemanticAnchorFilter returns the semantic anchor filter
func (c *PipelineCoordinator) GetSemanticAnchorFilter() *SemanticAnchorFilter {
	return c.semanticAnchorFilter
}

// GetAgentMemoryFilter returns the agent memory filter
func (c *PipelineCoordinator) GetAgentMemoryFilter() *AgentMemoryFilter {
	return c.agentMemoryFilter
}
