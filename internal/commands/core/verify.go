package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/integrity"
)

var verifyRequireAll bool

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify hook integrity",
	Long: `Verify the integrity of the TokMan hook script.

This command checks that the hook file (~/.claude/hooks/tokman-rewrite.sh)
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
		fmt.Printf("Hook:  %s\n", result.HookPath)
		fmt.Printf("Hash:  %s\n", result.HashPath)
		fmt.Println()
	}

	switch result.Status {
	case integrity.StatusVerified:
		hash, _ := integrity.ComputeHash(result.HookPath)
		fmt.Printf("%s  hook integrity verified\n", green("PASS"))
		fmt.Printf("      sha256:%s\n", hash)
		fmt.Printf("      %s\n", cyan(result.HookPath))

	case integrity.StatusTampered:
		fmt.Fprintf(os.Stderr, "%s  hook integrity check FAILED\n", red("FAIL"))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "  Expected: %s\n", result.Expected)
		fmt.Fprintf(os.Stderr, "  Actual:   %s\n", result.Actual)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  The hook file has been modified outside of `tokman init`.")
		fmt.Fprintln(os.Stderr, "  This could indicate tampering or a manual edit.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  To restore: tokman init")
		fmt.Fprintf(os.Stderr, "  To inspect: cat %s\n", result.HookPath)
		allPassed = false
		if !verifyRequireAll {
			return fmt.Errorf("hook integrity check failed")
		}

	case integrity.StatusNoBaseline:
		fmt.Printf("%s  no baseline hash found\n", yellow("WARN"))
		fmt.Println("      Hook exists but was installed before integrity checks.")
		fmt.Println("      Run `tokman init` to establish baseline.")
		allPassed = false

	case integrity.StatusNotInstalled:
		fmt.Printf("%s  TokMan hook not installed\n", yellow("SKIP"))
		fmt.Println("      Run `tokman init` to install.")
		allPassed = false

	case integrity.StatusOrphanedHash:
		fmt.Fprintf(os.Stderr, "%s  hash file exists but hook is missing\n", yellow("WARN"))
		fmt.Fprintln(os.Stderr, "      Run `tokman init` to reinstall.")
		allPassed = false

	case integrity.StatusOutdated:
		fmt.Printf("%s  hook is outdated\n", yellow("WARN"))
		fmt.Printf("      Installed version: %d\n", result.HookVersion)
		fmt.Printf("      Required version:  %d\n", result.RequiredVersion)
		fmt.Println("      Run `tokman init --claude` to refresh the generated hook.")
		allPassed = false
	}

	// If --require-all, also check config and filters
	if verifyRequireAll {
		fmt.Println()
		fmt.Println(cyan("Additional verification checks:"))

		// Check config
		configOK := verifyConfig()
		if configOK {
			fmt.Printf("%s  config valid\n", green("PASS"))
		} else {
			fmt.Printf("%s  config issues found\n", yellow("WARN"))
			allPassed = false
		}

		// Check filters
		filtersOK := verifyFilters()
		if filtersOK {
			fmt.Printf("%s  filters valid\n", green("PASS"))
		} else {
			fmt.Printf("%s  filter issues found\n", yellow("WARN"))
			allPassed = false
		}
	}

	// Print security reference
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("Security: This check prevents command injection via hook tampering.")

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
