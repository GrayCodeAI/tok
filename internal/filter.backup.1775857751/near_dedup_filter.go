package filter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "DART: Stop Looking for Important Tokens, Duplication Matters More"
// EMNLP 2025 — Kim et al., KAIST
//
// Key finding: aggressively collapsing near-duplicate content consistently
// outperforms importance-based selection across benchmarks, because LLMs
// are hurt more by seeing the same information N times than by losing one
// "important" token.
//
// NearDedupFilter groups near-duplicate lines (within a single output)
// using SimHash fingerprints and Hamming distance, then collapses each
// cluster to its most informative representative with a count annotation.
//
// Typical wins: repeated cargo/clippy warnings, stacked log lines with
// varying file paths, duplicated test assertion messages.
type NearDedupFilter struct {
	threshold  int // max Hamming distance to treat lines as near-duplicate
	minLineLen int // lines shorter than this are never clustered
	minCluster int // minimum cluster size before collapsing
}

// NewNearDedupFilter creates a new DART-inspired near-duplicate line filter.
func NewNearDedupFilter() *NearDedupFilter {
	return &NearDedupFilter{
		threshold:  8,
		minLineLen: 20,
		minCluster: 2,
	}
}

// Name returns the filter name.
func (f *NearDedupFilter) Name() string { return "22_near_dedup" }

// Apply collapses near-duplicate lines preserving the best representative.
func (f *NearDedupFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	threshold := f.threshold
	if mode == ModeAggressive {
		threshold = 12
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 4 {
		return input, 0
	}

	type lineInfo struct {
		line string
		hash uint64
	}

	infos := make([]lineInfo, len(lines))
	for i, line := range lines {
		infos[i] = lineInfo{line: line, hash: SimHash(line)}
	}

	// Union-find for clustering
	parent := make([]int, len(lines))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(x, y int) {
		px, py := find(x), find(y)
		if px != py {
			parent[px] = py
		}
	}

	// Only compare non-empty lines of sufficient length; look-ahead capped at 50
	for i := range infos {
		if len(strings.TrimSpace(infos[i].line)) < f.minLineLen {
			continue
		}
		limit := i + 50
		if limit > len(infos) {
			limit = len(infos)
		}
		for j := i + 1; j < limit; j++ {
			if len(strings.TrimSpace(infos[j].line)) < f.minLineLen {
				continue
			}
			if HammingDistance(infos[i].hash, infos[j].hash) <= threshold {
				union(i, j)
			}
		}
	}

	// Build clusters indexed by root
	clusters := make(map[int][]int)
	for i := range infos {
		if strings.TrimSpace(infos[i].line) == "" {
			continue
		}
		root := find(i)
		clusters[root] = append(clusters[root], i)
	}

	// For each cluster of ≥ minCluster, pick best representative (longest = most specific)
	suppressed := make(map[int]bool)
	annotation := make(map[int]string) // representative idx → " [+N similar]"

	for _, members := range clusters {
		if len(members) < f.minCluster {
			continue
		}
		sort.Ints(members)
		bestIdx := members[0]
		for _, idx := range members[1:] {
			if len(infos[idx].line) > len(infos[bestIdx].line) {
				bestIdx = idx
			}
		}
		for _, idx := range members {
			if idx != bestIdx {
				suppressed[idx] = true
			}
		}
		annotation[bestIdx] = fmt.Sprintf(" [+%d similar]", len(members)-1)
	}

	var result []string
	for i, li := range infos {
		if suppressed[i] {
			continue
		}
		line := li.line
		if ann, ok := annotation[i]; ok {
			line = strings.TrimRight(line, " \t") + ann
		}
		result = append(result, line)
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// SetThreshold overrides the Hamming distance threshold.
func (f *NearDedupFilter) SetThreshold(t int) { f.threshold = t }
