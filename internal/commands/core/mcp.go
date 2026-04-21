package core

import (
	"os"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/mcp"
)

func init() {
	registry.Add(func() { registry.Register(mcpCmd) })
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start tok as an MCP server",
	Long: `Start tok as a Model Context Protocol (MCP) server.

This allows AI assistants (Claude, Cursor, etc.) to use tok's filtering
capabilities as MCP tools instead of wrapping CLI commands.

The server exposes these tools:
  - tok_filter: Filter arbitrary text through compression pipeline
  - tok_compress_file: Read and compress file content
  - tok_analyze_output: Analyze structure without filtering
  - tok_get_stats: Get usage statistics
  - tok_explain_layers: Explain compression layers

Usage:
  # Start MCP server with stdio transport (default)
  tok mcp

  # Configure in Claude Code (.mcp.json)
  {
    "mcpServers": {
      "tok": {
        "command": "tok",
        "args": ["mcp"]
      }
    }
  }

  # Configure in Cursor (mcp.json)
  {
    "mcpServers": {
      "tok": {
        "command": "tok",
        "args": ["mcp"],
        "env": {
          "TOK_MODE": "balanced"
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
		server := mcp.NewServer("tok", "1.0.0", pipeline)

		// Run stdio server (blocking)
		if err := server.RunStdio(); err != nil {
			out.Global().Errorf("MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}
