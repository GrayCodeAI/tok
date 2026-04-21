package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
)

var reviewDiffCmd = &cobra.Command{
	Use:   "review-diff",
	Short: "Emit caveman-style one-line review comments for a diff",
	Long: `Scan a unified diff (or ` + "`git diff`" + ` output) and emit one-line
review comments with the form:

  <file>:<line> <severity> <problem>. <fix>.

Severities: 🔴 bug  🟡 risk  🔵 nit

Rule-based — no LLM. Catches common smells: TODO/FIXME, panics, ignored
errors, console.log, debugger, fmt.Println in non-main code, hard-coded
secrets patterns, > 120 char lines.`,
	RunE: runReviewDiff,
}

func runReviewDiff(cmd *cobra.Command, args []string) error {
	diff, err := readReviewInput()
	if err != nil {
		return err
	}
	findings := scanDiff(diff)
	if len(findings) == 0 {
		fmt.Println("no issues")
		return nil
	}
	for _, f := range findings {
		fmt.Println(f)
	}
	return nil
}

func readReviewInput() (string, error) {
	stat, _ := os.Stdin.Stat()
	if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		var sb strings.Builder
		s := bufio.NewScanner(os.Stdin)
		s.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for s.Scan() {
			sb.WriteString(s.Text())
			sb.WriteByte('\n')
		}
		return sb.String(), s.Err()
	}
	c := exec.Command("git", "diff", "--cached")
	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	return string(out), nil
}

type reviewRule struct {
	re       *regexp.Regexp
	severity string
	problem  string
	fix      string
}

var reviewRules = []reviewRule{
	{regexp.MustCompile(`\bTODO\b|\bFIXME\b|\bXXX\b`), "🟡 risk", "unresolved TODO/FIXME", "resolve or link issue"},
	{regexp.MustCompile(`\bpanic\(`), "🟡 risk", "panic in production code", "return error instead"},
	{regexp.MustCompile(`,\s*_\s*=\s*\w+\.(Marshal|Unmarshal|Write|Close|Read)`), "🟡 risk", "ignored error", "check and handle"},
	{regexp.MustCompile(`console\.log\(`), "🔵 nit", "console.log", "remove or use logger"},
	{regexp.MustCompile(`\bdebugger\b`), "🔴 bug", "debugger statement", "remove"},
	{regexp.MustCompile(`fmt\.Println\(`), "🔵 nit", "fmt.Println", "use structured logger"},
	{regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*["'][A-Za-z0-9_\-]{16,}`), "🔴 bug", "hard-coded credential", "load from env/secret store"},
	{regexp.MustCompile(`http://[^"'\s)]+`), "🔵 nit", "plain-HTTP URL", "use https"},
	{regexp.MustCompile(`\.only\(|\.skip\(|fdescribe\(|xdescribe\(`), "🟡 risk", "focused/skipped test", "remove before merge"},
	{regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*\}`), "🟡 risk", "empty catch", "log or rethrow"},
}

func scanDiff(diff string) []string {
	var out []string
	currentFile := ""
	newLine := 0
	inHunk := false

	lines := strings.Split(diff, "\n")
	reFileHdr := regexp.MustCompile(`^\+\+\+ b/(.+)$`)
	reHunk := regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

	for _, line := range lines {
		if m := reFileHdr.FindStringSubmatch(line); m != nil {
			currentFile = m[1]
			inHunk = false
			continue
		}
		if m := reHunk.FindStringSubmatch(line); m != nil {
			fmt.Sscanf(m[1], "%d", &newLine)
			inHunk = true
			continue
		}
		if !inHunk || currentFile == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			content := line[1:]
			for _, r := range reviewRules {
				if r.re.MatchString(content) {
					out = append(out, fmt.Sprintf("%s:%d %s %s. %s.",
						currentFile, newLine, r.severity, r.problem, r.fix))
				}
			}
			if len(content) > 120 {
				out = append(out, fmt.Sprintf("%s:%d %s line > 120 chars. wrap.",
					currentFile, newLine, "🔵 nit"))
			}
			newLine++
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			// deletion — skip line number advance
		default:
			newLine++
		}
	}
	return out
}

func init() {
	registry.Add(func() { registry.Register(reviewDiffCmd) })
}
