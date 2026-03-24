package pkgmgr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle [subcommand] [args...]",
	Short: "Bundler with filtered output",
	Long: `Bundler package manager with token-optimized output.

Subcommands:
  install   - Strip 'Using' lines, show only changes (90%+ savings)
  update    - Show only updated gems
  list      - Compact gem listing
  outdated  - Show outdated gems
  exec      - Passthrough with tracking

Examples:
  tokman bundle install
  tokman bundle update rails
  tokman bundle outdated`,
	DisableFlagParsing: true,
	RunE:               runBundle,
}

func init() {
	registry.Add(func() { registry.Register(bundleCmd) })
}

func runBundle(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"install"}
	}

	subcommand := args[0]

	c := exec.Command("bundle", args...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	switch subcommand {
	case "install":
		filtered = filterBundleInstall(output)
	case "update":
		filtered = filterBundleUpdate(output)
	case "list":
		filtered = filterBundleList(output)
	case "outdated":
		filtered = filterBundleOutdated(output)
	default:
		filtered = strings.TrimSpace(output) + "\n"
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("bundle %s", strings.Join(args, " ")), "tokman bundle", originalTokens, filteredTokens)

	if shared.Verbose > 0 {
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

func filterBundleInstall(output string) string {
	lines := strings.Split(output, "\n")

	var installed []string
	var removed []string
	var updated []string
	var errors []string
	usingCount := 0
	bundleComplete := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(trimmed, "Using ") {
			usingCount++
			continue
		}
		if strings.HasPrefix(trimmed, "Installing ") {
			installed = append(installed, trimmed)
			continue
		}
		if strings.HasPrefix(trimmed, "Removing ") {
			removed = append(removed, trimmed)
			continue
		}
		if strings.HasPrefix(trimmed, "Updating ") || strings.HasPrefix(trimmed, "Updated ") {
			updated = append(updated, trimmed)
			continue
		}
		if strings.Contains(lower, "bundle complete") || strings.Contains(lower, "bundle updated") {
			bundleComplete = true
			continue
		}
		if strings.Contains(lower, "error") || strings.Contains(lower, "could not find") ||
			strings.Contains(lower, "bundler::gemnotfound") {
			errors = append(errors, trimmed)
		}
	}

	if len(errors) > 0 {
		var result strings.Builder
		result.WriteString("bundle install: FAILED\n")
		for _, e := range errors {
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
		return result.String()
	}

	if bundleComplete && len(installed) == 0 && len(removed) == 0 && len(updated) == 0 {
		return fmt.Sprintf("ok bundle install: %d gems (all up to date)\n", usingCount)
	}

	var result strings.Builder
	totalChanges := len(installed) + len(removed) + len(updated)
	result.WriteString(fmt.Sprintf("bundle install: %d gems, %d changes\n", usingCount+totalChanges, totalChanges))

	if len(installed) > 0 {
		for _, line := range installed {
			result.WriteString(fmt.Sprintf("  + %s\n", line))
		}
	}
	if len(updated) > 0 {
		for _, line := range updated {
			result.WriteString(fmt.Sprintf("  ~ %s\n", line))
		}
	}
	if len(removed) > 0 {
		for _, line := range removed {
			result.WriteString(fmt.Sprintf("  - %s\n", line))
		}
	}

	return result.String()
}

func filterBundleUpdate(output string) string {
	lines := strings.Split(output, "\n")

	var updated []string
	usingCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "Using ") {
			usingCount++
			continue
		}
		if strings.HasPrefix(trimmed, "Installing ") || strings.HasPrefix(trimmed, "Updating ") {
			updated = append(updated, trimmed)
		}
	}

	if len(updated) == 0 {
		return fmt.Sprintf("ok bundle update: %d gems (no changes)\n", usingCount)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("bundle update: %d updated\n", len(updated)))
	for _, line := range updated {
		result.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return result.String()
}

func filterBundleList(output string) string {
	lines := strings.Split(output, "\n")
	var gems []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.HasPrefix(trimmed, "* ") {
			gems = append(gems, trimmed)
		}
	}

	if len(gems) == 0 {
		return strings.TrimSpace(output) + "\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("bundle list: %d gems\n", len(gems)))

	for i, gem := range gems {
		if i >= 20 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(gems)-20))
			break
		}
		result.WriteString(fmt.Sprintf("  %s\n", gem))
	}

	return result.String()
}

func filterBundleOutdated(output string) string {
	lines := strings.Split(output, "\n")
	var outdated []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "outdated gems") || strings.HasPrefix(trimmed, "Fetching") ||
			strings.HasPrefix(trimmed, "Resolving") {
			continue
		}
		if strings.Contains(trimmed, "(newest") || strings.Contains(trimmed, "(current:") {
			outdated = append(outdated, trimmed)
		}
	}

	if len(outdated) == 0 {
		return "ok bundle outdated: all gems up to date\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("bundle outdated: %d gems\n", len(outdated)))
	result.WriteString("═══════════════════════════════════════\n")

	for i, gem := range outdated {
		if i >= 15 {
			result.WriteString(fmt.Sprintf("\n... +%d more\n", len(outdated)-15))
			break
		}
		result.WriteString(fmt.Sprintf("  %s\n", gem))
	}

	return result.String()
}
