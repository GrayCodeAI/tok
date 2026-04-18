package filter

// PreviewMode quick compression preview
type PreviewMode struct {
	sampleSize int
}

func NewPreviewMode(size int) *PreviewMode {
	return &PreviewMode{sampleSize: size}
}

func (pm *PreviewMode) Preview(input string) string {
	if len(input) <= pm.sampleSize {
		return input
	}
	return input[:pm.sampleSize]
}

// LayerFusion combines compatible layers
type LayerFusion struct {
	fused map[string]Filter
}

func NewLayerFusion() *LayerFusion {
	return &LayerFusion{fused: make(map[string]Filter)}
}

// GPUAccelerator placeholder for GPU support
type GPUAccelerator struct {
	enabled bool
}

func NewGPUAccelerator() *GPUAccelerator {
	return &GPUAccelerator{enabled: false}
}

// RollingHash for semantic chunking
type RollingHash struct {
	window int
	hash   uint64
}

func NewRollingHash(window int) *RollingHash {
	return &RollingHash{window: window}
}

func (rh *RollingHash) Update(b byte) {
	rh.hash = rh.hash*31 + uint64(b)
}

func (rh *RollingHash) Value() uint64 {
	return rh.hash
}

// QualityCache caches quality scores
type QualityCache struct {
	scores map[string]float64
}

func NewQualityCache() *QualityCache {
	return &QualityCache{scores: make(map[string]float64)}
}

func (qc *QualityCache) Get(key string) (float64, bool) {
	score, ok := qc.scores[key]
	return score, ok
}

func (qc *QualityCache) Set(key string, score float64) {
	qc.scores[key] = score
}
