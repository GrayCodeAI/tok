package filter

import "strings"

const (
	LayerGateModeAll        = "all"
	LayerGateModeStableOnly = "stable-only"
)

// LayerGate controls layer execution policy.
type LayerGate struct {
	mode              string
	allowExperimental map[string]struct{}
	registry          *LayerRegistry
}

func NewLayerGate(mode string, allowExperimental []string, registry *LayerRegistry) *LayerGate {
	if strings.TrimSpace(mode) == "" {
		mode = LayerGateModeAll
	}
	allow := make(map[string]struct{}, len(allowExperimental))
	for _, id := range allowExperimental {
		allow[strings.TrimSpace(id)] = struct{}{}
	}
	return &LayerGate{
		mode:              mode,
		allowExperimental: allow,
		registry:          registry,
	}
}

func (g *LayerGate) Allows(layerID string) bool {
	if g == nil || g.mode == LayerGateModeAll {
		return true
	}
	meta, ok := g.registry.Get(layerID)
	if !ok {
		// Unknown layers are allowed by default to avoid breaking behavior.
		return true
	}
	if g.mode == LayerGateModeStableOnly {
		if meta.Tier == LayerTierStable || meta.Tier == LayerTierRecovery {
			return true
		}
		if _, ok := g.allowExperimental[layerID]; ok {
			return true
		}
		return false
	}
	return true
}
