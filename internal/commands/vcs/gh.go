package vcs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var ghCmd = &cobra.Command{
	Use:   "gh [args...]",
	Short: "GitHub CLI with token-optimized output",
	Long: `Execute GitHub CLI commands with compact output.

Provides specialized filtering for pr, issue, run, and repo commands.

Examples:
  tok gh pr list
  tok gh issue list --repo owner/repo
  tok gh run list`,
	DisableFlagParsing: true,
	RunE:               runGh,
}

func init() {
	registry.Add(func() { registry.Register(ghCmd) })
}

func runGh(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	// Route to specialized handlers
	switch args[0] {
	case "pr":
		return runGhPr(args[1:])
	case "issue":
		return runGhIssue(args[1:])
	case "run":
		return runGhRun(args[1:])
	case "repo":
		return runGhRepo(args[1:])
	case "release":
		return runGhRelease(args[1:])
	case "api":
		return runGhApi(args[1:])
	default:
		return runGhPassthrough(args)
	}
}

func runGhPr(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh pr %s\n", strings.Join(args, " "))
	}

	// Add --json for structured output if listing
	if len(args) > 0 && args[0] == "list" {
		args = append(args, "--json", "number,title,author,headRefName,state")
	} else if len(args) > 0 && args[0] == "view" {
		// Add JSON fields for view command
		args = append(args, "--json", "number,title,author,state,headRefName,baseRefName,additions,deletions,changedFiles,mergeable,mergeStateStatus,commits,files,statusCheckRollup")
	}

	execCmd := exec.Command("gh", append([]string{"pr"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhPrOutput(raw, args)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_pr", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh pr %s", strings.Join(args, " ")), "tok gh pr", originalTokens, filteredTokens)

	return err
}

func runGhIssue(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh issue %s\n", strings.Join(args, " "))
	}

	if len(args) > 0 && args[0] == "list" {
		args = append(args, "--json", "number,title,author,state")
	}

	execCmd := exec.Command("gh", append([]string{"issue"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhIssueOutput(raw, args)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_issue", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh issue %s", strings.Join(args, " ")), "tok gh issue", originalTokens, filteredTokens)

	return err
}

func runGhRun(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh run %s\n", strings.Join(args, " "))
	}

	// Add JSON output for list command
	if len(args) > 0 && args[0] == "list" {
		args = append(args, "--json", "databaseId,displayTitle,status,conclusion,createdAt,event")
	}

	execCmd := exec.Command("gh", append([]string{"run"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhRunOutput(raw, args)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_run", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh run %s", strings.Join(args, " ")), "tok gh run", originalTokens, filteredTokens)

	return err
}

func runGhRepo(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh repo %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gh", append([]string{"repo"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhRepoOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_repo", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh repo %s", strings.Join(args, " ")), "tok gh repo", originalTokens, filteredTokens)

	return err
}

func runGhPassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gh", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh %s", strings.Join(args, " ")), "tok gh", originalTokens, filteredTokens)

	return err
}

// Filter functions

type GhPR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	HeadRefName string `json:"headRefName"`
	State       string `json:"state"`
}

type GhPRView struct {
	Number            int          `json:"number"`
	Title             string       `json:"title"`
	Author            string       `json:"author"`
	State             string       `json:"state"`
	HeadRefName       string       `json:"headRefName"`
	BaseRefName       string       `json:"baseRefName"`
	Additions         int          `json:"additions"`
	Deletions         int          `json:"deletions"`
	ChangedFiles      int          `json:"changedFiles"`
	Mergeable         string       `json:"mergeable"`
	MergeStateStatus  string       `json:"mergeStateStatus"`
	Commits           int          `json:"commits"`
	Files             []GhPRFile   `json:"files"`
	StatusCheckRollup []GhCheckRun `json:"statusCheckRollup"`
}

type GhPRFile struct {
	Path string `json:"path"`
}

type GhCheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

func filterGhPrOutput(raw string, args []string) string {
	if shared.UltraCompact {
		if len(args) > 0 && args[0] == "list" {
			var prs []GhPR
			if err := json.Unmarshal([]byte(raw), &prs); err == nil {
				openCount := 0
				mergedCount := 0
				for _, pr := range prs {
					if pr.State == "OPEN" {
						openCount++
					} else if pr.State == "MERGED" {
						mergedCount++
					}
				}
				return fmt.Sprintf("%d PRs: %d open %d merged\n", len(prs), openCount, mergedCount)
			}
		}
		if len(args) > 0 && args[0] == "view" {
			var pr GhPRView
			if err := json.Unmarshal([]byte(raw), &pr); err == nil {
				state := "open"
				if pr.State == "MERGED" {
					state = "merged"
				} else if pr.State == "CLOSED" {
					state = "closed"
				}
				return fmt.Sprintf("#%d %s [%s] +%d -%d\n", pr.Number, shared.TruncateLine(pr.Title, 30), state, pr.Additions, pr.Deletions)
			}
		}
	}

	// Try JSON parsing for list command
	if len(args) > 0 && args[0] == "list" {
		var prs []GhPR
		if err := json.Unmarshal([]byte(raw), &prs); err == nil {
			var result []string
			result = append(result, fmt.Sprintf("Pull Requests (%d):", len(prs)))
			for i, pr := range prs {
				if i >= 15 {
					result = append(result, fmt.Sprintf("   ... +%d more", len(prs)-15))
					break
				}
				state := "○"
				if pr.State == "OPEN" {
					state = "●"
				} else if pr.State == "MERGED" {
					state = "✓"
				}
				result = append(result, fmt.Sprintf("   %s #%d: %s (%s)", state, pr.Number, shared.TruncateLine(pr.Title, 50), pr.Author))
			}
			return strings.Join(result, "\n")
		}
	}

	// Handle pr view command
	if len(args) > 0 && args[0] == "view" {
		var pr GhPRView
		if err := json.Unmarshal([]byte(raw), &pr); err == nil {
			return formatGhPRView(pr)
		}
	}

	return raw
}

func formatGhPRView(pr GhPRView) string {
	var result []string

	// Header with PR number and title
	state := "○"
	if pr.State == "OPEN" {
		state = "●"
	} else if pr.State == "MERGED" {
		state = "✓"
	} else if pr.State == "CLOSED" {
		state = "✗"
	}
	result = append(result, fmt.Sprintf("%s PR #%d: %s", state, pr.Number, shared.TruncateLine(pr.Title, 60)))

	// Branch info
	result = append(result, fmt.Sprintf("   %s → %s", pr.HeadRefName, pr.BaseRefName))

	// Author and stats
	result = append(result, fmt.Sprintf("   by %s | +%d -%d in %d files | %d commits", pr.Author, pr.Additions, pr.Deletions, pr.ChangedFiles, pr.Commits))

	// Merge status
	mergeStatus := "? unknown"
	if pr.Mergeable == "MERGEABLE" {
		mergeStatus = "OK mergeable"
	} else if pr.Mergeable == "CONFLICTING" {
		mergeStatus = "FAIL conflicts"
	} else if pr.Mergeable == "UNKNOWN" {
		mergeStatus = "...checking"
	}
	result = append(result, fmt.Sprintf("   Merge: %s", mergeStatus))

	// CI/Checks
	if len(pr.StatusCheckRollup) > 0 {
		passed := 0
		failed := 0
		pending := 0
		for _, check := range pr.StatusCheckRollup {
			if check.Status == "COMPLETED" {
				if check.Conclusion == "SUCCESS" {
					passed++
				} else if check.Conclusion == "FAILURE" {
					failed++
				}
			} else {
				pending++
			}
		}
		checks := fmt.Sprintf("   Checks: %d PASS", passed)
		if failed > 0 {
			checks += fmt.Sprintf(" %d FAIL", failed)
		}
		if pending > 0 {
			checks += fmt.Sprintf(" %d pending", pending)
		}
		result = append(result, checks)
	}

	// Files changed (show first 10)
	if len(pr.Files) > 0 {
		result = append(result, fmt.Sprintf("   Files (%d):", len(pr.Files)))
		for i, f := range pr.Files {
			if i >= 10 {
				result = append(result, fmt.Sprintf("      ... +%d more", len(pr.Files)-10))
				break
			}
			result = append(result, fmt.Sprintf("      %s", shared.TruncateLine(f.Path, 60)))
		}
	}

	return strings.Join(result, "\n")
}

type GhIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Author string `json:"author"`
	State  string `json:"state"`
}

func filterGhIssueOutput(raw string, args []string) string {
	if shared.UltraCompact {
		if len(args) > 0 && args[0] == "list" {
			var issues []GhIssue
			if err := json.Unmarshal([]byte(raw), &issues); err == nil {
				openCount := 0
				closedCount := 0
				for _, issue := range issues {
					if issue.State == "OPEN" {
						openCount++
					} else {
						closedCount++
					}
				}
				return fmt.Sprintf("%d issues: %d open %d closed\n", len(issues), openCount, closedCount)
			}
		}
	}

	if len(args) > 0 && args[0] == "list" {
		var issues []GhIssue
		if err := json.Unmarshal([]byte(raw), &issues); err == nil {
			var result []string
			result = append(result, fmt.Sprintf("Issues (%d):", len(issues)))
			for i, issue := range issues {
				if i >= 15 {
					result = append(result, fmt.Sprintf("   ... +%d more", len(issues)-15))
					break
				}
				state := "○"
				if issue.State == "OPEN" {
					state = "●"
				} else if issue.State == "CLOSED" {
					state = "✓"
				}
				result = append(result, fmt.Sprintf("   %s #%d: %s (%s)", state, issue.Number, shared.TruncateLine(issue.Title, 50), issue.Author))
			}
			return strings.Join(result, "\n")
		}
	}
	return raw
}

type GhRun struct {
	DatabaseId   int    `json:"databaseId"`
	DisplayTitle string `json:"displayTitle"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	CreatedAt    string `json:"createdAt"`
	Event        string `json:"event"`
}

func filterGhRunOutput(raw string, args []string) string {
	if shared.UltraCompact {
		if len(args) > 0 && args[0] == "list" {
			var runs []GhRun
			if err := json.Unmarshal([]byte(raw), &runs); err == nil {
				successCount := 0
				failCount := 0
				pendingCount := 0
				for _, run := range runs {
					if run.Status == "completed" && run.Conclusion == "success" {
						successCount++
					} else if run.Status == "completed" && run.Conclusion == "failure" {
						failCount++
					} else {
						pendingCount++
					}
				}
				return fmt.Sprintf("%d runs: %d ok %d fail %d pending\n", len(runs), successCount, failCount, pendingCount)
			}
		}
	}

	// Try JSON parsing for list command
	if len(args) > 0 && args[0] == "list" {
		var runs []GhRun
		if err := json.Unmarshal([]byte(raw), &runs); err == nil {
			var result []string
			result = append(result, fmt.Sprintf("Workflow Runs (%d):", len(runs)))
			for i, run := range runs {
				if i >= 15 {
					result = append(result, fmt.Sprintf("   ... +%d more", len(runs)-15))
					break
				}
				status := "○"
				if run.Status == "completed" {
					if run.Conclusion == "success" {
						status = "OK"
					} else if run.Conclusion == "failure" {
						status = "FAIL"
					} else {
						status = "WARN"
					}
				} else if run.Status == "in_progress" {
					status = "..."
				} else if run.Status == "queued" {
					status = "queued"
				}
				result = append(result, fmt.Sprintf("   %s #%d: %s (%s)", status, run.DatabaseId, shared.TruncateLine(run.DisplayTitle, 40), run.Event))
			}
			return strings.Join(result, "\n")
		}
	}

	// Fallback for non-list or parse errors
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, shared.TruncateLine(line, 100))
	}

	if len(result) > 20 {
		return strings.Join(result[:20], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-20)
	}
	return strings.Join(result, "\n")
}

func runGhRelease(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh release %s\n", strings.Join(args, " "))
	}

	// Add JSON output for list command
	if len(args) > 0 && args[0] == "list" {
		args = append(args, "--json", "tagName,name,createdAt,isDraft,isPrerelease")
	}

	execCmd := exec.Command("gh", append([]string{"release"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGhReleaseOutput(raw, args)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_release", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh release %s", strings.Join(args, " ")), "tok gh release", originalTokens, filteredTokens)

	return err
}

type GhRelease struct {
	TagName      string `json:"tagName"`
	Name         string `json:"name"`
	CreatedAt    string `json:"createdAt"`
	IsDraft      bool   `json:"isDraft"`
	IsPrerelease bool   `json:"isPrerelease"`
}

func filterGhReleaseOutput(raw string, args []string) string {
	// Try JSON parsing for list command
	if len(args) > 0 && args[0] == "list" {
		var releases []GhRelease
		if err := json.Unmarshal([]byte(raw), &releases); err == nil {
			var result []string
			result = append(result, fmt.Sprintf("Releases (%d):", len(releases)))
			for i, rel := range releases {
				if i >= 15 {
					result = append(result, fmt.Sprintf("   ... +%d more", len(releases)-15))
					break
				}
				status := "OK"
				if rel.IsDraft {
					status = "draft"
				} else if rel.IsPrerelease {
					status = "pre"
				}
				name := rel.Name
				if name == "" {
					name = rel.TagName
				}
				result = append(result, fmt.Sprintf("   %s %s (%s)", status, rel.TagName, shared.TruncateLine(name, 40)))
			}
			return strings.Join(result, "\n")
		}
	}

	// Fallback
	lines := strings.Split(raw, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}
	if len(result) > 20 {
		return strings.Join(result[:20], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-20)
	}
	return strings.Join(result, "\n")
}

func runGhApi(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gh api %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gh", append([]string{"api"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Try to parse as JSON and show structure
	filtered := filterGhApiOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gh_api", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gh api %s", strings.Join(args, " ")), "tok gh api", originalTokens, filteredTokens)

	return err
}

func filterGhApiOutput(raw string) string {
	// Try to detect JSON and show compact structure
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		// Use the JSON structure filter
		schema := shared.TryJSONSchema(trimmed, 10)
		if schema != "" {
			return "API Response:\n" + schema
		}
	}

	// Fallback: compact output
	lines := strings.Split(raw, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 120))
		}
	}
	if len(result) > 30 {
		return strings.Join(result[:30], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-30)
	}
	return strings.Join(result, "\n")
}

func filterGhRepoOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	if len(result) > 15 {
		return strings.Join(result[:15], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-15)
	}
	return strings.Join(result, "\n")
}

func filterGhOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 120))
		}
	}

	if len(result) > 30 {
		return strings.Join(result[:30], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-30)
	}
	return strings.Join(result, "\n")
}
