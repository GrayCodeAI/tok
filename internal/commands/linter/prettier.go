package linter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var prettierCmd = &cobra.Command{
	Use:   "prettier [args...]",
	Short: "Prettier formatter with filtered output",
	Long: `Prettier formatter with token-optimized output.

Shows files that need formatting in check mode.

Examples:
  tok prettier --check .
  tok prettier --write src/
  tok prettier --check "**/*.{ts,tsx}"`,
	RunE: runPrettier,
}

func init() {
	registry.Add(func() { registry.Register(prettierCmd) })
}

func runPrettier(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	prettierPath, err := exec.LookPath("prettier")
	if err != nil {
		prettierPath = ""
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

	hasOutput := strings.TrimSpace(stdout.String()) != ""
	if !hasOutput && err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			out.Global().Errorf("Error: prettier not found or produced no output")
		} else {
			out.Global().Error(msg)
		}
		return err
	}

	filtered := filterPrettierOutput(output)

	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("prettier %s", strings.Join(args, " ")), "tok prettier", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("prettier failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("prettier failed: %w", err)
	}
	return nil
}
