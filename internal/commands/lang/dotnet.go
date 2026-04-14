package lang

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func atoi(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		n = 0
	}
	return n
}

var dotnetCmd = &cobra.Command{
	Use:   "dotnet [command]",
	Short: ".NET commands with compact output",
	Long: `.NET commands with compact output.

Specialized filters for:
  - build: Show errors, warnings, and build summary
  - test: Show test results and failures
  - restore: Show restore summary
  - format: Show format check results
  - run: Show run output and errors
  - publish: Show publish summary

Examples:
  tokman dotnet build
  tokman dotnet test
  tokman dotnet run
  tokman dotnet publish -c Release`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
}

var dotnetBuildCmd = &cobra.Command{
	Use:                "build [args...]",
	Short:              "Build with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("build", args)
	},
}

var dotnetTestCmd = &cobra.Command{
	Use:                "test [args...]",
	Short:              "Test with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("test", args)
	},
}

var dotnetRestoreCmd = &cobra.Command{
	Use:                "restore [args...]",
	Short:              "Restore with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("restore", args)
	},
}

var dotnetFormatCmd = &cobra.Command{
	Use:                "format [args...]",
	Short:              "Format with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("format", args)
	},
}

var dotnetRunCmd = &cobra.Command{
	Use:                "run [args...]",
	Short:              "Run with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("run", args)
	},
}

var dotnetPublishCmd = &cobra.Command{
	Use:                "publish [args...]",
	Short:              "Publish with compact output",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDotnetSubcommand("publish", args)
	},
}

func init() {
	registry.Add(func() { registry.Register(dotnetCmd) })
	dotnetCmd.AddCommand(dotnetBuildCmd)
	dotnetCmd.AddCommand(dotnetTestCmd)
	dotnetCmd.AddCommand(dotnetRestoreCmd)
	dotnetCmd.AddCommand(dotnetFormatCmd)
	dotnetCmd.AddCommand(dotnetRunCmd)
	dotnetCmd.AddCommand(dotnetPublishCmd)
}

func runDotnetSubcommand(subCmd string, args []string) error {
	timer := tracking.Start()

	dotnetArgs := append([]string{subCmd}, args...)
	c := exec.Command("dotnet", dotnetArgs...)

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	switch subCmd {
	case "build":
		filtered = filterDotnetBuild(output)
	case "test":
		filtered = filterDotnetTest(output)
	case "restore":
		filtered = filterDotnetRestore(output)
	case "format":
		filtered = filterDotnetFormat(output)
	case "run":
		filtered = filterDotnetRun(output)
	case "publish":
		filtered = filterDotnetPublish(output)
	default:
		filtered = filterDotnetOutput(output)
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("dotnet %s", subCmd), "tokman dotnet", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func filterDotnetBuild(output string) string {
	var result strings.Builder
	var errors, warnings int
	var errorLines []string
	var warningLines []string
	var succeeded bool

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Build succeeded") {
			succeeded = true
			continue
		}
		if strings.Contains(trimmed, "Build FAILED") {
			continue
		}
		if strings.Contains(trimmed, "error ") || strings.Contains(trimmed, "Error(s)") {
			if strings.Contains(trimmed, " Error(s)") {
				errors = atoi(trimmed)
			} else {
				errors++
				errorLines = append(errorLines, shared.TruncateLine(trimmed, 100))
			}
			continue
		}
		if strings.Contains(trimmed, "warning ") || strings.Contains(trimmed, "Warning(s)") {
			if strings.Contains(trimmed, " Warning(s)") {
				warnings = atoi(trimmed)
			} else {
				warnings++
				if len(warningLines) < 5 {
					warningLines = append(warningLines, shared.TruncateLine(trimmed, 100))
				}
			}
			continue
		}
	}

	if succeeded && errors == 0 {
		result.WriteString("Build: succeeded\n")
	} else {
		result.WriteString(fmt.Sprintf("Build: %d error(s), %d warning(s)\n", errors, warnings))
	}

	if len(errorLines) > 0 {
		result.WriteString("\nErrors:\n")
		for i, e := range errorLines {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errorLines)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if len(warningLines) > 0 {
		result.WriteString(fmt.Sprintf("\nWarnings: %d total\n", warnings))
		for _, w := range warningLines {
			result.WriteString(fmt.Sprintf("  %s\n", w))
		}
	}

	if result.Len() == 0 {
		return filterDotnetOutput(output)
	}
	return result.String()
}

func filterDotnetTest(output string) string {
	var result []string
	var passed, failed, skipped int
	var failureDetails []string
	var inFailure bool
	var currentFailure []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Passed!") || strings.Contains(trimmed, "Total:") {
			result = append(result, trimmed)
		}

		if strings.Contains(trimmed, "passed") && !strings.Contains(trimmed, "0 passed") {
			passed = atoi(trimmed)
		}
		if strings.Contains(trimmed, "failed") && !strings.Contains(trimmed, "0 failed") {
			failed = atoi(trimmed)
		}
		if strings.Contains(trimmed, "skipped") && !strings.Contains(trimmed, "0 skipped") {
			skipped = atoi(trimmed)
		}

		if strings.Contains(trimmed, "Failed!") || strings.Contains(trimmed, "[FAIL]") {
			inFailure = true
			currentFailure = []string{shared.TruncateLine(trimmed, 80)}
		}

		if inFailure {
			if trimmed == "" || strings.Contains(trimmed, "Passed!") || strings.Contains(trimmed, "Total") {
				if len(currentFailure) > 0 {
					failureDetails = append(failureDetails, strings.Join(currentFailure, "\n"))
				}
				inFailure = false
				currentFailure = nil
			} else {
				currentFailure = append(currentFailure, shared.TruncateLine(trimmed, 80))
			}
		}

		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			result = append(result, shared.TruncateLine(trimmed, 100))
		}
	}

	var summary []string
	summary = append(summary, "Test Results:")
	if passed > 0 {
		summary = append(summary, fmt.Sprintf("  %d passed", passed))
	}
	if failed > 0 {
		summary = append(summary, fmt.Sprintf("  %d failed", failed))
	}
	if skipped > 0 {
		summary = append(summary, fmt.Sprintf("  %d skipped", skipped))
	}

	if len(failureDetails) > 0 {
		summary = append(summary, "")
		summary = append(summary, "Failures:")
		for i, f := range failureDetails {
			if i >= 5 {
				summary = append(summary, fmt.Sprintf("  ... +%d more", len(failureDetails)-5))
				break
			}
			for _, l := range strings.Split(f, "\n") {
				if len(l) > 3 {
					summary = append(summary, fmt.Sprintf("  %s", l))
				}
			}
		}
	}

	if len(summary) > 1 {
		return strings.Join(summary, "\n")
	}

	if len(result) > 0 {
		return strings.Join(result, "\n")
	}

	return filterDotnetOutput(output)
}

func filterDotnetRestore(output string) string {
	var result strings.Builder
	var restored bool
	var projectCount int
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Restored") {
			restored = true
			projectCount++
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if restored {
		result.WriteString(fmt.Sprintf("Restore: %d project(s) restored\n", projectCount))
	} else {
		result.WriteString("Restore: completed\n")
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	return result.String()
}

func filterDotnetFormat(output string) string {
	var result strings.Builder
	var formatted, filesNeedFormatting int
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Format complete") || strings.Contains(trimmed, "was formatted") {
			formatted++
		}
		if strings.Contains(trimmed, "needs formatting") || strings.Contains(trimmed, "would be formatted") {
			filesNeedFormatting++
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if formatted > 0 {
		result.WriteString(fmt.Sprintf("Format: %d file(s) formatted\n", formatted))
	} else if filesNeedFormatting > 0 {
		result.WriteString(fmt.Sprintf("Format check: %d file(s) need formatting\n", filesNeedFormatting))
	} else if len(errors) == 0 {
		result.WriteString("Format: all files OK\n")
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	return result.String()
}

func filterDotnetRun(output string) string {
	var result strings.Builder
	var errors []string
	var warnings []string
	var programOutput []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
			continue
		}
		if strings.Contains(trimmed, "warning") || strings.Contains(trimmed, "Warning") {
			warnings = append(warnings, shared.TruncateLine(trimmed, 80))
			continue
		}

		if strings.HasPrefix(trimmed, "Building") || strings.HasPrefix(trimmed, "Running") {
			continue
		}

		programOutput = append(programOutput, line)
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if len(warnings) > 0 {
		result.WriteString(fmt.Sprintf("Warnings: %d\n", len(warnings)))
	}

	if len(programOutput) > 0 {
		if len(programOutput) > 20 {
			for _, line := range programOutput[:10] {
				result.WriteString(line + "\n")
			}
			result.WriteString(fmt.Sprintf("... (%d more lines)\n", len(programOutput)-20))
			for _, line := range programOutput[len(programOutput)-10:] {
				result.WriteString(line + "\n")
			}
		} else {
			for _, line := range programOutput {
				result.WriteString(line + "\n")
			}
		}
	}

	if result.Len() == 0 {
		return "Run: completed\n"
	}
	return result.String()
}

func filterDotnetPublish(output string) string {
	var result strings.Builder
	var published bool
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "Published") || strings.Contains(trimmed, "publish") {
			published = true
			result.WriteString(shared.TruncateLine(trimmed, 100) + "\n")
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if !published && len(errors) == 0 {
		result.WriteString("Publish: completed\n")
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	return result.String()
}

func filterDotnetOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	var errors, warnings int

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.Contains(line, "Microsoft ") ||
			strings.Contains(line, "  Determining projects to restore") ||
			strings.Contains(line, "  Restored ") ||
			strings.Contains(line, "  dotnet ") ||
			(strings.HasPrefix(line, "  ") && !strings.Contains(line, "error") && !strings.Contains(line, "warning")) {
			continue
		}

		if strings.Contains(strings.ToLower(line), "error") {
			errors++
		}
		if strings.Contains(strings.ToLower(line), "warning") {
			warnings++
		}

		if len(line) > 100 {
			line = line[:97] + "..."
		}

		result = append(result, line)
	}

	if len(result) > 0 {
		result = append(result, fmt.Sprintf("---\n%d errors, %d warnings", errors, warnings))
	}

	return strings.Join(result, "\n")
}
