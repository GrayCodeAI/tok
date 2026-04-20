package vcs

import (
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

var gtCmd = &cobra.Command{
	Use:   "gt [args...]",
	Short: "Graphite (gt) stacked PR commands with compact output",
	Long: `Execute Graphite CLI commands with compact output.

Provides specialized filtering for log, submit, sync, restack, create, and branch.

Examples:
  tok gt log
  tok gt submit
  tok gt sync`,
	DisableFlagParsing: true,
	RunE:               runGt,
}

func init() {
	registry.Add(func() { registry.Register(gtCmd) })
}

func runGt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	// Route to specialized handlers
	switch args[0] {
	case "log":
		return runGtLog(args[1:])
	case "submit":
		return runGtSubmit(args[1:])
	case "sync":
		return runGtSync(args[1:])
	case "restack":
		return runGtRestack(args[1:])
	case "create":
		return runGtCreate(args[1:])
	case "branch":
		return runGtBranch(args[1:])
	default:
		return runGtPassthrough(args)
	}
}

func runGtLog(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt log %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"log"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtLogOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_log", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt log %s", strings.Join(args, " ")), "tok gt log", originalTokens, filteredTokens)

	return err
}

func runGtSubmit(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt submit %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"submit"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtSubmitOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_submit", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt submit %s", strings.Join(args, " ")), "tok gt submit", originalTokens, filteredTokens)

	return err
}

func runGtSync(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt sync %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"sync"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtSyncOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_sync", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt sync %s", strings.Join(args, " ")), "tok gt sync", originalTokens, filteredTokens)

	return err
}

func runGtRestack(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt restack %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"restack"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtRestackOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_restack", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt restack %s", strings.Join(args, " ")), "tok gt restack", originalTokens, filteredTokens)

	return err
}

func runGtCreate(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt create %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"create"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtCreateOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_create", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt create %s", strings.Join(args, " ")), "tok gt create", originalTokens, filteredTokens)

	return err
}

func runGtBranch(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt branch %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", append([]string{"branch"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtBranchOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt_branch", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt branch %s", strings.Join(args, " ")), "tok gt branch", originalTokens, filteredTokens)

	return err
}

func runGtPassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: gt %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("gt", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGtOutput(raw)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "gt", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("gt %s", strings.Join(args, " ")), "tok gt", originalTokens, filteredTokens)

	return err
}

// Filter functions

func filterGtLogOutput(raw string) string {
	if shared.UltraCompact {
		branches := 0
		for _, line := range strings.Split(raw, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && (strings.HasPrefix(line, "│") || strings.HasPrefix(line, "├") || strings.HasPrefix(line, "└") || len(line) > 2) {
				branches++
			}
		}
		return fmt.Sprintf("%d branches in stack\n", branches)
	}

	lines := strings.Split(raw, "\n")
	var result []string
	var branches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract branch info
		if strings.HasPrefix(line, "│") || strings.HasPrefix(line, "├") || strings.HasPrefix(line, "└") {
			// Tree structure - extract branch name
			branch := strings.TrimLeft(line, "│├└─ ")
			if branch != "" {
				branches = append(branches, shared.TruncateLine(branch, 50))
			}
		} else if line != "" && len(line) > 2 {
			branches = append(branches, shared.TruncateLine(line, 50))
		}
	}

	if len(branches) > 0 {
		result = append(result, fmt.Sprintf("Stack (%d branches):", len(branches)))
		for i, b := range branches {
			if i >= 15 {
				result = append(result, fmt.Sprintf("   ... +%d more", len(branches)-15))
				break
			}
			result = append(result, fmt.Sprintf("   %s", b))
		}
		return strings.Join(result, "\n")
	}
	return raw
}

func filterGtSubmitOutput(raw string) string {
	if shared.UltraCompact {
		if strings.Contains(strings.ToLower(raw), "error") {
			return "submit failed\n"
		}
		return "submit ok\n"
	}

	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Keep important lines
		if strings.Contains(line, "PR") || strings.Contains(line, "submitted") ||
			strings.Contains(line, "created") || strings.Contains(line, "updated") ||
			strings.Contains(line, "error") || strings.Contains(line, "success") {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	if len(result) == 0 {
		return "OK Submit completed"
	}
	return strings.Join(result, "\n")
}

func filterGtSyncOutput(raw string) string {
	if shared.UltraCompact {
		if strings.Contains(strings.ToLower(raw), "error") {
			return "sync failed\n"
		}
		return "sync ok\n"
	}

	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	if len(result) == 0 {
		return "OK Sync completed"
	}
	if len(result) > 10 {
		return strings.Join(result[:10], "\n") + fmt.Sprintf("\n... (%d more)", len(result)-10)
	}
	return strings.Join(result, "\n")
}

func filterGtRestackOutput(raw string) string {
	if shared.UltraCompact {
		if strings.Contains(strings.ToLower(raw), "error") {
			return "restack failed\n"
		}
		return "restack ok\n"
	}
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	if len(result) == 0 {
		return "OK Restacked completed"
	}
	if len(result) > 10 {
		return strings.Join(result[:10], "\n") + fmt.Sprintf("\n... (%d more)", len(result)-10)
	}
	return strings.Join(result, "\n")
}

func filterGtCreateOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	if len(result) == 0 {
		return "OK Branch created"
	}
	return strings.Join(result, "\n")
}

func filterGtBranchOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 80))
		}
	}

	if len(result) > 15 {
		return strings.Join(result[:15], "\n") + fmt.Sprintf("\n... (%d more)", len(result)-15)
	}
	return strings.Join(result, "\n")
}

func filterGtOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	if len(result) > 30 {
		return strings.Join(result[:30], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-30)
	}
	return strings.Join(result, "\n")
}
