package filter

import (
	"reflect"
	"strings"
	"testing"
)

type researchTestFilter struct {
	tag   string
	saved int
	calls *[]string
}

func (f researchTestFilter) Name() string { return f.tag }

func (f researchTestFilter) Apply(input string, _ Mode) (string, int) {
	*f.calls = append(*f.calls, f.tag)
	return input + "|" + f.tag, f.saved
}

func TestProcessResearchLayers_OrderIncludes31To43(t *testing.T) {
	calls := make([]string, 0, 23)
	p := &PipelineCoordinator{
		config: PipelineConfig{Mode: ModeMinimal, SessionTracking: true},
		layers: make([]filterLayer, 42),
	}

	// Non-nil pointers gate execution in processResearchLayers.
	p.marginalInfoGainFilter = &MarginalInfoGainFilter{}
	p.nearDedupFilter = &NearDedupFilter{}
	p.cotCompressFilter = &CoTCompressFilter{}
	p.codingAgentCtxFilter = &CodingAgentContextFilter{}
	p.perceptionCompressFilter = &PerceptionCompressFilter{}
	p.lightThinkerFilter = &LightThinkerFilter{}
	p.thinkSwitcherFilter = &ThinkSwitcherFilter{}
	p.gmsaFilter = &GMSAFilter{}
	p.carlFilter = &CARLFilter{}
	p.slimInferFilter = &SlimInferFilter{}
	p.diffAdaptFilter = &DiffAdaptFilter{}
	p.epicFilter = &EPiCFilter{}
	p.ssdpFilter = &SSDPFilter{}
	p.agentOCRFilter = &AgentOCRFilter{}
	p.s2madFilter = &S2MADFilter{}
	p.aconFilter = &ACONFilter{}
	p.latentCollabFilter = &LatentCollabFilter{}
	p.graphCoTFilter = &GraphCoTFilter{}
	p.roleBudgetFilter = &RoleBudgetFilter{}
	p.sweAdaptiveLoop = &SWEAdaptiveLoopFilter{}
	p.agentOCRHistory = &AgentOCRHistoryFilter{}
	p.planBudgetFilter = &PlanBudgetFilter{}
	p.lightMemFilter = &LightMemFilter{}

	names := []string{
		"21_marginal_info_gain", "22_near_dedup", "23_cot_compress", "24_coding_agent_ctx", "25_perception_compress",
		"26_lightthinker", "27_think_switcher", "28_gmsa", "29_carl", "30_slim_infer",
		"31_difft_adapt", "32_epic", "33_ssdp", "34_agent_ocr", "35_s2_mad",
		"36_acon",
		"37_latent_collab", "38_graph_cot", "39_role_budget",
		"40_swe_adaptive_loop", "41_agent_ocr_history", "42_plan_budget", "43_lightmem",
	}
	for i, name := range names {
		p.layers[19+i] = filterLayer{
			filter: researchTestFilter{tag: name, saved: 1, calls: &calls},
			name:   name,
		}
	}

	stats := &PipelineStats{OriginalTokens: 1000, LayerStats: map[string]LayerStat{}}
	out := p.processResearchLayers("start", stats)

	if !reflect.DeepEqual(calls, names) {
		t.Fatalf("unexpected order:\n got=%v\nwant=%v", calls, names)
	}
	for _, name := range names {
		if _, ok := stats.LayerStats[name]; !ok {
			t.Fatalf("missing layer stat for %s", name)
		}
		if !strings.Contains(out, name) {
			t.Fatalf("output missing marker for %s", name)
		}
	}
}

func TestProcessResearchLayers_EarlyExitStopsAfterBudgetMet(t *testing.T) {
	calls := make([]string, 0, 3)
	p := &PipelineCoordinator{
		config: PipelineConfig{
			Mode:            ModeMinimal,
			SessionTracking: true,
			Budget:          10,
		},
		layers: make([]filterLayer, 34),
	}

	p.marginalInfoGainFilter = &MarginalInfoGainFilter{}
	p.nearDedupFilter = &NearDedupFilter{}
	p.cotCompressFilter = &CoTCompressFilter{}

	p.layers[19] = filterLayer{
		filter: researchTestFilter{tag: "21_marginal_info_gain", saved: 95, calls: &calls},
		name:   "21_marginal_info_gain",
	}
	p.layers[20] = filterLayer{
		filter: researchTestFilter{tag: "22_near_dedup", saved: 1, calls: &calls},
		name:   "22_near_dedup",
	}
	p.layers[21] = filterLayer{
		filter: researchTestFilter{tag: "23_cot_compress", saved: 1, calls: &calls},
		name:   "23_cot_compress",
	}

	stats := &PipelineStats{OriginalTokens: 100, LayerStats: map[string]LayerStat{}}
	_ = p.processResearchLayers("start", stats)

	want := []string{"21_marginal_info_gain"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("early exit failed, calls=%v want=%v", calls, want)
	}
}
