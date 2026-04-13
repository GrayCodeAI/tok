{
  "id": "51be1eae",
  "title": "Implement MCP Server Mode for TokMan",
  "status": "draft",
  "created_at": "2026-04-12T17:02:14.193Z",
  "assigned_to_session": "22ec26ab-caed-4b78-95bf-ebd2f15d0700",
  "steps": [
    {
      "id": 1,
      "text": "Create internal/mcp/ package structure",
      "done": true
    },
    {
      "id": 2,
      "text": "Implement MCP protocol handlers (stdio/HTTP)",
      "done": true
    },
    {
      "id": 3,
      "text": "Create tool definitions and handlers",
      "done": true
    },
    {
      "id": 4,
      "text": "Integrate with existing filter pipeline",
      "done": true
    },
    {
      "id": 5,
      "text": "Add CLI command `tokman mcp` to start server",
      "done": true
    },
    {
      "id": 6,
      "text": "Write tests for MCP handlers",
      "done": true
    },
    {
      "id": 7,
      "text": "Update documentation",
      "done": false
    }
  ]
}

Add Model Context Protocol (MCP) server support to TokMan, exposing filter capabilities as MCP tools that AI assistants can call directly. This bridges the gap with Token Savior's architecture.

## Goals
- Create `internal/mcp/` package with MCP server implementation
- Expose TokMan filters as MCP tools
- Support stdio and HTTP transport
- Maintain CLI proxy functionality alongside MCP mode

## Tools to Expose
1. `tokman_filter` - Filter arbitrary text through pipeline
2. `tokman_compress_file` - Compress file content
3. `tokman_analyze_output` - Analyze structure without filtering
4. `tokman_get_stats` - Get savings statistics
5. `tokman_find_symbol` - Find symbol in codebase (index integration)
