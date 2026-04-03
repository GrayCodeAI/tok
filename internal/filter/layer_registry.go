package filter

import "time"

// LayerMetadata describes a compression layer's requirements and behavior.
// Used by the registry to schedule and orchestrate layers.
type LayerMetadata struct {
	// ID is a unique layer identifier (e.g., "entropy", "compaction").
	ID string
	// Name is a human-readable name for stats/logging.
	Name string
	// Group controls execution ordering. Layers in the same group
	// run sequentially (output chains). Different groups can run
	// concurrently if their inputs are independent.
	Group int
	// Cost estimates this layer's CPU cost (0=cheap, 1=moderate, 2=expensive).
	Cost int
	// MinTokens required to be useful. Below this the layer returns input unchanged.
	MinTokens int
	// MinLines required to be useful.
	MinLines int
	// RequiresQuery indicates this layer needs a QueryIntent to function.
	RequiresQuery bool
	// RequiresBudget indicates this layer needs a Budget > 0 to function.
	RequiresBudget bool
	// DisabledByDefault if true, layer must be explicitly enabled in config.
	DisabledByDefault bool
	// ContentTypeRestricts limits this layer to specific content types.
	// Empty means all content types.
	ContentTypeRestricts []ContentType
}

// RegisteredLayer pairs a filter with its metadata and an optional skip check.
type RegisteredLayer struct {
	Metadata LayerMetadata
	Filter   Filter
	// SkipFn returns true if this layer should be skipped for the given input.
	SkipFn func(input string) bool
	// EnabledFn returns true if this layer is enabled by the current config.
	EnabledFn func(cfg PipelineConfig) bool
}

// LayerRegistry manages layer registration and retrieval.
// This replaces the hardcoded 40+ field PipelineCoordinator struct
// with a dynamic, extensible registration system.
type LayerRegistry struct {
	layers []RegisteredLayer
}

// NewLayerRegistry creates a registry and registers all standard layers.
func NewLayerRegistry() *LayerRegistry {
	r := &LayerRegistry{}
	r.registerDefaults()
	return r
}

// registerDefaults registers all standard compression layers.
func (r *LayerRegistry) registerDefaults() {
	// Core layers (Groups 1-2)
	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "entropy", Name: "Entropy Filter", Group: 1, Cost: 1, MinTokens: 10, DisabledByDefault: false},
		Filter:   NewEntropyFilter(),
		SkipFn:   func(input string) bool { return len(input) < 50 },
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnableEntropy },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "perplexity", Name: "Perplexity Pruning", Group: 1, Cost: 2, MinTokens: 20, MinLines: 5, DisabledByDefault: false},
		Filter:   NewPerplexityFilter(),
		SkipFn:   func(input string) bool {
			lines := 0
			for i := 0; i < len(input); i++ {
				if input[i] == '\n' {
					lines++
				}
			}
			return lines < 5
		},
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnablePerplexity },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "goal_driven", Name: "Goal-Driven Selection", Group: 2, Cost: 1, RequiresQuery: true},
		SkipFn:   nil,
		EnabledFn: func(cfg PipelineConfig) bool {
			return cfg.EnableGoalDriven && cfg.QueryIntent != ""
		},
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "ast_preserve", Name: "AST Preservation", Group: 2, Cost: 2, MinTokens: 100},
		Filter:   NewASTPreserveFilter(),
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnableAST },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "contrastive", Name: "Contrastive Ranking", Group: 2, Cost: 1, RequiresQuery: true},
		SkipFn:   nil,
		EnabledFn: func(cfg PipelineConfig) bool {
			return cfg.EnableContrastive && cfg.QueryIntent != ""
		},
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "ngram", Name: "N-gram Abbreviation", Group: 3, Cost: 1, MinTokens: 50},
		SkipFn:   func(input string) bool { return len(input) < 100 },
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.NgramEnabled },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "evaluator", Name: "Evaluator Heads", Group: 3, Cost: 1, MinTokens: 100},
		Filter:   NewEvaluatorHeadsFilter(),
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnableEvaluator },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "gist", Name: "Gist Compression", Group: 3, Cost: 2, MinTokens: 100},
		SkipFn:   func(input string) bool { return len(input) < 100 },
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnableGist },
	})

	r.register(RegisteredLayer{
		Metadata: LayerMetadata{ID: "hierarchical", Name: "Hierarchical Summary", Group: 3, Cost: 2, MinTokens: 200, MinLines: 10},
		SkipFn:   func(input string) bool {
			lines := 0
			for i := 0; i < len(input); i++ {
				if input[i] == '\n' {
					lines++
				}
			}
			return lines < 10
		},
		EnabledFn: func(cfg PipelineConfig) bool { return cfg.EnableHierarchical },
	})
}

// register adds a layer to the registry.
func (r *LayerRegistry) register(layer RegisteredLayer) {
	// Set defaults for missing filter implementations
	if layer.Metadata.MinTokens <= 0 {
		layer.Metadata.MinTokens = 10
	}
	r.layers = append(r.layers, layer)
}

// GetEnabledLayers returns all layers that should run for this config.
// Layers are returned in registration order (which defines execution order).
func (r *LayerRegistry) GetEnabledLayers(cfg PipelineConfig) []RegisteredLayer {
	var enabled []RegisteredLayer
	for _, layer := range r.layers {
		if layer.EnabledFn == nil || layer.EnabledFn(cfg) {
			enabled = append(enabled, layer)
		}
	}
	return enabled
}

// GetLayer returns a layer by ID, or nil if not found.
func (r *LayerRegistry) GetLayer(id string) *RegisteredLayer {
	for i := range r.layers {
		if r.layers[i].Metadata.ID == id {
			return &r.layers[i]
		}
	}
	return nil
}

// GetLayerIDs returns all registered layer IDs.
func (r *LayerRegistry) GetLayerIDs() []string {
	ids := make([]string, len(r.layers))
	for i, l := range r.layers {
		ids[i] = l.Metadata.ID
	}
	return ids
}

// Count returns the number of registered layers.
func (r *LayerRegistry) Count() int {
	return len(r.layers)
}

// Register adds a custom layer to the registry.
// Layers are appended to the end of the execution order.
func (r *LayerRegistry) Register(layer RegisteredLayer) {
	if layer.Metadata.ID == "" {
		layer.Metadata.ID = "custom_" + time.Now().Format("20060102_150405")
	}
	r.layers = append(r.layers, layer)
}
