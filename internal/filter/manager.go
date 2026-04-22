package filter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GrayCodeAI/tok/internal/config"
)

// maxTeeInputSize limits the size of input saved to tee files (10 MiB).
const maxTeeInputSize = 10 * 1024 * 1024

// CommandContext is now imported from config package to avoid duplication.
// Use config.CommandContext instead of defining it here.

// PipelineManager handles resilient large-context processing.
// Supports streaming for inputs up to 2M tokens with automatic
// chunking, validation, and failure recovery.
type PipelineManager struct {
	config      ManagerConfig
	coordinator *PipelineCoordinator
	lruCache    *LRUCache // TTL-aware LRU cache for compression results
	teeDir      string
	mu          sync.RWMutex
}

// ManagerConfig configures the pipeline manager
type ManagerConfig struct {
	// Context limits
	MaxContextTokens int
	ChunkSize        int
	StreamThreshold  int

	// Resilience
	TeeOnFailure       bool
	FailSafeMode       bool
	ValidateOutput     bool
	ShortCircuitBudget bool

	// Performance
	CacheEnabled bool
	CacheMaxSize int

	// Layer config
	PipelineCfg PipelineConfig
}

// NewPipelineManager creates a new pipeline manager
func NewPipelineManager(cfg ManagerConfig) *PipelineManager {
	m := &PipelineManager{
		config: cfg,
	}

	// Create coordinator with pipeline config
	m.coordinator = NewPipelineCoordinator(&cfg.PipelineCfg)

	// Initialize cache
	if cfg.CacheEnabled {
		m.lruCache = NewLRUCache(cfg.CacheMaxSize, 5*time.Minute) // T101
	}

	// Set tee directory
	if cfg.TeeOnFailure {
		m.teeDir = os.TempDir()
	}

	return m
}

// ProcessResult contains the result of processing
type ProcessResult struct {
	Output           string
	OriginalTokens   int
	FinalTokens      int
	SavedTokens      int
	ReductionPercent float64
	LayerStats       map[string]LayerStat
	CacheHit         bool
	Chunks           int
	Validated        bool
	TeeFile          string // If failure occurred
	Warning          string
}

// Process processes input with full resilience and large context support.
// For inputs > StreamThreshold, uses streaming chunk processing.
func (m *PipelineManager) Process(input string, mode Mode, ctx config.CommandContext) (*ProcessResult, error) {
	coord := m.coordinatorForRequest(mode, ctx.Intent)
	return m.processWithCoordinator(input, mode, ctx, coord)
}

// coordinatorForRequest builds a fresh coordinator for the given mode and query.
// The returned coordinator is independent and safe to use without locking.
func (m *PipelineManager) coordinatorForRequest(mode Mode, query string) *PipelineCoordinator {
	m.mu.RLock()
	cfg := m.config.PipelineCfg
	m.mu.RUnlock()

	cfg.Mode = mode
	cfg.QueryIntent = query
	return NewPipelineCoordinator(&cfg)
}

// chunkInput splits large input into processable chunks
func (m *PipelineManager) chunkInput(input string, maxTokens int) []string {
	tokens := EstimateTokens(input)
	if tokens <= maxTokens {
		return []string{input}
	}

	// Split by logical boundaries
	lines := strings.Split(input, "\n")
	var chunks []string
	var currentChunk []string
	currentTokens := 0

	for _, line := range lines {
		lineTokens := EstimateTokens(line)

		// Check if adding this line exceeds chunk size
		if currentTokens+lineTokens > maxTokens && len(currentChunk) > 0 {
			chunks = append(chunks, strings.Join(currentChunk, "\n"))
			currentChunk = nil
			currentTokens = 0
		}

		currentChunk = append(currentChunk, line)
		currentTokens += lineTokens
	}

	// Add remaining chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, "\n"))
	}

	return chunks
}

// validateOutput checks if output is valid
func (m *PipelineManager) validateOutput(output, original string, ctx config.CommandContext) bool {
	// Check for empty output when original was not empty
	if output == "" && original != "" {
		return false
	}

	// Check for reasonable compression (not more than 99% unless aggressive)
	if len(original) == 0 {
		return output == ""
	}
	compressionRatio := float64(len(output)) / float64(len(original))
	if compressionRatio < 0.01 {
		// Suspicious - output is less than 1% of original
		// This might indicate corruption
		return false
	}

	// Check that important content is preserved
	if ctx.IsError || ctx.ExitCode != 0 {
		// For error output, ensure error markers are preserved
		if strings.Contains(original, "error") && !strings.Contains(strings.ToLower(output), "error") {
			return false
		}
		if strings.Contains(original, "Error:") && !strings.Contains(output, "Error:") {
			return false
		}
	}

	// Check for structural integrity (balanced brackets, etc.)
	if !m.checkStructure(output) {
		return false
	}

	return true
}

// looksLikeJSON detects JSON-like content more accurately than
// checking for individual characters (which produces false positives on prose).
func looksLikeJSON(s string) bool {
	// Find first non-whitespace character
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	if i >= len(s) {
		return false
	}
	first := s[i]
	if first != '{' && first != '[' {
		return false
	}

	// Arrays: starting with '[' are treated as JSON
	if first == '[' {
		return true
	}

	// Objects: must contain both quotes and colons to be key-value JSON
	hasQuotes := false
	hasColons := false
	for _, c := range s {
		if c == '"' {
			hasQuotes = true
		}
		if c == ':' {
			hasColons = true
		}
		if hasQuotes && hasColons {
			return true
		}
	}
	return false
}

// checkStructure performs basic structural validation.
// For JSON-like content, requires exact brace balance. For general text,
// allows moderate imbalance (code snippets may be truncated).
func (m *PipelineManager) checkStructure(s string) bool {
	parens := 0
	brackets := 0
	braces := 0
	isJSON := looksLikeJSON(s)

	// Quick scan to count brackets
	for _, c := range s {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			brackets++
		case ']':
			brackets--
		case '{':
			braces++
		case '}':
			braces--
		}

		// Hard fail on severe negative imbalance
		if parens < -10 || brackets < -10 || braces < -10 {
			return false
		}
	}

	if isJSON {
		// JSON requires exact balance
		if parens != 0 || brackets != 0 || braces != 0 {
			return false
		}
	} else {
		// General text: allow moderate imbalance
		if parens > 50 || brackets > 50 || braces > 50 {
			return false
		}
	}

	return true
}

// saveTee saves raw output to a file for recovery
func (m *PipelineManager) saveTee(input string, ctx config.CommandContext, reason string) string {
	// Validate input size
	if len(input) > maxTeeInputSize {
		fmt.Fprintf(os.Stderr, "warning: tee input too large (%d > %d), skipping\n", len(input), maxTeeInputSize)
		return ""
	}

	timestamp := time.Now().Format("20060102-150405")
	// Sanitize command to prevent path traversal in filename
	safeCommand := strings.NewReplacer("/", "_", "\\", "_", "..", "_", " ", "_").Replace(ctx.Command)
	filename := fmt.Sprintf("tok-tee-%s-%s-%s.txt", timestamp, safeCommand, reason)
	path := filepath.Join(m.teeDir, filename)

	// Validate the resolved path is within tee directory (resolve symlinks)
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: invalid tee path, skipping\n")
		return ""
	}
	safeDir, err := filepath.EvalSymlinks(m.teeDir)
	if err != nil {
		safeDir = m.teeDir
	}
	safePath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		safePath = absPath
	}
	if !strings.HasPrefix(safePath, safeDir+string(filepath.Separator)) {
		fmt.Fprintf(os.Stderr, "warning: invalid tee path, skipping\n")
		return ""
	}

	data := struct {
		Timestamp  time.Time
		Command    string
		Subcommand string
		Reason     string
		Input      string
		Context    config.CommandContext
	}{
		Timestamp:  time.Now(),
		Command:    ctx.Command,
		Subcommand: ctx.Subcommand,
		Reason:     reason,
		Input:      input,
		Context:    ctx,
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to marshal tee data: %v\n", err)
		return ""
	}
	if err := os.WriteFile(path, content, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write %s: %v\n", path, err)
		return ""
	}

	return path
}

// cacheKey generates a cache key for the input
func (m *PipelineManager) cacheKey(input string, mode Mode, ctx config.CommandContext) string {
	m.mu.RLock()
	budget := m.coordinator.config.Budget
	m.mu.RUnlock()

	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%s-%s-%s-%s-%d",
		hex.EncodeToString(hash[:]),
		mode,
		ctx.Command,
		ctx.Intent,
		budget,
	)
}

// CachedResult represents a cached compression result
type CachedResult struct {
	Output   string
	Tokens   int
	CachedAt time.Time
}

// ProcessWithBudget processes with a specific token budget.
func (m *PipelineManager) ProcessWithBudget(input string, mode Mode, budget int, ctx config.CommandContext) (*ProcessResult, error) {
	m.mu.RLock()
	cfg := m.config.PipelineCfg
	m.mu.RUnlock()

	cfg.Mode = mode
	cfg.Budget = budget
	coord := NewPipelineCoordinator(&cfg)

	return m.processWithCoordinator(input, mode, ctx, coord)
}

// ProcessWithQuery processes with query-aware compression.
func (m *PipelineManager) ProcessWithQuery(input string, mode Mode, query string, ctx config.CommandContext) (*ProcessResult, error) {
	ctx.Intent = query
	coord := m.coordinatorForRequest(mode, query)
	return m.processWithCoordinator(input, mode, ctx, coord)
}

// processWithCoordinator runs the pipeline using the provided coordinator.
// The coordinator must be fully configured before calling this method.
func (m *PipelineManager) processWithCoordinator(input string, mode Mode, ctx config.CommandContext, coord *PipelineCoordinator) (*ProcessResult, error) {
	result := &ProcessResult{
		LayerStats: make(map[string]LayerStat),
	}

	tokens := EstimateTokens(input)
	result.OriginalTokens = tokens

	m.mu.RLock()
	maxCtx := m.config.MaxContextTokens
	streamThreshold := m.config.StreamThreshold
	m.mu.RUnlock()

	if tokens > maxCtx {
		return nil, fmt.Errorf("input exceeds max context tokens (%d > %d)", tokens, maxCtx)
	}

	// Check cache
	cacheKey := m.cacheKey(input, mode, ctx)
	if m.lruCache != nil {
		if val, ok := m.lruCache.Get(cacheKey); ok && val != nil {
			cached, typeOk := val.(*CachedResult)
			if typeOk {
				result.Output = cached.Output
				result.FinalTokens = EstimateTokens(result.Output)
				result.SavedTokens = result.OriginalTokens - result.FinalTokens
				result.CacheHit = true
				return result, nil
			}
			// Corrupted cache entry: treat as miss
		}
	}

	if tokens > streamThreshold {
		return m.processStreamingWithCoordinator(input, mode, ctx, result, coord)
	}

	return m.processSingleWithCoordinator(input, ctx, result, coord)
}

// processSingleWithCoordinator processes input in a single pass with a pre-built coordinator.
func (m *PipelineManager) processSingleWithCoordinator(input string, ctx config.CommandContext, result *ProcessResult, coord *PipelineCoordinator) (*ProcessResult, error) {
	output, stats := coord.Process(input)

	m.mu.RLock()
	validateOutput := m.config.ValidateOutput
	failSafeMode := m.config.FailSafeMode
	teeOnFailure := m.config.TeeOnFailure
	m.mu.RUnlock()

	if validateOutput {
		if !m.validateOutput(output, input, ctx) {
			if failSafeMode {
				result.Output = input
				result.Warning = "output validation failed, returning original"
			} else {
				result.Output = output
				result.Warning = "output may be corrupted"
			}
		} else {
			result.Output = output
			result.Validated = true
		}
	} else {
		result.Output = output
	}

	if result.Output == "" && input != "" {
		if teeOnFailure {
			teeFile := m.saveTee(input, ctx, "empty_output")
			result.TeeFile = teeFile
			result.Warning = "pipeline produced empty output, original saved to tee file"
		}
		if failSafeMode {
			result.Output = input
		}
	}

	result.FinalTokens = stats.FinalTokens
	result.SavedTokens = stats.TotalSaved
	result.ReductionPercent = stats.ReductionPercent
	result.LayerStats = stats.LayerStats

	if !result.CacheHit && m.lruCache != nil {
		cacheKey := m.cacheKey(input, coord.config.Mode, ctx)
		cached := &CachedResult{
			Output:   result.Output,
			Tokens:   result.FinalTokens,
			CachedAt: time.Now(),
		}
		m.lruCache.Set(cacheKey, cached)
	}

	return result, nil
}

// processStreamingWithCoordinator processes large input in chunks using a pre-built coordinator.
func (m *PipelineManager) processStreamingWithCoordinator(input string, mode Mode, ctx config.CommandContext, result *ProcessResult, coord *PipelineCoordinator) (*ProcessResult, error) {
	m.mu.RLock()
	chunkSize := m.config.ChunkSize
	failSafeMode := m.config.FailSafeMode
	shortCircuit := m.config.ShortCircuitBudget
	budget := coord.config.Budget
	m.mu.RUnlock()

	chunks := m.chunkInput(input, chunkSize)
	result.Chunks = len(chunks)

	var processedChunks []string
	totalSaved := 0

	for i, chunk := range chunks {
		chunkResult, err := m.processSingleWithCoordinator(chunk, ctx, &ProcessResult{
			LayerStats: make(map[string]LayerStat),
		}, coord)
		if err != nil {
			if failSafeMode {
				processedChunks = append(processedChunks, chunk)
				continue
			}
			return nil, fmt.Errorf("chunk %d failed: %w", i, err)
		}

		processedChunks = append(processedChunks, chunkResult.Output)
		totalSaved += chunkResult.SavedTokens

		if shortCircuit && budget > 0 {
			currentTokens := EstimateTokens(strings.Join(processedChunks, "\n"))
			if currentTokens <= budget {
				break
			}
		}
	}

	result.Output = strings.Join(processedChunks, "\n---\n")
	result.FinalTokens = EstimateTokens(result.Output)

	if result.OriginalTokens >= result.FinalTokens {
		result.SavedTokens = result.OriginalTokens - result.FinalTokens
	} else {
		result.SavedTokens = 0
	}

	if result.OriginalTokens > 0 {
		result.ReductionPercent = float64(result.SavedTokens) / float64(result.OriginalTokens) * 100
		if result.ReductionPercent < 0 {
			result.ReductionPercent = 0
		} else if result.ReductionPercent > 100 {
			result.ReductionPercent = 100
		}
	}

	return result, nil
}
