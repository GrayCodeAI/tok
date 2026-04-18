package filter

// HotPathOptimizer caches frequently used paths
type HotPathOptimizer struct {
	fastCache map[string]string
	hitCount  map[string]int
	threshold int
}

func NewHotPathOptimizer(threshold int) *HotPathOptimizer {
	return &HotPathOptimizer{
		fastCache: make(map[string]string),
		hitCount:  make(map[string]int),
		threshold: threshold,
	}
}

func (hpo *HotPathOptimizer) Get(key string) (string, bool) {
	if val, ok := hpo.fastCache[key]; ok {
		hpo.hitCount[key]++
		return val, true
	}
	return "", false
}

func (hpo *HotPathOptimizer) Set(key, value string) {
	hpo.hitCount[key]++
	if hpo.hitCount[key] >= hpo.threshold {
		hpo.fastCache[key] = value
	}
}
