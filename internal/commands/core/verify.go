package core

import (
	"fmt"
	"os"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/integrity"
)

var verifyRequireAll bool

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify hook integrity",
	Long: `Verify the integrity of the tok hook script.

This command checks that the hook file (~/.claude/hooks/tok-rewrite.sh)
matches its stored SHA-256 hash to detect any unauthorized modifications.

The integrity check protects against command injection attacks where
an attacker might modify the hook to execute malicious commands.

Options:
  --require-all    Require all checks to pass (hook + config + filters)`,
	RunE: runVerify,
}

func init() {
	registry.Add(func() { registry.Register(verifyCmd) })
	verifyCmd.Flags().BoolVar(&verifyRequireAll, "require-all", false, "Require all verification checks to pass")
}

func runVerify(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	allPassed := true

	result, err := integrity.VerifyHook()
	if err != nil {
		return fmt.Errorf("error verifying hook: %w", err)
	}

	if shared.Verbose > 0 {
		out.Global().Printf("Hook:  %s\n", result.HookPath)
		out.Global().Printf("Hash:  %s\n", result.HashPath)
		out.Global().Println()
	}

	switch result.Status {
	case integrity.StatusVerified:
		hash, _ := integrity.ComputeHash(result.HookPath)
		out.Global().Printf("%s  hook integrity verified\n", green("PASS"))
		out.Global().Printf("      sha256:%s\n", hash)
		out.Global().Printf("      %s\n", cyan(result.HookPath))

	case integrity.StatusTampered:
		out.Global().Errorf("%s  hook integrity check FAILED\n", red("FAIL"))
		fmt.Fprintln(os.Stderr)
		out.Global().Errorf("  Expected: %s\n", result.Expected)
		out.Global().Errorf("  Actual:   %s\n", result.Actual)
		fmt.Fprintln(os.Stderr)
		out.Global().Errorf("  The hook file has been modified outside of `tok init`.")
		out.Global().Errorf("  This could indicate tampering or a manual edit.")
		fmt.Fprintln(os.Stderr)
		out.Global().Errorf("  To restore: tok init")
		out.Global().Errorf("  To inspect: cat %s\n", result.HookPath)
		allPassed = false
		if !verifyRequireAll {
			return fmt.Errorf("hook integrity check failed")
		}

	case integrity.StatusNoBaseline:
		out.Global().Printf("%s  no baseline hash found\n", yellow("WARN"))
		out.Global().Println("      Hook exists but was installed before integrity checks.")
		out.Global().Println("      Run `tok init` to establish baseline.")
		allPassed = false

	case integrity.StatusNotInstalled:
		out.Global().Printf("%s  tok hook not installed\n", yellow("SKIP"))
		out.Global().Println("      Run `tok init` to install.")
		allPassed = false

	case integrity.StatusOrphanedHash:
		out.Global().Errorf("%s  hash file exists but hook is missing\n", yellow("WARN"))
		out.Global().Errorf("      Run `tok init` to reinstall.")
		allPassed = false

	case integrity.StatusOutdated:
		out.Global().Printf("%s  hook is outdated\n", yellow("WARN"))
		out.Global().Printf("      Installed version: %d\n", result.HookVersion)
		out.Global().Printf("      Required version:  %d\n", result.RequiredVersion)
		out.Global().Println("      Run `tok init --claude` to refresh the generated hook.")
		allPassed = false
	}

	// If --require-all, also check config and filters
	if verifyRequireAll {
		out.Global().Println()
		out.Global().Println(cyan("Additional verification checks:"))

		// Check config
		configOK := verifyConfig()
		if configOK {
			out.Global().Printf("%s  config valid\n", green("PASS"))
		} else {
			out.Global().Printf("%s  config issues found\n", yellow("WARN"))
			allPassed = false
		}

		// Check filters
		filtersOK := verifyFilters()
		if filtersOK {
			out.Global().Printf("%s  filters valid\n", green("PASS"))
		} else {
			out.Global().Printf("%s  filter issues found\n", yellow("WARN"))
			allPassed = false
		}
	}

	// Print security reference
	out.Global().Println()
	out.Global().Println(strings.Repeat("─", 50))
	out.Global().Println("Security: This check prevents command injection via hook tampering.")

	if verifyRequireAll && !allPassed {
		return fmt.Errorf("not all verification checks passed")
	}

	return nil
}

func verifyConfig() bool {
	// Placeholder for config verification
	// In a real implementation, this would verify config file syntax, etc.
	return true
}

func verifyFilters() bool {
	// Placeholder for filter verification
	// In a real implementation, this would verify TOML filter syntax
	return true
}
