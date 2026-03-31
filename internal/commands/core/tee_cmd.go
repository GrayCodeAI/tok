package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/tee"
)

var teeCmd = &cobra.Command{
	Use:   "tee",
	Short: "Manage full output recovery (tee)",
	Long:  `Save and retrieve full command output for recovery when compression fails.`,
}

var teeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved tee files",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := tee.List(tee.DefaultConfig())
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No saved tee files.")
			return nil
		}
		fmt.Printf("%-20s  %-30s  %s\n", "Date", "Command", "File")
		fmt.Println("────────────────────────────────────────────────────────────────")
		for _, e := range entries {
			fmt.Printf("%-20s  %-30s  %s\n",
				e.Timestamp.Format("2006-01-02 15:04"),
				truncCmd(e.Command, 30),
				e.Filename)
		}
		return nil
	},
}

var teeReadCmd = &cobra.Command{
	Use:   "read [filename]",
	Short: "Read a saved tee file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := tee.Read(args[0], tee.DefaultConfig())
		if err != nil {
			return err
		}
		fmt.Print(content)
		return nil
	},
}

func init() {
	registry.Add(func() { registry.Register(teeCmd) })
	registry.Add(func() { registry.Register(teeListCmd) })
	registry.Add(func() { registry.Register(teeReadCmd) })
	teeCmd.AddCommand(teeListCmd, teeReadCmd)
}

func truncCmd(s string, n int) string {
	if len(s) > n {
		return s[:n-3] + "..."
	}
	return s
}
