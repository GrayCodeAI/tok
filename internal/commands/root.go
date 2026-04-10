package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/integrity"
	"github.com/GrayCodeAI/tokman/internal/utils"

	// CLI commands
	_ "github.com/GrayCodeAI/tokman/internal/commands/build"
	_ "github.com/GrayCodeAI/tokman/internal/commands/compression"
	_ "github.com/GrayCodeAI/tokman/internal/commands/configcmd"
	_ "github.com/GrayCodeAI/tokman/internal/commands/container"
	_ "github.com/GrayCodeAI/tokman/internal/commands/core"
	_ "github.com/GrayCodeAI/tokman/internal/commands/filtercmd"
	_ "github.com/GrayCodeAI/tokman/internal/commands/hooks"
	_ "github.com/GrayCodeAI/tokman/internal/commands/lang"
	_ "github.com/GrayCodeAI/tokman/internal/commands/linter"
	_ "github.com/GrayCodeAI/tokman/internal/commands/output"
	_ "github.com/GrayCodeAI/tokman/internal/commands/pattern"
	_ "github.com/GrayCodeAI/tokman/internal/commands/pkgmgr"
	_ "github.com/GrayCodeAI/tokman/internal/commands/session"
	_ "github.com/GrayCodeAI/tokman/internal/commands/system"
	_ "github.com/GrayCodeAI/tokman/internal/commands/test"
	_ "github.com/GrayCodeAI/tokman/internal/commands/vcs"
)

var (
	cfgFile      string
	verbose      int // Count-based: -v, -vv, -vvv
	dryRun       bool
	ultraCompact bool
	skipEnv      bool
	queryIntent  string   // Query intent for query-aware compression
	llmEnabled   bool     // Enable LLM-based compression
	tokenBudget  int      // Token budget for compression (0 = unlimited)
	fallbackArgs []string // Args for fallback handler
	layerPreset  string   // Pipeline preset: fast/balanced/full (T90)
	layerProfile string   // Compression tier: surface/trim/extract/core/code/log/thread
	outputFile   string   // R35: Write output to file
	quietMode    bool     // R36: Suppress all non-essential output
	jsonOutput   bool     // R37: Machine-readable JSON output

	// Remote mode flags (Phase 4)
	remoteMode      bool
	compressionAddr string
	analyticsAddr   string
	remoteTimeout   int // seconds

	// Compaction flags (Layer 11)
	compactionEnabled    bool
	compactionThreshold  int
	compactionPreserve   int
	compactionMaxTokens  int
	compactionSnapshot   bool
	compactionAutoDetect bool

	// Reversible compression (R1: claw-compactor style)
	reversibleEnabled bool

	// Custom layer configuration (Task 5: Layer enable/disable)
	enableLayers     []string // Layers to explicitly enable
	disableLayers    []string // Layers to explicitly disable
	streamMode       bool     // Enable streaming for large inputs
	policyRouter     bool     // Enable policy-based routing
	extractive       bool     // Enable extractive prefilter
	extractiveMax    int      // Max lines before extractive prefilter triggers
	extractiveHead   int      // Head lines to preserve
	extractiveTail   int      // Tail lines to preserve
	extractiveSignal int      // Signal lines to preserve
	qualityGuardrail bool     // Enable quality guardrail auto-fallback
	diffAdapt        bool     // Enable DiffAdapt layer
	epic             bool     // Enable EPiC layer
	ssdp             bool     // Enable SSDP layer
	agentOCR         bool     // Enable AgentOCR layer
	s2mad            bool     // Enable S2-MAD layer
	acon             bool     // Enable ACON layer
	researchPack     bool     // Enable research layer pack (31-36)
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tokman",
		Version: shared.Version, // Set from shared.Version (injected via ldflags)
		Short:   "Token-aware CLI proxy",
		Long: `TokMan intercepts CLI commands and filters verbose output
to reduce token usage in LLM interactions.

It acts as a transparent proxy that executes commands, captures their
output, applies intelligent filtering, and tracks token savings.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			shared.SetRootCmd(cmd)
			// Version is already set in shared.Version via ldflags
			shared.SetFlags(shared.FlagConfig{
				Verbose:              verbose,
				DryRun:               dryRun,
				UltraCompact:         ultraCompact,
				SkipEnv:              skipEnv,
				QueryIntent:          queryIntent,
				LLMEnabled:           llmEnabled,
				TokenBudget:          tokenBudget,
				FallbackArgs:         fallbackArgs,
				LayerPreset:          layerPreset,
				LayerProfile:         layerProfile,
				OutputFile:           outputFile,
				QuietMode:            quietMode,
				JSONOutput:           jsonOutput,
				RemoteMode:           remoteMode,
				CompressionAddr:      compressionAddr,
				AnalyticsAddr:        analyticsAddr,
				RemoteTimeout:        remoteTimeout,
				CompactionEnabled:    compactionEnabled,
				CompactionThreshold:  compactionThreshold,
				CompactionPreserve:   compactionPreserve,
				CompactionMaxTokens:  compactionMaxTokens,
				CompactionSnapshot:   compactionSnapshot,
				CompactionAutoDetect: compactionAutoDetect,
				ReversibleEnabled:    reversibleEnabled,
				EnableLayers:         enableLayers,
				DisableLayers:        disableLayers,
				StreamMode:           streamMode,
				PolicyRouter:         policyRouter,
				Extractive:           extractive,
				ExtractiveMax:        extractiveMax,
				ExtractiveHead:       extractiveHead,
				ExtractiveTail:       extractiveTail,
				ExtractiveSignal:     extractiveSignal,
				QualityGuardrail:     qualityGuardrail,
				DiffAdapt:            diffAdapt,
				EPiC:                 epic,
				SSDP:                 ssdp,
				AgentOCR:             agentOCR,
				S2MAD:                s2mad,
				ACON:                 acon,
				ResearchPack:         researchPack,
			})
			shared.SetConfigFile(cfgFile)

			// SkipEnv is tracked in AppState for child process env propagation.
			// Rather than mutating the global process environment with os.Setenv
			// (thread-unsafe), child processes inherit this via cmd.Env.
			// See shared.AppState for the SkipEnv field.
			_ = skipEnv // consumed by executor to pass env to child processes

			if isOperationalCommand(cmd) {
				if err := integrity.RuntimeCheck(); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			fallback := shared.GetFallback()
			output, handled, err := fallback.Handle(args)

			if !handled {
				return fmt.Errorf("unknown command: %s", args[0])
			}

			fmt.Print(output)
			return err
		},
	}
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// Unknown commands are handled by the TOML filter fallback system.
func Execute() int {
	// Enable unknown command handling
	rootCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
	rootCmd.TraverseChildren = true

	_, err := rootCmd.ExecuteC()
	if err != nil {
		// Check if this is an unknown command error
		if isUnknownCommandError(err) {
			// Extract the unknown command from args
			args := extractUnknownCommandArgs()
			if len(args) > 0 {
				fallback := shared.GetFallback()
				output, handled, ferr := fallback.Handle(args)
				if handled {
					fmt.Print(output)
					if ferr != nil {
						return exitCodeForError(ferr)
					}
					return 0
				}
			}
		}
		fmt.Fprintln(os.Stderr, err)
		return exitCodeForError(err)
	}

	return 0
}

// ExecuteContext runs the CLI with a context for graceful cancellation
func ExecuteContext(ctx context.Context) int {
	rootCmd.SetContext(ctx)
	return Execute()
}

func exitCodeForError(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return 1
}

// isUnknownCommandError checks if the error is an unknown command error
func isUnknownCommandError(err error) bool {
	return strings.Contains(err.Error(), "unknown command") ||
		strings.Contains(err.Error(), "unknown shorthand flag")
}

// extractUnknownCommandArgs extracts args for the fallback handler
func extractUnknownCommandArgs() []string {
	if len(fallbackArgs) == 0 && len(os.Args) > 1 {
		return os.Args[1:]
	}
	return fallbackArgs
}

func init() {
	registry.Init(rootCmd)

	cobra.OnInitialize(initConfig)

	// Version is already set in newRootCmd() from shared.Version
	rootCmd.SetVersionTemplate("TokMan {{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is %s)", config.ConfigPath()))
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v",
		"verbosity level (-v, -vv, -vvv)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false,
		"show what would be filtered without executing")
	rootCmd.PersistentFlags().BoolVarP(&ultraCompact, "ultra-compact", "u", true,
		"ultra-compact mode: ASCII icons, inline format (default: true)")
	rootCmd.PersistentFlags().BoolVar(&skipEnv, "skip-env", false,
		"set SKIP_ENV_VALIDATION=1 for child processes")
	rootCmd.PersistentFlags().StringVar(&queryIntent, "query", "",
		"query intent for compression (debug/review/deploy/search)")
	rootCmd.PersistentFlags().BoolVar(&llmEnabled, "llm", false,
		"enable LLM-based compression (requires Ollama/LM Studio)")
	rootCmd.PersistentFlags().IntVar(&tokenBudget, "budget", 0,
		"token budget for output (0 = unlimited, e.g., --budget 2000)")
	rootCmd.PersistentFlags().StringVar(&layerPreset, "preset", "",
		"pipeline preset: fast, balanced, or full (T90)")
	rootCmd.PersistentFlags().StringVar(&layerProfile, "profile", "",
		"compression mode: surface, trim, extract, core, adaptive, code, log, thread (auto-detects if unset)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "",
		"write output to file instead of stdout")
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false,
		"suppress all non-essential output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false,
		"machine-readable JSON output")

	// Compaction flags (Layer 11 - Semantic compression)
	rootCmd.PersistentFlags().BoolVar(&compactionEnabled, "compaction", true,
		"enable semantic compaction for chat/conversation content (default: true)")
	rootCmd.PersistentFlags().IntVar(&compactionThreshold, "compaction-threshold", 500,
		"minimum tokens to trigger compaction (default: 500)")
	rootCmd.PersistentFlags().IntVar(&compactionPreserve, "compaction-preserve", 10,
		"recent conversation turns to preserve verbatim (default: 10)")
	rootCmd.PersistentFlags().IntVar(&compactionMaxTokens, "compaction-max-tokens", 5000,
		"maximum tokens for compaction summary (default: 5000)")
	rootCmd.PersistentFlags().BoolVar(&compactionSnapshot, "compaction-snapshot", true,
		"use state snapshot format (4-section XML)")
	rootCmd.PersistentFlags().BoolVar(&compactionAutoDetect, "compaction-auto-detect", true,
		"auto-detect conversation content for compaction")

	// Reversible compression flag (R1)
	rootCmd.PersistentFlags().BoolVar(&reversibleEnabled, "reversible", false,
		"store original output for later restoration (use 'tokman restore' to retrieve)")

	// Remote mode flags (Phase 4 - Microservice)
	rootCmd.PersistentFlags().BoolVar(&remoteMode, "remote", false,
		"enable remote mode - connect to TokMan services via gRPC")
	rootCmd.PersistentFlags().StringVar(&compressionAddr, "compression-addr", "localhost:50051",
		"compression service address (default: localhost:50051)")
	rootCmd.PersistentFlags().StringVar(&analyticsAddr, "analytics-addr", "localhost:50053",
		"analytics service address (default: localhost:50053)")
	rootCmd.PersistentFlags().IntVar(&remoteTimeout, "remote-timeout", 30,
		"remote operation timeout in seconds (default: 30)")

	// Bind viper to flags — errors are non-fatal (flags are defined above).
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("query", rootCmd.PersistentFlags().Lookup("query"))
	_ = viper.BindPFlag("llm", rootCmd.PersistentFlags().Lookup("llm"))
	_ = viper.BindPFlag("budget", rootCmd.PersistentFlags().Lookup("budget"))
	_ = viper.BindPFlag("pipeline.enable_compaction", rootCmd.PersistentFlags().Lookup("compaction"))
	_ = viper.BindPFlag("pipeline.compaction_threshold", rootCmd.PersistentFlags().Lookup("compaction-threshold"))
	_ = viper.BindPFlag("pipeline.compaction_preserve_turns", rootCmd.PersistentFlags().Lookup("compaction-preserve"))
	_ = viper.BindPFlag("pipeline.compaction_max_tokens", rootCmd.PersistentFlags().Lookup("compaction-max-tokens"))
	_ = viper.BindPFlag("pipeline.compaction_state_snapshot", rootCmd.PersistentFlags().Lookup("compaction-snapshot"))
	_ = viper.BindPFlag("pipeline.compaction_auto_detect", rootCmd.PersistentFlags().Lookup("compaction-auto-detect"))

	// Custom layer configuration flags (Task 5)
	rootCmd.PersistentFlags().StringSliceVar(&enableLayers, "enable-layer", []string{},
		"enable specific layers (comma-separated: entropy,perplexity,h2o,etc.)")
	rootCmd.PersistentFlags().StringSliceVar(&disableLayers, "disable-layer", []string{},
		"disable specific layers (comma-separated: entropy,perplexity,h2o,etc.)")
	rootCmd.PersistentFlags().BoolVar(&streamMode, "stream", false,
		"enable streaming mode for large inputs (>500K tokens)")
	rootCmd.PersistentFlags().BoolVar(&policyRouter, "policy-router", false,
		"enable policy router to infer query intent from output")
	rootCmd.PersistentFlags().BoolVar(&extractive, "extractive-prefilter", false,
		"enable extractive prefilter for large outputs")
	rootCmd.PersistentFlags().IntVar(&extractiveMax, "extractive-max-lines", 400,
		"max lines before extractive prefilter triggers")
	rootCmd.PersistentFlags().IntVar(&extractiveHead, "extractive-head-lines", 80,
		"head lines to preserve in extractive prefilter")
	rootCmd.PersistentFlags().IntVar(&extractiveTail, "extractive-tail-lines", 60,
		"tail lines to preserve in extractive prefilter")
	rootCmd.PersistentFlags().IntVar(&extractiveSignal, "extractive-signal-lines", 120,
		"signal lines to preserve in extractive prefilter")
	rootCmd.PersistentFlags().BoolVar(&qualityGuardrail, "quality-guardrail", false,
		"enable quality guardrail with safe fallback when critical context is lost")
	rootCmd.PersistentFlags().BoolVar(&diffAdapt, "diff-adapt", false,
		"enable DiffAdapt difficulty-adaptive compression layer")
	rootCmd.PersistentFlags().BoolVar(&epic, "epic", false,
		"enable EPiC causal-edge preservation layer")
	rootCmd.PersistentFlags().BoolVar(&ssdp, "ssdp", false,
		"enable SSDP tree-of-thought branch pruning layer")
	rootCmd.PersistentFlags().BoolVar(&agentOCR, "agent-ocr", false,
		"enable AgentOCR multi-turn content-density compression layer")
	rootCmd.PersistentFlags().BoolVar(&s2mad, "s2-mad", false,
		"enable S2-MAD agreement-collapse layer for debate traces")
	rootCmd.PersistentFlags().BoolVar(&acon, "acon", false,
		"enable ACON adaptive context optimization layer")
	rootCmd.PersistentFlags().BoolVar(&researchPack, "research-pack", false,
		"enable research layer pack (31-36): DiffAdapt, EPiC, SSDP, AgentOCR, S2-MAD, ACON")

	_ = viper.BindPFlag("layers.enable", rootCmd.PersistentFlags().Lookup("enable-layer"))
	_ = viper.BindPFlag("layers.disable", rootCmd.PersistentFlags().Lookup("disable-layer"))
	_ = viper.BindPFlag("pipeline.streaming", rootCmd.PersistentFlags().Lookup("stream"))
	_ = viper.BindPFlag("pipeline.enable_policy_router", rootCmd.PersistentFlags().Lookup("policy-router"))
	_ = viper.BindPFlag("pipeline.enable_extractive_prefilter", rootCmd.PersistentFlags().Lookup("extractive-prefilter"))
	_ = viper.BindPFlag("pipeline.extractive_max_lines", rootCmd.PersistentFlags().Lookup("extractive-max-lines"))
	_ = viper.BindPFlag("pipeline.extractive_head_lines", rootCmd.PersistentFlags().Lookup("extractive-head-lines"))
	_ = viper.BindPFlag("pipeline.extractive_tail_lines", rootCmd.PersistentFlags().Lookup("extractive-tail-lines"))
	_ = viper.BindPFlag("pipeline.extractive_signal_lines", rootCmd.PersistentFlags().Lookup("extractive-signal-lines"))
	_ = viper.BindPFlag("pipeline.enable_quality_guardrail", rootCmd.PersistentFlags().Lookup("quality-guardrail"))
	_ = viper.BindPFlag("pipeline.enable_difft_adapt", rootCmd.PersistentFlags().Lookup("diff-adapt"))
	_ = viper.BindPFlag("pipeline.enable_epic", rootCmd.PersistentFlags().Lookup("epic"))
	_ = viper.BindPFlag("pipeline.enable_ssdp", rootCmd.PersistentFlags().Lookup("ssdp"))
	_ = viper.BindPFlag("pipeline.enable_agent_ocr", rootCmd.PersistentFlags().Lookup("agent-ocr"))
	_ = viper.BindPFlag("pipeline.enable_s2_mad", rootCmd.PersistentFlags().Lookup("s2-mad"))
	_ = viper.BindPFlag("pipeline.enable_acon", rootCmd.PersistentFlags().Lookup("acon"))
	_ = viper.BindPFlag("pipeline.enable_research_pack", rootCmd.PersistentFlags().Lookup("research-pack"))

	registry.RegisterAll()
}

// initConfig reads in config file and ENV variables if set.
// Delegates to config.Load() for a single source of truth.
func initConfig() {
	if _, err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}

	// Initialize logger
	logLevel := utils.LevelInfo
	if verbose > 0 {
		logLevel = utils.LevelDebug
	}
	if err := utils.InitLogger(config.LogPath(), logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
	}
}

// integrityExemptAnnotation is the key used in cobra.Command.Annotations to
// mark commands that should skip hook integrity verification. New commands
// should set Annotations[integrityExemptAnnotation] = "true" rather than
// relying on the name-based fallback.
const integrityExemptAnnotation = "tokman:skip_integrity"

// metaCommandNames is a backward-compatibility fallback for commands that
// predate the annotation-based approach. New commands should use annotations.
var metaCommandNames = map[string]bool{
	"init": true, "verify": true, "config": true, "economics": true,
	"status": true, "report": true, "summary": true, "ccusage": true,
	"help": true, "version": true, "rewrite": true, "deps": true,
	"gain": true, "hook-audit": true, "discover": true, "learn": true, "err": true,
}

// isOperationalCommand returns true for commands that process CLI output
// and need runtime integrity verification. Meta commands are excluded
// via cobra.Annotations["tokman:skip_integrity"] = "true" (preferred)
// or by name in the metaCommandNames fallback list.
func isOperationalCommand(cmd *cobra.Command) bool {
	// Check annotation-based exemption (preferred for new commands)
	if cmd.Annotations[integrityExemptAnnotation] == "true" {
		return false
	}

	// Walk up parent chain -- if any parent is exempt, child is too
	for p := cmd.Parent(); p != nil; p = p.Parent() {
		if p.Annotations[integrityExemptAnnotation] == "true" {
			return false
		}
	}

	// Name-based fallback for commands registered before annotations
	name := cmd.Name()
	if metaCommandNames[name] {
		return false
	}
	for p := cmd.Parent(); p != nil; p = p.Parent() {
		if metaCommandNames[p.Name()] {
			return false
		}
	}

	return true
}
