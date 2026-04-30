package compressor

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// maxRegexCacheSize limits the number of compiled regex patterns kept in memory.
// This prevents unbounded growth when dynamic/unique patterns are passed.
const maxRegexCacheSize = 1000

// Pre-compiled regex patterns for performance
var (
	regexCache   = make(map[string]*regexp.Regexp)
	regexCacheMu sync.RWMutex
)

func getRegex(pattern string) *regexp.Regexp {
	regexCacheMu.RLock()
	if re, ok := regexCache[pattern]; ok {
		regexCacheMu.RUnlock()
		return re
	}
	regexCacheMu.RUnlock()

	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()
	// Double-check after acquiring write lock
	if re, ok := regexCache[pattern]; ok {
		return re
	}

	// Evict a random entry if cache is at capacity to prevent memory exhaustion.
	if len(regexCache) >= maxRegexCacheSize {
		for k := range regexCache {
			delete(regexCache, k)
			break // remove just one entry
		}
	}

	re := regexp.MustCompile(pattern)
	regexCache[pattern] = re
	return re
}

// Compress takes input text and mode and returns compressed text.
func Compress(text, mode string) (string, error) {
	switch mode {
	case "lite":
		return compressLite(text), nil
	case "full":
		return compressFull(text), nil
	case "ultra":
		return compressUltra(text), nil
	case "wenyan-lite":
		return compressWenyanLite(text), nil
	case "wenyan", "wenyan-full":
		return compressWenyanFull(text), nil
	case "wenyan-ultra":
		return compressWenyanUltra(text), nil
	default:
		return "", fmt.Errorf("unsupported mode: %s (use lite, full, ultra, wenyan-lite, wenyan, wenyan-ultra)", mode)
	}
}

func compressLite(text string) string {
	// Remove filler words, keep grammar
	rules := []struct {
		pattern string
		replace string
	}{
		// Remove filler adverbs (case-insensitive)
		{`(?i)\b(just|really|basically|actually|simply)\b`, ""},
		// Remove pleasantries (case-insensitive)
		{`(?i)\b(sure|certainly|of course|happy to|glad to)\b`, ""},
		// Remove hedging (case-insensitive)
		{`(?i)\b(perhaps|maybe|might|could|possibly)\b`, ""},
		// Trim extra spaces
		{`\s+`, " "},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}

func compressFull(text string) string {
	// Remove articles and more aggressive compression
	rules := []struct {
		pattern string
		replace string
	}{
		// Remove articles
		{`\b(the|a|an)\b`, ""},
		// Remove filler words (more aggressive)
		{`\b(just|really|basically|actually|simply|perhaps|maybe|might|could|possibly)\b`, ""},
		// Remove pleasantries
		{`\b(sure|certainly|of course|happy to|glad to|please|thank you|thanks)\b`, ""},
		// Remove verbose phrases
		{`\b(I would like to|I want to|I need to|I am going to)\b`, ""},
		{`\b(in order to|so that|such that)\b`, "to"},
		// Replace longer words with shorter synonyms
		{`\b(utilize|utilise)\b`, "use"},
		{`\b(additional)\b`, "more"},
		{`\b(implement a solution for)\b`, "fix"},
		// Trim extra spaces
		{`\s+`, " "},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}

func compressUltra(text string) string {
	// Ultra compression: telegraphic style
	rules := []struct {
		pattern string
		replace string
	}{
		// Remove articles and common words
		{`\b(the|a|an|and|or|but|so|because|if|then)\b`, ""},
		// Remove all filler
		{`\b(just|really|basically|actually|simply|perhaps|maybe|might|could|possibly|probably)\b`, ""},
		// Remove pleasantries
		{`\b(sure|certainly|of course|happy to|glad to|please|thank you|thanks|sorry)\b`, ""},
		// Replace with abbreviations
		{`\b(database)\b`, "DB"},
		{`\b(authentication)\b`, "auth"},
		{`\b(configuration)\b`, "config"},
		{`\b(request)\b`, "req"},
		{`\b(response)\b`, "res"},
		{`\b(function)\b`, "fn"},
		{`\b(implementation)\b`, "impl"},
		// Replace arrows for causality
		{`\bcauses\b|\bleads to\b|\bresults in\b`, "→"},
		{`\btherefore\b|\bthus\b|\bhence\b`, "∴"},
		// Remove unnecessary punctuation at end of words
		{`[.,;:]+\s*`, " "},
		// Trim extra spaces
		{`\s+`, " "},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}

// Wenyan modes - Classical Chinese compression
func compressWenyanLite(text string) string {
	// Semi-classical: Drop filler/hedging but keep grammar structure
	rules := []struct {
		pattern string
		replace string
	}{
		// Remove filler
		{`\b(just|really|basically|actually|simply|perhaps|maybe)\b`, ""},
		// Replace common phrases with semi-classical equivalents
		{`\bcreate a new\b`, "new"},
		{`\bmake a\b`, ""},
		{`\bthis is\b`, "this"},
		{`\bthere is\b`, "exists"},
		// Trim spaces
		{`\s+`, " "},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}

func compressWenyanFull(text string) string {
	// Maximum classical terseness with Chinese characters
	rules := []struct {
		pattern string
		replace string
	}{
		// Replace common patterns with Chinese equivalents
		{`\bnew object reference\b`, "物出新參照"},
		{`\bre-render\b`, "重繪"},
		{`\bwrap in\b`, "Wrap之"},
		{`\bdatabase\b`, "庫"},
		{`\bconfiguration\b`, "配置"},
		{`\bconnection\b`, "連接"},
		{`\breuse\b`, "復用"},
		{`\bopen\b`, "開"},
		{`\bper request\b`, "每請求"},
		{`\bskip\b`, "skip"},
		{`\boverhead\b`, "overhead"},
		// Remove articles and fillers
		{`\b(the|a|an|and|is|are|to)\b`, ""},
		// Trim spaces
		{`\s+`, " "},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}

func compressWenyanUltra(text string) string {
	// Extreme abbreviation with classical Chinese feel
	rules := []struct {
		pattern string
		replace string
	}{
		// Maximum compression
		{`\bnew\b`, "新"},
		{`\breference\b`, "參"},
		{`\bobject\b`, "物"},
		{`\bcause\b`, "致"},
		{`\bresult\b`, "果"},
		{`\buse\b`, "用"},
		{`\bfix\b`, "修"},
		{`\bbug\b`, "蟲"},
		{`\berror\b`, "錯"},
		{`\bconnection\b`, "連"},
		{`\bdatabase\b`, "庫"},
		{`\bconfiguration\b`, "配"},
		{`\bauthentication\b`, "驗"},
		// Remove all small words
		{`\b(the|a|an|and|or|is|are|to|of|in|on|at|for|with)\b`, ""},
		// Trim spaces aggressively
		{`\s+`, ""},
	}

	result := text
	for _, rule := range rules {
		result = getRegex(rule.pattern).ReplaceAllString(result, rule.replace)
	}
	return strings.TrimSpace(result)
}
