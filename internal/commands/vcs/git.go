package vcs

import (
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
)

// Git status codes
const (
	GitStaged     = "staged"
	GitModified   = "modified"
	GitUntracked  = "untracked"
	GitDeleted    = "deleted"
	GitConflicted = "conflicted"
)

// Global git flags (persisted to all subcommands)
var (
	gitDir          string
	gitWorkTree     string
	gitDirectory    string // -C flag
	gitNoPager      bool
	gitNoOptLocks   bool
	gitBare         bool
	gitLiteralPaths bool
	gitConfigOpts   []string // -c key=value options
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git command wrappers with output filtering",
	Long: `Wrap git commands with intelligent output filtering to reduce
token usage while preserving important information.

Global flags (applied before subcommand):
  -C, --directory <path>      Run git in specified directory
  --git-dir <path>            Set the .git directory path
  --work-tree <path>          Set the working tree path
  --no-pager                  Disable pager
  --no-optional-locks         Skip optional locks
  --bare                      Treat repository as bare
  --literal-pathspecs         Treat pathspecs literally
  -c, --config <key=value>    Set git config option

Note: All git commands are supported. Commands with explicit tok
filters (status, diff, log, etc.) use optimized output. Other commands
pass through to git with global flags applied.`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	TraverseChildren:   true,
	SilenceUsage:       true,
	SilenceErrors:      true,
	Args:               cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand specified, show help
		if len(args) == 0 {
			cmd.Help()
			return
		}
		// Pass through to real git
		runGitPassthrough(args)
	},
}

// buildGitCmd builds a git command with global flags prepended
func buildGitCmd(subCmd string, args ...string) *exec.Cmd {
	gitArgs := []string{}

	// Add global flags
	if gitDirectory != "" {
		gitArgs = append(gitArgs, "-C", gitDirectory)
	}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	if gitWorkTree != "" {
		gitArgs = append(gitArgs, "--work-tree", gitWorkTree)
	}
	if gitNoPager {
		gitArgs = append(gitArgs, "--no-pager")
	}
	if gitNoOptLocks {
		gitArgs = append(gitArgs, "--no-optional-locks")
	}
	if gitBare {
		gitArgs = append(gitArgs, "--bare")
	}
	if gitLiteralPaths {
		gitArgs = append(gitArgs, "--literal-pathspecs")
	}
	for _, opt := range gitConfigOpts {
		gitArgs = append(gitArgs, "-c", opt)
	}

	gitArgs = append(gitArgs, subCmd)
	gitArgs = append(gitArgs, args...)

	return exec.Command("git", gitArgs...)
}

// extractGitArgs filters tok-specific flags from args, leaving only git-compatible ones
func extractGitArgs(args []string) []string {
	var gitArgs []string
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		// Skip tok-specific flags that take values
		if arg == "--query" || arg == "--budget" || arg == "--preset" ||
			arg == "--output" || arg == "-o" ||
			arg == "--compaction-threshold" || arg == "--compaction-preserve" ||
			arg == "--compaction-max-tokens" {
			skipNext = true
			continue
		}
		// Skip tok-specific boolean flags
		if strings.HasPrefix(arg, "--ultra-compact") ||
			arg == "-u" ||
			strings.HasPrefix(arg, "--verbose") ||
			arg == "-v" ||
			strings.HasPrefix(arg, "-vv") ||
			strings.HasPrefix(arg, "-vvv") ||
			arg == "--dry-run" ||
			arg == "--llm" ||
			arg == "--skip-env" ||
			arg == "--quiet" ||
			arg == "-q" ||
			arg == "--json" ||
			arg == "--compaction" ||
			arg == "--compaction-snapshot" ||
			arg == "--compaction-auto-detect" {
			continue
		}
		// Skip tok flags with values
		if strings.HasPrefix(arg, "--query=") ||
			strings.HasPrefix(arg, "--config=") ||
			strings.HasPrefix(arg, "-c=") ||
			strings.HasPrefix(arg, "--budget=") ||
			strings.HasPrefix(arg, "--preset=") ||
			strings.HasPrefix(arg, "--output=") ||
			strings.HasPrefix(arg, "-o=") ||
			strings.HasPrefix(arg, "--compaction-threshold=") ||
			strings.HasPrefix(arg, "--compaction-preserve=") ||
			strings.HasPrefix(arg, "--compaction-max-tokens=") {
			continue
		}
		// Skip tok flags with values (positional form)
		if (arg == "--budget" || arg == "--preset" || arg == "--output" || arg == "-o" ||
			arg == "--compaction-threshold" || arg == "--compaction-preserve" ||
			arg == "--compaction-max-tokens") && i+1 < len(args) {
			skipNext = true
			continue
		}
		gitArgs = append(gitArgs, arg)
	}
	return gitArgs
}

// runGitPassthrough executes an unknown git subcommand directly
func runGitPassthrough(args []string) {
	gitArgs := []string{}

	// Add global flags first
	if gitDirectory != "" {
		gitArgs = append(gitArgs, "-C", gitDirectory)
	}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	if gitWorkTree != "" {
		gitArgs = append(gitArgs, "--work-tree", gitWorkTree)
	}
	if gitNoPager {
		gitArgs = append(gitArgs, "--no-pager")
	}
	if gitNoOptLocks {
		gitArgs = append(gitArgs, "--no-optional-locks")
	}
	if gitBare {
		gitArgs = append(gitArgs, "--bare")
	}
	if gitLiteralPaths {
		gitArgs = append(gitArgs, "--literal-pathspecs")
	}
	for _, opt := range gitConfigOpts {
		gitArgs = append(gitArgs, "-c", opt)
	}

	// Add the subcommand and args
	gitArgs = append(gitArgs, args...)

	// Execute git directly
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	cmd.Run()
}

// gray returns a gray-colored string
func gray(s string) string {
	dim := color.New(color.FgHiBlack).SprintFunc()
	return dim(s)
}

// isHexHash checks if the first 7 characters of a string are hex digits
func isHexHash(s string) bool {
	if len(s) < 7 {
		return false
	}
	for _, c := range s[:7] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func init() {
	registry.Add(func() { registry.Register(gitCmd) })

	// Global flags for git command
	gitCmd.PersistentFlags().StringVarP(&gitDirectory, "directory", "C", "", "Run git in specified directory")
	gitCmd.PersistentFlags().StringVar(&gitDir, "git-dir", "", "Set the .git directory path")
	gitCmd.PersistentFlags().StringVar(&gitWorkTree, "work-tree", "", "Set the working tree path")
	gitCmd.PersistentFlags().BoolVar(&gitNoPager, "no-pager", false, "Disable pager")
	gitCmd.PersistentFlags().BoolVar(&gitNoOptLocks, "no-optional-locks", false, "Skip optional locks")
	gitCmd.PersistentFlags().BoolVar(&gitBare, "bare", false, "Treat repository as bare")
	gitCmd.PersistentFlags().BoolVar(&gitLiteralPaths, "literal-pathspecs", false, "Treat pathspecs literally")
	gitCmd.PersistentFlags().StringArrayVarP(&gitConfigOpts, "config", "c", nil, "Set git config option (key=value)")

	// Add subcommands
	gitCmd.AddCommand(gitStatusCmd)
	gitCmd.AddCommand(gitDiffCmd)
	gitCmd.AddCommand(gitLogCmd)
	gitCmd.AddCommand(gitShowCmd)
	gitCmd.AddCommand(gitAddCmd)
	gitCmd.AddCommand(gitCommitCmd)
	gitCmd.AddCommand(gitPushCmd)
	gitCmd.AddCommand(gitPullCmd)
	gitCmd.AddCommand(gitBranchCmd)
	gitCmd.AddCommand(gitFetchCmd)
	gitCmd.AddCommand(gitStashCmd)
	gitCmd.AddCommand(gitWorktreeCmd)

	// Git log specific flags
	gitLogCmd.Flags().IntVarP(&gitLogCount, "number", "n", 0, "Number of commits to show")

	// Add passthrough subcommands for common git operations
	addGitPassthroughCommands()
}

// addGitPassthroughCommands adds passthrough subcommands for git operations without explicit filters
func addGitPassthroughCommands() {
	passthroughCommands := []string{
		// All official git commands (excluding the 12 with explicit tok filters)
		"am", "annotate", "apply", "archimport", "archive", "backfill",
		"bisect", "blame", "bugreport", "bundle",
		"checkout", "cherry", "cherry-pick", "citool", "clean", "clone", "column",
		"commit-graph", "commit-tree", "config", "count-objects",
		"credential", "credential-cache", "credential-store",
		"cvsexportcommit", "cvsimport", "cvsserver",
		"daemon", "describe", "diagnose", "diff-files", "diff-index",
		"diff-pairs", "diff-tree", "difftool",
		"fast-export", "fast-import", "fetch-pack", "filter-branch",
		"fmt-merge-msg", "for-each-ref", "for-each-repo", "format-patch", "fsck",
		"gc", "get-tar-commit-id", "gitk", "grep", "gui",
		"hash-object", "help", "hook", "http-backend", "http-fetch", "http-push",
		"imap-send", "index-pack", "init", "init-db", "instaweb", "interpret-trailers",
		"ls-files", "ls-remote", "ls-tree",
		"mailinfo", "mailsplit", "maintenance", "merge", "merge-base", "merge-file",
		"merge-index", "merge-one-file", "merge-tree", "mergetool", "mktag", "mktree",
		"multi-pack-index", "mv",
		"name-rev", "notes",
		"pack-objects", "pack-redundant", "pack-refs", "p4", "patch-id", "prune",
		"prune-packed",
		"quiltimport",
		"range-diff", "read-tree", "rebase", "receive-pack", "reflog", "refs",
		"remote", "repack", "replace", "replay", "request-pull", "rerere",
		"reset", "restore", "rev-list", "rev-parse", "revert", "rm",
		"scalar", "send-email", "send-pack", "sh-i18n", "sh-setup", "shell",
		"shortlog", "show-branch", "show-index", "show-ref",
		"sparse-checkout", "stage", "stripspace", "submodule",
		"subtree", "svn", "switch", "symbolic-ref",
		"tag",
		"unpack-file", "unpack-objects", "update-index", "update-ref",
		"update-server-info", "upload-archive", "upload-pack",
		"var", "verify-commit", "verify-pack", "verify-tag", "version",
		"whatchanged", "write-tree",
	}

	for _, cmdName := range passthroughCommands {
		name := cmdName // capture for closure
		cmd := &cobra.Command{
			Use:                name + " [args...]",
			Short:              "Git " + name + " (passthrough to git)",
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				gitArgs := []string{name}
				gitArgs = append(gitArgs, args...)
				return runGitPassthroughReturn(gitArgs)
			},
		}
		gitCmd.AddCommand(cmd)
	}
}

// runGitPassthroughReturn executes git with args and returns error
func runGitPassthroughReturn(args []string) error {
	gitArgs := []string{}

	// Add global flags
	if gitDirectory != "" {
		gitArgs = append(gitArgs, "-C", gitDirectory)
	}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	if gitWorkTree != "" {
		gitArgs = append(gitArgs, "--work-tree", gitWorkTree)
	}
	if gitNoPager {
		gitArgs = append(gitArgs, "--no-pager")
	}
	if gitNoOptLocks {
		gitArgs = append(gitArgs, "--no-optional-locks")
	}
	if gitBare {
		gitArgs = append(gitArgs, "--bare")
	}
	if gitLiteralPaths {
		gitArgs = append(gitArgs, "--literal-pathspecs")
	}
	for _, opt := range gitConfigOpts {
		gitArgs = append(gitArgs, "-c", opt)
	}

	gitArgs = append(gitArgs, args...)

	// Execute git directly
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}
