package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Show TokMan adoption across sessions",
	Long:  `Display session history, adoption rate, and token savings per session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSession()
	},
}

func runSession() error {
	dbPath := tracking.DatabasePath()
	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return err
	}
	defer tracker.Close()

	recent, err := tracker.GetRecentCommands("", 500)
	if err != nil {
		return err
	}

	if len(recent) == 0 {
		fmt.Println("No session data available.")
		return nil
	}

	totalCmds := len(recent)
	totalSaved := 0
	totalOriginal := 0
	for _, r := range recent {
		totalSaved += r.SavedTokens
		totalOriginal += r.OriginalTokens
	}

	avgSavings := 0.0
	if totalOriginal > 0 {
		avgSavings = float64(totalSaved) / float64(totalOriginal) * 100
	}

	fmt.Printf("Session Summary (last %d commands)\n", totalCmds)
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("  Commands:     %d\n", totalCmds)
	fmt.Printf("  Tokens saved: %s\n", formatTokens(totalSaved))
	fmt.Printf("  Avg savings:  %.1f%%\n", avgSavings)
	fmt.Printf("  Adoption:     %.0f%%\n", float64(totalSaved)/float64(totalOriginal)*100)
	return nil
}

func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func init() {
	registry.Add(func() { registry.Register(sessionCmd) })
}
