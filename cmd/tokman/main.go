package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/GrayCodeAI/tokman/internal/commands"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var version = "dev"

func main() {
	shared.Version = version

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		if tracker := tracking.GetGlobalTracker(); tracker != nil {
			tracker.Close()
		}
		os.Exit(0)
	}()

	commands.Execute()

	if tracker := tracking.GetGlobalTracker(); tracker != nil {
		tracker.Close()
	}
}
