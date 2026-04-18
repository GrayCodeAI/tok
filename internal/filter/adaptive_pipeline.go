package filter

// AdaptivePipeline selects layers based on content characteristics
type AdaptivePipeline struct {
	allLayers []Filter
	selector  *LayerSelector
}

type LayerSelector struct {
	history map[string][]int // content type -> effective layer indices
}

func NewAdaptivePipeline(layers []Filter) *AdaptivePipeline {
	return &AdaptivePipeline{
		allLayers: layers,
		selector:  &LayerSelector{history: make(map[string][]int)},
	}
}

func (ap *AdaptivePipeline) Process(input string) (string, int) {
	contentType := detectContentType(input)
	layerIndices := ap.selector.SelectLayers(contentType, len(input))
	
	output := input
	totalSaved := 0
	
	for _, idx := range layerIndices {
		if idx >= len(ap.allLayers) {
			continue
		}
		result, saved := ap.allLayers[idx].Apply(output, ModeMinimal)
		output = result
		totalSaved += saved
	}
	
	return output, totalSaved
}

func (ls *LayerSelector) SelectLayers(contentType string, inputSize int) []int {
	if cached, ok := ls.history[contentType]; ok {
		return cached
	}
	
	// Default: use first 10 layers for unknown content
	indices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	ls.history[contentType] = indices
	return indices
}

func detectContentType(input string) string {
	if len(input) == 0 {
		return "empty"
	}
	if input[0] == '{' || input[0] == '[' {
		return "json"
	}
	if hasCodePatterns(input) {
		return "code"
	}
	return "text"
}

func hasCodePatterns(s string) bool {
	return len(s) > 0 && (s[0] == '/' || s[0] == '#' || s[0] == '<')
}
