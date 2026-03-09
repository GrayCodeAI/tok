package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var formatCmd = &cobra.Command{
	Use:   "format [args...]",
	Short: "Universal format checker (prettier, black, ruff format, biome)",
	Long: `Auto-detect and run the appropriate formatter for your project.

Detects formatter from project files:
- .prettierrc, .prettierrc.json, package.json → prettier
- pyproject.toml with [tool.ruff] → ruff format
- pyproject.toml with [tool.black] → black
- biome.json → biome
- go.mod → gofmt

Examples:
  tokman format --check .
  tokman format --write .
  tokman format prettier .      # Explicit formatter
  tokman format ruff --check .`,
	DisableFlagParsing: true,
	RunE:               runFormat,
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

func runFormat(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Detect if first arg is an explicit formatter
	var formatter string
	var formatterArgs []string

	if len(args) > 0 {
		firstArg := args[0]
		switch firstArg {
		case "prettier", "black", "ruff", "biome", "gofmt":
			formatter = firstArg
			formatterArgs = args[1:]
		default:
			formatter = detectFormatter()
			formatterArgs = args
		}
	} else {
		formatter = detectFormatter()
		formatterArgs = []string{"--check", "."}
	}

	if verbose > 0 {
		fmt.Fprintf(os.Stderr, "Detected formatter: %s\n", formatter)
	}

	// Build command based on formatter
	var execCmd *exec.Cmd
	switch formatter {
	case "prettier":
		execCmd = exec.Command("prettier", formatterArgs...)
	case "black":
		// Inject --check if not present for check mode
		hasCheck := false
		for _, arg := range formatterArgs {
			if arg == "--check" || arg == "--diff" {
				hasCheck = true
				break
			}
		}
		if !hasCheck {
			formatterArgs = append([]string{"--check"}, formatterArgs...)
		}
		execCmd = exec.Command("black", formatterArgs...)
	case "ruff":
		// Add "format" subcommand if not present
		if len(formatterArgs) == 0 || !strings.HasPrefix(formatterArgs[0], "format") {
			formatterArgs = append([]string{"format"}, formatterArgs...)
		}
		execCmd = exec.Command("ruff", formatterArgs...)
	case "biome":
		// Use npx biome or direct biome
		execCmd = packageManagerExec("biome", append([]string{"format"}, formatterArgs...)...)
	case "gofmt":
		execCmd = exec.Command("gofmt", formatterArgs...)
	default:
		execCmd = exec.Command("prettier", formatterArgs...)
	}

	// Default to current directory if no path specified
	hasPath := false
	for _, arg := range formatterArgs {
		if !strings.HasPrefix(arg, "-") {
			hasPath = true
			break
		}
	}
	if !hasPath && (formatter == "prettier" || formatter == "biome") {
		execCmd.Args = append(execCmd.Args, ".")
	}

	if verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: %s %s\n", formatter, strings.Join(formatterArgs, " "))
	}

	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Filter output based on formatter
	filtered := filterFormatOutput(raw, formatter)
	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("%s %s", formatter, strings.Join(formatterArgs, " ")), "tokman format", originalTokens, filteredTokens)

	// Preserve exit code for CI/CD
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}

	return nil
}

// packageManagerExec tries npx, pnpm exec, or direct command
func packageManagerExec(cmd string, args ...string) *exec.Cmd {
	// Try npx first
	npxCmd := exec.Command("npx", append([]string{cmd}, args...)...)
	if _, err := exec.LookPath("npx"); err == nil {
		return npxCmd
	}
	// Try pnpm exec
	pnpmCmd := exec.Command("pnpm", append([]string{"exec", cmd}, args...)...)
	if _, err := exec.LookPath("pnpm"); err == nil {
		return pnpmCmd
	}
	// Fallback to direct command
	return exec.Command(cmd, args...)
}

func detectFormatter() string {
	// Check for biome config first
	if _, err := os.Stat("biome.json"); err == nil {
		return "biome"
	}

	// Check for prettier config
	if _, err := os.Stat(".prettierrc"); err == nil {
		return "prettier"
	}
	if _, err := os.Stat(".prettierrc.json"); err == nil {
		return "prettier"
	}
	if _, err := os.Stat(".prettierrc.js"); err == nil {
		return "prettier"
	}

	// Check for Python formatters (priority: ruff > black)
	if _, err := os.Stat("pyproject.toml"); err == nil {
		content, _ := os.ReadFile("pyproject.toml")
		contentStr := string(content)
		// Check for [tool.ruff] or [tool.ruff.format]
		if strings.Contains(contentStr, "[tool.ruff") {
			return "ruff"
		}
		// Check for [tool.black]
		if strings.Contains(contentStr, "[tool.black]") {
			return "black"
		}
		// Default to ruff for Python projects
		return "ruff"
	}
	if _, err := os.Stat("setup.cfg"); err == nil {
		return "black"
	}
	if _, err := os.Stat(".python-version"); err == nil {
		return "ruff"
	}

	// Check for Go
	if _, err := os.Stat("go.mod"); err == nil {
		return "gofmt"
	}

	// Check for package.json (prettier is common for JS/TS)
	if _, err := os.Stat("package.json"); err == nil {
		return "prettier"
	}

	// Default to prettier
	return "prettier"
}

func filterFormatOutput(raw, formatter string) string {
	if raw == "" {
		return "✓ Format: All files formatted correctly"
	}

	switch formatter {
	case "black":
		return filterBlackOutput(raw)
	case "ruff":
		return filterRuffFormatOutput(raw)
	case "biome":
		return filterBiomeOutput(raw)
	default:
		return filterPrettierOutput(raw)
	}
}

func filterBlackOutput(output string) string {
	var filesToFormat []string
	var filesUnchanged int
	var filesWouldReformat int
	var allDone bool
	var ohNo bool

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Check for "would reformat" lines
		if strings.Contains(lower, "would reformat:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				filesToFormat = append(filesToFormat, strings.TrimSpace(parts[1]))
			}
		}

		// Parse summary line
		if strings.Contains(lower, "would be reformatted") || strings.Contains(lower, "left unchanged") {
			// Parse counts from summary
			if strings.Contains(lower, "would be reformatted") {
				fields := strings.Fields(trimmed)
				for i, f := range fields {
					if f == "file" || f == "files" {
						if i > 0 {
							fmt.Sscanf(fields[i-1], "%d", &filesWouldReformat)
						}
						break
					}
				}
			}
			if strings.Contains(lower, "left unchanged") {
				fields := strings.Fields(trimmed)
				for i, f := range fields {
					if f == "file" || f == "files" {
						if i > 0 {
							fmt.Sscanf(fields[i-1], "%d", &filesUnchanged)
						}
						break
					}
				}
			}
		}

		// Check for success/failure indicators
		if strings.Contains(lower, "all done") {
			allDone = true
		}
		if strings.Contains(lower, "oh no") {
			ohNo = true
		}
	}

	needsFormatting := len(filesToFormat) > 0 || filesWouldReformat > 0 || ohNo

	if !needsFormatting && (allDone || filesUnchanged > 0) {
		result := "✓ Format (black): All files formatted"
		if filesUnchanged > 0 {
			result += fmt.Sprintf(" (%d files checked)", filesUnchanged)
		}
		return result
	}

	if needsFormatting {
		count := len(filesToFormat)
		if count == 0 {
			count = filesWouldReformat
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Format (black): %d files need formatting\n", count))
		result.WriteString("═══════════════════════════════════════\n")

		for i, file := range filesToFormat {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("\n... +%d more files\n", len(filesToFormat)-10))
				break
			}
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, compactPath(file)))
		}

		if filesUnchanged > 0 {
			result.WriteString(fmt.Sprintf("\n✓ %d files already formatted\n", filesUnchanged))
		}

		result.WriteString("\n💡 Run `black .` to format these files\n")
		return result.String()
	}

	return output
}

func filterRuffFormatOutput(output string) string {
	var filesToFormat []string
	var filesUnchanged int

	outputLower := strings.ToLower(output)

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Check for "would reformat" lines
		if strings.Contains(lower, "would reformat:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				filesToFormat = append(filesToFormat, strings.TrimSpace(parts[1]))
			}
		}
	}

	// Check for all formatted
	if len(filesToFormat) == 0 && strings.Contains(outputLower, "left unchanged") {
		// Parse count
		fields := strings.Fields(output)
		for i, f := range fields {
			if f == "file" || f == "files" {
				if i > 0 {
					fmt.Sscanf(fields[i-1], "%d", &filesUnchanged)
				}
				break
			}
		}
		return fmt.Sprintf("✓ Ruff format: All files formatted correctly (%d files)", filesUnchanged)
	}

	// Files need formatting
	if len(filesToFormat) > 0 || strings.Contains(outputLower, "would reformat") {
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Ruff format: %d files need formatting\n", len(filesToFormat)))
		result.WriteString("═══════════════════════════════════════\n")

		for i, file := range filesToFormat {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("\n... +%d more files\n", len(filesToFormat)-10))
				break
			}
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, compactPath(file)))
		}

		result.WriteString("\n💡 Run `ruff format` to format these files\n")
		return result.String()
	}

	// Write mode
	if strings.Contains(outputLower, "reformatted") {
		return "✓ Ruff format: Files formatted"
	}

	return output
}

func filterBiomeOutput(output string) string {
	// Biome outputs similar format to prettier
	return filterPrettierOutput(output)
}

func filterPrettierOutput(output string) string {
	var filesToFormat []string
	var filesChecked int

	// Check for success message first
	outputLower := strings.ToLower(output)
	if strings.Contains(outputLower, "all matched files use prettier") {
		return "✓ Prettier: All files formatted correctly"
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)

		// Skip info lines
		if strings.Contains(lower, "checking") || strings.Contains(lower, "parsing") {
			continue
		}

		// Count files that need formatting
		if !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "+") {
			// Check if it's a file path
			if !strings.Contains(lower, "error") && !strings.Contains(lower, "warning") {
				filesToFormat = append(filesToFormat, trimmed)
				filesChecked++
			}
		}
	}

	// All formatted
	if len(filesToFormat) == 0 {
		return "✓ Prettier: All files formatted correctly"
	}

	// Check mode output
	if len(filesToFormat) > 0 {
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Prettier: %d files need formatting\n", len(filesToFormat)))
		result.WriteString("═══════════════════════════════════════\n")

		for i, file := range filesToFormat {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("\n... +%d more files\n", len(filesToFormat)-10))
				break
			}
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, compactPath(file)))
		}

		if filesChecked > len(filesToFormat) {
			result.WriteString(fmt.Sprintf("\n✓ %d files already formatted\n", filesChecked-len(filesToFormat)))
		}

		result.WriteString("\n💡 Run `prettier --write .` to format these files\n")
		return result.String()
	}

	// Write mode
	if strings.Contains(strings.ToLower(output), "modified") || strings.Contains(strings.ToLower(output), "formatted") {
		return "✓ Prettier: Files formatted"
	}

	return "✓ Format: All files formatted correctly"
}

// compactPath shortens file paths
func compactPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	// Try to find src, lib, tests prefix
	prefixes := []string{"/src/", "/lib/", "/tests/", "/test/"}
	for _, prefix := range prefixes {
		if idx := strings.LastIndex(path, prefix); idx >= 0 {
			return path[idx+1:]
		}
	}

	// Just return filename
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}

	return path
}
