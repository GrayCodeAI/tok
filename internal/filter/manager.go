package filter

import (
	"container/list"
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

// CommandContext is now imported from config package to avoid duplication.
// Use config.CommandContext instead of defining it here.

// PipelineManager handles resilient large-context processing.
// Supports streaming for inputs up to 2M tokens with automatic
// chunking, validation, and failure recovery.
type PipelineManager struct {
	config      ManagerConfig
	coordinator *PipelineCoordinator
	cache       *CompressionCache
	lruCache    *LRUCache // LRU cache for better eviction
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
	m.coordinator = NewPipelineCoordinator(cfg.PipelineCfg)

	// Initialize cache
	if cfg.CacheEnabled {
		m.cache = NewCompressionCache(cfg.CacheMaxSize)
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
	result := &ProcessResult{
		LayerStats: make(map[string]LayerStat),
	}

	// Validate context size
	tokens := EstimateTokens(input)
	result.OriginalTokens = tokens

	m.mu.RLock()
	maxCtx := m.config.MaxContextTokens
	streamThreshold := m.config.StreamThreshold
	m.mu.RUnlock()

	if tokens > maxCtx {
		return nil, fmt.Errorf("input exceeds max context tokens (%d > %d)", tokens, maxCtx)
	}

	// Check cache (T101: LRU cache with TTL)
	cacheKey := m.cacheKey(input, mode, ctx)

	// Try LRU cache first (faster, TTL-aware)
	if m.lruCache != nil {
		if val, ok := m.lruCache.Get(cacheKey); ok && val != nil {
			cached := val.(*CachedResult)
			result.Output = cached.Output
			result.FinalTokens = EstimateTokens(result.Output)
			result.SavedTokens = result.OriginalTokens - result.FinalTokens
			result.CacheHit = true
			return result, nil
		}
	}

	// Fall back to legacy cache
	if m.cache != nil {
		if cached, ok := m.cache.Get(cacheKey); ok {
			result.Output = cached.Output
			result.FinalTokens = EstimateTokens(result.Output)
			result.SavedTokens = result.OriginalTokens - result.FinalTokens
			result.CacheHit = true
			return result, nil
		}
	}

	// Choose processing strategy based on size
	if tokens > streamThreshold {
		return m.processStreaming(input, mode, ctx, result)
	}

	return m.processSingle(input, mode, ctx, result)
}

// processSingle processes input in a single pass
func (m *PipelineManager) processSingle(input string, mode Mode, ctx config.CommandContext, result *ProcessResult) (*ProcessResult, error) {
	// Set query intent and process under write lock to prevent races on
	// coordinator.config between concurrent goroutines.
	m.mu.Lock()
	m.syncCoordinatorForRequest(mode, ctx.Intent)
	output, stats := m.coordinator.Process(input)
	m.mu.Unlock()

	m.mu.RLock()
	validateOutput := m.config.ValidateOutput
	failSafeMode := m.config.FailSafeMode
	teeOnFailure := m.config.TeeOnFailure
	m.mu.RUnlock()

	// Validate output
	if validateOutput {
		if !m.validateOutput(output, input, ctx) {
			// Output validation failed
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

	// Check for empty output (failure)
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

	// Copy stats
	result.FinalTokens = stats.FinalTokens
	result.SavedTokens = stats.TotalSaved
	result.ReductionPercent = stats.ReductionPercent
	result.LayerStats = stats.LayerStats

	// Cache result - prefer LRU cache (has TTL eviction), fall back to legacy
	if !result.CacheHit {
		cacheKey := m.cacheKey(input, mode, ctx)
		cached := &CachedResult{
			Output:   result.Output,
			Tokens:   result.FinalTokens,
			CachedAt: time.Now(),
		}
		if m.lruCache != nil {
			m.lruCache.Set(cacheKey, cached)
		} else if m.cache != nil {
			m.cache.Set(cacheKey, cached)
		}
	}

	return result, nil
}

func (m *PipelineManager) syncCoordinatorForRequest(mode Mode, query string) {
	m.coordinator.config.Mode = mode
	if m.coordinator.config.QueryIntent == query {
		return
	}

	m.coordinator.config.QueryIntent = query

	if query == "" {
		m.coordinator.goalDrivenFilter = nil
		m.coordinator.contrastiveFilter = nil
		m.coordinator.buildLayers()
		return
	}

	m.coordinator.goalDrivenFilter = NewGoalDrivenFilter(query)
	m.coordinator.contrastiveFilter = NewContrastiveFilter(query)
	m.coordinator.buildLayers()
}

// processStreaming processes large input in chunks
func (m *PipelineManager) processStreaming(input string, mode Mode, ctx config.CommandContext, result *ProcessResult) (*ProcessResult, error) {
	m.mu.RLock()
	chunkSize := m.config.ChunkSize
	failSafeMode := m.config.FailSafeMode
	shortCircuit := m.config.ShortCircuitBudget
	budget := m.coordinator.config.Budget
	m.mu.RUnlock()

	// Split into processable chunks
	chunks := m.chunkInput(input, chunkSize)
	result.Chunks = len(chunks)

	var processedChunks []string
	totalSaved := 0

	for i, chunk := range chunks {
		chunkResult, err := m.processSingle(chunk, mode, ctx, &ProcessResult{
			LayerStats: make(map[string]LayerStat),
		})
		if err != nil {
			// Handle chunk failure
			if failSafeMode {
				processedChunks = append(processedChunks, chunk)
				continue
			}
			return nil, fmt.Errorf("chunk %d failed: %w", i, err)
		}

		processedChunks = append(processedChunks, chunkResult.Output)
		totalSaved += chunkResult.SavedTokens

		// Short-circuit if budget met
		if shortCircuit && budget > 0 {
			currentTokens := EstimateTokens(strings.Join(processedChunks, "\n"))
			if currentTokens <= budget {
				break
			}
		}
	}

	// Combine chunks with minimal delimiter to avoid token inflation
	result.Output = strings.Join(processedChunks, "\n---\n")
	result.FinalTokens = EstimateTokens(result.Output)

	// Safely calculate saved tokens with overflow protection
	if result.OriginalTokens >= result.FinalTokens {
		result.SavedTokens = result.OriginalTokens - result.FinalTokens
	} else {
		result.SavedTokens = 0
	}

	if result.OriginalTokens > 0 {
		result.ReductionPercent = float64(result.SavedTokens) / float64(result.OriginalTokens) * 100
		// Clamp to valid range [0, 100]
		if result.ReductionPercent < 0 {
			result.ReductionPercent = 0
		} else if result.ReductionPercent > 100 {
			result.ReductionPercent = 100
		}
	}

	return result, nil
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

// checkStructure performs basic structural validation
func (m *PipelineManager) checkStructure(s string) bool {
	// Check balanced brackets
	parens := 0
	brackets := 0
	braces := 0

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

		// Allow some imbalance (code snippets, etc.)
		if parens < -10 || brackets < -10 || braces < -10 {
			return false
		}
	}

	// Allow moderate imbalance
	if parens > 50 || brackets > 50 || braces > 50 {
		return false
	}

	return true
}

// saveTee saves raw output to a file for recovery
func (m *PipelineManager) saveTee(input string, ctx config.CommandContext, reason string) string {
	// Validate input size (limit to 10MB)
	if len(input) > 10*1024*1024 {
		fmt.Fprintf(os.Stderr, "warning: tee input too large, skipping\n")
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
		return path
	}
	if err := os.WriteFile(path, content, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write %s: %v\n", path, err)
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

// CompressionCache provides caching for compression results with O(1) eviction
type CompressionCache struct {
	maxSize int
	entries map[string]*cacheEntry
	order   *list.List
	mu      sync.RWMutex
}

type cacheEntry struct {
	result  *CachedResult
	element *list.Element
}

// CachedResult represents a cached compression result
type CachedResult struct {
	Output   string
	Tokens   int
	CachedAt time.Time
}

// NewCompressionCache creates a new compression cache
func NewCompressionCache(maxSize int) *CompressionCache {
	return &CompressionCache{
		maxSize: maxSize,
		entries: make(map[string]*cacheEntry),
		order:   list.New(),
	}
}

// Get retrieves a cached result
func (c *CompressionCache) Get(key string) (*CachedResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return entry.result, ok
}

// Set stores a result in cache with O(1) eviction
func (c *CompressionCache) Set(key string, result *CachedResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict old entries if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	element := c.order.PushBack(key)
	c.entries[key] = &cacheEntry{
		result:  result,
		element: element,
	}
}

// evictOldest removes the oldest cache entry in O(1)
func (c *CompressionCache) evictOldest() {
	front := c.order.Front()
	if front == nil {
		return
	}

	key := front.Value.(string)
	c.order.Remove(front)
	delete(c.entries, key)
}

// Size returns the number of cached entries
func (c *CompressionCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// ProcessWithBudget processes with a specific token budget.
// NOTE: Sets the coordinator budget and calls Process sequentially.
// In tok's CLI context, each invocation is isolated per process,
// so concurrent budget races are not a practical concern.
func (m *PipelineManager) ProcessWithBudget(input string, mode Mode, budget int, ctx config.CommandContext) (*ProcessResult, error) {
	m.mu.Lock()
	m.coordinator.config.Budget = budget
	if m.coordinator.budgetEnforcer == nil {
		m.coordinator.budgetEnforcer = NewBudgetEnforcer(budget)
	} else {
		m.coordinator.budgetEnforcer.SetBudget(budget)
	}
	m.mu.Unlock()

	return m.Process(input, mode, ctx)
}

// ProcessWithQuery processes with query-aware compression
func (m *PipelineManager) ProcessWithQuery(input string, mode Mode, query string, ctx config.CommandContext) (*ProcessResult, error) {
	m.mu.Lock()
	ctx.Intent = query
	m.syncCoordinatorForRequest(mode, query)
	m.mu.Unlock()

	return m.Process(input, mode, ctx)
}
