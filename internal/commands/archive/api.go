package archive

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var apiAddr string

func init() {
	registry.Add(func() {
		registry.Register(apiCmd)
	})
}

var apiCmd = &cobra.Command{
	Use:   "archive-api",
	Short: "Start archive REST API server",
	Long:  `Start a REST API server for managing archives.`,
	Example: `  tokman archive-api
  tokman archive-api --addr=:8081`,
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().StringVar(&apiAddr, "addr", ":8081", "API server address")
}

func runAPI(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, err := archive.NewArchiveManager(archive.DefaultArchiveConfig())
	if err != nil {
		return fmt.Errorf("failed to create archive manager: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	api := archive.NewAPIServer(mgr, apiAddr)

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		api.Stop(ctx)
	}()

	fmt.Printf("Archive API server starting on %s\n", apiAddr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET    /api/v1/archives       - List archives")
	fmt.Println("  POST   /api/v1/archives       - Create archive")
	fmt.Println("  GET    /api/v1/archives/{hash} - Get archive")
	fmt.Println("  DELETE /api/v1/archives/{hash} - Delete archive")
	fmt.Println("  GET    /api/v1/archives/search?q={query} - Search")
	fmt.Println("  GET    /api/v1/archives/stats  - Statistics")
	fmt.Println("  POST   /api/v1/archives/export - Export")
	fmt.Println("  POST   /api/v1/archives/import - Import")
	fmt.Println("  GET    /health                 - Health check")

	return api.Start()
}
