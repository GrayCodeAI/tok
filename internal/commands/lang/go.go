package lang

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var goCmd = &cobra.Command{
	Use:   "go [args...]",
	Short: "Go commands with compact output",
	Long: `Execute Go commands with token-optimized output.

Provides compact output for test, build, vet, and other go commands.

Examples:
  tok go test ./...
  tok go build ./...
  tok go vet ./...`,
	DisableFlagParsing: true,
	RunE:               runGo,
}

func init() {
	registry.Add(func() { registry.Register(goCmd) })
}

func runGo(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"help"}
	}

	// Route to specialized handlers
	switch args[0] {
	case "test":
		return runGoTestCmd(args[1:])
	case "build":
		return runGoBuildCmd(args[1:])
	case "vet":
		return runGoVet(args[1:])
	case "mod":
		return runGoMod(args[1:])
	case "doc":
		return runGoDoc(args[1:])
	case "list":
		return runGoList(args[1:])
	case "env":
		return runGoEnv(args)
	default:
		return runGoPassthrough(args)
	}
}

func runGoTestCmd(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: go test %s\n", strings.Join(args, " "))
	}

	// Use -json for structured output
	jsonArgs := append([]string{"test", "-json"}, args...)
	execCmd := exec.Command("go", jsonArgs...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGoTestOutput(raw)

	// Add tee hint on failure
	if err != nil {
		if hint := shared.TeeOnFailure(raw, "go_test", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go test %s", strings.Join(args, " ")), "tok go test", originalTokens, filteredTokens)

	return err
}

func runGoBuildCmd(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: go build %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("go", append([]string{"build"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGoBuildOutput(raw)

	// Add tee hint on failure
	if err != nil {
		if hint := shared.TeeOnFailure(raw, "go_build", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go build %s", strings.Join(args, " ")), "tok go build", originalTokens, filteredTokens)

	return err
}

func runGoVet(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: go vet %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("go", append([]string{"vet"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGoVetOutput(raw)
	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go vet %s", strings.Join(args, " ")), "tok go vet", originalTokens, filteredTokens)

	return err
}

func runGoPassthrough(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: go %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("go", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Basic filtering
	filtered := filterGoOutput(raw)
	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go %s", strings.Join(args, " ")), "tok go", originalTokens, filteredTokens)

	return err
}

// Filter functions

type GoTestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

func filterGoTestOutput(raw string) string {
	var passed, failed, skipped int
	var failures []string
	var packageResults = make(map[string][]string)

	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			continue
		}

		var event GoTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Action {
		case "pass":
			if event.Test == "" {
				// Package pass
				packageResults[event.Package] = append(packageResults[event.Package], "PASS")
			} else {
				passed++
			}
		case "fail":
			if event.Test == "" {
				// Package fail
				packageResults[event.Package] = append(packageResults[event.Package], "FAIL")
			} else {
				failed++
				failures = append(failures, fmt.Sprintf("%s.%s", event.Package, event.Test))
			}
		case "skip":
			skipped++
		}
	}

	// Ultra-compact mode
	if shared.UltraCompact {
		return filterGoTestOutputUltraCompact(passed, failed, skipped, failures, packageResults)
	}

	var result []string
	result = append(result, "Go Test Results:")
	result = append(result, fmt.Sprintf("   OK %d passed", passed))
	if failed > 0 {
		result = append(result, fmt.Sprintf("   FAIL %d failed", failed))
	}
	if skipped > 0 {
		result = append(result, fmt.Sprintf("   SKIP %d skipped", skipped))
	}

	// Package summary
	if len(packageResults) > 0 {
		result = append(result, "")
		result = append(result, "Packages:")
		for pkg, status := range packageResults {
			result = append(result, fmt.Sprintf("   %s: %s", pkg, strings.Join(status, ", ")))
		}
	}

	if len(failures) > 0 {
		result = append(result, "")
		result = append(result, "Failures:")
		for i, f := range failures {
			if i >= 10 {
				result = append(result, fmt.Sprintf("   ... +%d more", len(failures)-10))
				break
			}
			result = append(result, fmt.Sprintf("   • %s", f))
		}
	}

	return strings.Join(result, "\n")
}

func filterGoTestOutputUltraCompact(passed, failed, skipped int, failures []string, packageResults map[string][]string) string {
	var parts []string

	// Summary on one line
	parts = append(parts, fmt.Sprintf("P:%d", passed))
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("F:%d", failed))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("S:%d", skipped))
	}

	var result []string
	result = append(result, strings.Join(parts, " "))

	// Package status (one per line, limited)
	pkgCount := 0
	for pkg, status := range packageResults {
		if pkgCount >= 5 {
			result = append(result, fmt.Sprintf("... +%d more pkgs", len(packageResults)-5))
			break
		}
		statusStr := "PASS"
		for _, s := range status {
			if strings.Contains(s, "FAIL") {
				statusStr = "FAIL"
				break
			}
		}
		// Shorten package path
		shortPkg := pkg
		if idx := strings.LastIndex(pkg, "/"); idx >= 0 {
			shortPkg = pkg[idx+1:]
		}
		result = append(result, fmt.Sprintf("%s: %s", shortPkg, statusStr))
		pkgCount++
	}

	// Failures (limited to 5)
	if len(failures) > 0 {
		for i, f := range failures {
			if i >= 5 {
				result = append(result, fmt.Sprintf("... +%d more failures", len(failures)-5))
				break
			}
			// Shorten failure name
			parts := strings.Split(f, ".")
			if len(parts) >= 2 {
				result = append(result, fmt.Sprintf("FAIL: %s", parts[len(parts)-1]))
			} else {
				result = append(result, fmt.Sprintf("FAIL: %s", f))
			}
		}
	}

	return strings.Join(result, "\n")
}

func filterGoBuildOutput(raw string) string {
	if raw == "" {
		return "OK Build successful"
	}

	lines := strings.Split(raw, "\n")
	var errors []string
	var warnings []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") {
			errors = append(errors, shared.TruncateLine(line, 100))
		} else if strings.Contains(lower, "warning") {
			warnings = append(warnings, shared.TruncateLine(line, 100))
		}
	}

	var result []string
	if len(errors) > 0 {
		result = append(result, fmt.Sprintf("FAIL Errors (%d):", len(errors)))
		for _, e := range errors {
			result = append(result, fmt.Sprintf("   %s", e))
		}
	}

	if len(warnings) > 0 {
		result = append(result, fmt.Sprintf("WARN Warnings (%d):", len(warnings)))
		for _, w := range warnings {
			result = append(result, fmt.Sprintf("   %s", w))
		}
	}

	if len(result) == 0 && raw != "" {
		// No errors/warnings detected, but output exists
		return raw
	}

	if len(result) == 0 {
		return "OK Build successful"
	}
	return strings.Join(result, "\n")
}

func filterGoVetOutput(raw string) string {
	if raw == "" {
		return "OK No vet issues found"
	}

	lines := strings.Split(raw, "\n")
	var issues []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			issues = append(issues, shared.TruncateLine(line, 100))
		}
	}

	if len(issues) == 0 {
		return "OK No vet issues found"
	}

	var result []string
	result = append(result, fmt.Sprintf("WARN Vet Issues (%d):", len(issues)))
	for i, issue := range issues {
		if i >= 15 {
			result = append(result, fmt.Sprintf("   ... +%d more", len(issues)-15))
			break
		}
		result = append(result, fmt.Sprintf("   %s", issue))
	}
	return strings.Join(result, "\n")
}

func filterGoOutput(raw string) string {
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

func runGoMod(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"help"}
	}

	execCmd := exec.Command("go", append([]string{"mod"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	filtered := filterGoModOutput(raw, args)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go mod %s", strings.Join(args, " ")), "tok go mod", originalTokens, filteredTokens)

	return err
}

func filterGoModOutput(raw string, args []string) string {
	if len(args) == 0 {
		return raw
	}

	switch args[0] {
	case "tidy":
		if strings.TrimSpace(raw) == "" {
			return "go mod tidy: clean\n"
		}
		return filterGoOutput(raw)
	case "verify":
		if strings.Contains(raw, "all modules verified") {
			return "All modules verified\n"
		}
		return filterGoOutput(raw)
	case "graph":
		var result strings.Builder
		lines := strings.Split(raw, "\n")
		for i, line := range lines {
			if i >= 30 {
				result.WriteString(fmt.Sprintf("... (%d more lines)\n", len(lines)-30))
				break
			}
			if strings.TrimSpace(line) != "" {
				result.WriteString(shared.TruncateLine(line, 80) + "\n")
			}
		}
		return result.String()
	case "why":
		return shared.TruncateLine(raw, 200) + "\n"
	default:
		return raw
	}
}

func runGoDoc(args []string) error {
	timer := tracking.Start()

	execCmd := exec.Command("go", append([]string{"doc"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	var result strings.Builder
	lineCount := 0
	for _, line := range strings.Split(raw, "\n") {
		if lineCount >= 40 {
			result.WriteString(fmt.Sprintf("... (%d more lines)\n", strings.Count(raw, "\n")-40))
			break
		}
		result.WriteString(line + "\n")
		lineCount++
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go doc %s", strings.Join(args, " ")), "tok go doc", originalTokens, filteredTokens)

	return err
}

func runGoList(args []string) error {
	timer := tracking.Start()

	execCmd := exec.Command("go", append([]string{"list"}, args...)...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) > 30 {
		filtered := strings.Join(lines[:30], "\n") + fmt.Sprintf("\n... +%d more", len(lines)-30)
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track(fmt.Sprintf("go list %s", strings.Join(args, " ")), "tok go list", originalTokens, filteredTokens)
		return err
	}

	out.Global().Print(raw)
	originalTokens := filter.EstimateTokens(raw)
	timer.Track(fmt.Sprintf("go list %s", strings.Join(args, " ")), "tok go list", originalTokens, originalTokens)

	return err
}

func runGoEnv(args []string) error {
	timer := tracking.Start()

	execCmd := exec.Command("go", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	var result strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		result.WriteString(trimmed + "\n")
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("go %s", strings.Join(args, " ")), "tok go", originalTokens, filteredTokens)

	return err
}
