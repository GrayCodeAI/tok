package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var (
	envFilter  string
	envShowAll bool
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment variables (filtered, sensitive masked)",
	Long: `Show environment variables with sensitive values masked.

Categorizes variables and filters noise. Sensitive values (keys, tokens, passwords)
are masked unless --show-all is specified.

Examples:
  tokman env
  tokman env --filter AWS
  tokman env --show-all`,
	RunE: runEnv,
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.Flags().StringVarP(&envFilter, "filter", "f", "", "Filter by name (e.g. PATH, AWS)")
	envCmd.Flags().BoolVar(&envShowAll, "show-all", false, "Show all (include sensitive)")
}

func runEnv(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	// Get environment variables
	env := os.Environ()
	sort.Strings(env)

	// Categorize variables
	var pathVars []string
	var langVars []string
	var cloudVars []string
	var toolVars []string
	var otherVars []string

	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name, value := parts[0], parts[1]

		// Apply filter
		if envFilter != "" && !strings.Contains(strings.ToUpper(name), strings.ToUpper(envFilter)) {
			continue
		}

		// Mask sensitive values
		if !envShowAll && isSensitive(name) {
			value = maskSensitive(value)
		} else if len(value) > 100 && !envShowAll {
			// Truncate long values
			value = value[:50] + fmt.Sprintf("... (%d chars)", len(value))
		}

		entry := fmt.Sprintf("  %s=%s", name, value)

		// Categorize
		switch {
		case strings.Contains(name, "PATH"):
			pathVars = append(pathVars, formatPathEntry(name, value))
		case isLangVar(name):
			langVars = append(langVars, entry)
		case isCloudVar(name):
			cloudVars = append(cloudVars, entry)
		case isToolVar(name):
			toolVars = append(toolVars, entry)
		case envFilter != "" || isInterestingVar(name):
			otherVars = append(otherVars, entry)
		}
	}

	// Output categorized
	var result strings.Builder

	if len(pathVars) > 0 {
		result.WriteString("📂 PATH Variables:\n")
		for _, v := range pathVars {
			result.WriteString(v + "\n")
		}
	}

	if len(langVars) > 0 {
		result.WriteString("\n🔧 Language/Runtime:\n")
		for _, v := range langVars {
			result.WriteString(v + "\n")
		}
	}

	if len(cloudVars) > 0 {
		result.WriteString("\n☁️  Cloud/Services:\n")
		for _, v := range cloudVars {
			result.WriteString(v + "\n")
		}
	}

	if len(toolVars) > 0 {
		result.WriteString("\n🛠️  Tools:\n")
		for _, v := range toolVars {
			result.WriteString(v + "\n")
		}
	}

	if len(otherVars) > 0 {
		result.WriteString("\n📋 Other:\n")
		for i, v := range otherVars {
			if i >= 20 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(otherVars)-20))
				break
			}
			result.WriteString(v + "\n")
		}
	}

	// Summary
	total := len(env)
	shown := len(pathVars) + len(langVars) + len(cloudVars) + len(toolVars)
	if len(otherVars) > 20 {
		shown += 20
	} else {
		shown += len(otherVars)
	}

	if envFilter == "" {
		result.WriteString(fmt.Sprintf("\n📊 Total: %d vars (showing %d relevant)\n", total, shown))
	}

	output := result.String()
	fmt.Print(output)

	originalTokens := filter.EstimateTokens(strings.Join(env, "\n"))
	filteredTokens := filter.EstimateTokens(output)
	timer.Track("env", "tokman env", originalTokens, filteredTokens)

	return nil
}

// formatPathEntry formats PATH variable with split entries
func formatPathEntry(name, value string) string {
	if name == "PATH" {
		paths := strings.Split(value, ":")
		var result strings.Builder
		result.WriteString(fmt.Sprintf("  PATH (%d entries):\n", len(paths)))
		for i, p := range paths {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("    ... +%d more\n", len(paths)-5))
				break
			}
			result.WriteString(fmt.Sprintf("    %s\n", p))
		}
		return result.String()
	}
	return fmt.Sprintf("  %s=%s", name, value)
}

func isSensitive(name string) bool {
	name = strings.ToUpper(name)
	sensitive := []string{"KEY", "SECRET", "TOKEN", "PASSWORD", "PASS", "API", "AUTH", "CRED", "PRIVATE", "ACCESS_KEY", "APIKEY", "JWT"}
	for _, s := range sensitive {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}

func maskSensitive(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

// isLangVar checks if variable is language/runtime related
func isLangVar(name string) bool {
	patterns := []string{"RUST", "CARGO", "PYTHON", "PIP", "NODE", "NPM", "YARN", "DENO", "BUN", "JAVA", "MAVEN", "GRADLE", "GO", "GOPATH", "GOROOT", "RUBY", "GEM", "PERL", "PHP", "DOTNET", "NUGET"}
	upper := strings.ToUpper(name)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

// isCloudVar checks if variable is cloud/services related
func isCloudVar(name string) bool {
	patterns := []string{"AWS", "AZURE", "GCP", "GOOGLE_CLOUD", "DOCKER", "KUBERNETES", "K8S", "HELM", "TERRAFORM", "VAULT", "CONSUL", "NOMAD"}
	upper := strings.ToUpper(name)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

// isToolVar checks if variable is tool related
func isToolVar(name string) bool {
	patterns := []string{"EDITOR", "VISUAL", "SHELL", "TERM", "GIT", "SSH", "GPG", "BREW", "HOMEBREW", "XDG", "CLAUDE", "ANTHROPIC"}
	upper := strings.ToUpper(name)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

// isInterestingVar checks if variable is generally interesting
func isInterestingVar(name string) bool {
	patterns := []string{"HOME", "USER", "LANG", "LC_", "TZ", "PWD", "OLDPWD"}
	upper := strings.ToUpper(name)
	for _, p := range patterns {
		if strings.HasPrefix(upper, p) {
			return true
		}
	}
	return false
}
