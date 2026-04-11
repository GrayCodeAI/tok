package filter

import "testing"

func TestLayerGate_AllModeAllowsAll(t *testing.T) {
	reg := NewLayerRegistry()
	g := NewLayerGate(LayerGateModeAll, nil, reg)
	if !g.Allows("15_meta_token") {
		t.Fatal("all mode should allow experimental layer")
	}
}

func TestLayerGate_StableOnlyBlocksExperimental(t *testing.T) {
	reg := NewLayerRegistry()
	g := NewLayerGate(LayerGateModeStableOnly, nil, reg)
	if g.Allows("15_meta_token") {
		t.Fatal("stable-only should block experimental layer by default")
	}
	if !g.Allows("13_h2o") {
		t.Fatal("stable-only should allow stable layer")
	}
}

func TestLayerGate_StableOnlyAllowList(t *testing.T) {
	reg := NewLayerRegistry()
	g := NewLayerGate(LayerGateModeStableOnly, []string{"15_meta_token"}, reg)
	if !g.Allows("15_meta_token") {
		t.Fatal("allow-listed experimental layer should be allowed")
	}
}
