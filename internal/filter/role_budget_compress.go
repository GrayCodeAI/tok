package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// RoleBudgetFilter allocates compression budget by multi-agent role.
type RoleBudgetFilter struct {
	targetRatio float64
}

// NewRoleBudgetFilter creates a role-aware budget filter.
func NewRoleBudgetFilter() *RoleBudgetFilter {
	return &RoleBudgetFilter{targetRatio: 0.60}
}

// Name returns the filter name.
func (f *RoleBudgetFilter) Name() string { return "39_role_budget" }

// Apply keeps more lines from high-priority roles (executor/planner) and trims low-value roles.
func (f *RoleBudgetFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	turns := parseRoleTurns(lines)
	if len(turns) < 2 {
		return input, 0
	}

	targetRatio := f.targetRatio
	if mode == ModeAggressive {
		targetRatio = 0.45
	}
	targetTotal := int(math.Ceil(float64(len(lines)) * targetRatio))
	if targetTotal < 4 {
		targetTotal = 4
	}

	totalWeight := 0.0
	weighted := make([]float64, len(turns))
	for i, t := range turns {
		w := rolePriorityWeight(t.role) * float64(t.end-t.start+1)
		weighted[i] = w
		totalWeight += w
	}
	if totalWeight <= 0 {
		return input, 0
	}

	keep := make(map[int]bool, targetTotal)
	for i, t := range turns {
		quota := int(math.Ceil(weighted[i] / totalWeight * float64(targetTotal)))
		if quota < 1 {
			quota = 1
		}
		roleKeepLines(lines, t.start, t.end, quota, keep)
	}
	pruneToTarget(lines, targetTotal, keep)

	var out []string
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}
	if len(out) >= len(lines) {
		return input, 0
	}
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

type roleTurn struct {
	start int
	end   int
	role  string
}

func parseRoleTurns(lines []string) []roleTurn {
	headerIdx := make([]int, 0, 8)
	roles := make([]string, 0, 8)
	for i, line := range lines {
		if role, ok := detectRoleHeader(line); ok {
			headerIdx = append(headerIdx, i)
			roles = append(roles, role)
		}
	}
	if len(headerIdx) < 2 {
		return nil
	}
	turns := make([]roleTurn, 0, len(headerIdx))
	for i := range headerIdx {
		end := len(lines) - 1
		if i+1 < len(headerIdx) {
			end = headerIdx[i+1] - 1
		}
		turns = append(turns, roleTurn{start: headerIdx[i], end: end, role: roles[i]})
	}
	return turns
}

func detectRoleHeader(line string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(line))
	rolePrefixes := []struct {
		role   string
		prefix string
	}{
		{"user", "user:"},
		{"assistant", "assistant:"},
		{"planner", "planner:"},
		{"critic", "critic:"},
		{"executor", "executor:"},
		{"reviewer", "reviewer:"},
		{"agent", "agent:"},
		{"tool", "tool:"},
		{"system", "system:"},
	}
	for _, rp := range rolePrefixes {
		if strings.HasPrefix(lower, rp.prefix) {
			return rp.role, true
		}
	}
	return "", false
}

func rolePriorityWeight(role string) float64 {
	switch role {
	case "executor":
		return 1.35
	case "planner":
		return 1.25
	case "critic", "reviewer":
		return 1.0
	case "assistant", "agent":
		return 0.9
	case "tool":
		return 0.7
	default:
		return 1.0
	}
}

func roleKeepLines(lines []string, start, end, quota int, keep map[int]bool) {
	type cand struct {
		idx   int
		score float64
	}
	cands := make([]cand, 0, end-start+1)
	for i := start; i <= end; i++ {
		line := lines[i]
		if i == start || isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
			keep[i] = true
			continue
		}
		score := 0.0
		if isReasoningLine(line) || epicIsCausalEdge(line) {
			score += 1.0
		}
		score += float64(len(ltTokenize(line))) / 10.0
		cands = append(cands, cand{idx: i, score: score})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	for _, c := range cands {
		if quota <= 0 {
			break
		}
		if !keep[c.idx] {
			keep[c.idx] = true
			quota--
		}
	}
}

func pruneToTarget(lines []string, target int, keep map[int]bool) {
	if len(keep) <= target {
		return
	}

	isRequired := func(i int) bool {
		if _, ok := detectRoleHeader(lines[i]); ok {
			return true
		}
		return isErrorLine(lines[i]) || isWarningLine(lines[i]) || isCodeLine(lines[i])
	}

	type cand struct {
		idx   int
		score float64
	}
	cands := make([]cand, 0, len(keep))
	for i := range keep {
		if isRequired(i) {
			continue
		}
		score := 0.0
		if isReasoningLine(lines[i]) || epicIsCausalEdge(lines[i]) {
			score += 1.0
		}
		score += float64(len(ltTokenize(lines[i]))) / 10.0
		cands = append(cands, cand{idx: i, score: score})
	}

	sort.Slice(cands, func(i, j int) bool { return cands[i].score < cands[j].score })
	for _, c := range cands {
		if len(keep) <= target {
			break
		}
		delete(keep, c.idx)
	}
}

func jaccardOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	union := make(map[string]bool, len(a)+len(b))
	for k := range a {
		union[k] = true
		if b[k] {
			inter++
		}
	}
	for k := range b {
		union[k] = true
	}
	return float64(inter) / float64(len(union))
}
