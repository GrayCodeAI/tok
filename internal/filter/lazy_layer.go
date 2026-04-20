package filter

import "sync"

// LazyLayer wraps a filter with lazy evaluation and memoization
type LazyLayer struct {
	filter Filter
	cache  sync.Map
}

func NewLazyLayer(f Filter) *LazyLayer {
	return &LazyLayer{filter: f}
}

func (l *LazyLayer) Apply(input string, mode Mode) (string, int) {
	key := input + string(mode)

	if cached, ok := l.cache.Load(key); ok {
		result := cached.(lazyResult)
		return result.output, result.tokens
	}

	output, tokens := l.filter.Apply(input, mode)
	l.cache.Store(key, lazyResult{output, tokens})

	return output, tokens
}

type lazyResult struct {
	output string
	tokens int
}
