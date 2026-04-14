package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

var (
	tuiCmd = &cobra.Command{
		Use:   "tui",
		Short: "Start interactive Terminal UI",
		Long: `Start TokMan's interactive TUI for managing workspaces, skills, and commands.

Slash Commands:
  /help, /h     - Show help
  /skills, /s   - List skills  
  /workspace, /w - Show workspace
  /stats, /st   - Show statistics
  /quit, /q    - Exit`,
	}

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the TUI",
		RunE:  runTUI,
	}

	listenCmd = &cobra.Command{
		Use:   "listen",
		Short: "Listen mode - filter stdin/stdout",
		RunE:  runListen,
	}

	prompt string
)

func init() {
	tuiCmd.AddCommand(startCmd)
	tuiCmd.AddCommand(listenCmd)

	startCmd.Flags().StringVarP(&prompt, "prompt", "p", "> ", "Input prompt")

	registry.Add(func() {
		registry.Register(tuiCmd)
	})
}

func runTUI(cmd *cobra.Command, args []string) error {
	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║         TokMan TUI v0.28.2            ║")
	fmt.Println("╠═══════════════════════════════════════╣")
	fmt.Println("║                                       ║")
	fmt.Println("║  Commands:                            ║")
	fmt.Println("║    /skills - Manage skills            ║")
	fmt.Println("║    /workspace - Workspace manager     ║")
	fmt.Println("║    /server - Start API server          ║")
	fmt.Println("║    /stats - View statistics           ║")
	fmt.Println("║                                       ║")
	fmt.Println("║  Type a command to filter it          ║")
	fmt.Println("║  Type /help for all commands          ║")
	fmt.Println("║                                       ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println()

	reader := os.Stdin
	buffer := make([]byte, 1024)

	for {
		fmt.Print(prompt)
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}

		input := strings.TrimSpace(string(buffer[:n]))
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			handleSlashCommand(input)
			continue
		}

		filtered := filterCommand(input)
		fmt.Println(filtered)
	}
	// nolint:nilerr
	return nil
}

func runListen(cmd *cobra.Command, args []string) error {
	fmt.Println("TokMan Listen Mode")
	fmt.Println("Commands will be automatically filtered.")
	fmt.Println("Press Ctrl+C to exit")

	reader := os.Stdin
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}

		input := string(buffer[:n])
		filtered := filterCommand(input)
		fmt.Print(filtered)
	}
	// nolint:nilerr
	return nil
}

func handleSlashCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	if !strings.HasPrefix(cmd, "/") {
		return
	}

	cmd = cmd[1:]

	switch cmd {
	case "help", "h":
		printHelp()
	case "skills", "s":
		listSkills()
	case "workspace", "w", "ws":
		showWorkspace()
	case "server", "srv":
		startServer()
	case "stats", "st":
		showStats()
	case "quit", "q", "exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: /%s\n", cmd)
		fmt.Println("Type /help for available commands")
	}
}

func printHelp() {
	fmt.Println(`
=== TokMan Help ===

/help, /h       - Show this help
/skills, /s     - List and manage skills  
/workspace, /w  - Show workspace info
/server, /srv    - Start API server
/stats, /st      - Show statistics
/quit, /q       - Exit TUI`)
}

func filterCommand(input string) string {
	engine := filter.NewEngine(filter.ModeMinimal)
	output, _ := engine.Process(input)
	return output
}

func listSkills() {
	skillsDir := filepath.Join(os.Getenv("HOME"), ".config", "tokman", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		fmt.Println("No skills installed")
		fmt.Println("Create one with: tokman skills create <name>")
		return
	}

	fmt.Println("=== Skills ===")
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		fmt.Printf("/%s\n", name)
	}
}

func showWorkspace() {
	dataDir := filepath.Join(os.Getenv("HOME"), ".config", "tokman", "workspaces")
	entries, _ := os.ReadDir(dataDir)

	fmt.Println("=== Workspace ===")
	fmt.Printf("Active: default\n")
	fmt.Printf("Workspaces: %d\n", len(entries))
}

func startServer() {
	fmt.Println("Starting API server on port 8081...")
	fmt.Println("Run: tokman server serve --port 8081")
}

func showStats() {
	fmt.Println("=== Statistics ===")
	fmt.Println("Total commands: 0")
	fmt.Println("Tokens saved: 0")
	fmt.Println("Compression: 0%")
}
