package pkgmgr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var npxCmd = &cobra.Command{
	Use:   "npx [args...]",
	Short: "npx with intelligent routing to specialized filters",
	Long: `Execute npx with intelligent command routing.

Routes common tools to specialized filters:
- tsc, typescript → tok tsc
- eslint → tok lint
- prettier → tok prettier
- prisma → specialized prisma filter
- next → tok next

Examples:
  tok npx tsc --noEmit
  tok npx eslint src/
  tok npx prisma generate`,
	DisableFlagParsing: true,
	RunE:               runNpx,
}

func init() {
	registry.Add(func() { registry.Register(npxCmd) })
}

func runNpx(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("npx requires a command argument")
	}

	switch args[0] {
	case "tsc", "typescript":
		return runTscCommand(args[1:])
	case "eslint":
		return runLintCommand(args[1:])
	case "prettier":
		return runPrettierCommand(args[1:])
	case "prisma":
		return runPrismaCommand(args[1:])
	case "next":
		return runNextCommand(args[1:])
	default:
		return runNpxPassthrough(args)
	}
}

func runTscCommand(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx tsc %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"tsc"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterTscOutput(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx tsc %s", strings.Join(args, " ")), "tok npx tsc", originalTokens, filteredTokens)

	return err
}

func runLintCommand(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx eslint %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"eslint"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterEslintJSON(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx eslint %s", strings.Join(args, " ")), "tok npx eslint", originalTokens, filteredTokens)

	return err
}

func runPrettierCommand(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx prettier %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"prettier"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterPrettierOutput(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx prettier %s", strings.Join(args, " ")), "tok npx prettier", originalTokens, filteredTokens)

	return err
}

func runPrismaCommand(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx prisma %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"prisma"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterPrismaOutputCompact(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx prisma %s", strings.Join(args, " ")), "tok npx prisma", originalTokens, filteredTokens)

	return err
}

func runNextCommand(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx next %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", append([]string{"next"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterNextOutputCompact(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx next %s", strings.Join(args, " ")), "tok npx next", originalTokens, filteredTokens)

	return err
}

func runNpxPassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: npx %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("npx", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterNpmOutput(raw)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npx %s", strings.Join(args, " ")), "tok npx", originalTokens, filteredTokens)

	return err
}
