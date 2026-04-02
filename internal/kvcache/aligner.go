package kvcache

import (
	"crypto/sha256"
	"fmt"
)

type KVCacheAligner struct {
	prefixCache map[string]int
	hits        int
	misses      int
}

func NewKVCacheAligner() *KVCacheAligner {
	return &KVCacheAligner{
		prefixCache: make(map[string]int),
	}
}

func (a *KVCacheAligner) AnalyzePrefix(messages []string) map[string]interface{} {
	prefixes := make(map[string]int)
	for _, msg := range messages {
		for length := 10; length <= len(msg) && length <= 200; length += 10 {
			prefix := msg[:length]
			prefixes[prefix]++
		}
	}

	var bestPrefix string
	var bestCount int
	for prefix, count := range prefixes {
		if count > bestCount {
			bestPrefix = prefix
			bestCount = count
		}
	}

	return map[string]interface{}{
		"best_prefix":     bestPrefix,
		"best_prefix_len": len(bestPrefix),
		"prefix_count":    bestCount,
		"cacheable_ratio": float64(bestCount) / float64(len(messages)),
	}
}

func (a *KVCacheAligner) Fingerprint(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))[:16]
}

func (a *KVCacheAligner) CheckCache(content string) bool {
	fp := a.Fingerprint(content)
	if _, ok := a.prefixCache[fp]; ok {
		a.hits++
		return true
	}
	a.misses++
	a.prefixCache[fp] = 1
	return false
}

func (a *KVCacheAligner) HitRate() float64 {
	total := a.hits + a.misses
	if total == 0 {
		return 0
	}
	return float64(a.hits) / float64(total) * 100
}

func (a *KVCacheAligner) OptimizeMessageOrder(messages []string) []string {
	if len(messages) == 0 {
		return messages
	}

	systemMsgs := []string{}
	otherMsgs := []string{}

	for _, msg := range messages {
		if len(msg) > 0 && msg[0] == '<' {
			systemMsgs = append(systemMsgs, msg)
		} else {
			otherMsgs = append(otherMsgs, msg)
		}
	}

	result := append(systemMsgs, otherMsgs...)
	return result
}

func (a *KVCacheAligner) Stats() map[string]interface{} {
	return map[string]interface{}{
		"hits":       a.hits,
		"misses":     a.misses,
		"hit_rate":   a.HitRate(),
		"cache_size": len(a.prefixCache),
	}
}
