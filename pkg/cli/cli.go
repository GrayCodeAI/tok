package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/GrayCodeAI/tok/internal/commands"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/core"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// Run executes tok CLI in-process with the provided args.
// Args should not include the binary name.
func Run(args []string, version string) int {
	shared.Version = version
	core.WarmupBPETokenizer()

	origArgs := os.Args
	os.Args = append([]string{"tok"}, args...)
	defer func() { os.Args = origArgs }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var closeTrackerOnce sync.Once
	closeTracker := func() {
		closeTrackerOnce.Do(func() {
			_ = tracking.CloseGlobalTracker()
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		cancel()
		closeTracker()
	}()

	exitCode := commands.ExecuteContext(ctx)
	closeTracker()
	return exitCode
}
