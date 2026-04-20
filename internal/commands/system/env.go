package system

import (
	"fmt"
	"os"
	"sort"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var envShowAll bool

var envCmd = &cobra.Command{
	Use:   "env [filter]",
	Short: "Environment variables with sensitive value masking",
	Long: `Show environment variables with categorization and sensitive value masking.

Categorizes variables into PATH, Language/Runtime, Cloud/Services, Tools, and Other.
Masks sensitive values (keys, secrets, passwords, tokens) by default.

Examples:
  tok env
  tok env AWS
  tok env --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEnv,
}

func init() {
	registry.Add(func() { registry.Register(envCmd) })
	envCmd.Flags().BoolVar(&envShowAll, "all", false, "Show unmasked sensitive values")
}

var sensitivePatterns = []string{
	"key", "secret", "password", "token", "credential",
	"auth", "private", "api_key", "apikey", "access_key", "jwt",
}

func runEnv(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	var filterStr string
	if len(args) > 0 {
		filterStr = args[0]
	}

	vars := os.Environ()
	sort.Strings(vars)

	var pathVars, langVars, cloudVars, toolVars, otherVars []envEntry
	totalVars := len(vars)

	for _, v := range vars {
		idx := strings.IndexByte(v, '=')
		if idx < 0 {
			continue
		}
		key := v[:idx]
		value := v[idx+1:]

		// Apply filter
		if filterStr != "" && !strings.Contains(strings.ToLower(key), strings.ToLower(filterStr)) {
			continue
		}

		// Check sensitivity
		isSensitive := false
		keyLower := strings.ToLower(key)
		for _, p := range sensitivePatterns {
			if strings.Contains(keyLower, p) {
				isSensitive = true
				break
			}
		}

		displayValue := value
		if isSensitive && !envShowAll {
			displayValue = maskValue(value)
		} else if len(value) > 100 {
			runes := []rune(value)
			displayValue = fmt.Sprintf("%s... (%d chars)", string(runes[:50]), len(runes))
		}

		entry := envEntry{key: key, value: displayValue}

		if strings.Contains(key, "PATH") {
			pathVars = append(pathVars, entry)
		} else if isLangVar(key) {
			langVars = append(langVars, entry)
		} else if isCloudVar(key) {
			cloudVars = append(cloudVars, entry)
		} else if isToolVar(key) {
			toolVars = append(toolVars, entry)
		} else if filterStr != "" || isInterestingVar(key) {
			otherVars = append(otherVars, entry)
		}
	}

	var result strings.Builder

	if len(pathVars) > 0 {
		result.WriteString("PATH Variables:\n")
		for _, e := range pathVars {
			if e.key == "PATH" {
				paths := strings.Split(e.value, ":")
				result.WriteString(fmt.Sprintf("  PATH (%d entries):\n", len(paths)))
				for i, p := range paths {
					if i >= 5 {
						result.WriteString(fmt.Sprintf("    ... +%d more\n", len(paths)-5))
						break
					}
					result.WriteString(fmt.Sprintf("    %s\n", p))
				}
			} else {
				result.WriteString(fmt.Sprintf("  %s=%s\n", e.key, e.value))
			}
		}
	}

	if len(langVars) > 0 {
		result.WriteString("\nLanguage/Runtime:\n")
		for _, e := range langVars {
			result.WriteString(fmt.Sprintf("  %s=%s\n", e.key, e.value))
		}
	}

	if len(cloudVars) > 0 {
		result.WriteString("\nCloud/Services:\n")
		for _, e := range cloudVars {
			result.WriteString(fmt.Sprintf("  %s=%s\n", e.key, e.value))
		}
	}

	if len(toolVars) > 0 {
		result.WriteString("\nTools:\n")
		for _, e := range toolVars {
			result.WriteString(fmt.Sprintf("  %s=%s\n", e.key, e.value))
		}
	}

	if len(otherVars) > 0 {
		result.WriteString("\nOther:\n")
		for i, e := range otherVars {
			if i >= 20 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(otherVars)-20))
				break
			}
			result.WriteString(fmt.Sprintf("  %s=%s\n", e.key, e.value))
		}
	}

	shown := len(pathVars) + len(langVars) + len(cloudVars) + len(toolVars)
	if len(otherVars) > 20 {
		shown += 20
	} else {
		shown += len(otherVars)
	}

	if filterStr == "" {
		result.WriteString(fmt.Sprintf("\nTotal: %d vars (showing %d relevant)\n", totalVars, shown))
	}

	output := result.String()
	out.Global().Print(output)

	raw := strings.Join(vars, "\n")
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(output)
	timer.Track("env", "tok env", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return nil
}

type envEntry struct {
	key   string
	value string
}

func maskValue(value string) string {
	runes := []rune(value)
	if len(runes) <= 4 {
		return "****"
	}
	return string(runes[:2]) + "****" + string(runes[len(runes)-2:])
}

func isLangVar(key string) bool {
	patterns := []string{
		"RUST", "CARGO", "PYTHON", "PIP", "NODE", "NPM", "YARN",
		"DENO", "BUN", "JAVA", "MAVEN", "GRADLE", "GO", "GOPATH",
		"GOROOT", "RUBY", "GEM", "PERL", "PHP", "DOTNET", "NUGET",
	}
	upper := strings.ToUpper(key)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

func isCloudVar(key string) bool {
	patterns := []string{
		"AWS", "AZURE", "GCP", "GOOGLE_CLOUD", "DOCKER",
		"KUBERNETES", "K8S", "HELM", "TERRAFORM", "VAULT",
		"CONSUL", "NOMAD",
	}
	upper := strings.ToUpper(key)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

func isToolVar(key string) bool {
	patterns := []string{
		"EDITOR", "VISUAL", "SHELL", "TERM", "GIT", "SSH",
		"GPG", "BREW", "HOMEBREW", "XDG", "CLAUDE", "ANTHROPIC",
	}
	upper := strings.ToUpper(key)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

func isInterestingVar(key string) bool {
	patterns := []string{"HOME", "USER", "LANG", "LC_", "TZ", "PWD", "OLDPWD"}
	upper := strings.ToUpper(key)
	for _, p := range patterns {
		if strings.HasPrefix(upper, p) {
			return true
		}
	}
	return false
}
