package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var prettierCmd = &cobra.Command{
	Use:   "prettier [args...]",
	Short: "Prettier formatter with filtered output",
	Long: `Prettier formatter with token-optimized output.

Shows files that need formatting in check mode.

Examples:
  tokman prettier --check .
  tokman prettier --write src/
  tokman prettier --check "**/*.{ts,tsx}"`,
	RunE: runPrettier,
}

func init() {
	rootCmd.AddCommand(prettierCmd)
}

func runPrettier(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Use package manager to run prettier
	prettierPath, err := exec.LookPath("prettier")
	if err != nil {
		prettierPath = "" // Will use npx
	}

	var c *exec.Cmd
	if prettierPath != "" {
		c = exec.Command(prettierPath, args...)
	} else {
		npxArgs := append([]string{"prettier"}, args...)
		c = exec.Command("npx", npxArgs...)
	}
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()
	output := stdout.String() + stderr.String()

	// Handle case where prettier not installed or no output
	hasOutput := strings.TrimSpace(stdout.String()) != ""
	if !hasOutput && err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			fmt.Fprintln(os.Stderr, "Error: prettier not found or produced no output")
		} else {
			fmt.Fprintln(os.Stderr, msg)
		}
		return err
	}

	filtered := filterPrettierOutput(output)

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("prettier %s", strings.Join(args, " ")), "tokman prettier", originalTokens, filteredTokens)

	if verbose > 0 {
		fmt.Fprintf(os.Stderr, "Tokens saved: %d\n", originalTokens-filteredTokens)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
	return nil
}

func parseInt(s string) (int, bool) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err == nil
}
