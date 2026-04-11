package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// PathShortenFilter aliases repeated long paths/identifiers for compactness.
type PathShortenFilter struct{}

func NewPathShortenFilter() *PathShortenFilter { return &PathShortenFilter{} }

func (f *PathShortenFilter) Name() string { return "44_path_shorten" }

var (
	pathTokenPattern = regexp.MustCompile(`(?:[A-Za-z0-9_.-]+/){2,}[A-Za-z0-9_.-]+`)
	longIdentPattern = regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_]{24,}\b`)
)

func (f *PathShortenFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 8 {
		return input, 0
	}

	pathAlias := map[string]string{}
	identAlias := map[string]string{}
	pathN, identN := 1, 1
	seenPath := map[string]int{}
	seenIdent := map[string]int{}

	// First pass: count repeated candidates.
	for _, line := range lines {
		for _, p := range pathTokenPattern.FindAllString(line, -1) {
			seenPath[p]++
		}
		for _, id := range longIdentPattern.FindAllString(line, -1) {
			seenIdent[id]++
		}
	}

	var out []string
	changed := false
	for _, line := range lines {
		replaced := line
		for _, p := range pathTokenPattern.FindAllString(replaced, -1) {
			if seenPath[p] < 2 {
				continue
			}
			alias, ok := pathAlias[p]
			if !ok {
				alias = "@p" + itoa(pathN)
				pathN++
				pathAlias[p] = alias
				continue
			}
			replaced = strings.ReplaceAll(replaced, p, alias)
		}
		for _, id := range longIdentPattern.FindAllString(replaced, -1) {
			if seenIdent[id] < 2 {
				continue
			}
			alias, ok := identAlias[id]
			if !ok {
				alias = "@id" + itoa(identN)
				identN++
				identAlias[id] = alias
				continue
			}
			replaced = strings.ReplaceAll(replaced, id, alias)
		}
		if replaced != line {
			changed = true
		}
		out = append(out, replaced)
	}

	if !changed {
		return input, 0
	}

	if len(pathAlias) > 0 || len(identAlias) > 0 {
		out = append(out, "[path-shorten: aliases active]")
	}

	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}
