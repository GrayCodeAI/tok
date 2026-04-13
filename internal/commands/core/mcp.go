package core

import (
	"fmt"
	"os"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/mcp"
	"github.com/spf13/cobra"
)

func init() {
	registry.Add(func() { registry.Register(mcpCmd) })
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start TokMan as an MCP server",
	Long: `Start TokMan as a Model Context Protocol (MCP) server.

This allows AI assistants (Claude, Cursor, etc.) to use TokMan's filtering
capabilities as MCP tools instead of wrapping CLI commands.

The server exposes these tools:
  - tokman_filter: Filter arbitrary text through compression pipeline
  - tokman_compress_file: Read and compress file content
  - tokman_analyze_output: Analyze structure without filtering
  - tokman_get_stats: Get usage statistics
  - tokman_explain_layers: Explain compression layers

Usage:
  # Start MCP server with stdio transport (default)
  tokman mcp

  # Configure in Claude Code (.mcp.json)
  {
    "mcpServers": {
      "tokman": {
        "command": "tokman",
        "args": ["mcp"]
      }
    }
  }

  # Configure in Cursor (mcp.json)
  {
    "mcpServers": {
      "tokman": {
        "command": "tokman",
        "args": ["mcp"],
        "env": {
          "TOKMAN_MODE": "balanced"
        }
      }
    }
  }
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create default pipeline
		cfg := filter.PipelineConfig{
			Mode: filter.ModeMinimal,
		}
		pipeline := filter.NewPipelineCoordinator(cfg)

		// Create and run MCP server
		server := mcp.NewServer("tokman", "1.0.0", pipeline)

		// Run stdio server (blocking)
		if err := server.RunStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}
