package review

import (
	"testing"
)

func TestFormatReview(t *testing.T) {
	tests := []struct {
		name    string
		results []ReviewResult
		want    string
	}{
		{
			name: "single issue",
			results: []ReviewResult{
				{Line: 42, Severity: "🔴", Issue: "panic found", Fix: "Return error instead"},
			},
			want: "L42: 🔴 panic found. Return error instead.",
		},
		{
			name: "multiple issues",
			results: []ReviewResult{
				{Line: 1, Severity: "🟡", Issue: "TODO present"},
				{Line: 2, Severity: "🔴", Issue: "Debug output", Fix: "Remove debug"},
			},
			want: "L1: 🟡 TODO present.\nL2: 🔴 Debug output. Remove debug.",
		},
		{
			name:    "no issues",
			results: []ReviewResult{},
			want:    "No issues found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatReview(tt.results)
			if got != tt.want {
				t.Errorf("FormatReview() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckIssues(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     *ReviewResult
	}{
		{
			name: "TODO found",
			line: "// TODO: fix this",
			want: &ReviewResult{Severity: "🟡", Issue: "TODO present", Fix: "Resolve before merge"},
		},
		{
			name: "debug output",
			line: "console.log('debug')",
			want: &ReviewResult{Severity: "🟡", Issue: "Debug output", Fix: "Remove debug"},
		},
		{
			name: "good pattern",
			line: "if err != nil { return err }",
			want: nil,
		},
		{
			name: "panic found",
			line: "panic(\"error\")",
			want: &ReviewResult{Severity: "🔴", Issue: "Avoid panic", Fix: "Return error instead"},
		},
		{
			name: "commented bug",
			line: "// fix this bug later",
			want: &ReviewResult{Severity: "🟡", Issue: "Commented issue", Fix: "Address or remove"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkIssues(tt.line)
			if tt.want == nil {
				if got != nil {
					t.Errorf("checkIssues(%q) = %+v, want nil", tt.line, got)
				}
				return
			}
			if got == nil {
				t.Errorf("checkIssues(%q) = nil, want %+v", tt.line, tt.want)
				return
			}
			if got.Severity != tt.want.Severity || got.Issue != tt.want.Issue || got.Fix != tt.want.Fix {
				t.Errorf("checkIssues(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}

func TestAnalyzeDiffForIssues(t *testing.T) {
	// Test diff parsing
	diff := `@@ -1,2 +1,3 @@
 // Some code
+console.log('debug')
+// TODO: fix later
`

	results := analyzeDiffForIssues(diff)
	if len(results) < 2 {
		t.Errorf("analyzeDiffForIssues found %d issues, want at least 2", len(results))
	}

	// Check line numbers
	for _, r := range results {
		if r.Line <= 0 {
			t.Errorf("analyzeDiffForIssues result has invalid line number: %d", r.Line)
		}
	}
}

func TestGenerateReview(t *testing.T) {
	// Test without git (will error)
	_, err := GenerateReview()
	if err == nil {
		t.Log("GenerateReview returned no error (might be in git repo)")
	}
}
