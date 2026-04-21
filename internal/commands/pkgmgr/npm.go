package pkgmgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var npmJSON bool

func formatAsJSONnpm(output string) string {
	return fmt.Sprintf(`{"output": %s}`, strconv.Quote(output))
}

func atoi(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		n = 0
	}
	return n
}

var npmCmd = &cobra.Command{
	Use:   "npm [args...]",
	Short: "npm run with filtered output",
	Long: `npm run with token-optimized output.

Strips boilerplate and progress bars from npm output.
Special handling for npm test with 90% token reduction.

Examples:
  tok npm run build
  tok npm install
  tok npm test
  tok npm test -- --coverage`,
	DisableFlagParsing: true,
	RunE:               runNpm,
}

func init() {
	registry.Add(func() { registry.Register(npmCmd) })
	npmCmd.Flags().BoolVarP(&npmJSON, "json", "j", false, "Output as JSON")
}

func runNpm(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	if len(args) > 0 {
		switch args[0] {
		case "test":
			return runNpmTest(args[1:])
		case "install", "i", "add":
			return runNpmInstall(args[1:])
		case "ls", "list":
			return runNpmList(args[1:])
		case "outdated":
			return runNpmOutdated(args[1:])
		case "run":
			return runNpmRun(args[1:])
		case "audit":
			return runNpmAudit(args[1:])
		case "publish":
			return runNpmPublish(args[1:])
		}
	}

	npmArgs := append([]string{}, args...)

	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmOutput(output)

	if npmJSON {
		out.Global().Println(formatAsJSONnpm(output))
		originalTokens := filter.EstimateTokens(output)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track(fmt.Sprintf("npm %s", strings.Join(args, " ")), "tok npm", originalTokens, filteredTokens)
		return err
	}

	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm %s", strings.Join(args, " ")), "tok npm", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return err
}

func runNpmTest(args []string) error {
	timer := tracking.Start()

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: npm test %s\n", strings.Join(args, " "))
	}

	npmArgs := append([]string{"test"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmTestOutput(output)

	if err != nil {
		if hint := shared.TeeOnFailure(output, "npm_test", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm test %s", strings.Join(args, " ")), "tok npm test", originalTokens, filteredTokens)

	return err
}

func filterNpmOutput(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "\\") || strings.Contains(trimmed, "|") || strings.Contains(trimmed, "/") {
			continue
		}
		if strings.HasPrefix(trimmed, "npm WARN") && !strings.Contains(trimmed, "deprecated") {
			continue
		}
		if trimmed == "" {
			continue
		}

		result.WriteString(line + "\n")
	}
	return result.String()
}

func filterNpmTestOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	var passed, failed, skipped int
	var failures []string
	var inFailure bool
	var currentFailure []string
	var testSuitesPassed, testSuitesFailed int

	for _, line := range lines {
		origLine := line
		line = strings.TrimSpace(line)

		if strings.Contains(line, "PASS") {
			testSuitesPassed++
		}
		if strings.Contains(line, "FAIL") {
			testSuitesFailed++
			inFailure = true
			currentFailure = []string{origLine}
		}

		if strings.Contains(line, "Tests:") || strings.Contains(line, "passed") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "passed" || p == "passing" {
					if i > 0 {
						if _, err := fmt.Sscanf(parts[i-1], "%d", &passed); err != nil {
							passed = 0
						}
					}
				}
				if p == "failed" || p == "failing" {
					if i > 0 {
						if _, err := fmt.Sscanf(parts[i-1], "%d", &failed); err != nil {
							failed = 0
						}
					}
				}
				if p == "skipped" || p == "pending" {
					if i > 0 {
						if _, err := fmt.Sscanf(parts[i-1], "%d", &skipped); err != nil {
							skipped = 0
						}
					}
				}
			}
		}

		if inFailure {
			if strings.HasPrefix(line, "●") || strings.Contains(line, "expect(") ||
				strings.Contains(line, "AssertionError") || strings.Contains(line, "Error:") {
				currentFailure = append(currentFailure, origLine)
			} else if line == "" && len(currentFailure) > 1 {
				failures = append(failures, strings.Join(currentFailure, "\n"))
				inFailure = false
				currentFailure = nil
			}
		}

		if strings.Contains(line, "1) ") || strings.Contains(line, "2) ") || strings.Contains(line, "3) ") {
			failures = append(failures, line)
		}
	}

	if shared.UltraCompact {
		return filterNpmTestOutputUltraCompact(passed, failed, skipped, testSuitesPassed, testSuitesFailed, failures)
	}

	result = append(result, "npm test Results:")
	if testSuitesPassed > 0 || testSuitesFailed > 0 {
		result = append(result, fmt.Sprintf("   %d suites passed, %d suites failed", testSuitesPassed, testSuitesFailed))
	}
	if passed > 0 {
		result = append(result, fmt.Sprintf("   OK %d tests passed", passed))
	}
	if failed > 0 {
		result = append(result, fmt.Sprintf("   FAIL %d tests failed", failed))
	}
	if skipped > 0 {
		result = append(result, fmt.Sprintf("   SKIP %d tests skipped", skipped))
	}

	if len(failures) > 0 {
		result = append(result, "")
		result = append(result, "Failures:")
		for i, f := range failures {
			if i >= 5 {
				result = append(result, fmt.Sprintf("   ... +%d more failures", len(failures)-5))
				break
			}
			for _, l := range strings.Split(f, "\n") {
				if len(strings.TrimSpace(l)) > 3 {
					result = append(result, fmt.Sprintf("   %s", shared.TruncateLine(strings.TrimSpace(l), 80)))
				}
			}
		}
	}

	if passed == 0 && failed == 0 && len(result) <= 2 {
		result = result[:1]
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result = append(result, shared.TruncateLine(strings.TrimSpace(line), 100))
				if len(result) > 20 {
					result = append(result, fmt.Sprintf("   ... (%d more lines)", len(lines)-20))
					break
				}
			}
		}
	}

	return strings.Join(result, "\n")
}

func filterNpmTestOutputUltraCompact(passed, failed, skipped, suitesPassed, suitesFailed int, failures []string) string {
	var parts []string

	if suitesPassed > 0 || suitesFailed > 0 {
		parts = append(parts, fmt.Sprintf("S:%d/%d", suitesPassed, suitesPassed+suitesFailed))
	}

	parts = append(parts, fmt.Sprintf("P:%d", passed))
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("F:%d", failed))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("S:%d", skipped))
	}

	var result []string
	result = append(result, strings.Join(parts, " "))

	if len(failures) > 0 {
		for i, f := range failures {
			if i >= 3 {
				result = append(result, fmt.Sprintf("... +%d more", len(failures)-3))
				break
			}
			lines := strings.Split(f, "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l != "" && len(l) > 3 {
					result = append(result, shared.TruncateLine(l, 60))
					break
				}
			}
		}
	}

	return strings.Join(result, "\n")
}

func runNpmInstall(args []string) error {
	timer := tracking.Start()

	npmArgs := append([]string{"install"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmInstallOutput(output)
	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm install %s", strings.Join(args, " ")), "tok npm install", originalTokens, filteredTokens)

	return err
}

func filterNpmInstallOutput(output string) string {
	var added, removed, changed, audited int
	var vulnerabilities []string
	var warnings []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "added ") {
			added = atoi(strings.TrimPrefix(trimmed, "added "))
		}
		if strings.HasPrefix(trimmed, "removed ") {
			removed = atoi(strings.TrimPrefix(trimmed, "removed "))
		}
		if strings.HasPrefix(trimmed, "changed ") {
			changed = atoi(strings.TrimPrefix(trimmed, "changed "))
		}
		if strings.HasPrefix(trimmed, "audited ") {
			audited = atoi(strings.TrimPrefix(trimmed, "audited "))
		}
		if strings.Contains(trimmed, "vulnerabilities") {
			vulnerabilities = append(vulnerabilities, trimmed)
		}
		if strings.HasPrefix(trimmed, "npm WARN") && strings.Contains(trimmed, "deprecated") {
			warnings = append(warnings, trimmed)
		}
	}

	var result []string
	result = append(result, "Install Summary:")
	if added > 0 {
		result = append(result, fmt.Sprintf("  + %d added", added))
	}
	if removed > 0 {
		result = append(result, fmt.Sprintf("  - %d removed", removed))
	}
	if changed > 0 {
		result = append(result, fmt.Sprintf("  ~ %d changed", changed))
	}
	if audited > 0 {
		result = append(result, fmt.Sprintf("  %d audited", audited))
	}

	if len(vulnerabilities) > 0 {
		result = append(result, "")
		for _, v := range vulnerabilities {
			result = append(result, fmt.Sprintf("  %s", v))
		}
	}

	if len(warnings) > 0 {
		result = append(result, fmt.Sprintf("  %d deprecation warnings", len(warnings)))
	}

	if len(result) == 1 {
		return "Install complete"
	}
	return strings.Join(result, "\n")
}

func runNpmList(args []string) error {
	timer := tracking.Start()

	npmArgs := append([]string{"ls", "--depth=0"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmListOutput(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm ls %s", strings.Join(args, " ")), "tok npm ls", originalTokens, filteredTokens)

	return err
}

func filterNpmListOutput(output string) string {
	var deps []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "(") || strings.Contains(trimmed, "ERR!") {
			continue
		}
		if strings.HasPrefix(trimmed, "├──") || strings.HasPrefix(trimmed, "└──") {
			pkg := strings.TrimPrefix(trimmed, "├── ")
			pkg = strings.TrimPrefix(pkg, "└── ")
			pkg = strings.TrimSpace(pkg)
			if pkg != "" && len(pkg) < 80 {
				deps = append(deps, pkg)
			}
		}
	}

	if len(deps) == 0 {
		return output
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d packages:\n", len(deps)))
	for i, dep := range deps {
		if i >= 20 {
			break
		}
		result.WriteString(fmt.Sprintf("  %s\n", shared.TruncateLine(dep, 60)))
	}
	if len(deps) > 20 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(deps)-20))
	}

	return result.String()
}

func runNpmOutdated(args []string) error {
	timer := tracking.Start()

	npmArgs := append([]string{"outdated"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	output, _ := c.CombinedOutput()
	raw := string(output)

	filtered := filterNpmOutdatedOutput(raw)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm outdated %s", strings.Join(args, " ")), "tok npm outdated", originalTokens, filteredTokens)

	return nil
}

func filterNpmOutdatedOutput(output string) string {
	if strings.Contains(output, "ERR!") {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= 1 {
		return "All packages up to date"
	}

	var result []string
	for i, line := range lines {
		if i == 0 {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 {
			pkg := fields[0]
			current := fields[1]
			latest := fields[2]
			if len(fields) >= 4 {
				latest = fields[3]
			}
			result = append(result, fmt.Sprintf("  %s: %s -> %s", pkg, current, latest))
		}
	}

	if len(result) == 0 {
		return "All packages up to date"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d outdated packages:\n", len(result)))
	for _, r := range result {
		sb.WriteString(r + "\n")
	}
	return sb.String()
}

func runNpmRun(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		return runNpmPassthrough(append([]string{"run"}, args...))
	}

	npmArgs := append([]string{"run"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmRunOutput(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm run %s", strings.Join(args, " ")), "tok npm run", originalTokens, filteredTokens)

	return err
}

func runNpmPassthrough(args []string) error {
	c := exec.Command("npm", args...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmOutput(output)
	out.Global().Print(filtered)

	return err
}

func filterNpmRunOutput(output string) string {
	var result strings.Builder
	var errors []string

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "ERR!") {
			errors = append(errors, shared.TruncateLine(trimmed, 120))
			continue
		}
		if strings.HasPrefix(trimmed, ">") || strings.Contains(trimmed, "watching") || strings.Contains(trimmed, "ready") || strings.Contains(trimmed, "compiled") {
			result.WriteString(trimmed + "\n")
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return filterNpmOutput(output)
	}
	return result.String()
}

func runNpmAudit(args []string) error {
	timer := tracking.Start()

	npmArgs := append([]string{"audit", "--json"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	output, _ := c.CombinedOutput()
	raw := string(output)

	filtered := filterNpmAuditOutput(raw)
	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm audit %s", strings.Join(args, " ")), "tok npm audit", originalTokens, filteredTokens)

	return nil
}

func filterNpmAuditOutput(output string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return filterNpmOutput(output)
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return filterNpmOutput(output)
	}

	var result strings.Builder

	if metadata, ok := m["metadata"].(map[string]interface{}); ok {
		if vulnerabilities, ok := metadata["vulnerabilities"].(map[string]interface{}); ok {
			info := ""
			if critical, ok := vulnerabilities["critical"].(float64); ok && critical > 0 {
				info += fmt.Sprintf(" critical=%d", int(critical))
			}
			if high, ok := vulnerabilities["high"].(float64); ok && high > 0 {
				info += fmt.Sprintf(" high=%d", int(high))
			}
			if moderate, ok := vulnerabilities["moderate"].(float64); ok && moderate > 0 {
				info += fmt.Sprintf(" moderate=%d", int(moderate))
			}
			if low, ok := vulnerabilities["low"].(float64); ok && low > 0 {
				info += fmt.Sprintf(" low=%d", int(low))
			}
			if info != "" {
				result.WriteString(fmt.Sprintf("Vulnerabilities:%s\n", info))
			} else {
				result.WriteString("No vulnerabilities found\n")
			}
		}
	}

	if vulns, ok := m["vulnerabilities"].(map[string]interface{}); ok {
		result.WriteString(fmt.Sprintf("\n%d affected packages:\n", len(vulns)))
		count := 0
		for name, adv := range vulns {
			if count >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(vulns)-10))
				break
			}
			if am, ok := adv.(map[string]interface{}); ok {
				severity, _ := am["severity"].(string)
				title, _ := am["title"].(string)
				if len(title) > 60 {
					title = title[:57] + "..."
				}
				result.WriteString(fmt.Sprintf("  [%s] %s: %s\n", severity, name, title))
			}
			count++
		}
	}

	if result.Len() == 0 {
		return "No vulnerabilities found"
	}
	return result.String()
}

func runNpmPublish(args []string) error {
	timer := tracking.Start()

	npmArgs := append([]string{"publish"}, args...)
	c := exec.Command("npm", npmArgs...)
	c.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	output := stdout.String() + stderr.String()

	filtered := filterNpmPublishOutput(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("npm publish %s", strings.Join(args, " ")), "tok npm publish", originalTokens, filteredTokens)

	return err
}

func filterNpmPublishOutput(output string) string {
	var result strings.Builder
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "published") || strings.Contains(trimmed, "+") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "ERR!") || strings.Contains(trimmed, "403") {
			result.WriteString(trimmed + "\n")
		}
	}
	if result.Len() == 0 {
		return "Publish complete"
	}
	return result.String()
}
