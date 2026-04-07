# TokMan MCP Tools Reference

## Overview

TokMan provides **27 MCP (Model Context Protocol) tools** for intelligent context management, file operations, and token optimization. These tools can be used by Claude Desktop, Cursor, and other MCP-compatible clients.

## Tool Categories

### Core Context Tools

**ctx_read** - Read file content with 7 different modes
```typescript
ctx_read({
  path: "/path/to/file.ts",
  mode: "full" | "map" | "outline" | "symbols" | "imports" | "types" | "exports",
  max_tokens: 5000
})
```

**ctx_delta** - Get diff from previous file version
```typescript
ctx_delta({
  path: "/path/to/file.ts",
  base_hash: "abc123"
})
```

**ctx_grep** - Search files with regex pattern
```typescript
ctx_grep({
  path: "/directory",
  pattern: "TODO|FIXME|BUG",
  context: 2
})
```

**ctx_hash** - Compute SHA-256 hash of content
```typescript
ctx_hash({ content: "text content" })
```

**ctx_cache_info** - Get cache statistics
```typescript
ctx_cache_info()
```

**ctx_invalidate** - Invalidate cache entries
```typescript
ctx_invalidate({ pattern: "*.ts" })
```

**ctx_compact** - Compress content using TokMan filters
```typescript
ctx_compact({ content: "large output" })
```

**ctx_summary** - Get file summary/preview
```typescript
ctx_summary({ path: "/path/to/file.ts" })
```

### Memory Tools

**ctx_remember** - Store a memory entry
```typescript
ctx_remember({
  key: "project-notes",
  value: "Important notes here",
  expires_in_hours: 24
})
```

**ctx_recall** - Retrieve a memory entry
```typescript
ctx_recall({ key: "project-notes" })
```

**ctx_search_memory** - Search memory entries
```typescript
ctx_search_memory({ query: "notes" })
```

### Bundle Tools

**ctx_bundle** - Create a bundle of multiple files
```typescript
ctx_bundle({
  paths: ["file1.ts", "file2.ts", "README.md"]
})
```

**ctx_bundle_changed** - Bundle files changed in git
```typescript
ctx_bundle_changed()
```

**ctx_bundle_summary** - Get bundle statistics
```typescript
ctx_bundle_summary()
```

### Utility Tools

**ctx_exec** - Execute a shell command safely
```typescript
ctx_exec({ command: "ls -la" })
```

**ctx_tldr** - Get command help from TLDR pages
```typescript
ctx_tldr({ command: "git" })
```

**ctx_patterns** - List available hook patterns
```typescript
ctx_patterns()
```

**ctx_modes** - List available context modes
```typescript
ctx_modes()
```

*ctx_mode** - Set default context mode
```typescript
ctx_mode({ mode: "outline" })
```

**ctx_status** - Get MCP server status
```typescript
ctx_status()
```

**ctx_config** - Get or set configuration
```typescript
ctx_config({ action: "get" | "set", key: "...", value: "..." })
```

**ctx_mcp** - Export MCP configuration for clients
```typescript
ctx_mcp()
```

## MCP Configuration

### Claude Desktop

```json
{
  "mcpServers": {
    "tokman": {
      "command": "/path/to/tokman",
      "args": ["mcp", "start", "--transport", "stdio"]
    }
  }
}
```

### Cursor

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "tokman": {
      "command": "/path/to/tokman",
      "args": ["mcp", "start", "--transport", "stdio"]
    }
  }
}
```

## Token Savings

When used with MCP tools, TokMan can achieve:

| Tool | Typical Savings | Notes |
|------|----------------|-------|
| ctx_read | 80% | Mode-based reading |
| ctx_delta | 85% | Diff-only updates |
| ctx_grep | 90% | Match-only output |
| ctx_bundle | 75% | Multi-file compression |
| ctx_compact | 60-90% | Pipeline compression |

## Setup

1. Install TokMan: `brew install tokman` or `go install`
2. Start MCP server: `tokman mcp start`
3. Add configuration to your AI tool
4. Verify: `ctx_status`

## Best Practices

1. Use `ctx_read` with appropriate mode instead of full file reads
2. Check `ctx_cache_info` before re-reading files
3. Use `ctx_bundle` for multi-file operations
4. Leverage `ctx_memor
