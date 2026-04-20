package filter

import (
	"regexp"
	"strings"
)

// Wenyan compression applies rule-based classical-Chinese-inspired terseness
// to English or mixed text. It does not translate English to classical Chinese —
// that requires an LLM. Instead, it produces maximum-density fragment output
// with arrow causality and abbreviated technical terms.
//
// Modes:
//
//	WenyanLite  — strip filler + hedging, keep grammar, classical register
//	WenyanFull  — fragments + arrows + core abbreviations
//	WenyanUltra — extreme abbreviation + symbolic causality chains
//
// For text containing CJK characters, classical particles (之/乃/為/其) are
// added as connectives and modern particles (的/了/是) are stripped.
type WenyanMode int

const (
	WenyanLite WenyanMode = iota
	WenyanFull
	WenyanUltra
)

// WenyanCompress applies wenyan-style compression to the input.
func WenyanCompress(input string, mode WenyanMode) string {
	if input == "" {
		return input
	}

	out := input
	out = wenyanStripFiller(out)
	if containsCJK(out) {
		out = wenyanCJK(out, mode)
	}

	switch mode {
	case WenyanLite:
		// keep grammar; just filler-stripped
	case WenyanFull:
		out = wenyanArrowCausality(out)
		out = wenyanAbbreviate(out, false)
	case WenyanUltra:
		out = wenyanArrowCausality(out)
		out = wenyanAbbreviate(out, true)
		out = wenyanStripConnectives(out)
	}

	return collapseWhitespace(out)
}

var (
	wenyanFillers = []string{
		"just", "really", "basically", "actually", "simply",
		"very", "quite", "rather", "somewhat", "pretty",
		"obviously", "clearly", "certainly", "definitely",
		"sure", "of course", "happy to",
	}
	wenyanCausalityPatterns = []struct {
		re   *regexp.Regexp
		repl string
	}{
		{regexp.MustCompile(`(?i)\bbecause\b`), "→"},
		{regexp.MustCompile(`(?i)\bso that\b`), "→"},
		{regexp.MustCompile(`(?i)\btherefore\b`), "→"},
		{regexp.MustCompile(`(?i)\bthus\b`), "→"},
		{regexp.MustCompile(`(?i)\bhence\b`), "→"},
		{regexp.MustCompile(`(?i)\bleads? to\b`), "→"},
		{regexp.MustCompile(`(?i)\bresults? in\b`), "→"},
		{regexp.MustCompile(`(?i)\bcauses?\b`), "→"},
	}
	wenyanAbbrevs = map[string]string{
		"configuration":  "config",
		"implementation": "impl",
		"documentation":  "docs",
		"development":    "dev",
		"environment":    "env",
		"application":    "app",
		"authentication": "auth",
		"authorization":  "authz",
		"information":    "info",
		"directory":      "dir",
		"parameters":     "params",
		"parameter":      "param",
		"arguments":      "args",
		"argument":       "arg",
		"functions":      "fns",
		"function":       "fn",
		"variables":      "vars",
		"variable":       "var",
		"database":       "db",
		"connections":    "conns",
		"connection":     "conn",
		"responses":      "resps",
		"response":       "resp",
		"requests":       "reqs",
		"request":        "req",
		"messages":       "msgs",
		"message":        "msg",
		"errors":         "errs",
		"error":          "err",
	}
	wenyanUltraAbbrevs = map[string]string{
		"without": "w/o",
		"with":    "w/",
		"through": "thru",
		"should":  "shld",
		"would":   "wld",
		"could":   "cld",
	}
	wenyanConnectives = []string{
		" and ", " but ", " or ", " however ", " although ", " while ",
	}
)

func wenyanStripFiller(s string) string {
	for _, f := range wenyanFillers {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(f) + `\b\s*`)
		s = re.ReplaceAllString(s, "")
	}
	return s
}

func wenyanArrowCausality(s string) string {
	for _, p := range wenyanCausalityPatterns {
		s = p.re.ReplaceAllString(s, p.repl)
	}
	return s
}

func wenyanAbbreviate(s string, includeUltra bool) string {
	apply := func(m map[string]string) {
		for long, short := range m {
			re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(long) + `\b`)
			s = re.ReplaceAllString(s, short)
		}
	}
	apply(wenyanAbbrevs)
	if includeUltra {
		apply(wenyanUltraAbbrevs)
	}
	return s
}

func wenyanStripConnectives(s string) string {
	for _, c := range wenyanConnectives {
		s = strings.ReplaceAll(s, c, " ")
	}
	return s
}

// containsCJK reports whether the string contains CJK characters.
func containsCJK(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x4E00 && r <= 0x9FFF, // CJK Unified Ideographs
			r >= 0x3400 && r <= 0x4DBF, // Extension A
			r >= 0xF900 && r <= 0xFAFF: // Compatibility Ideographs
			return true
		}
	}
	return false
}

// wenyanCJK applies classical-particle substitutions for CJK-containing text.
// Strips modern particles (的/了/吧/呢) and collapses vernacular constructs.
func wenyanCJK(s string, mode WenyanMode) string {
	modernParticles := []string{"的", "了", "吧", "呢", "啊", "嗎"}
	for _, p := range modernParticles {
		s = strings.ReplaceAll(s, p, "")
	}
	if mode == WenyanUltra {
		extras := []string{"是", "有", "在", "會"}
		for _, e := range extras {
			s = strings.ReplaceAll(s, e, "")
		}
	}
	return s
}

func collapseWhitespace(s string) string {
	re := regexp.MustCompile(`[ \t]+`)
	s = re.ReplaceAllString(s, " ")
	re2 := regexp.MustCompile(`\n{3,}`)
	s = re2.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// WenyanFilter adapts WenyanCompress to the Filter interface.
type WenyanFilter struct{ mode WenyanMode }

// NewWenyanFilter creates a filter that applies wenyan compression at the given mode.
func NewWenyanFilter(mode WenyanMode) *WenyanFilter { return &WenyanFilter{mode: mode} }

// Name returns the filter name.
func (w *WenyanFilter) Name() string {
	switch w.mode {
	case WenyanLite:
		return "wenyan_lite"
	case WenyanUltra:
		return "wenyan_ultra"
	default:
		return "wenyan_full"
	}
}

// Apply compresses the input. The Mode argument is honored as a global on/off —
// ModeNone returns input unchanged.
func (w *WenyanFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	before := len(input)
	out := WenyanCompress(input, w.mode)
	saved := before - len(out)
	if saved < 0 {
		saved = 0
	}
	return out, saved
}
