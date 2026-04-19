package filter

import (
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// Paper: "GMSA: Enhancing Context Compression via Group Merging and Layer Semantic Alignment"
// arXiv:2505.12215 — 2025
//
// GMSAFilter operates at paragraph/chunk level (blank-line separated blocks),
// complementing NearDedupFilter (line-level) and PerceptionCompressFilter (window-level).
//
// Two phases:
//  1. Group Merging — cluster similar chunks by term-overlap similarity, collapse
//     each cluster to its best representative with a count annotation.
//  2. Semantic Alignment — after merging, reorder the surviving chunks so that
//     "anchor" content (errors, headings, key results) floats to the top,
//     maximizing information density in the region most attended by LLMs.
//
// Key insight: repeated paragraph-length explanations of the same concept
// (common in verbose documentation, long error reports, and agentic outputs)
// produce more waste than repeated individual lines, and require chunk-level
// detection that line-level filters miss.
type GMSAFilter struct {
	similarityThreshold float64 // min term-overlap fraction to group chunks
	minChunkLines       int     // chunks shorter than this are never merged
	alignEnabled        bool    // whether to apply semantic alignment phase
}

// NewGMSAFilter creates a new GMSA group-merge + semantic-alignment filter.
func NewGMSAFilter() *GMSAFilter {
	return &GMSAFilter{
		similarityThreshold: 0.40, // containment-based; lower = more aggressive merging
		minChunkLines:       3,
		alignEnabled:        true,
	}
}

// Name returns the filter name.
func (f *GMSAFilter) Name() string { return "28_gmsa" }

// Apply applies group merging and semantic alignment.
func (f *GMSAFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	thresh := f.similarityThreshold
	if mode == ModeAggressive {
		thresh = 0.40
	}

	chunks := f.splitChunks(input)
	if len(chunks) < 2 {
		return input, 0
	}

	// Phase 1: Group Merging
	merged := f.groupMerge(chunks, thresh)

	// Phase 2: Semantic Alignment
	if f.alignEnabled {
		merged = f.semanticAlign(merged)
	}

	output := f.joinChunks(merged)
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

type textChunk struct {
	lines      []string
	termSet    map[string]bool
	anchor     float64 // structural importance score
	suppressed bool
	annotation string // e.g. "[+2 similar chunks merged]"
}

// splitChunks splits input on blank lines into paragraph-sized chunks.
func (f *GMSAFilter) splitChunks(input string) []*textChunk {
	lines := strings.Split(input, "\n")
	var chunks []*textChunk
	var cur []string

	flush := func() {
		if len(cur) >= f.minChunkLines {
			c := &textChunk{lines: append([]string{}, cur...)}
			c.termSet = gmsaTermSet(cur)
			c.anchor = gmsaAnchorScore(cur)
			chunks = append(chunks, c)
		} else if len(cur) > 0 {
			// short chunk: add as single-line pass-through with special flag
			c := &textChunk{lines: append([]string{}, cur...), anchor: 1.0}
			c.termSet = gmsaTermSet(cur)
			chunks = append(chunks, c)
		}
		cur = cur[:0]
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
		} else {
			cur = append(cur, line)
		}
	}
	flush()
	return chunks
}

// groupMerge clusters similar chunks and keeps best representatives.
func (f *GMSAFilter) groupMerge(chunks []*textChunk, threshold float64) []*textChunk {
	n := len(chunks)
	parent := make([]int, n)
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

	// Only consider chunks that are large enough to merge
	for i := 0; i < n; i++ {
		if len(chunks[i].lines) < f.minChunkLines {
			continue
		}
		// Look ahead up to 10 chunks
		limit := i + 10
		if limit > n {
			limit = n
		}
		for j := i + 1; j < limit; j++ {
			if len(chunks[j].lines) < f.minChunkLines {
				continue
			}
			if gmsaOverlap(chunks[i].termSet, chunks[j].termSet) >= threshold {
				union(i, j)
			}
		}
	}

	// Build cluster groups
	groups := make(map[int][]int)
	for i := range chunks {
		root := find(i)
		groups[root] = append(groups[root], i)
	}

	// For clusters with ≥ 2 members, keep the best (highest anchor score + most terms)
	for _, members := range groups {
		if len(members) < 2 {
			continue
		}
		bestIdx := members[0]
		for _, idx := range members[1:] {
			c := chunks[idx]
			best := chunks[bestIdx]
			if c.anchor > best.anchor || (c.anchor == best.anchor && len(c.termSet) > len(best.termSet)) {
				bestIdx = idx
			}
		}
		count := len(members) - 1
		chunks[bestIdx].annotation = "[+" + itoa(count) + " similar chunks merged]"
		for _, idx := range members {
			if idx != bestIdx {
				chunks[idx].suppressed = true
			}
		}
	}

	var result []*textChunk
	for _, c := range chunks {
		if !c.suppressed {
			result = append(result, c)
		}
	}
	return result
}

// semanticAlign reorders surviving chunks: anchors (errors/headings/results) first.
func (f *GMSAFilter) semanticAlign(chunks []*textChunk) []*textChunk {
	// Stable sort: high-anchor chunks move to front, low-anchor to back
	// Use insertion sort to preserve relative order within tiers
	n := len(chunks)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && chunks[j].anchor > chunks[j-1].anchor+0.5; j-- {
			chunks[j], chunks[j-1] = chunks[j-1], chunks[j]
		}
	}
	return chunks
}

// joinChunks reassembles chunks into a string with blank-line separators.
func (f *GMSAFilter) joinChunks(chunks []*textChunk) string {
	var parts []string
	for _, c := range chunks {
		block := strings.Join(c.lines, "\n")
		if c.annotation != "" {
			block += "\n" + c.annotation
		}
		parts = append(parts, block)
	}
	return strings.Join(parts, "\n\n")
}

// -- helpers --

func gmsaTermSet(lines []string) map[string]bool {
	set := make(map[string]bool)
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			set[t] = true
		}
	}
	return set
}

func gmsaOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	shared := 0
	for t := range a {
		if b[t] {
			shared++
		}
	}
	// Containment similarity: shared / min(|A|, |B|)
	// Better than Jaccard for detecting when one chunk is a paraphrase of another,
	// since paraphrases use synonyms and the smaller set is more constrained.
	smaller := len(a)
	if len(b) < smaller {
		smaller = len(b)
	}
	if smaller == 0 {
		return 0
	}
	return float64(shared) / float64(smaller)
}

func gmsaAnchorScore(lines []string) float64 {
	score := 0.0
	for _, line := range lines {
		if isErrorLine(line) || isWarningLine(line) {
			score += 3.0
		} else if isHeadingLine(line) {
			score += 2.0
		} else if isCodeLine(line) {
			score += 0.5
		}
	}
	if len(lines) > 0 {
		score /= float64(len(lines))
	}
	return score
}
