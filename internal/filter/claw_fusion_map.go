package filter

// FusionStageMap describes Claw-style 14-stage coverage mapped onto TokMan layers.
type FusionStageMap struct {
	Stage    string
	LayerIDs []string
}

// ClawFusionStageCoverage returns the 14-stage compatibility mapping.
func ClawFusionStageCoverage() []FusionStageMap {
	return []FusionStageMap{
		{Stage: "QuantumLock", LayerIDs: []string{"17_semantic_cache", "43_lightmem"}},
		{Stage: "Cortex", LayerIDs: []string{"0_policy_router"}},
		{Stage: "Photon", LayerIDs: []string{"44_path_shorten"}},
		{Stage: "RLE", LayerIDs: []string{"6_ngram"}},
		{Stage: "SemanticDedup", LayerIDs: []string{"22_near_dedup", "37_latent_collab"}},
		{Stage: "Ionizer", LayerIDs: []string{"45_json_sampler"}},
		{Stage: "LogCrunch", LayerIDs: []string{"46_log_crunch"}},
		{Stage: "SearchCrunch", LayerIDs: []string{"47_search_crunch"}},
		{Stage: "DiffCrunch", LayerIDs: []string{"48_diff_crunch"}},
		{Stage: "StructuralCollapse", LayerIDs: []string{"49_structural_collapse"}},
		{Stage: "Neurosyntax", LayerIDs: []string{"4_ast_preserve"}},
		{Stage: "Nexus", LayerIDs: []string{"7_evaluator"}},
		{Stage: "TokenOpt", LayerIDs: []string{"42_plan_budget", "31_difft_adapt"}},
		{Stage: "Abbrev", LayerIDs: []string{"6_ngram", "15_meta_token"}},
	}
}
