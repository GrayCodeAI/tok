package compressor

import (
	"fmt"
	"regexp"
	"strings"
)

// rule pairs a compiled regex with its replacement string.
type rule struct {
	re      *regexp.Regexp
	replace string
}

// newRules compiles a slice of (pattern, replacement) pairs into rules.
// Every pattern is automatically wrapped with (?i) for case-insensitive
// matching so behaviour is consistent across all compression modes.
func newRules(pairs ...[2]string) []rule {
	rules := make([]rule, 0, len(pairs))
	for _, p := range pairs {
		pat := p[0]
		// Add case-insensitive flag if the pattern doesn't already have it.
		if !strings.HasPrefix(pat, "(?i)") {
			pat = "(?i:" + pat + ")"
		}
		re := regexp.MustCompile(pat)
		rules = append(rules, rule{re: re, replace: p[1]})
	}
	return rules
}

var (
	liteRules   []rule
	fullRules   []rule
	ultraRules  []rule
	wenyanLite  []rule
	wenyanFull  []rule
	wenyanUltra []rule
)

func init() {
	liteRules = newRules(
		[2]string{`\b(just|really|basically|actually|simply)\b`, ""},
		[2]string{`\b(sure|certainly|of course|happy to|glad to)\b`, ""},
		[2]string{`\b(perhaps|maybe|might|could|possibly)\b`, ""},
		[2]string{`\s+`, " "},
	)

	fullRules = newRules(
		[2]string{`\b(the|a|an)\b`, ""},
		[2]string{`\b(just|really|basically|actually|simply|perhaps|maybe|might|could|possibly)\b`, ""},
		[2]string{`\b(sure|certainly|of course|happy to|glad to|please|thank you|thanks)\b`, ""},
		[2]string{`\b(I would like to|I want to|I need to|I am going to)\b`, ""},
		[2]string{`\b(in order to|so that|such that)\b`, "to"},
		[2]string{`\b(utilize|utilise)\b`, "use"},
		[2]string{`\b(additional)\b`, "more"},
		[2]string{`\b(implement a solution for)\b`, "fix"},
		[2]string{`\s+`, " "},
	)

	ultraRules = newRules(
		[2]string{`\b(the|a|an|and|or|but|so|because|if|then)\b`, ""},
		[2]string{`\b(just|really|basically|actually|simply|perhaps|maybe|might|could|possibly|probably)\b`, ""},
		[2]string{`\b(sure|certainly|of course|happy to|glad to|please|thank you|thanks|sorry)\b`, ""},
		[2]string{`\b(database)\b`, "DB"},
		[2]string{`\b(authentication)\b`, "auth"},
		[2]string{`\b(configuration)\b`, "config"},
		[2]string{`\b(request)\b`, "req"},
		[2]string{`\b(response)\b`, "res"},
		[2]string{`\b(function)\b`, "fn"},
		[2]string{`\b(implementation)\b`, "impl"},
		[2]string{`\bcauses\b|\bleads to\b|\bresults in\b`, "→"},
		[2]string{`\btherefore\b|\bthus\b|\bhence\b`, "∴"},
		[2]string{`[.,;:]+
\s*`, " "},
		[2]string{`\s+`, " "},
	)

	wenyanLite = newRules(
		[2]string{`\b(just|really|basically|actually|simply|perhaps|maybe)\b`, ""},
		[2]string{`\bcreate a new\b`, "new"},
		[2]string{`\bmake a\b`, ""},
		[2]string{`\bthis is\b`, "this"},
		[2]string{`\bthere is\b`, "exists"},
		[2]string{`\s+`, " "},
	)

	wenyanFull = newRules(
		[2]string{`\bnew object reference\b`, "物出新參照"},
		[2]string{`\bre-render\b`, "重繪"},
		[2]string{`\bwrap in\b`, "Wrap之"},
		[2]string{`\bdatabase\b`, "庫"},
		[2]string{`\bconfiguration\b`, "配置"},
		[2]string{`\bconnection\b`, "連接"},
		[2]string{`\breuse\b`, "復用"},
		[2]string{`\bopen\b`, "開"},
		[2]string{`\bper request\b`, "每請求"},
		[2]string{`\bskip\b`, "skip"},
		[2]string{`\boverhead\b`, "overhead"},
		[2]string{`\b(the|a|an|and|is|are|to)\b`, ""},
		[2]string{`\s+`, " "},
	)

	wenyanUltra = newRules(
		[2]string{`\bnew\b`, "新"},
		[2]string{`\breference\b`, "參"},
		[2]string{`\bobject\b`, "物"},
		[2]string{`\bcause\b`, "致"},
		[2]string{`\bresult\b`, "果"},
		[2]string{`\buse\b`, "用"},
		[2]string{`\bfix\b`, "修"},
		[2]string{`\bbug\b`, "蟲"},
		[2]string{`\berror\b`, "錯"},
		[2]string{`\bconnection\b`, "連"},
		[2]string{`\bdatabase\b`, "庫"},
		[2]string{`\bconfiguration\b`, "配"},
		[2]string{`\bauthentication\b`, "驗"},
		[2]string{`\b(the|a|an|and|or|is|are|to|of|in|on|at|for|with)\b`, ""},
		[2]string{`\s+`, ""},
	)
}

// Compress takes input text and mode and returns compressed text.
func Compress(text, mode string) (string, error) {
	switch mode {
	case "lite":
		return applyRules(text, liteRules), nil
	case "full":
		return applyRules(text, fullRules), nil
	case "ultra":
		return applyRules(text, ultraRules), nil
	case "wenyan-lite":
		return applyRules(text, wenyanLite), nil
	case "wenyan", "wenyan-full":
		return applyRules(text, wenyanFull), nil
	case "wenyan-ultra":
		return applyRules(text, wenyanUltra), nil
	default:
		return "", fmt.Errorf("unsupported mode: %s (use lite, full, ultra, wenyan-lite, wenyan, wenyan-ultra)", mode)
	}
}

// Backward-compatible wrappers used by tests.
func compressLite(text string) string   { return applyRules(text, liteRules) }
func compressFull(text string) string   { return applyRules(text, fullRules) }
func compressUltra(text string) string  { return applyRules(text, ultraRules) }
func compressWenyanLite(text string) string { return applyRules(text, wenyanLite) }
func compressWenyanFull(text string) string  { return applyRules(text, wenyanFull) }
func compressWenyanUltra(text string) string { return applyRules(text, wenyanUltra) }

// applyRules runs a slice of pre-compiled rules over the input text.
func applyRules(text string, rules []rule) string {
	result := text
	for _, r := range rules {
		result = r.re.ReplaceAllString(result, r.replace)
	}
	return strings.TrimSpace(result)
}
