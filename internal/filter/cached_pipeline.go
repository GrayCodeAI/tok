package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lakshmanpatel/tok/internal/cache"
)

// CachedPipeline wraps a PipelineCoordinator with caching
type CachedPipeline struct {
	pipeline   *PipelineCoordinator
	cache      *cache.QueryCache
	gitWatcher *cache.GitWatcher
	enabled    bool
}

// CachedPipelineConfig for creating a cached pipeline
type CachedPipelineConfig struct {
	Pipeline   *PipelineCoordinator
	Cache      *cache.QueryCache
	GitWatcher *cache.GitWatcher
	Enabled    bool
}

// NewCachedPipeline creates a pipeline with caching support
func NewCachedPipeline(cfg CachedPipelineConfig) *CachedPipeline {
	return &CachedPipeline{
		pipeline:   cfg.Pipeline,
		cache:      cfg.Cache,
		gitWatcher: cfg.GitWatcher,
		enabled:    cfg.Enabled,
	}
}

// CachedProcessResult contains the result of a cached process
type CachedProcessResult struct {
	Output         string
	OriginalOutput string
	Stats          *PipelineStats
	FromCache      bool
	CacheKey       string
	CacheHitCount  int
}

// Process executes command with caching
func (cp *CachedPipeline) Process(command string, args []string, getOutput func() (string, error)) (*CachedProcessResult, error) {
	if !cp.enabled || cp.cache == nil {
		// Cache disabled, just run normally
		output, err := getOutput()
		if err != nil {
			return nil, err
		}

		filtered, stats := cp.pipeline.Process(output)
		return &CachedProcessResult{
			Output:         filtered,
			OriginalOutput: output,
			Stats:          stats,
			FromCache:      false,
		}, nil
	}

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// Check for git changes and invalidate if needed
	if cp.gitWatcher != nil && cache.IsGitRepo(workingDir) {
		_ = cp.gitWatcher.InvalidateChanged(workingDir)
	}

	// Get relevant files for cache key
	relevantFiles := cp.detectRelevantFiles(workingDir, command, args)

	// Get file hashes
	var fileHashes map[string]string
	if cp.gitWatcher != nil {
		fileHashes, _ = cp.gitWatcher.GetFileHashes(workingDir, relevantFiles)
	}

	// Generate cache key
	cacheKey := cache.GenerateKey(command, args, workingDir, fileHashes)

	// Try to get from cache
	if entry, found := cp.cache.Get(cacheKey); found {
		return &CachedProcessResult{
			Output:         entry.FilteredOutput,
			OriginalOutput: "", // Not stored to save space
			Stats: &PipelineStats{
				OriginalTokens:   entry.OriginalTokens,
				FinalTokens:      entry.FilteredTokens,
				TotalSaved:       entry.OriginalTokens - entry.FilteredTokens,
				ReductionPercent: entry.CompressionRatio * 100,
				CacheHit:         true,
			},
			FromCache:     true,
			CacheKey:      cacheKey,
			CacheHitCount: entry.HitCount,
		}, nil
	}

	// Cache miss - execute command
	output, err := getOutput()
	if err != nil {
		return nil, err
	}

	// Filter output
	filtered, stats := cp.pipeline.Process(output)

	// Store in cache
	_ = cp.cache.Set(
		cacheKey,
		command,
		args,
		workingDir,
		fileHashes,
		filtered,
		stats.OriginalTokens,
		stats.FinalTokens,
	)

	return &CachedProcessResult{
		Output:         filtered,
		OriginalOutput: output,
		Stats:          stats,
		FromCache:      false,
		CacheKey:       cacheKey,
	}, nil
}

// detectRelevantFiles detects files relevant to a command
// This is a simple heuristic - can be improved
func (cp *CachedPipeline) detectRelevantFiles(workingDir string, command string, args []string) []string {
	var files []string

	// Check args for file paths
	for _, arg := range args {
		// Skip flags
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// Check if it's a file
		path := filepath.Join(workingDir, arg)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files = append(files, arg)
			continue
		}

		// Check if it's a glob pattern
		if matches, err := filepath.Glob(path); err == nil {
			for _, match := range matches {
				rel, _ := filepath.Rel(workingDir, match)
				if rel != "" && rel != "." {
					files = append(files, rel)
				}
			}
		}
	}

	// Command-specific detection
	switch command {
	case "go":
		// Go commands often depend on go.mod and source files
		files = append(files, "go.mod", "go.sum")
	case "npm", "pnpm", "yarn":
		// Node commands depend on package.json
		files = append(files, "package.json", "package-lock.json", "pnpm-lock.yaml")
	case "docker":
		// Docker commands depend on Dockerfile
		files = append(files, "Dockerfile", "docker-compose.yml")
	}

	return uniqueStrings(files)
}

// uniqueStrings removes duplicates
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// Stats returns cache statistics
func (cp *CachedPipeline) Stats() (*cache.CacheStats, error) {
	if cp.cache == nil {
		return &cache.CacheStats{}, nil
	}
	return cp.cache.Stats()
}

// Close closes the cache
func (cp *CachedPipeline) Close() error {
	if cp.cache != nil {
		return cp.cache.Close()
	}
	return nil
}
