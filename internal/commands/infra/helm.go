package infra

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var helmCmd = &cobra.Command{
	Use:   "helm [subcommand] [args...]",
	Short: "Helm CLI with compact output",
	Long: `Helm CLI with token-optimized output.

Specialized filters for common commands:
  - helm list: Compact release listing
  -helm status: Compact release status
  - helm history: Compact release history
  - helm search: Compact search results
  - helm values: Compact values output

Examples:
  tok helm list
  tok helm status my-release
  tok helm search repo stable
  tok helm values my-chart`,
	DisableFlagParsing: true,
	RunE:               runHelm,
}

func init() {
	registry.Add(func() { registry.Register(helmCmd) })
}

func runHelm(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	switch args[0] {
	case "list", "ls":
		return runHelmList(args[1:])
	case "status":
		return runHelmStatus(args[1:])
	case "history":
		return runHelmHistory(args[1:])
	case "search":
		return runHelmSearch(args[1:])
	case "values":
		return runHelmValues(args[1:])
	case "install":
		return runHelmInstall(args[1:])
	case "upgrade":
		return runHelmUpgrade(args[1:])
	default:
		return runHelmPassthrough(args)
	}
}

func runHelmPassthrough(args []string) error {
	timer := tracking.Start()

	c := exec.Command("helm", args...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmOutput(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm %s", strings.Join(args, " ")), "tok helm", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runHelmList(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"list", "--output", "json"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	if err != nil {
		helmArgs := append([]string{"list"}, args...)
		c := exec.Command("helm", helmArgs...)
		output, _ = c.CombinedOutput()
		raw = string(output)
		filtered := filterHelmListText(raw)
		fmt.Print(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("helm list", "tok helm list", originalTokens, filteredTokens)
		return err
	}

	filtered := filterHelmListJSON(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("helm list", "tok helm list", originalTokens, filteredTokens)

	return nil
}

func runHelmStatus(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"status"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmStatus(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm status %s", strings.Join(args, " ")), "tok helm status", originalTokens, filteredTokens)

	return err
}

func runHelmHistory(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"history"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmHistory(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm history %s", strings.Join(args, " ")), "tok helm history", originalTokens, filteredTokens)

	return err
}

func runHelmSearch(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"search"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmSearch(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm search %s", strings.Join(args, " ")), "tok helm search", originalTokens, filteredTokens)

	return err
}

func runHelmValues(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"values"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	var result strings.Builder
	lineCount := 0
	for _, line := range strings.Split(raw, "\n") {
		lineCount++
		if lineCount > 100 {
			result.WriteString(fmt.Sprintf("... (%d more lines)\n", strings.Count(raw, "\n")-100))
			break
		}
		result.WriteString(line + "\n")
	}

	filtered := result.String()
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm values %s", strings.Join(args, " ")), "tok helm values", originalTokens, filteredTokens)

	return err
}

func runHelmInstall(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"install"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, _ := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmInstall(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm install %s", strings.Join(args, " ")), "tok helm install", originalTokens, filteredTokens)

	return nil
}

func runHelmUpgrade(args []string) error {
	timer := tracking.Start()

	helmArgs := append([]string{"upgrade"}, args...)
	c := exec.Command("helm", helmArgs...)
	output, _ := c.CombinedOutput()
	raw := string(output)

	filtered := filterHelmInstall(raw)
	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("helm upgrade %s", strings.Join(args, " ")), "tok helm upgrade", originalTokens, filteredTokens)

	return nil
}

// --- Filter functions ---

func filterHelmOutput(raw string) string {
	var result strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		result.WriteString(shared.TruncateLine(line, 120) + "\n")
	}
	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

type HelmRelease struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

func filterHelmListJSON(raw string) string {
	var releases []HelmRelease
	if err := json.Unmarshal([]byte(raw), &releases); err != nil {
		return filterHelmListText(raw)
	}

	if len(releases) == 0 {
		return "No releases found\n"
	}

	if shared.UltraCompact {
		result := fmt.Sprintf("%d releases:", len(releases))
		for i, r := range releases {
			if i >= 5 {
				break
			}
			result += fmt.Sprintf(" %s(%s)", r.Name, r.Status)
		}
		if len(releases) > 5 {
			result += fmt.Sprintf(" +%d", len(releases)-5)
		}
		return result + "\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d releases:\n", len(releases)))
	for i, r := range releases {
		if i >= 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(releases)-20))
			break
		}
		ns := r.Namespace
		if ns == "" || ns == "default" {
			ns = ""
		} else {
			ns = ns + "/"
		}
		chart := r.Chart
		if len(chart) > 30 {
			chart = chart[:27] + "..."
		}
		result.WriteString(fmt.Sprintf("  %s%s %s [%s] %s\n", ns, r.Name, r.Status, r.Revision, chart))
	}
	return result.String()
}

func filterHelmListText(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return "No releases found\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d releases:\n", len(lines)-1))

	for i, line := range lines {
		if i == 0 {
			continue
		}
		if i > 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(lines)-21))
			break
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			name := fields[0]
			ns := ""
			rev := ""
			status := ""
			chart := ""
			if len(fields) >= 1 {
				name = fields[0]
			}
			if len(fields) >= 2 {
				ns = fields[1]
			}
			if len(fields) >= 3 {
				rev = fields[2]
			}
			if len(fields) >= 4 {
				status = fields[len(fields)-2]
				chart = fields[len(fields)-1]
			}
			result.WriteString(fmt.Sprintf("  %s/%s rev=%s %s %s\n", ns, name, rev, status, chart))
		} else {
			result.WriteString(shared.TruncateLine(line, 100) + "\n")
		}
	}

	return result.String()
}

func filterHelmStatus(raw string) string {
	var result strings.Builder
	var resources []string
	var hooks []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "STATUS:") || strings.Contains(trimmed, "STATUS") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "REVISION:") || strings.Contains(trimmed, "REVISION") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "LAST DEPLOYED:") || strings.Contains(trimmed, "LAST DEPLOYED") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
		if strings.Contains(trimmed, "NOTES:") {
			break
		}

		if strings.HasPrefix(trimmed, "v1/") || strings.HasPrefix(trimmed, "v1beta") ||
			strings.HasPrefix(trimmed, "apps/v1") || strings.HasPrefix(trimmed, "batch/v1") {
			resources = append(resources, shared.TruncateLine(trimmed, 80))
		}
		if strings.Contains(trimmed, "hook") {
			hooks = append(hooks, shared.TruncateLine(trimmed, 80))
		}
	}

	if len(resources) > 0 {
		result.WriteString(fmt.Sprintf("\nResources (%d):\n", len(resources)))
		for i, r := range resources {
			if i >= 15 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(resources)-15))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", r))
		}
	}

	if len(hooks) > 0 {
		result.WriteString(fmt.Sprintf("\nHooks (%d):\n", len(hooks)))
		for i, h := range hooks {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(hooks)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", h))
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterHelmHistory(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return "No history\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("History (%d revisions):\n", len(lines)-1))

	for i, line := range lines {
		if i == 0 {
			continue
		}
		if i > 15 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(lines)-16))
			break
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			rev := fields[0]
			status := ""
			chart := ""
			appVer := ""
			if len(fields) >= 2 {
				status = fields[len(fields)-2]
				chart = fields[len(fields)-1]
			}
			for _, f := range fields {
				if strings.HasPrefix(f, "v") || strings.Contains(f, "app-") {
					appVer = f
				}
			}
			result.WriteString(fmt.Sprintf("  #%s %s %s %s\n", rev, status, chart, appVer))
		} else {
			result.WriteString(shared.TruncateLine(line, 100) + "\n")
		}
	}

	return result.String()
}

func filterHelmSearch(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return "No results\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Charts (%d results):\n", len(lines)-1))

	for i, line := range lines {
		if i == 0 {
			continue
		}
		if i > 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(lines)-21))
			break
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			name := fields[0]
			chartVer := fields[1]
			appVer := ""
			if len(fields) >= 3 {
				appVer = fields[2]
			}
			desc := ""
			if len(fields) >= 4 {
				desc = strings.Join(fields[3:], " ")
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
			}
			result.WriteString(fmt.Sprintf("  %s %s (app: %s) %s\n", name, chartVer, appVer, desc))
		} else {
			result.WriteString(shared.TruncateLine(line, 100) + "\n")
		}
	}

	return result.String()
}

func filterHelmInstall(raw string) string {
	var result strings.Builder
	var releaseName string
	var namespace string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "NAME:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				releaseName = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(trimmed, "NAMESPACE:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				namespace = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(trimmed, "STATUS:") || strings.Contains(trimmed, "STATUS") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "NOTES:") {
			result.WriteString("\nNotes:\n")
			continue
		}
		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
	}

	if releaseName != "" {
		result.WriteString(fmt.Sprintf("Release: %s\n", releaseName))
		if namespace != "" {
			result.WriteString(fmt.Sprintf("Namespace: %s\n", namespace))
		}
	}

	if result.Len() == 0 {
		return "Install/Upgrade complete\n"
	}
	return result.String()
}
