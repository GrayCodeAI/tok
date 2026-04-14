package haskell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func init() {
	registry.Add(func() {
		registry.Register(haskellCmd)
	})
}

var haskellCmd = &cobra.Command{
	Use:   "haskell [args...]",
	Short: "Haskell/Stack/Cabal with compact output",
	Long:  `Haskell toolchain commands with token-optimized output.`,
}

func runHaskell(args []string) error {
	if len(args) == 0 {
		return haskellCmd.Help()
	}

	timer := tracking.Start()
	bin := "ghc"
	if len(args) > 0 && (args[0] == "build" || args[0] == "hadrian") {
		bin = "stack"
	} else if len(args) > 0 && (args[0] == "new" || args[0] == "init" || args[0] == "install") {
		bin = "cabal"
	}

	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	raw := string(output)

	filtered := filterHaskell(raw, args)
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)

	timer.Track("haskell", "tokman haskell", originalTokens, filteredTokens)

	if shared.IsUltraCompact() {
		filtered = compactOutput(filtered)
	}

	fmt.Print(filtered)

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "haskell", err); hint != "" {
			fmt.Print("\n" + hint)
		}
	}

	return err
}

func filterHaskell(raw string, args []string) string {
	lines := strings.Split(raw, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "Error:") {
			filtered = append(filtered, line)
		} else if strings.Contains(line, "warning:") || strings.Contains(line, "Warning:") {
			filtered = append(filtered, line)
		} else if strings.HasPrefix(line, "Linking") || strings.HasPrefix(line, "Compiling") || strings.HasPrefix(line, "Preprocessing") {
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
