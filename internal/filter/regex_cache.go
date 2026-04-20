package filter

import (
	"regexp"
	"sync"
)

// RegexCache caches compiled regexes
type RegexCache struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

var globalRegexCache = &RegexCache{cache: make(map[string]*regexp.Regexp)}

func CompileRegex(pattern string) *regexp.Regexp {
	globalRegexCache.mu.RLock()
	if re, ok := globalRegexCache.cache[pattern]; ok {
		globalRegexCache.mu.RUnlock()
		return re
	}
	globalRegexCache.mu.RUnlock()

	globalRegexCache.mu.Lock()
	defer globalRegexCache.mu.Unlock()

	re := regexp.MustCompile(pattern)
	globalRegexCache.cache[pattern] = re
	return re
}
