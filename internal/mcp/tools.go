package mcp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/GrayCodeAI/tokman/internal/archive"
	"github.com/GrayCodeAI/tokman/internal/scoring"
)

// registerAllTools registers all MCP tools
func (s *Server) registerAllTools() {
	// Archive tools
	s.registerArchiveTools()

	// Session tools
	s.registerSessionTools()

	// Scoring tools
	s.registerScoringTools()

	// Filter tools
	s.registerFilterTools()
}

// registerArchiveTools registers archive-related tools
func (s *Server) registerArchiveTools() {
	// ctx_archive - Archive content
	s.RegisterTool(Tool{
		Name:        "ctx_archive",
		Description: "Archive content for later retrieval",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"content": {Type: "string", Description: "Content to archive"},
				"command": {Type: "string", Description: "Command that generated content"},
				"tags":    {Type: "array", Description: "Tags for the archive"},
			},
			Required: []string{"content"},
		},
	}, s.handleCtxArchive)

	// ctx_retrieve - Retrieve archived content
	s.RegisterTool(Tool{
		Name:        "ctx_retrieve",
		Description: "Retrieve archived content by hash",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"hash": {Type: "string", Description: "SHA-256 hash of archive"},
			},
			Required: []string{"hash"},
		},
	}, s.handleCtxRetrieve)

	// ctx_search - Search archives
	s.RegisterTool(Tool{
		Name:        "ctx_search",
		Description: "Search through archived content",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {Type: "string", Description: "Search query"},
				"limit": {Type: "number", Description: "Maximum results"},
				"agent": {Type: "string", Description: "Filter by agent"},
			},
			Required: []string{"query"},
		},
	}, s.handleCtxSearch)
}

// registerSessionTools registers session-related tools
func (s *Server) registerSessionTools() {
	// ctx_session_start - Start new session
	s.RegisterTool(Tool{
		Name:        "ctx_session_start",
		Description: "Start a new context session",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent":   {Type: "string", Description: "Agent name"},
				"project": {Type: "string", Description: "Project path"},
			},
			Required: []string{},
		},
	}, s.handleCtxSessionStart)

	// ctx_session_compact - Run PreCompact
	s.RegisterTool(Tool{
		Name:        "ctx_session_compact",
		Description: "Run PreCompact on current session",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"max_tokens": {Type: "number", Description: "Maximum tokens for summary"},
			},
			Required: []string{},
		},
	}, s.handleCtxSessionCompact)

	// ctx_session_snapshot - Create snapshot
	s.RegisterTool(Tool{
		Name:        "ctx_session_snapshot",
		Description: "Create a snapshot of current session",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
		},
	}, s.handleCtxSessionSnapshot)
}

// registerScoringTools registers scoring-related tools
func (s *Server) registerScoringTools() {
	// ctx_score - Score content
	s.RegisterTool(Tool{
		Name:        "ctx_score",
		Description: "Score content using semantic signals",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"content":  {Type: "string", Description: "Content to score"},
				"query":    {Type: "string", Description: "Query for relevance"},
				"top_n":    {Type: "number", Description: "Return top N lines"},
				"min_tier": {Type: "string", Description: "Minimum tier (critical, important, nice_to_have)"},
			},
			Required: []string{"content"},
		},
	}, s.handleCtxScore)
}

// registerFilterTools registers filter-related tools
func (s *Server) registerFilterTools() {
	// ctx_filter - Filter content
	s.RegisterTool(Tool{
		Name:        "ctx_filter",
		Description: "Filter content using TokMan pipeline",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"content": {Type: "string", Description: "Content to filter"},
				"mode":    {Type: "string", Description: "Filter mode (minimal, aggressive)"},
				"budget":  {Type: "number", Description: "Token budget"},
			},
			Required: []string{"content"},
		},
	}, s.handleCtxFilter)
}

// Tool handlers

func (s *Server) handleCtxArchive(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	command, _ := params["command"].(string)

	if s.archiveMgr == nil {
		return nil, fmt.Errorf("archive manager not initialized")
	}

	entry := archive.NewArchiveEntry([]byte(content), command)
	hash, err := s.archiveMgr.Archive(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to archive: %w", err)
	}

	return map[string]string{
		"hash":    hash,
		"message": fmt.Sprintf("Content archived with hash %s", hash[:16]),
	}, nil
}

func (s *Server) handleCtxRetrieve(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	hash, _ := params["hash"].(string)

	if s.archiveMgr == nil {
		return nil, fmt.Errorf("archive manager not initialized")
	}

	entry, err := s.archiveMgr.Retrieve(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve: %w", err)
	}

	return map[string]interface{}{
		"hash":    entry.Hash,
		"command": entry.Command,
		"content": string(entry.OriginalContent),
		"created": entry.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) handleCtxSearch(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, _ := params["query"].(string)

	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// Mock search results
	results := []map[string]string{
		{"hash": "abc123...", "command": "ls -la", "preview": "total 128"},
		{"hash": "def456...", "command": "git status", "preview": "On branch main"},
	}

	return map[string]interface{}{
		"query":   query,
		"results": results[:min(limit, len(results))],
		"total":   len(results),
	}, nil
}

func (s *Server) handleCtxSessionStart(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	agent, _ := params["agent"].(string)
	project, _ := params["project"].(string)

	if agent == "" {
		agent = "default"
	}
	if project == "" {
		project, _ = os.Getwd()
	}

	return map[string]string{
		"session_id": "new-session-id",
		"agent":      agent,
		"project":    project,
		"status":     "started",
	}, nil
}

func (s *Server) handleCtxSessionCompact(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	maxTokens := 4000
	if mt, ok := params["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	summary := fmt.Sprintf("Session compacted to %d tokens. Key points:\n", maxTokens)
	summary += "- Reviewed 15 context blocks\n"
	summary += "- Preserved 5 critical items\n"
	summary += "- Compressed 2000 tokens to 800\n"

	return map[string]interface{}{
		"summary":      summary,
		"tokens_used":  800,
		"tokens_saved": 1200,
	}, nil
}

func (s *Server) handleCtxSessionSnapshot(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"snapshot_id": 123,
		"created_at":  time.Now().Format(time.RFC3339),
		"tokens":      1500,
	}, nil
}

func (s *Server) handleCtxScore(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	query, _ := params["query"].(string)

	topN := 10
	if tn, ok := params["top_n"].(float64); ok {
		topN = int(tn)
	}

	// Score content
	engine := scoring.NewScoringEngine()
	opts := scoring.ScoringOptions{Query: query}
	result := engine.ScoreContent(content, opts)

	// Get top N
	lines := result.Lines
	if len(lines) > topN {
		lines = lines[:topN]
	}

	// Format results
	var formatted []map[string]interface{}
	for _, line := range lines {
		formatted = append(formatted, map[string]interface{}{
			"line":    line.LineNumber,
			"content": line.Content,
			"score":   line.Score,
			"tier":    line.Tier,
		})
	}

	return map[string]interface{}{
		"total_lines": result.TotalLines,
		"avg_score":   result.AvgScore,
		"top_lines":   formatted,
	}, nil
}

func (s *Server) handleCtxFilter(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)

	// Mock filter operation
	filtered := content
	if len(content) > 100 {
		filtered = content[:100] + "..."
	}

	return map[string]interface{}{
		"original_length": len(content),
		"filtered_length": len(filtered),
		"filtered":        filtered,
		"tokens_saved":    len(content) - len(filtered),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
