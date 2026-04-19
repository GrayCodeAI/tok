package lang

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

var rubyCmd = &cobra.Command{
	Use:   "ruby [args...]",
	Short: "Ruby commands with compact output",
	Long: `Execute Ruby commands with token-optimized output.

Provides compact output for rspec, rubocop, rake, bundle, and rails commands.

Examples:
  tok ruby rspec
  tok ruby rubocop
  tok ruby rake test
  tok ruby bundle install
  tok ruby rails test`,
	DisableFlagParsing: true,
	RunE:               runRuby,
}

func init() {
	registry.Add(func() { registry.Register(rubyCmd) })
}

func runRuby(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	// Route to specialized handlers
	switch args[0] {
	case "rspec":
		return runRspecCmd(args[1:])
	case "rubocop":
		return runRubocopCmd(args[1:])
	case "rake":
		return runRakeCmd(args[1:])
	case "bundle":
		return runBundleCmd(args[1:])
	case "rails":
		return runRailsCmd(args[1:])
	default:
		return runRubyPassthrough(args)
	}
}

// =============================================================================
// Ruby Passthrough
// =============================================================================

func runRubyPassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: ruby %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("ruby", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterRubyOutput(raw)
	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("ruby %s", strings.Join(args, " ")), "tok ruby", originalTokens, filteredTokens)

	return err
}

func filterRubyOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, shared.TruncateLine(line, 120))
		}
	}

	if len(result) > 30 {
		return strings.Join(result[:30], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-30)
	}
	return strings.Join(result, "\n")
}
