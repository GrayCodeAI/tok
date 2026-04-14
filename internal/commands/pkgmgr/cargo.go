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

var cargoCmd = &cobra.Command{
	Use:   "cargo [subcommand] [args...]",
	Short: "Cargo commands with compact output",
	Long: `Cargo commands with token-optimized output.

Subcommands:
  build   - Build with compact output (strip Compiling lines, keep errors)
  test    - Test with failures-only output
  nextest - Nextest test runner with failures-only output
  check   - Check with compact output
  clippy  - Clippy with warnings grouped by lint rule
  run     - Run with compact output
  doc     - Doc generation with compact output
  fmt     - Format check with compact output
  clean   - Clean build artifacts

Examples:
  tokman cargo build
  tokman cargo test --lib
  tokman cargo clippy -- -W clippy::all
  tokman cargo run
  tokman cargo doc --open`,
	RunE: runCargo,
}

func init() {
	registry.Add(func() { registry.Register(cargoCmd) })
}

func runCargo(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	subcommand := args[0]
	cargoArgs := append([]string{}, args...)

	c := exec.Command("cargo", cargoArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	var filtered string
	switch subcommand {
	case "build", "check":
		if shared.UltraCompact {
			filtered = filterCargoBuildUltraCompact(output)
		} else {
			filtered = filterCargoBuild(output)
		}
	case "test":
		if shared.UltraCompact {
			filtered = filterCargoTestUltraCompact(output)
		} else {
			filtered = filterCargoTest(output)
		}
	case "nextest":
		filtered = filterCargoNextest(output)
	case "clippy":
		filtered = filterCargoClippy(output)
	case "install":
		filtered = filterCargoInstall(output)
	case "run":
		filtered = filterCargoRun(output)
	case "doc":
		filtered = filterCargoDoc(output)
	case "fmt":
		filtered = filterCargoFmt(output)
	case "clean":
		filtered = filterCargoClean(output)
	default:
		filtered = output
	}

	if err != nil {
		if hint := shared.TeeOnFailure(output, "cargo_"+subcommand, err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	fmt.Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("cargo %s", strings.Join(args, " ")), "tokman cargo", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func filterCargoBuild(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Compiling") {
			continue
		}
		if strings.HasPrefix(line, "Finished") {
			result.WriteString("✓ build complete\n")
			continue
		}
		if strings.Contains(line, "error") || strings.Contains(line, "warning") {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}

func filterCargoTest(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "test result:") {
			result.WriteString(line + "\n")
			continue
		}
		if strings.Contains(line, "FAILED") || strings.Contains(line, "----") {
			result.WriteString(line + "\n")
			continue
		}
		if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}

func filterCargoNextest(output string) string {
	var result strings.Builder
	var failures []string
	var summary string

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "test result:") || strings.Contains(line, "passed") && strings.Contains(line, "failed") {
			summary = line
			continue
		}
		if strings.HasPrefix(line, "FAIL") || strings.Contains(line, "[FAIL]") {
			failures = append(failures, line)
		}
		if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			failures = append(failures, line)
		}
	}

	if len(failures) > 0 {
		result.WriteString(fmt.Sprintf("Failures (%d):\n", len(failures)))
		for _, f := range failures {
			result.WriteString("  " + shared.TruncateLine(f, 80) + "\n")
		}
	}

	if summary != "" {
		result.WriteString(summary + "\n")
	} else if result.Len() == 0 {
		result.WriteString("✓ all tests passed\n")
	}

	return result.String()
}

func filterCargoClippy(output string) string {
	warnings := make(map[string][]string)
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "warning:") {
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 4 {
				warnType := strings.TrimSpace(parts[3])
				warnings[warnType] = append(warnings[warnType], line)
			}
		} else if strings.Contains(line, "error:") {
			errors = append(errors, line)
		}
	}

	var result strings.Builder
	if len(warnings) > 0 {
		result.WriteString(fmt.Sprintf("Warnings (%d types):\n", len(warnings)))
		for wtype, lines := range warnings {
			result.WriteString(fmt.Sprintf("  %s: %d occurrences\n", wtype, len(lines)))
		}
	}
	for _, e := range errors {
		result.WriteString(e + "\n")
	}

	return result.String()
}

func filterCargoBuildUltraCompact(output string) string {
	errors := 0
	warnings := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "error") {
			errors++
		}
		if strings.Contains(line, "warning") {
			warnings++
		}
	}
	if errors > 0 {
		return fmt.Sprintf("build failed: %d errors, %d warnings", errors, warnings)
	}
	if warnings > 0 {
		return fmt.Sprintf("build ok: %d warnings", warnings)
	}
	return "build ok"
}

func filterCargoTestUltraCompact(output string) string {
	passed := 0
	failed := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "test result:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "passed" && i > 0 {
					if _, err := fmt.Sscanf(parts[i-1], "%d", &passed); err != nil {
						passed = 0
					}
				}
				if p == "failed" && i > 0 {
					if _, err := fmt.Sscanf(parts[i-1], "%d", &failed); err != nil {
						failed = 0
					}
				}
			}
		}
	}
	if failed > 0 {
		return fmt.Sprintf("tests: %d passed, %d failed", passed, failed)
	}
	return fmt.Sprintf("tests: %d passed", passed)
}

func filterCargoInstall(output string) string {
	lines := strings.Split(output, "\n")
	var installed []string
	var updated []string
	var compiling []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Installing") {
			pkg := strings.TrimPrefix(line, "Installing ")
			installed = append(installed, pkg)
		} else if strings.HasPrefix(line, "Updating") {
			pkg := strings.TrimPrefix(line, "Updating ")
			updated = append(updated, pkg)
		} else if strings.HasPrefix(line, "Compiling") {
			// Skip compilation noise
			continue
		} else if strings.HasPrefix(line, "Finished") || strings.HasPrefix(line, "error") {
			// Keep these important lines
			compiling = append(compiling, line)
		}
	}

	var result strings.Builder

	if len(installed) > 0 {
		result.WriteString(fmt.Sprintf("Installed %d package(s):\n", len(installed)))
		for _, pkg := range installed {
			result.WriteString(fmt.Sprintf("  ✓ %s\n", pkg))
		}
	}

	if len(updated) > 0 {
		result.WriteString(fmt.Sprintf("🔄 Updated %d package(s):\n", len(updated)))
		for _, pkg := range updated {
			result.WriteString(fmt.Sprintf("  ✓ %s\n", pkg))
		}
	}

	if len(compiling) > 0 {
		for _, line := range compiling {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	if result.Len() == 0 {
		return "OK Install complete"
	}

	return result.String()
}

func filterCargoRun(output string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	var errors []string
	var warnings []string
	var programOutput []string
	inProgramOutput := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip compilation lines
		if strings.HasPrefix(trimmed, "Compiling") ||
			strings.HasPrefix(trimmed, "Finished") ||
			strings.HasPrefix(trimmed, "Running") {
			continue
		}

		// Capture errors
		if strings.Contains(line, "error[") || strings.Contains(line, "error:") {
			errors = append(errors, line)
			continue
		}

		// Capture warnings
		if strings.Contains(line, "warning:") {
			warnings = append(warnings, line)
			continue
		}

		// Everything else is program output
		programOutput = append(programOutput, line)
		inProgramOutput = true
	}

	// Show errors first
	if len(errors) > 0 {
		result.WriteString("FAIL Build Errors:\n")
		for _, err := range errors {
			result.WriteString(fmt.Sprintf("  %s\n", err))
		}
		result.WriteString("\n")
	}

	// Show warnings
	if len(warnings) > 0 {
		result.WriteString(fmt.Sprintf("WARN %d warning(s)\n", len(warnings)))
		if len(warnings) <= 5 {
			for _, warn := range warnings {
				result.WriteString(fmt.Sprintf("  %s\n", warn))
			}
		}
		result.WriteString("\n")
	}

	// Show program output (limited)
	if inProgramOutput && len(programOutput) > 0 {
		if len(programOutput) > 20 {
			for _, line := range programOutput[:10] {
				result.WriteString(line + "\n")
			}
			result.WriteString(fmt.Sprintf("... (%d more lines) ...\n", len(programOutput)-20))
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
		return "OK Program executed successfully"
	}

	return result.String()
}

func filterCargoDoc(output string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	var documented int
	var errors []string
	var docPath string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Count documented items
		if strings.HasPrefix(trimmed, "Documenting") {
			documented++
			continue
		}

		// Skip compilation lines
		if strings.HasPrefix(trimmed, "Compiling") ||
			strings.HasPrefix(trimmed, "Finished") {
			continue
		}

		// Capture errors
		if strings.Contains(line, "error[") || strings.Contains(line, "error:") {
			errors = append(errors, line)
			continue
		}

		// Capture doc path
		if strings.Contains(line, "file://") {
			docPath = line
		}
	}

	if documented > 0 {
		result.WriteString(fmt.Sprintf("Documented %d crate(s)\n", documented))
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nFAIL %d error(s):\n", len(errors)))
		for _, err := range errors {
			result.WriteString(fmt.Sprintf("  %s\n", err))
		}
	}

	if docPath != "" {
		result.WriteString(fmt.Sprintf("\nDocumentation: %s\n", docPath))
	}

	if result.Len() == 0 {
		return "OK Documentation generated"
	}

	return result.String()
}

func filterCargoFmt(output string) string {
	lines := strings.Split(output, "\n")
	var formatted []string
	var errors []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for formatted files
		if strings.Contains(line, "Formatting") || strings.Contains(line, "formatted") {
			formatted = append(formatted, line)
			continue
		}

		// Check for errors
		if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			errors = append(errors, line)
		}
	}

	if len(errors) > 0 {
		return fmt.Sprintf("FAIL Format check failed: %d error(s)", len(errors))
	}

	if len(formatted) > 0 {
		return fmt.Sprintf("OK Formatted %d file(s)", len(formatted))
	}

	return "OK All files formatted correctly"
}

func filterCargoClean(output string) string {
	if strings.Contains(output, "error") || strings.Contains(output, "Error") {
		return "FAIL Clean failed"
	}
	return "OK Build artifacts cleaned"
}
