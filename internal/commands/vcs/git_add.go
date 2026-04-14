package vcs

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/shared"
)

var gitAddCmd = &cobra.Command{
	Use:   "add [args...]",
	Short: "Add files to staging (compact output)",
	Long:  `Add files to git staging area with filtered output. Supports all git add options including -A, --all, -u, etc.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return shared.ExecuteAndRecord("git add", func() (string, string, error) {
			return runGitAdd(args)
		})
	},
}

func runGitAdd(args []string) (string, string, error) {
	addArgs := args

	hasAllFlag := false
	hasUpdateFlag := false
	for _, arg := range args {
		if arg == "-A" || arg == "--all" || arg == "--no-ignore-removal" {
			hasAllFlag = true
		}
		if arg == "-u" || arg == "--update" {
			hasUpdateFlag = true
		}
	}

	if len(addArgs) == 0 || (!hasAllFlag && !hasUpdateFlag && len(args) == 0) {
		addArgs = []string{"-A"}
	}

	addCmd := buildGitCmd("add", addArgs...)
	output, err := addCmd.CombinedOutput()
	raw := string(output)
	if err != nil {
		return raw, "", fmt.Errorf("git add failed: %w\n%s", err, output)
	}

	statCmd := buildGitCmd("diff", "--cached", "--stat", "--shortstat")
	statOut, _ := statCmd.Output()
	stat := strings.TrimSpace(string(statOut))

	if stat == "" {
		return raw, "ok (nothing to add)", nil
	}

	lines := strings.Split(stat, "\n")
	lastLine := lines[len(lines)-1]
	return raw, fmt.Sprintf("ok %s", lastLine), nil
}
