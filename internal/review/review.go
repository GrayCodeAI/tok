package review

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ReviewResult represents a code review comment
type ReviewResult struct {
	Line     int
	Severity string // 🔴 🟡 🟢
	Issue    string
	Fix      string
}

// GenerateReview analyzes git diff and returns terse review comments
func GenerateReview() ([]ReviewResult, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	diff := string(output)
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no changes to review")
	}

	return analyzeDiffForIssues(diff), nil
}

func analyzeDiffForIssues(diff string) []ReviewResult {
	var results []ReviewResult
	lines := strings.Split(diff, "\n")

	lineNum := 0
	for i, line := range lines {
		// Track line numbers in the new file
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header: @@ -old_start,old_len +new_start,new_len @@
			if match := regexp.MustCompile(`\+([0-9]+)`).FindStringSubmatch(line); len(match) > 1 {
				fmt.Sscanf(match[1], "%d", &lineNum)
			}
			continue
		}

		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}

		content := line[1:] // Remove the '+' prefix
		lineNum++

		// Check for common issues
		if issue := checkIssues(content); issue != nil {
			issue.Line = lineNum
			results = append(results, *issue)
		}

		// Limit output
		if len(results) >= 10 {
			break
		}
		_ = i
	}

	return results
}

func checkIssues(line string) *ReviewResult {
	// Check for common code issues
	patterns := []struct {
		pattern *regexp.Regexp
		sev     string
		issue   string
		fix     string
	}{
		{regexp.MustCompile(`TODO|FIXME`), "🟡", "TODO present", "Resolve before merge"},
		{regexp.MustCompile(`console\.log|fmt\.Println`), "🟡", "Debug output", "Remove debug"},
		{regexp.MustCompile(`defer.*Close\(\)`), "🟢", "Good: defer close", ""},
		{regexp.MustCompile(`if err != nil`), "🟢", "Good: error check", ""},
		{regexp.MustCompile(`panic\(`), "🔴", "Avoid panic", "Return error instead"},
		{regexp.MustCompile(`// .*bug|// .*fix`), "🟡", "Commented issue", "Address or remove"},
	}

	for _, p := range patterns {
		if p.pattern.MatchString(line) {
			// Skip good practices for actual issues only
			if p.sev == "🟢" {
				continue
			}
			return &ReviewResult{
				Severity: p.sev,
				Issue:    p.issue,
				Fix:      p.fix,
			}
		}
	}

	return nil
}

// FormatReview formats review results in terse style
func FormatReview(results []ReviewResult) string {
	if len(results) == 0 {
		return "No issues found"
	}

	var lines []string
	for _, r := range results {
		if r.Fix != "" {
			lines = append(lines, fmt.Sprintf("L%d: %s %s. %s.", r.Line, r.Severity, r.Issue, r.Fix))
		} else {
			lines = append(lines, fmt.Sprintf("L%d: %s %s.", r.Line, r.Severity, r.Issue))
		}
	}

	return strings.Join(lines, "\n")
}
