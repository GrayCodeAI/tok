package filter

import "strings"

// QualityGuardrail checks whether critical context was preserved.
type QualityGuardrail struct{}

func NewQualityGuardrail() *QualityGuardrail {
	return &QualityGuardrail{}
}

type GuardrailResult struct {
	Passed bool
	Reason string
}

func (g *QualityGuardrail) Validate(before, after string) GuardrailResult {
	checks := []struct {
		name   string
		tokens []string
	}{
		{
			name:   "errors",
			tokens: []string{"error", "failed", "panic", "exception", "traceback", "fatal"},
		},
		{
			name:   "diff_markers",
			tokens: []string{"diff --git", "@@ ", "--- ", "+++ "},
		},
		{
			name:   "test_assertions",
			tokens: []string{"assert", "expected", "actual", "FAIL", "FAILED"},
		},
		{
			name:   "file_refs",
			tokens: []string{".go:", ".ts:", ".py:", ".js:", ".java:", ".rb:", ".rs:"},
		},
	}

	b := strings.ToLower(before)
	a := strings.ToLower(after)
	for _, chk := range checks {
		if hasAny(b, chk.tokens...) && !hasAny(a, chk.tokens...) {
			return GuardrailResult{
				Passed: false,
				Reason: "dropped_" + chk.name,
			}
		}
	}
	return GuardrailResult{Passed: true}
}

func hasAny(s string, tokens ...string) bool {
	for _, t := range tokens {
		if strings.Contains(s, strings.ToLower(t)) {
			return true
		}
	}
	return false
}
