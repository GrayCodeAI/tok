package archive

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// PipelineIntegration provides integration between the filter pipeline and archive system
type PipelineIntegration struct {
	manager *ArchiveManager
	enabled bool
	mu      sync.RWMutex
}

// NewPipelineIntegration creates a new pipeline integration
func NewPipelineIntegration(manager *ArchiveManager) *PipelineIntegration {
	return &PipelineIntegration{
		manager: manager,
		enabled: true,
	}
}

// Enabled returns whether archiving is enabled
func (pi *PipelineIntegration) Enabled() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.enabled
}

// SetEnabled enables or disables archiving
func (pi *PipelineIntegration) SetEnabled(enabled bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.enabled = enabled
}

// ArchiveFiltered archives the filtered output from the pipeline
// This should be called after filtering is complete
func (pi *PipelineIntegration) ArchiveFiltered(ctx context.Context, original []byte, filtered []byte, command string, opts ArchiveOptions) (string, error) {
	if !pi.Enabled() {
		return "", nil
	}

	if pi.manager == nil {
		return "", fmt.Errorf("archive manager not initialized")
	}

	// Create archive entry
	entry := NewArchiveEntry(original, command).
		WithFiltered(filtered)

	// Apply options
	if opts.Category != "" {
		entry.WithCategory(opts.Category)
	}
	if opts.Agent != "" {
		entry.WithAgent(opts.Agent)
	}
	if opts.WorkingDir != "" {
		entry.WithWorkingDirectory(opts.WorkingDir)
	}
	if opts.ProjectPath != "" {
		entry.ProjectPath = opts.ProjectPath
	}
	if len(opts.Tags) > 0 {
		entry.WithTags(opts.Tags...)
	}
	if opts.Expiration > 0 {
		entry.WithExpiration(opts.Expiration)
	}

	// Archive it
	hash, err := pi.manager.Archive(ctx, entry)
	if err != nil {
		return "", fmt.Errorf("failed to archive: %w", err)
	}

	slog.Debug("archived filtered content",
		"hash", hash[:8],
		"command", command,
		"original_size", len(original),
		"filtered_size", len(filtered),
	)

	return hash, nil
}

// ArchiveOptions contains options for archiving
type ArchiveOptions struct {
	Category    ArchiveCategory
	Agent       string
	WorkingDir  string
	ProjectPath string
	Tags        []string
	Expiration  time.Duration
	Async       bool
}

// DefaultArchiveOptions returns default archive options
func DefaultArchiveOptions() ArchiveOptions {
	return ArchiveOptions{
		Category:   CategoryCommand,
		Expiration: 90 * 24 * time.Hour, // 90 days default
	}
}

// PipelineArchiveMiddleware wraps a filter pipeline with archiving capability
type PipelineArchiveMiddleware struct {
	pipeline filter.Pipeline
	archive  *PipelineIntegration
	opts     ArchiveOptions
}

// NewPipelineArchiveMiddleware creates a new archiving middleware
func NewPipelineArchiveMiddleware(pipeline filter.Pipeline, archive *PipelineIntegration, opts ArchiveOptions) *PipelineArchiveMiddleware {
	return &PipelineArchiveMiddleware{
		pipeline: pipeline,
		archive:  archive,
		opts:     opts,
	}
}

// Process runs the pipeline and archives the result
func (pam *PipelineArchiveMiddleware) Process(ctx context.Context, input string, command string) (string, *filter.PipelineStats, string, error) {
	// Run the pipeline
	output, stats := pam.pipeline.Process(input)

	// Archive if enabled
	var archiveHash string
	if pam.archive.Enabled() {
		hash, err := pam.archive.ArchiveFiltered(ctx, []byte(input), []byte(output), command, pam.opts)
		if err != nil {
			// Log error but don't fail the pipeline
			slog.Error("failed to archive pipeline output", "error", err)
		} else {
			archiveHash = hash
		}
	}

	return output, stats, archiveHash, nil
}

// ArchiveHook is a hook that can be registered with the pipeline
type ArchiveHook struct {
	integration *PipelineIntegration
	opts        ArchiveOptions
}

// NewArchiveHook creates a new archive hook
func NewArchiveHook(integration *PipelineIntegration, opts ArchiveOptions) *ArchiveHook {
	return &ArchiveHook{
		integration: integration,
		opts:        opts,
	}
}

// BeforeFiltering is called before filtering starts
func (ah *ArchiveHook) BeforeFiltering(command string, input []byte) {
	// Could log or track here
}

// AfterFiltering is called after filtering completes
func (ah *ArchiveHook) AfterFiltering(ctx context.Context, command string, original []byte, filtered []byte) (string, error) {
	if ah.integration == nil || !ah.integration.Enabled() {
		return "", nil
	}

	return ah.integration.ArchiveFiltered(ctx, original, filtered, command, ah.opts)
}

// GlobalPipelineIntegration is the global instance for use across the application
var globalPipelineIntegration *PipelineIntegration
var globalPiOnce sync.Once

// InitGlobalPipelineIntegration initializes the global pipeline integration
func InitGlobalPipelineIntegration(manager *ArchiveManager) {
	globalPiOnce.Do(func() {
		globalPipelineIntegration = NewPipelineIntegration(manager)
	})
}

// GetGlobalPipelineIntegration returns the global pipeline integration
func GetGlobalPipelineIntegration() *PipelineIntegration {
	return globalPipelineIntegration
}

// ArchivePipelineOutput is a convenience function to archive pipeline output
func ArchivePipelineOutput(ctx context.Context, original []byte, filtered []byte, command string, opts ArchiveOptions) (string, error) {
	if globalPipelineIntegration == nil {
		return "", fmt.Errorf("global pipeline integration not initialized")
	}
	return globalPipelineIntegration.ArchiveFiltered(ctx, original, filtered, command, opts)
}

// ArchiveConfig contains configuration for pipeline archiving
type ArchiveConfig struct {
	Enabled           bool            `mapstructure:"enabled"`
	Category          ArchiveCategory `mapstructure:"category"`
	Expiration        time.Duration   `mapstructure:"expiration"`
	Tags              []string        `mapstructure:"tags"`
	ExcludeCmds       []string        `mapstructure:"exclude_commands"`
	MinSize           int64           `mapstructure:"min_size"`           // Minimum content size to archive
	MaxSize           int64           `mapstructure:"max_size"`           // Maximum content size to archive
	Async             bool            `mapstructure:"async"`              // Archive asynchronously
	EnableCompression bool            `mapstructure:"enable_compression"` // Enable Brotli compression
}

// DefaultArchiveConfig returns default archive configuration
func DefaultArchiveConfig() ArchiveConfig {
	return ArchiveConfig{
		Enabled:           true,
		Category:          CategoryCommand,
		Expiration:        90 * 24 * time.Hour,
		Tags:              []string{},
		ExcludeCmds:       []string{"tokman archive", "tokman retrieve"},
		MinSize:           100,               // Don't archive tiny outputs
		MaxSize:           100 * 1024 * 1024, // 100MB max
		Async:             true,
		EnableCompression: true, // Enable Brotli compression by default
	}
}

// ShouldArchive checks if the given command and content should be archived
func (ac *ArchiveConfig) ShouldArchive(command string, contentSize int64) bool {
	if !ac.Enabled {
		return false
	}

	// Check excluded commands
	for _, exclude := range ac.ExcludeCmds {
		if command == exclude {
			return false
		}
	}

	// Check size constraints
	if contentSize < ac.MinSize {
		return false
	}
	if contentSize > ac.MaxSize {
		return false
	}

	return true
}

// ToOptions converts config to options
func (ac *ArchiveConfig) ToOptions() ArchiveOptions {
	return ArchiveOptions{
		Category:   ac.Category,
		Expiration: ac.Expiration,
		Tags:       ac.Tags,
		Async:      ac.Async,
	}
}

// AsyncArchiveResult contains the result of an async archive operation
type AsyncArchiveResult struct {
	Hash  string
	Error error
}

// ArchiveAsync performs archiving asynchronously
func (pi *PipelineIntegration) ArchiveAsync(ctx context.Context, original []byte, filtered []byte, command string, opts ArchiveOptions) <-chan AsyncArchiveResult {
	result := make(chan AsyncArchiveResult, 1)

	go func() {
		defer close(result)

		hash, err := pi.ArchiveFiltered(ctx, original, filtered, command, opts)
		result <- AsyncArchiveResult{
			Hash:  hash,
			Error: err,
		}
	}()

	return result
}

// PipelineStatsWithArchive extends PipelineStats with archive information
type PipelineStatsWithArchive struct {
	*filter.PipelineStats
	ArchiveHash    string        `json:"archive_hash,omitempty"`
	Archived       bool          `json:"archived"`
	ArchiveError   string        `json:"archive_error,omitempty"`
	ArchiveLatency time.Duration `json:"archive_latency_ms,omitempty"`
}

// EnrichStats adds archive information to pipeline stats
func EnrichStats(stats *filter.PipelineStats, hash string, err error, latency time.Duration) *PipelineStatsWithArchive {
	enriched := &PipelineStatsWithArchive{
		PipelineStats:  stats,
		ArchiveHash:    hash,
		Archived:       hash != "",
		ArchiveLatency: latency,
	}

	if err != nil {
		enriched.ArchiveError = err.Error()
	}

	return enriched
}
