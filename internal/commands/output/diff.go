package output

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/core"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> [file2]",
	Short: "Ultra-condensed diff (only changed lines)",
	Long: `Show diff with only changed lines for token efficiency.

If file2 is omitted, reads from stdin.
Strips context lines, shows only additions/deletions.

Examples:
  tok diff file1.go file2.go
  git diff | tok diff -`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDiff,
}

func init() {
	registry.Add(func() { registry.Register(diffCmd) })
}

func runDiff(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	var output string
	var err error

	if len(args) == 1 || args[0] == "-" {
		var buf bytes.Buffer
		_, err = buf.ReadFrom(cmd.InOrStdin())
		output = buf.String()
	} else {
		c := exec.Command("diff", args...)
		var stdout, stderr bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stderr
		err = c.Run()
		output = stdout.String()
		if output == "" && stderr.Len() > 0 {
			output = stderr.String()
		}
	}

	filtered := compactDiff(output)

	fmt.Print(filtered)

	originalTokens := core.EstimateTokens(output)
	filteredTokens := core.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("diff %s", strings.Join(args, " ")), "tok diff", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func compactDiff(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "+") && !strings.HasPrefix(trimmed, "+++") {
			result.WriteString(line + "\n")
		} else if strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "---") {
			result.WriteString(line + "\n")
		} else if strings.HasPrefix(trimmed, "@@") {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}
