// Package app provides the main application logic for tok.
package app

import (
	"fmt"

	"github.com/lakshmanpatel/tok/internal/app/input"
	"github.com/lakshmanpatel/tok/internal/app/output"
	"github.com/lakshmanpatel/tok/internal/app/unified"
)

const version = "0.1.0"

// App represents the tok application.
type App struct {
	inputHandler  *input.Handler
	outputHandler *output.Handler
	unifiedHandler *unified.Handler
}

// New creates a new App instance.
func New() *App {
	inputHandler := input.New()
	outputHandler := output.New("embedded")
	unifiedHandler := unified.New(inputHandler, outputHandler, version)

	return &App{
		inputHandler:   inputHandler,
		outputHandler:  outputHandler,
		unifiedHandler: unifiedHandler,
	}
}

// Run executes the application with the given arguments.
func (a *App) Run(args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("no command provided")
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "input", "in":
		return a.inputHandler.Run(cmdArgs)
	case "output", "out":
		return a.outputHandler.Run(cmdArgs)
	case "both":
		return a.unifiedHandler.Both(cmdArgs)
	case "doctor":
		return a.unifiedHandler.Doctor()
	case "status":
		return a.unifiedHandler.Status()
	case "statusline":
		return a.inputHandler.Statusline()
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "-v", "--version":
		return a.unifiedHandler.Version()
	default:
		// Check if it's a top-level input command
		if isTopLevelInputCommand(cmd) {
			return a.inputHandler.Run(append([]string{cmd}, cmdArgs...))
		}
		// Default to output mode
		return a.outputHandler.Run(args)
	}
}

// isTopLevelInputCommand checks if cmd should be routed to input handler.
func isTopLevelInputCommand(cmd string) bool {
	switch cmd {
	case "compress", "c",
		"terse", "t", "compact", "cmp", "activate", "on",
		"mode",
		"deactivate", "off", "normal", "stop", "stop-terse",
		"statusline",
		"template", "agent-template",
		"commit", "cm", "terse-commit", "tc", "compact-commit", "cc",
		"review", "rv", "terse-review", "tr", "compact-review", "cr",
		"compress-file",
		"compress-memory", "tf", "compact-file", "cf",
		"restore",
		"install-agents", "agents-install",
		"uninstall-agents", "agents-uninstall",
		"hooks-install", "hook-install",
		"hooks-uninstall", "hook-uninstall",
		"hooks-install-pwsh", "hook-install-pwsh",
		"hooks-uninstall-pwsh", "hook-uninstall-pwsh":
		return true
	default:
		return false
	}
}

func printUsage() {
	fmt.Print(`tok - unified token optimization CLI

Usage:
  tok <command> [args]

Commands:
  input, in     Input namespace (compress text)
  output, out   Output namespace (filter commands)
  both          Run both input and output stages
  doctor        Unified diagnostics
  status        Unified status
  version       Show version

Top-level input commands (no namespace needed):
  compress|c, terse|t, compact|cmp, on, mode, off
  template, commit, review, compress-file, compress-memory
  install-agents, uninstall-agents, hooks-install, hooks-uninstall

All other commands route to output automatically.
Examples: 
  tok git status
  tok npm test
  tok compress -mode ultra -input "text"
  tok on full
  tok commit
`)
}
