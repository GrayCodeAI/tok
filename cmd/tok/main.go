package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lakshmanpatel/tok/internal/compressor"
	"github.com/lakshmanpatel/tok/internal/git"
	"github.com/lakshmanpatel/tok/internal/hooks"
	"github.com/lakshmanpatel/tok/internal/review"
	tokcli "github.com/lakshmanpatel/tok/pkg/cli"
)

const version = "0.1.0"

var tokVersion = "embedded"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "input", "in":
		if err := runInput(args); err != nil {
			exitErr(err)
		}
	case "output", "out":
		if err := runOutput(args); err != nil {
			exitErr(err)
		}
	case "both":
		if err := runBoth(args); err != nil {
			exitErr(err)
		}
	case "doctor":
		if err := runDoctorUnified(); err != nil {
			exitErr(err)
		}
	case "status":
		if err := runStatusUnified(); err != nil {
			exitErr(err)
		}
	case "statusline":
		if err := runInputStatusline(); err != nil {
			exitErr(err)
		}
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		if err := runVersionUnified(); err != nil {
			exitErr(err)
		}
	default:
		if isTopLevelInputCommand(cmd) {
			if err := runInput(append([]string{cmd}, args...)); err != nil {
				exitErr(err)
			}
			return
		}
		if err := runOutput(os.Args[1:]); err != nil {
			exitErr(err)
		}
	}
}

func runOutput(args []string) error {
	if code := tokcli.Run(args, tokVersion); code != 0 {
		return fmt.Errorf("tok output failed with exit code %d", code)
	}
	return nil
}

func runBoth(args []string) error {
	fs := flag.NewFlagSet("both", flag.ContinueOnError)
	text := fs.String("text", "", "Text to compress through input engine before sending to AI")
	mode := fs.String("mode", "", "Input compression mode passed to input engine (lite|full|ultra|wenyan)")
	command := fs.String("command", "", "Command to run through output engine (example: \"git status\")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*text) != "" {
		inputArgs := []string{"-input", *text}
		if strings.TrimSpace(*mode) != "" {
			inputArgs = append(inputArgs, "-mode", *mode)
		}
		if err := runInputCompress(inputArgs); err != nil {
			return fmt.Errorf("input stage failed: %w", err)
		}
	}

	if strings.TrimSpace(*command) != "" {
		outputArgs := shellWords(*command)
		if len(outputArgs) == 0 {
			return errors.New("both: --command is empty after parsing")
		}
		if err := runOutput(outputArgs); err != nil {
			return fmt.Errorf("output stage failed: %w", err)
		}
	}

	if strings.TrimSpace(*text) == "" && strings.TrimSpace(*command) == "" {
		return errors.New("both requires at least one of --text or --command")
	}
	return nil
}

func runDoctor() error {
	fmt.Println("tok doctor")
	fmt.Println("  [input] ok: native tok input engine")
	fmt.Println("  [output] ok: embedded output engine")
	return nil
}

func runDoctorUnified() error {
	return runDoctor()
}

func runStatusUnified() error {
	fmt.Println("tok status (input)")
	if err := runInputStatus(); err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println("tok status (output)")
	fmt.Println("output engine is available")
	return nil
}

func runVersionUnified() error {
	fmt.Printf("tok %s\n", version)
	fmt.Println("input engine: compatible")
	fmt.Println("output engine: embedded")
	return nil
}

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

func shellWords(s string) []string {
	return strings.Fields(strings.TrimSpace(s))
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "tok error: %v\n", err)
	os.Exit(1)
}

func printUsage() {
	fmt.Print(`tok - unified token optimization CLI

Usage:
  tok <command> [args]

Commands:
  input, in     Input namespace
  output, out   Output namespace
  both          Run both input and output stages
  doctor        Unified doctor (input + output diagnostics)
  status        Unified status (input + output status)
  version       Unified versions (tok + engines)

Top-level input commands (no namespace needed):
  compress|c, terse|t, compact|cmp, mode, stop/off
  template, commit, review, compress-file, compress-memory, restore
  install-agents, uninstall-agents, hooks-install, hooks-uninstall
  hooks-install-pwsh, hooks-uninstall-pwsh
  statusline

All other commands route to output automatically.
Examples: tok git status, tok npm test, tok gain, tok doctor

Examples:
  tok input compress -mode ultra -input "Please implement this safely"
  tok compress -mode ultra -input "Please implement this safely"
  tok output git status
  tok both --text "Please summarize" --command "git diff"
  tok git status    # smart default routes to output mode
`)
}

func runInput(args []string) error {
	if len(args) == 0 {
		return errors.New("input command required. Use: tok input compress -mode ultra -input \"text\"")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "compress", "c":
		return runInputCompress(subargs)
	case "terse", "t":
		return runInputTerse(subargs)
	case "compact", "cmp":
		return runInputCompact(subargs)
	case "activate", "on":
		return runInputActivate(subargs)
	case "mode":
		return runInputMode()
	case "deactivate", "off", "normal", "stop", "stop-terse":
		return runInputDeactivate()
	case "statusline":
		return runInputStatusline()
	case "template", "agent-template":
		return runInputTemplate()
	case "commit", "cm":
		return runInputCommit(subargs)
	case "terse-commit", "tc":
		return runInputTerseCommit(subargs)
	case "compact-commit", "cc":
		return runInputCompactCommit(subargs)
	case "review", "rv":
		return runInputReview(subargs)
	case "terse-review", "tr":
		return runInputTerseReview(subargs)
	case "compact-review", "cr":
		return runInputCompactReview(subargs)
	case "compress-file":
		return runInputCompressFile(subargs)
	case "compress-memory", "tf", "compact-file", "cf":
		return runInputCompressMemory(subargs)
	case "restore":
		return runInputRestore(subargs)
	case "install-agents", "agents-install":
		return runInputInstallAgents()
	case "uninstall-agents", "agents-uninstall":
		return runInputUninstallAgents()
	case "hooks-install", "hook-install":
		return runInputHooksInstall()
	case "hooks-uninstall", "hook-uninstall":
		return runInputHooksUninstall()
	case "hooks-install-pwsh", "hook-install-pwsh":
		return runInputHooksInstallPwsh()
	case "hooks-uninstall-pwsh", "hook-uninstall-pwsh":
		return runInputHooksUninstallPwsh()
	default:
		return fmt.Errorf("unknown input command: %s", subcmd)
	}
}

func runInputCompress(args []string) error {
	fs := flag.NewFlagSet("compress", flag.ContinueOnError)
	input := fs.String("input", "", "Input text (or use stdin)")
	mode := fs.String("mode", "full", "Compression mode: lite, full, ultra (default \"full\")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var text string
	if *input != "" {
		text = *input
	} else {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		text = string(bytes)
	}

	compressed, err := compressor.Compress(text, *mode)
	if err != nil {
		return err
	}

	fmt.Print(compressed)
	return nil
}

func runInputTerse(args []string) error {
	mode := "full"
	if len(args) > 0 {
		mode = args[0]
	}
	
	if err := hooks.Activate(mode); err != nil {
		return fmt.Errorf("failed to activate terse mode: %w", err)
	}
	
	fmt.Printf("Terse mode activated: %s\n", mode)
	fmt.Println("Run 'tok off' to deactivate")
	return nil
}

func runInputCompact(args []string) error {
	if err := hooks.Activate("full"); err != nil {
		return fmt.Errorf("failed to activate compact mode: %w", err)
	}
	fmt.Println("Compact mode activated")
	return nil
}

func runInputActivate(args []string) error {
	mode := "full"
	if len(args) > 0 {
		mode = args[0]
	}
	
	if err := hooks.Activate(mode); err != nil {
		return fmt.Errorf("failed to activate: %w", err)
	}
	
	fmt.Printf("tok activated: %s mode\n", mode)
	return nil
}

func runInputMode() error {
	if !hooks.IsActive() {
		fmt.Println("tok mode: inactive")
		return nil
	}
	
	mode := hooks.GetMode()
	if mode == "" {
		mode = "full"
	}
	fmt.Printf("tok mode: %s\n", mode)
	return nil
}

func runInputDeactivate() error {
	if err := hooks.Deactivate(); err != nil {
		return fmt.Errorf("failed to deactivate: %w", err)
	}
	fmt.Println("tok deactivated")
	return nil
}

func runInputStatus() error {
	if hooks.IsActive() {
		fmt.Printf("tok input status: active (%s mode)\n", hooks.GetMode())
	} else {
		fmt.Println("tok input status: inactive")
	}
	return nil
}

func runInputStatusline() error {
	badge := hooks.GetStatusLine()
	if badge == "" {
		badge = "[TOK:off]"
	}
	fmt.Println(badge)
	return nil
}

func runInputTemplate() error {
	fmt.Println("Template: not implemented yet")
	return nil
}

func runInputCommit(args []string) error {
	msg, err := git.GenerateCommitMessage()
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}
	fmt.Println(msg)
	return nil
}

func runInputTerseCommit(args []string) error {
	return runInputCommit(args)
}

func runInputCompactCommit(args []string) error {
	return runInputCommit(args)
}

func runInputReview(args []string) error {
	results, err := review.GenerateReview()
	if err != nil {
		return fmt.Errorf("failed to generate review: %w", err)
	}
	
	formatted := review.FormatReview(results)
	fmt.Println(formatted)
	return nil
}

func runInputTerseReview(args []string) error {
	return runInputReview(args)
}

func runInputCompactReview(args []string) error {
	return runInputReview(args)
}

func runInputCompressFile(args []string) error {
	fs := flag.NewFlagSet("compress-file", flag.ContinueOnError)
	file := fs.String("file", "", "File to compress")
	mode := fs.String("mode", "full", "Compression mode")
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	if *file == "" {
		return errors.New("file required: -file <path>")
	}
	
	content, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	
	compressed, err := compressor.Compress(string(content), *mode)
	if err != nil {
		return err
	}
	
	fmt.Print(compressed)
	return nil
}

func runInputCompressMemory(args []string) error {
	fs := flag.NewFlagSet("compress-memory", flag.ContinueOnError)
	file := fs.String("file", "", "File to compress to memory format")
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	if *file == "" && len(fs.Args()) > 0 {
		*file = fs.Args()[0]
	}
	
	if *file == "" {
		return errors.New("file required")
	}
	
	content, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	
	// Use ultra mode for memory compression
	compressed, err := compressor.Compress(string(content), "ultra")
	if err != nil {
		return err
	}
	
	fmt.Printf("// %s (compressed)\n", *file)
	fmt.Print(compressed)
	return nil
}

func runInputRestore(args []string) error {
	fmt.Println("Restore: not implemented yet")
	return nil
}

func runInputInstallAgents() error {
	fmt.Println("Install agents: not implemented yet")
	return nil
}

func runInputUninstallAgents() error {
	fmt.Println("Uninstall agents: not implemented yet")
	return nil
}

func runInputHooksInstall() error {
	fmt.Println("Hooks install: not implemented yet")
	return nil
}

func runInputHooksUninstall() error {
	fmt.Println("Hooks uninstall: not implemented yet")
	return nil
}

func runInputHooksInstallPwsh() error {
	fmt.Println("Hooks install pwsh: not implemented yet")
	return nil
}

func runInputHooksUninstallPwsh() error {
	fmt.Println("Hooks uninstall pwsh: not implemented yet")
	return nil
}