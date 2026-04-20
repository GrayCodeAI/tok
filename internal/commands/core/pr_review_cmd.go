package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
)

var (
	prReviewBase string
	prReviewPR   string
)

var prReviewCmd = &cobra.Command{
	Use:   "pr-review",
	Short: "Batch review all changes in a PR or branch",
	Long: `Run review-diff logic across every file changed in a PR (or against a base
branch) and group findings by file. Rule-based — no LLM.

Modes:
  --base <branch>   review HEAD...<branch> locally (default: main)
  --pr <number>     fetch PR #N via gh and review its diff

Output format matches review-diff: one-line findings with severity badge.`,
	Example: `  tok pr-review --base main
  tok pr-review --base develop
  tok pr-review --pr 42`,
	RunE: runPRReview,
}

func runPRReview(cmd *cobra.Command, args []string) error {
	diff, err := fetchPRDiff()
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		fmt.Println("no changes to review")
		return nil
	}

	findings := scanDiff(diff)
	if len(findings) == 0 {
		fmt.Println("no issues")
		return nil
	}

	byFile := groupByFile(findings)
	files := sortedKeys(byFile)
	for _, f := range files {
		fmt.Printf("\n%s\n", f)
		fmt.Println(strings.Repeat("─", len(f)))
		for _, line := range byFile[f] {
			fmt.Println("  " + line)
		}
	}

	fmt.Printf("\n%d finding(s) across %d file(s)\n", len(findings), len(files))
	return nil
}

func fetchPRDiff() (string, error) {
	if prReviewPR != "" {
		out, err := exec.Command("gh", "pr", "diff", prReviewPR).Output()
		if err != nil {
			return "", fmt.Errorf("gh pr diff %s: %w", prReviewPR, err)
		}
		return string(out), nil
	}

	base := prReviewBase
	if base == "" {
		base = "main"
	}
	out, err := exec.Command("git", "diff", base+"...HEAD").Output()
	if err != nil {
		// Fallback: try master
		if base == "main" {
			out, err = exec.Command("git", "diff", "master...HEAD").Output()
		}
		if err != nil {
			return "", fmt.Errorf("git diff %s...HEAD: %w", base, err)
		}
	}
	return string(out), nil
}

func groupByFile(findings []string) map[string][]string {
	out := make(map[string][]string, len(findings))
	for _, f := range findings {
		// finding format: "path:line <severity> <message>"
		colonIdx := strings.IndexByte(f, ':')
		if colonIdx < 0 {
			continue
		}
		file := f[:colonIdx]
		// strip file prefix from finding to reduce noise in grouped view
		rest := f[colonIdx+1:]
		out[file] = append(out[file], rest)
	}
	return out
}

func sortedKeys(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// insertion sort — 2-30 files typical
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

func init() {
	registry.Add(func() { registry.Register(prReviewCmd) })
	prReviewCmd.Flags().StringVar(&prReviewBase, "base", "main", "base branch to diff against")
	prReviewCmd.Flags().StringVar(&prReviewPR, "pr", "", "GitHub PR number (uses gh CLI)")
}
