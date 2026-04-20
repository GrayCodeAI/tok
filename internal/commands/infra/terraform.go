package infra

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

var terraformCmd = &cobra.Command{
	Use:   "terraform [subcommand] [args...]",
	Short: "Terraform CLI with compact output",
	Long: `Terraform CLI with token-optimized output.

Specialized filters for common commands:
  - terraform plan: Compact plan summary
  - terraform apply: Compact apply result
  - terraform show: Compact state/planned output
  - terraform output: Compact output values
  - terraform state list: Compact resource listing
  - terraform import: Import result
  - terraform init: Compact initialization summary
  - terraform validate: Compact validation result
  - terraform graph: Compact graph output

Also handles 'tf' as an alias.

Examples:
  tok terraform plan
  tok terraform apply
  tok terraform state list`,
	Aliases:            []string{"tf"},
	DisableFlagParsing: true,
	RunE:               runTerraform,
}

func init() {
	registry.Add(func() { registry.Register(terraformCmd) })
}

func runTerraform(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	if len(args) > 0 {
		switch args[0] {
		case "plan":
			return runTerraformPlan(args[1:])
		case "apply":
			return runTerraformApply(args[1:])
		case "show":
			return runTerraformShow(args[1:])
		case "output":
			return runTerraformOutput(args[1:])
		case "state":
			return runTerraformState(args[1:])
		case "import":
			return runTerraformImport(args[1:])
		case "init":
			return runTerraformInit(args[1:])
		case "validate":
			return runTerraformValidate(args[1:])
		case "graph":
			return runTerraformGraph(args[1:])
		}
	}

	return runTerraformPassthrough(args)
}

func runTerraformPassthrough(args []string) error {
	timer := tracking.Start()

	c := exec.Command("terraform", args...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformOutput(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("terraform %s", strings.Join(args, " ")), "tok terraform", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runTerraformPlan(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"plan"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformPlan(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform plan", "tok terraform plan", originalTokens, filteredTokens)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "terraform_plan", err); hint != "" {
			filtered = filtered + "\n" + hint
			out.Global().Print(filtered)
		}
	}
	return err
}

func runTerraformApply(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"apply"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformApply(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform apply", "tok terraform apply", originalTokens, filteredTokens)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "terraform_apply", err); hint != "" {
			filtered = filtered + "\n" + hint
			out.Global().Print(filtered)
		}
	}
	return err
}

func runTerraformShow(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"show"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformShow(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("terraform show %s", strings.Join(args, " ")), "tok terraform show", originalTokens, filteredTokens)

	return err
}

func runTerraformOutput(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"output", "-json"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformOutputJSON(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform output", "tok terraform output", originalTokens, filteredTokens)

	return err
}

func runTerraformState(args []string) error {
	timer := tracking.Start()

	if len(args) > 0 && args[0] == "list" {
		rest := args[1:]
		tfArgs := append([]string{"state", "list"}, rest...)
		c := exec.Command("terraform", tfArgs...)
		output, err := c.CombinedOutput()
		raw := string(output)

		filtered := filterTerraformStateList(raw)
		out.Global().Print(filtered)

		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("terraform state list", "tok terraform state list", originalTokens, filteredTokens)

		return err
	}

	return runTerraformPassthrough(append([]string{"state"}, args...))
}

func runTerraformImport(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"import"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformImport(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("terraform import %s", strings.Join(args, " ")), "tok terraform import", originalTokens, filteredTokens)

	return err
}

func runTerraformInit(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"init"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformInit(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform init", "tok terraform init", originalTokens, filteredTokens)

	return err
}

func runTerraformValidate(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"validate"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformValidate(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform validate", "tok terraform validate", originalTokens, filteredTokens)

	return err
}

func runTerraformGraph(args []string) error {
	timer := tracking.Start()

	tfArgs := append([]string{"graph"}, args...)
	c := exec.Command("terraform", tfArgs...)
	output, err := c.CombinedOutput()
	raw := string(output)

	filtered := filterTerraformGraph(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("terraform graph", "tok terraform graph", originalTokens, filteredTokens)

	return err
}

// --- Filter functions ---

func filterTerraformOutput(raw string) string {
	var result strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "...") {
			continue
		}
		result.WriteString(shared.TruncateLine(line, 120) + "\n")
	}
	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterTerraformPlan(raw string) string {
	var result strings.Builder
	var additions, changes, destroys int
	var vars []string
	var errors []string
	inPlan := false

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "will be created") {
			fmt.Sscanf(trimmed, "%d", &additions)
		}
		if strings.Contains(trimmed, "will be updated") || strings.Contains(trimmed, "will be changed") {
			fmt.Sscanf(trimmed, "%d", &changes)
		}
		if strings.Contains(trimmed, "will be destroyed") || strings.Contains(trimmed, "will be deleted") {
			fmt.Sscanf(trimmed, "%d", &destroys)
		}

		if strings.Contains(trimmed, "No changes") {
			result.WriteString("Plan: No changes\n")
			inPlan = true
		}

		if strings.Contains(trimmed, "Plan:") || strings.Contains(trimmed, "will be") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}

		if strings.HasPrefix(trimmed, "var.") || strings.HasPrefix(trimmed, "-var") {
			vars = append(vars, trimmed)
		}

		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if !inPlan && (additions > 0 || changes > 0 || destroys > 0) {
		result.WriteString(fmt.Sprintf("Plan: %d add, %d change, %d destroy\n", additions, changes, destroys))
	}

	if len(vars) > 0 {
		result.WriteString(fmt.Sprintf("Variables: %d\n", len(vars)))
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

	if result.Len() == 0 {
		return filterTerraformOutput(raw)
	}
	return result.String()
}

func filterTerraformApply(raw string) string {
	var result strings.Builder
	var additions, changes, destroys int
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "will be created") {
			additions = atoi(trimmed)
		}
		if strings.Contains(trimmed, "will be updated") || strings.Contains(trimmed, "will be changed") {
			changes = atoi(trimmed)
		}
		if strings.Contains(trimmed, "will be destroyed") || strings.Contains(trimmed, "will be deleted") {
			destroys = atoi(trimmed)
		}

		if strings.Contains(trimmed, "Apply complete!") {
			result.WriteString(trimmed + "\n")
		}

		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if additions > 0 || changes > 0 || destroys > 0 {
		result.WriteString(fmt.Sprintf("Apply: %d add, %d change, %d destroy\n", additions, changes, destroys))
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

	if result.Len() == 0 {
		return filterTerraformOutput(raw)
	}
	return result.String()
}

func filterTerraformShow(raw string) string {
	var result strings.Builder
	lineCount := 0
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lineCount++
		if lineCount > 50 {
			result.WriteString(fmt.Sprintf("\n... (%d more lines)\n", strings.Count(raw, "\n")-50))
			break
		}
		result.WriteString(shared.TruncateLine(line, 120) + "\n")
	}
	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterTerraformOutputJSON(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return "No outputs"
	}

	var result strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "{" || trimmed == "}" || trimmed == "{}" {
			continue
		}
		if strings.Contains(trimmed, `"value"`) {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
	}

	if result.Len() == 0 {
		return raw
	}
	return result.String()
}

func filterTerraformStateList(raw string) string {
	resources := strings.Split(strings.TrimSpace(raw), "\n")
	if len(resources) == 0 || (len(resources) == 1 && strings.TrimSpace(resources[0]) == "") {
		return "No resources in state"
	}

	types := make(map[string]int)
	for _, r := range resources {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		parts := strings.Split(r, ".")
		if len(parts) > 0 {
			types[parts[0]]++
		}
	}

	if shared.UltraCompact {
		var parts []string
		for t, c := range types {
			parts = append(parts, fmt.Sprintf("%s:%d", t, c))
		}
		return fmt.Sprintf("%d resources: %s\n", len(resources), strings.Join(parts, " "))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d resources:\n", len(resources)))
	for t, c := range types {
		result.WriteString(fmt.Sprintf("  %s (%d)\n", t, c))
	}
	if len(resources) > 30 {
		for _, r := range resources[:20] {
			result.WriteString(fmt.Sprintf("    %s\n", r))
		}
		result.WriteString(fmt.Sprintf("    ... +%d more\n", len(resources)-20))
	} else {
		for _, r := range resources {
			result.WriteString(fmt.Sprintf("    %s\n", r))
		}
	}
	return result.String()
}

func filterTerraformImport(raw string) string {
	var result strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "Import") || strings.Contains(trimmed, "Importing") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
	}
	if result.Len() == 0 {
		return "Import complete"
	}
	return result.String()
}

func filterTerraformInit(raw string) string {
	var result strings.Builder
	var initialized bool
	var providerCount int
	var backend string
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Terraform has been successfully initialized") {
			initialized = true
		}
		if strings.Contains(trimmed, "Installing") && strings.Contains(trimmed, "provider") {
			providerCount++
		}
		if strings.Contains(trimmed, "backend") {
			backend = trimmed
		}
		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if initialized {
		result.WriteString("Init: successful\n")
		if providerCount > 0 {
			result.WriteString(fmt.Sprintf("  Providers: %d installed\n", providerCount))
		}
		if backend != "" {
			result.WriteString(fmt.Sprintf("  %s\n", backend))
		}
	} else {
		result.WriteString("Init: checking...\n")
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return filterTerraformOutput(raw)
	}
	return result.String()
}

func filterTerraformValidate(raw string) string {
	var result strings.Builder
	var errors []string
	var warnings []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Success!") {
			result.WriteString("Validation: passed\n")
			return result.String()
		}
		if strings.Contains(trimmed, "Error:") || strings.Contains(trimmed, "error:") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
		if strings.Contains(trimmed, "Warning:") || strings.Contains(trimmed, "warning:") {
			warnings = append(warnings, shared.TruncateLine(trimmed, 100))
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Validation: %d error(s)\n", len(errors)))
		for i, e := range errors {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if len(warnings) > 0 {
		result.WriteString(fmt.Sprintf("Warnings: %d\n", len(warnings)))
	}

	if result.Len() == 0 {
		return filterTerraformOutput(raw)
	}
	return result.String()
}

func filterTerraformGraph(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 1 {
		return "Empty graph"
	}

	resources := make(map[string]int)
	for _, line := range lines {
		if strings.Contains(line, "[label=") {
			if idx := strings.Index(line, `label="`); idx != -1 {
				end := strings.Index(line[idx+7:], `"`)
				if end > 0 {
					label := line[idx+7 : idx+7+end]
					resources[label]++
				}
			}
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Graph: %d nodes, %d edges\n", len(resources), len(lines)-2))
	for label, count := range resources {
		if count > 1 {
			result.WriteString(fmt.Sprintf("  %s (x%d)\n", label, count))
		} else {
			result.WriteString(fmt.Sprintf("  %s\n", label))
		}
		if result.Len() > 2000 {
			result.WriteString("  ... (truncated)\n")
			break
		}
	}
	return result.String()
}
