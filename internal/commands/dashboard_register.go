package commands

import (
	"github.com/GrayCodeAI/tokman/internal/dashboard"
)

func init() {
	rootCmd.AddCommand(dashboard.Cmd())
}
