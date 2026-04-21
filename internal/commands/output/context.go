package output

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/contextread"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var (
	contextMode         string
	contextLevel        string
	contextMaxLines     int
	contextMaxTokens    int
	contextLineNumbers  bool
	contextStartLine    int
	contextEndLine      int
	contextSaveSnapshot bool
	contextRelatedFiles int
)

var contextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Context tools and usage analysis",
	Long:    `Analyze context usage and generate smart file context for AI agents.`,
	RunE:    runContext,
}

var contextReadCmd = &cobra.Command{
	Use:   "read <file>",
	Short: "Produce smart, budgeted file context",
	Long: `Read a file using tok's smart context modes.

Examples:
  tok ctx read main.go --mode auto
  tok ctx read main.go --mode signatures --max-tokens 300
  tok ctx read main.go --start-line 20 --end-line 80
  tok ctx read main.go --mode graph --related-files 4`,
	Args: cobra.ExactArgs(1),
	RunE: runContextRead,
}

var contextDeltaCmd = &cobra.Command{
	Use:   "delta <file>",
	Short: "Show what changed since the last saved file snapshot",
	Long: `Compare a file to tok's last saved snapshot and emit a compact delta.

Examples:
  tok ctx delta main.go
  tok ctx delta main.go --max-tokens 200`,
	Args: cobra.ExactArgs(1),
	RunE: runContextDelta,
}

func init() {
	addContextReadFlags(contextReadCmd)
	addContextReadFlags(contextDeltaCmd)

	contextCmd.AddCommand(contextReadCmd, contextDeltaCmd)
	registry.Add(func() { registry.Register(contextCmd) })
}

func runContext(cmd *cobra.Command, args []string) error {
	tracker, err := shared.OpenTracker()
	if err != nil {
		return fmt.Errorf("tracking not available: %w", err)
	}
	defer tracker.Close()

	projectPath := shared.GetProjectPath()
	savings, err := tracker.GetSavings(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get context data: %w", err)
	}

	out.Global().Println("Context Window Analysis")
	out.Global().Println("======================")
	out.Global().Println()

	if savings.TotalCommands == 0 {
		out.Global().Println("No data yet. Run some commands through tok first.")
		return nil
	}

	out.Global().Printf("Commands analyzed: %d\n", savings.TotalCommands)
	out.Global().Printf("Original context:  %d tokens\n", savings.TotalOriginal)
	out.Global().Printf("Filtered context:  %d tokens\n", savings.TotalFiltered)
	out.Global().Printf("Tokens saved:      %d tokens\n", savings.TotalSaved)
	out.Global().Printf("Reduction:         %.1f%%\n\n", savings.ReductionPct)

	readSavings, err := tracker.GetSavingsForContextReads(projectPath, "", "")
	if err != nil {
		return fmt.Errorf("failed to get smart read data: %w", err)
	}
	if readSavings.TotalCommands > 0 {
		out.Global().Println("Smart context reads")
		out.Global().Println("-------------------")
		out.Global().Printf("Reads analyzed:    %d\n", readSavings.TotalCommands)
		out.Global().Printf("Original context:  %d tokens\n", readSavings.TotalOriginal)
		out.Global().Printf("Delivered context: %d tokens\n", readSavings.TotalFiltered)
		out.Global().Printf("Tokens saved:      %d tokens\n", readSavings.TotalSaved)
		out.Global().Printf("Reduction:         %.1f%%\n\n", readSavings.ReductionPct)
	}

	contextSizes := []struct {
		name  string
		limit int
	}{
		{"GPT-4o-mini (128K)", 128000},
		{"GPT-4o (128K)", 128000},
		{"Claude 3.5 (200K)", 200000},
		{"Claude 3 Opus (200K)", 200000},
		{"Gemini 1.5 (1M)", 1000000},
	}

	out.Global().Println("Context window capacity with tok:")
	out.Global().Printf("%-25s %12s %12s %10s\n", "Model", "Without", "With", "Extra")
	out.Global().Printf("%-25s %12s %12s %10s\n", "─────────────────────────", "────────────", "────────────", "──────────")

	for _, cs := range contextSizes {
		without := cs.limit / savings.TotalOriginal
		if without == 0 {
			without = 1
		}
		with := cs.limit / savings.TotalFiltered
		if with == 0 {
			with = 1
		}
		extra := with - without
		out.Global().Printf("%-25s %10dx %10dx +%dx\n", cs.name, without, with, extra)
	}

	return nil
}

func addContextReadFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&contextMode, "mode", "auto", "Context mode: auto, full, map, signatures, aggressive, entropy, lines, delta, graph")
	cmd.Flags().StringVar(&contextLevel, "level", "minimal", "Legacy filter level: none, minimal, aggressive")
	cmd.Flags().IntVar(&contextMaxLines, "max-lines", 0, "Maximum lines to emit (0 = no limit)")
	cmd.Flags().IntVar(&contextMaxTokens, "max-tokens", 0, "Approximate token budget for emitted context (0 = no limit)")
	cmd.Flags().BoolVar(&contextLineNumbers, "line-numbers", false, "Include line numbers in output")
	cmd.Flags().IntVar(&contextStartLine, "start-line", 0, "Start line for line-oriented reads")
	cmd.Flags().IntVar(&contextEndLine, "end-line", 0, "End line for line-oriented reads")
	cmd.Flags().BoolVar(&contextSaveSnapshot, "save-snapshot", true, "Persist a snapshot for future delta reads")
	cmd.Flags().IntVar(&contextRelatedFiles, "related-files", 3, "Number of related files to include in graph mode")
}

func runContextRead(cmd *cobra.Command, args []string) error {
	return emitContextFile(args[0], contextread.Options{
		Level:             contextLevel,
		Mode:              contextMode,
		MaxLines:          contextMaxLines,
		MaxTokens:         contextMaxTokens,
		LineNumbers:       contextLineNumbers,
		StartLine:         contextStartLine,
		EndLine:           contextEndLine,
		SaveSnapshot:      contextSaveSnapshot,
		RelatedFilesCount: contextRelatedFiles,
	})
}

func runContextDelta(cmd *cobra.Command, args []string) error {
	mode := contextMode
	if mode == "" || mode == "auto" {
		mode = "delta"
	}
	return emitContextFile(args[0], contextread.Options{
		Level:             contextLevel,
		Mode:              mode,
		MaxLines:          contextMaxLines,
		MaxTokens:         contextMaxTokens,
		LineNumbers:       contextLineNumbers,
		StartLine:         contextStartLine,
		EndLine:           contextEndLine,
		SaveSnapshot:      contextSaveSnapshot,
		RelatedFilesCount: contextRelatedFiles,
	})
}

func emitContextFile(path string, opts contextread.Options) error {
	start := time.Now()
	content, rawContent, originalTokens, filteredTokens, err := buildContextFile(path, opts)
	if err != nil {
		return err
	}

	out.Global().Print(content)
	if content != "" && !strings.HasSuffix(content, "\n") {
		out.Global().Println()
	}

	commandName := "tok ctx read"
	if strings.EqualFold(opts.Mode, "delta") {
		commandName = "tok ctx delta"
	}
	recordContextRead(commandName, path, rawContent, opts, originalTokens, filteredTokens, time.Since(start).Milliseconds())
	return nil
}

func buildContextFile(path string, opts contextread.Options) (string, string, int, int, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}

	content, originalTokens, filteredTokens, err := contextread.Build(cleanPath, string(data), "auto", opts)
	if err != nil {
		return "", "", 0, 0, err
	}
	return content, string(data), originalTokens, filteredTokens, nil
}

func recordContextRead(commandName, path, rawContent string, opts contextread.Options, originalTokens, filteredTokens int, execTimeMs int64) {
	tracker, err := shared.OpenTracker()
	if err != nil {
		return
	}
	defer tracker.Close()

	projectPath := shared.GetProjectPath()

	savedTokens := originalTokens - filteredTokens
	if savedTokens < 0 {
		savedTokens = 0
	}

	if err := tracker.Record(&tracking.CommandRecord{
		Command:             fmt.Sprintf("%s %s", commandName, filepath.Clean(path)),
		OriginalTokens:      originalTokens,
		FilteredTokens:      filteredTokens,
		SavedTokens:         savedTokens,
		ProjectPath:         projectPath,
		ExecTimeMs:          execTimeMs,
		ParseSuccess:        true,
		ContextKind:         "read",
		ContextMode:         opts.Mode,
		ContextResolvedMode: opts.Mode,
		ContextTarget:       path,
		ContextBundle:       false,
	}); err != nil {
		log.Printf("failed to record context read: %v", err)
	}
}
