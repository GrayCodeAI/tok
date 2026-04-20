// tok - unified token optimization CLI
//
// tok combines input compression (for human-written text) and output filtering
// (for terminal/tool output) into a single CLI tool.
//
// Usage:
//
//	tok <command> [args]
//
// Examples:
//
//	tok compress -mode ultra -input "text to compress"
//	tok git status
//	tok doctor
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lakshmanpatel/tok/internal/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	code := commands.ExecuteContext(ctx)
	if code != 0 {
		os.Exit(code)
	}
}
