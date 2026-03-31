package linter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var rubocopCmd = &cobra.Command{
	Use:   "rubocop [args...]",
	Short: "RuboCop linter with filtered output",
	Long: `RuboCop linter with token-optimized output.

Injects --format json for structured parsing, groups offenses by
file and severity. Falls back to text parsing for autocorrect mode.

Examples:
  tokman rubocop
  tokman rubocop app/models/
  tokman rubocop -A`,
	DisableFlagParsing: true,
	RunE:               runRubocop,
}

func init() {
	registry.Add(func() { registry.Register(rubocopCmd) })
}

// JSON structures for RuboCop --format json output

type RubocopOutput struct {
	Files   []RubocopFile  `json:"files"`
	Summary RubocopSummary `json:"summary"`
}

type RubocopFile struct {
	Path     string           `json:"path"`
	Offenses []RubocopOffense `json:"offenses"`
}

type RubocopOffense struct {
	CopName     string          `json:"cop_name"`
	Severity    string          `json:"severity"`
	Message     string          `json:"message"`
	Correctable bool            `json:"correctable"`
	Location    RubocopLocation `json:"location"`
}

type RubocopLocation struct {
	StartLine int `json:"start_line"`
}

type RubocopSummary struct {
	OffenseCount            int `json:"offense_count"`
	InspectedFileCount      int `json:"inspected_file_count"`
	CorrectableOffenseCount int `json:"correctable_offense_count"`
}

func runRubocop(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Detect autocorrect mode
	isAutocorrect := false
	for _, a := range args {
		if a == "-a" || a == "-A" || a == "--auto-correct" || a == "--auto-correct-all" {
			isAutocorrect = true
			break
		}
	}

	// Detect if user already specified a format
	hasFormat := false
	for _, a := range args {
		if strings.HasPrefix(a, "--format") || a == "-f" || strings.HasPrefix(a, "-f") && len(a) > 2 {
			hasFormat = true
			break
		}
	}

	c := rubyExec("rubocop")
	if !hasFormat && !isAutocorrect {
		c.Args = append(c.Args, "--format", "json")
	}
	c.Args = append(c.Args, args...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	if stdout.String() == "" && err != nil {
		filtered = "RuboCop: FAILED (no stdout, see stderr below)\n"
		if stderr.String() != "" {
			filtered += stderr.String()
		}
	} else if hasFormat || isAutocorrect {
		filtered = filterRubocopText(stdout.String())
	} else {
		filtered = filterRubocopJSON(stdout.String())
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("rubocop %s", strings.Join(args, " ")), "tokman rubocop", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("rubocop failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("rubocop failed: %w", err)
	}
	return nil
}

// rubyExec returns a command that uses "bundle exec" if Gemfile exists.
func rubyExec(tool string) *exec.Cmd {
	if _, err := os.Stat("Gemfile"); err == nil {
		if bundlePath, err := exec.LookPath("bundle"); err == nil {
			cmd := exec.Command(bundlePath, "exec", tool)
			return cmd
		}
	}
	return exec.Command(tool)
}

// severityRank returns a sort priority: lower = more severe.
func severityRank(severity string) int {
	switch severity {
	case "fatal", "error":
		return 0
	case "warning":
		return 1
	case "convention", "refactor", "info":
		return 2
	default:
		return 3
	}
}

func filterRubocopJSON(output string) string {
	if strings.TrimSpace(output) == "" {
		return "RuboCop: No output\n"
	}

	var rubocop RubocopOutput
	if err := json.Unmarshal([]byte(output), &rubocop); err != nil {
		return fmt.Sprintf("RuboCop (JSON parse failed): %s\n", fallbackTail(output, 5))
	}

	s := rubocop.Summary
	if s.OffenseCount == 0 {
		return fmt.Sprintf("ok ✓ rubocop (%d files)\n", s.InspectedFileCount)
	}

	// Count correctable offenses
	correctableCount := s.CorrectableOffenseCount
	if correctableCount == 0 {
		for _, f := range rubocop.Files {
			for _, o := range f.Offenses {
				if o.Correctable {
					correctableCount++
				}
			}
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("rubocop: %d offenses (%d files)\n", s.OffenseCount, s.InspectedFileCount))

	// Collect files with offenses, sorted by worst severity then path
	var filesWithOffenses []RubocopFile
	for _, f := range rubocop.Files {
		if len(f.Offenses) > 0 {
			filesWithOffenses = append(filesWithOffenses, f)
		}
	}

	sort.Slice(filesWithOffenses, func(i, j int) bool {
		aWorst := 3
		for _, o := range filesWithOffenses[i].Offenses {
			if r := severityRank(o.Severity); r < aWorst {
				aWorst = r
			}
		}
		bWorst := 3
		for _, o := range filesWithOffenses[j].Offenses {
			if r := severityRank(o.Severity); r < bWorst {
				bWorst = r
			}
		}
		if aWorst != bWorst {
			return aWorst < bWorst
		}
		return filesWithOffenses[i].Path < filesWithOffenses[j].Path
	})

	maxFiles := 10
	maxOffensesPerFile := 5

	for idx, file := range filesWithOffenses {
		if idx >= maxFiles {
			break
		}

		result.WriteString(fmt.Sprintf("\n%s\n", compactRubyPath(file.Path)))

		// Sort offenses: by severity, then by line number
		sortedOffenses := make([]RubocopOffense, len(file.Offenses))
		copy(sortedOffenses, file.Offenses)
		sort.Slice(sortedOffenses, func(i, j int) bool {
			ri := severityRank(sortedOffenses[i].Severity)
			rj := severityRank(sortedOffenses[j].Severity)
			if ri != rj {
				return ri < rj
			}
			return sortedOffenses[i].Location.StartLine < sortedOffenses[j].Location.StartLine
		})

		for oi, offense := range sortedOffenses {
			if oi >= maxOffensesPerFile {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(sortedOffenses)-maxOffensesPerFile))
				break
			}
			firstLine := offense.Message
			if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
				firstLine = firstLine[:idx]
			}
			result.WriteString(fmt.Sprintf("  :%d %s — %s\n", offense.Location.StartLine, offense.CopName, firstLine))
		}
	}

	if len(filesWithOffenses) > maxFiles {
		result.WriteString(fmt.Sprintf("\n... +%d more files\n", len(filesWithOffenses)-maxFiles))
	}

	if correctableCount > 0 {
		result.WriteString(fmt.Sprintf("\n(%d correctable, run `rubocop -A`)\n", correctableCount))
	}

	return result.String()
}

func filterRubocopText(output string) string {
	// Check for Ruby/Bundler errors first
	for _, line := range strings.Split(output, "\n") {
		t := strings.TrimSpace(line)
		if strings.Contains(t, "cannot load such file") ||
			strings.Contains(t, "Bundler::GemNotFound") ||
			strings.Contains(t, "Gem::MissingSpecError") ||
			strings.HasPrefix(t, "rubocop: command not found") ||
			strings.HasPrefix(t, "rubocop: No such file") {
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) > 20 {
				return fmt.Sprintf("RuboCop error:\n%s\n... (%d more lines)\n", strings.Join(lines[:20], "\n"), len(lines)-20)
			}
			return fmt.Sprintf("RuboCop error:\n%s\n", strings.TrimSpace(output))
		}
	}

	// Detect autocorrect summary
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if strings.Contains(t, "inspected") && strings.Contains(t, "autocorrected") {
			files := extractLeadingNumber(t)
			corrected := extractAutocorrectCount(t)
			if files > 0 && corrected > 0 {
				return fmt.Sprintf("ok ✓ rubocop -A (%d files, %d autocorrected)\n", files, corrected)
			}
			return fmt.Sprintf("RuboCop: %s\n", t)
		}
		if strings.Contains(t, "inspected") && (strings.Contains(t, "offense") || strings.Contains(t, "no offenses")) {
			if strings.Contains(t, "no offenses") {
				files := extractLeadingNumber(t)
				if files > 0 {
					return fmt.Sprintf("ok ✓ rubocop (%d files)\n", files)
				}
				return "ok ✓ rubocop (no offenses)\n"
			}
			return fmt.Sprintf("RuboCop: %s\n", t)
		}
	}

	// Last resort: last 5 lines
	return fallbackTail(output, 5)
}

func extractLeadingNumber(s string) int {
	words := strings.Fields(s)
	if len(words) == 0 {
		return 0
	}
	n, err := strconv.Atoi(words[0])
	if err != nil {
		return 0
	}
	return n
}

func extractAutocorrectCount(s string) int {
	parts := strings.Split(s, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		t := strings.TrimSpace(parts[i])
		if strings.Contains(t, "autocorrected") {
			return extractLeadingNumber(t)
		}
	}
	return 0
}

func compactRubyPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	prefixes := []string{
		"app/models/", "app/controllers/", "app/views/",
		"app/helpers/", "app/services/", "app/jobs/",
		"app/mailers/", "lib/", "spec/", "test/", "config/",
	}
	for _, prefix := range prefixes {
		if idx := strings.Index(path, prefix); idx >= 0 {
			return path[idx:]
		}
	}

	if idx := strings.LastIndex(path, "/app/"); idx >= 0 {
		return path[idx+1:]
	}
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func fallbackTail(output string, n int) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= n {
		return strings.TrimSpace(output) + "\n"
	}
	return strings.Join(lines[len(lines)-n:], "\n") + "\n"
}
