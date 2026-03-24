package initpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const opencodePluginContent = `// TokMan plugin for OpenCode
// Rewrites shell commands through tokman for 60-90% token savings.
import { definePlugin } from "opencode";

export default definePlugin({
  name: "tokman",
  version: "1.0.0",
  hooks: {
    "tool.execute.before": async (ctx) => {
      if (ctx.tool !== "bash" && ctx.tool !== "shell") return ctx;
      const cmd = ctx.args?.command || ctx.args?.cmd || "";
      if (!cmd || cmd.startsWith("tokman ")) return ctx;

      // Rewrite known commands through tokman
      const prefixes = [
        "git ", "gh ", "gt ", "cargo ", "npm ", "pnpm ", "npx ",
        "pip ", "ruff ", "pytest ", "go ", "docker ", "kubectl ",
        "ls ", "tree ", "grep ", "find ", "diff ", "cat ",
        "rspec ", "rubocop ", "rake ", "bundle ", "rails ",
        "jest ", "vitest ", "playwright ", "tsc ", "prettier ",
        "eslint ", "mypy ", "golangci-lint ", "next ", "prisma ",
        "curl ", "wget ", "wc ", "env", "aws ", "psql ",
      ];

      for (const prefix of prefixes) {
        if (cmd.startsWith(prefix) || cmd === prefix.trim()) {
          ctx.args = { ...ctx.args, command: "tokman " + cmd };
          break;
        }
      }
      return ctx;
    },
  },
});
`

// runOpenCodeInit sets up OpenCode integration
func runOpenCodeInit(global bool) {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if !global {
		fmt.Fprintf(os.Stderr, "OpenCode support is global-only. Use: tokman init -g --opencode\n")
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	pluginsDir := filepath.Join(homeDir, ".config", "opencode", "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating OpenCode plugins directory: %v\n", err)
		return
	}

	pluginPath := filepath.Join(pluginsDir, "tokman.ts")
	if err := writeIfChanged(pluginPath, opencodePluginContent, "tokman.ts"); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing OpenCode plugin: %v\n", err)
		return
	}

	fmt.Printf("\n%s\n\n", green("TokMan configured for OpenCode."))
	fmt.Printf("  Plugin: %s\n", cyan(pluginPath))
	fmt.Println("  Restart OpenCode to activate. Test with: git status")
	fmt.Println()
}

// uninstallOpenCode removes OpenCode artifacts
func uninstallOpenCode() []string {
	var removed []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return removed
	}

	pluginPath := filepath.Join(homeDir, ".config", "opencode", "plugins", "tokman.ts")
	if data, err := os.ReadFile(pluginPath); err == nil {
		if strings.Contains(string(data), "tokman") {
			if err := os.Remove(pluginPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", pluginPath, err)
			}
			removed = append(removed, fmt.Sprintf("OpenCode plugin: %s", pluginPath))
		}
	}

	return removed
}
