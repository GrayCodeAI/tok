package core

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/server"
)

var serverAddr string
var serverPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server",
	Long: `Start the TokMan HTTP API server for remote compression.
		
The server provides REST endpoints for:
- /health - Health check
- /api/v1/compress - Compress text
- /api/v1/metrics - Get metrics
- /api/v1/config - Get/update configuration
		
Examples:
  tokman server                    # Start on default port 8080
  tokman server --port 9000       # Start on port 9000
  tokman server --addr :9000      # Custom address`,
	RunE: runServer,
}

func init() {
	serverCmd.Flags().StringVar(&serverAddr, "addr", "", "server address (default: :8080)")
	serverCmd.Flags().IntVar(&serverPort, "port", 8080, "server port (deprecated, use --addr)")
	registry.Add(func() { registry.Register(serverCmd) })
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := shared.GetCachedConfig()
	if err != nil {
		return err
	}

	addr := serverAddr
	if addr == "" {
		if serverPort != 8080 {
			addr = ":" + string(rune(serverPort))
		} else {
			addr = ":8080"
		}
	}

	srv := server.New(cfg, shared.Version)

	// Use command context for graceful shutdown on signal
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return srv.Run(ctx, addr)
}
