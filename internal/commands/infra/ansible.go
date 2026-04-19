package infra

import (
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

func atoi(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		n = 0
	}
	return n
}

var ansibleCmd = &cobra.Command{
	Use:   "ansible [subcommand] [args...]",
	Short: "Ansible CLI with compact output",
	Long: `Ansible CLI with token-optimized output.

Specialized filters for common commands:
  - ansible-playbook: Play summary with task results
  - ansible-inventory: Compact inventory listing
  - ansible ad-hoc: Compact ad-hoc output

Examples:
  tok ansible-playbook site.yml
  tok ansible all -m ping
  tok ansible-inventory --list`,
	DisableFlagParsing: true,
	RunE:               runAnsible,
}

func init() {
	registry.Add(func() {
		registry.Register(ansibleCmd)
		registry.Register(ansiblePlaybookCmd)
		registry.Register(ansibleInventoryCmd)
	})
}

func runAnsible(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return runAnsiblePassthrough(args)
}

var ansiblePlaybookCmd = &cobra.Command{
	Use:   "ansible-playbook [args...]",
	Short: "Ansible playbook with compact output",
	Long: `Execute ansible-playbook with token-optimized output.

Shows play summary, task results, and errors only.
Strips verbose task headers and ok/changed detail lines.

Examples:
  tok ansible-playbook site.yml
  tok ansible-playbook playbook.yml --limit production`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnsiblePlaybook(args)
	},
}

var ansibleInventoryCmd = &cobra.Command{
	Use:   "ansible-inventory [args...]",
	Short: "Ansible inventory with compact output",
	Long: `Ansible inventory with token-optimized output.

Examples:
  tok ansible-inventory --list
  tok ansible-inventory --host myhost`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnsibleInventory(args)
	},
}

func runAnsiblePassthrough(args []string) error {
	timer := tracking.Start()

	c := exec.Command("ansible", args...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterAnsibleOutput(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("ansible %s", strings.Join(args, " ")), "tok ansible", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runAnsiblePlaybook(args []string) error {
	timer := tracking.Start()

	c := exec.Command("ansible-playbook", args...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterAnsiblePlaybook(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("ansible-playbook %s", strings.Join(args, " ")), "tok ansible-playbook", originalTokens, filteredTokens)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "ansible_playbook", err); hint != "" {
			filtered = filtered + "\n" + hint
			out.Global().Print(filtered)
		}
	}
	return err
}

func runAnsibleInventory(args []string) error {
	timer := tracking.Start()

	c := exec.Command("ansible-inventory", args...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterAnsibleInventory(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("ansible-inventory %s", strings.Join(args, " ")), "tok ansible-inventory", originalTokens, filteredTokens)

	return err
}

// --- Filter functions ---

func filterAnsibleOutput(raw string) string {
	var result strings.Builder
	var okCount, changedCount, failedCount, unreachableCount int

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, " | SUCCESS ") || strings.Contains(trimmed, " | ") && strings.Contains(trimmed, " pong") {
			okCount++
			continue
		}
		if strings.Contains(trimmed, " | CHANGED ") {
			changedCount++
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
			continue
		}
		if strings.Contains(trimmed, " | FAILED ") || strings.Contains(trimmed, " | UNREACHABLE ") {
			failedCount++
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
			continue
		}
		if strings.Contains(trimmed, "fatal:") || strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
	}

	if okCount > 0 || changedCount > 0 || failedCount > 0 || unreachableCount > 0 {
		result.WriteString(fmt.Sprintf("\nSummary: %d ok, %d changed", okCount, changedCount))
		if failedCount > 0 {
			result.WriteString(fmt.Sprintf(", %d failed", failedCount))
		}
		result.WriteString("\n")
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterAnsiblePlaybook(raw string) string {
	var result strings.Builder
	var playName string
	var okCount, changedCount, failedCount, skippedCount, rescuedCount int
	var errors []string
	var tasks []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "PLAY [") {
			start := strings.Index(trimmed, "[")
			end := strings.Index(trimmed, "]")
			if start != -1 && end != -1 && end > start {
				playName = trimmed[start+1 : end]
				result.WriteString(fmt.Sprintf("Play: %s\n", playName))
			}
			continue
		}

		if strings.Contains(trimmed, "TASK [") {
			start := strings.Index(trimmed, "[")
			end := strings.Index(trimmed, "]")
			if start != -1 && end != -1 && end > start {
				taskName := trimmed[start+1 : end]
				if len(tasks) < 30 {
					tasks = append(tasks, taskName)
				}
			}
			continue
		}

		if strings.Contains(trimmed, "PLAY RECAP") || strings.Contains(trimmed, "RECAP") {
			result.WriteString("\nRecap:\n")
			continue
		}

		if strings.Contains(trimmed, "ok=") && strings.Contains(trimmed, "changed=") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
			parts := strings.Fields(trimmed)
			for _, p := range parts {
				if strings.HasPrefix(p, "ok=") {
					okCount = atoi(p[3:])
				}
				if strings.HasPrefix(p, "changed=") {
					changedCount = atoi(p[8:])
				}
				if strings.HasPrefix(p, "failed=") {
					failedCount = atoi(p[7:])
				}
				if strings.HasPrefix(p, "skipped=") {
					skippedCount = atoi(p[8:])
				}
				if strings.HasPrefix(p, "rescued=") {
					rescuedCount = atoi(p[8:])
				}
			}
			continue
		}

		if strings.Contains(trimmed, "fatal:") || strings.Contains(trimmed, "FAILED") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
			continue
		}
	}

	if shared.UltraCompact {
		return fmt.Sprintf("ok=%d changed=%d failed=%d skipped=%d tasks=%d\n", okCount, changedCount, failedCount, skippedCount, len(tasks))
	}

	if len(tasks) > 0 && len(tasks) <= 10 {
		result.WriteString(fmt.Sprintf("\nTasks (%d):\n", len(tasks)))
		for _, t := range tasks {
			result.WriteString(fmt.Sprintf("  %s\n", t))
		}
	} else if len(tasks) > 10 {
		result.WriteString(fmt.Sprintf("\nTasks: %d executed\n", len(tasks)))
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if okCount > 0 || changedCount > 0 || failedCount > 0 {
		result.WriteString(fmt.Sprintf("\nSummary: ok=%d changed=%d failed=%d", okCount, changedCount, failedCount))
		if skippedCount > 0 {
			result.WriteString(fmt.Sprintf(" skipped=%d", skippedCount))
		}
		if rescuedCount > 0 {
			result.WriteString(fmt.Sprintf(" rescued=%d", rescuedCount))
		}
		result.WriteString("\n")
	}

	if result.Len() == 0 {
		return filterAnsibleOutput(raw)
	}
	return result.String()
}

func filterAnsibleInventory(raw string) string {
	var result strings.Builder
	lineCount := 0

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lineCount++
		if lineCount > 50 {
			totalLines := strings.Count(raw, "\n")
			result.WriteString(fmt.Sprintf("... (%d more lines)\n", totalLines-50))
			break
		}
		result.WriteString(shared.TruncateLine(line, 120) + "\n")
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}
