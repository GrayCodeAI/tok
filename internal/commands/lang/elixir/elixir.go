package elixir

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

func init() {
	registry.Add(func() {
		registry.Register(elixirCmd)
	})
}

var elixirCmd = &cobra.Command{
	Use:   "elixir [args...]",
	Short: "Elixir/Erlang with compact output",
	Long:  `Elixir and Erlang commands with token-optimized output.`,
}

func runElixir(args []string) error {
	if len(args) == 0 {
		return elixirCmd.Help()
	}

	timer := tracking.Start()

	cmd := exec.Command("elixir", args...)
	output, err := cmd.CombinedOutput()
	raw := string(output)

	filtered := filterElixir(raw, args)
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)

	timer.Track("elixir", "tok elixir", originalTokens, filteredTokens)

	if shared.IsUltraCompact() {
		filtered = compactOutput(filtered)
	}

	out.Global().Print(filtered)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "elixir", err); hint != "" {
			out.Global().Print("\n" + hint)
		}
	}

	return err
}

func filterElixir(raw string, args []string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "**") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "warning:") {
			filtered = append(filtered, line)
		} else if strings.HasPrefix(line, "===") || strings.HasPrefix(line, "Compiling") || strings.HasPrefix(line, "Generated") {
			filtered = append(filtered, line)
		} else if !shared.IsUltraCompact() && line != "" {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func compactOutput(filtered string) string {
	lines := strings.Split(filtered, "\n")
	if len(lines) > 8 {
		return strings.Join(lines[:8], "\n") + fmt.Sprintf("\n... (+%d lines)", len(lines)-8)
	}
	return filtered
}
