package archive

import (
	"encoding/json"
	"fmt"
	"time"
)

// ArchiveCategory represents the type of archived content
type ArchiveCategory string

const (
	CategoryCommand  ArchiveCategory = "command"
	CategorySession  ArchiveCategory = "session"
	CategoryUser     ArchiveCategory = "user"
	CategorySystem   ArchiveCategory = "system"
	CategoryProject  ArchiveCategory = "project"
	CategoryPipeline ArchiveCategory = "pipeline"
)

// CompressionType represents the compression algorithm used
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGzip   CompressionType = "gzip"
	CompressionBrotli CompressionType = "brotli"
	CompressionZstd   CompressionType = "zstd"
)

// ArchiveEntry represents a single archived content entry
type ArchiveEntry struct {
	// Core identification
	ID      int64  `json:"id" db:"id"`
	Hash    string `json:"hash" db:"hash"`
	Version int    `json:"version" db:"version"`

	// Content storage (stored as bytes in DB, may be compressed)
	OriginalContent []byte `json:"-" db:"original_content"`
	FilteredContent []byte `json:"-" db:"filtered_content"`

	// Content metadata
	OriginalSize   int64           `json:"original_size" db:"original_size"`
	CompressedSize int64           `json:"compressed_size" db:"compressed_size"`
	Compression    CompressionType `json:"compression" db:"compression_type"`

	// Context information
	Command          string          `json:"command" db:"command"`
	WorkingDirectory string          `json:"working_directory" db:"working_directory"`
	ProjectPath      string          `json:"project_path" db:"project_path"`
	Agent            string          `json:"agent" db:"agent"`
	Category         ArchiveCategory `json:"category" db:"category"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	AccessedAt *time.Time `json:"accessed_at" db:"accessed_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`

	// Access tracking
	AccessCount int `json:"access_count" db:"access_count"`

	// Categorization
	Tags []string `json:"tags" db:"-"`

	// Flexible metadata storage
	Metadata ArchiveMetadata `json:"metadata" db:"metadata"`
}

// ArchiveMetadata contains flexible metadata for an archive entry
type ArchiveMetadata struct {
	// User information
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`

	// System information
	Hostname string `json:"hostname,omitempty"`
	OS       string `json:"os,omitempty"`
	Shell    string `json:"shell,omitempty"`
	Terminal string `json:"terminal,omitempty"`

	// Git information
	GitBranch string `json:"git_branch,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
	GitRemote string `json:"git_remote,omitempty"`
	GitDirty  bool   `json:"git_dirty,omitempty"`

	// Environment
	Environment string `json:"environment,omitempty"` // dev, staging, prod
	CI          bool   `json:"ci,omitempty"`
	CIProvider  string `json:"ci_provider,omitempty"`

	// TokMan specific
	TokManVersion  string   `json:"tokman_version,omitempty"`
	PipelineUsed   []string `json:"pipeline_used,omitempty"`
	Compression    float64  `json:"compression_ratio,omitempty"`
	TokensSaved    int      `json:"tokens_saved,omitempty"`
	OriginalTokens int      `json:"original_tokens,omitempty"`
	FilteredTokens int      `json:"filtered_tokens,omitempty"`

	// Custom fields
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// ArchiveListOptions provides filtering and pagination for listing archives
type ArchiveListOptions struct {
	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Filtering
	Category    ArchiveCategory `json:"category,omitempty"`
	Agent       string          `json:"agent,omitempty"`
	ProjectPath string          `json:"project_path,omitempty"`
	Tags        []string        `json:"tags,omitempty"`

	// Time filtering
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// Sorting
	SortBy    string `json:"sort_by"`    // "created_at", "accessed_at", "size"
	SortOrder string `json:"sort_order"` // "asc", "desc"

	// Search
	Query string `json:"query,omitempty"`
}

// DefaultListOptions returns default options for listing
func DefaultListOptions() ArchiveListOptions {
	return ArchiveListOptions{
		Limit:     100,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// CompressionRatio returns the compression ratio for this entry
func (e *ArchiveEntry) CompressionRatio() float64 {
	if e.OriginalSize == 0 {
		return 1.0
	}
	return float64(e.CompressedSize) / float64(e.OriginalSize)
}

// SpaceSaved returns bytes saved by compression
func (e *ArchiveEntry) SpaceSaved() int64 {
	return e.OriginalSize - e.CompressedSize
}

// IsExpired checks if the archive has expired
func (e *ArchiveEntry) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

// MarkAccessed updates the access timestamp and count
func (e *ArchiveEntry) MarkAccessed() {
	now := time.Now()
	e.AccessedAt = &now
	e.AccessCount++
}

// Content returns the original content (decompressed if necessary)
// Note: This is a placeholder - actual decompression would be handled by the manager
func (e *ArchiveEntry) Content() []byte {
	return e.OriginalContent
}

// Filtered returns the filtered content (decompressed if necessary)
func (e *ArchiveEntry) Filtered() []byte {
	return e.FilteredContent
}

// AddTag adds a tag to the entry
func (e *ArchiveEntry) AddTag(tag string) {
	// Check if tag already exists
	for _, t := range e.Tags {
		if t == tag {
			return
		}
	}
	e.Tags = append(e.Tags, tag)
}

// RemoveTag removes a tag from the entry
func (e *ArchiveEntry) RemoveTag(tag string) {
	var newTags []string
	for _, t := range e.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	e.Tags = newTags
}

// HasTag checks if the entry has a specific tag
func (e *ArchiveEntry) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// MarshalMetadata converts metadata to JSON string for database storage
func (e *ArchiveEntry) MarshalMetadata() (string, error) {
	data, err := json.Marshal(e.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return string(data), nil
}

// UnmarshalMetadata parses JSON metadata from database
func (e *ArchiveEntry) UnmarshalMetadata(data string) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), &e.Metadata)
}

// Summary returns a short summary of the archive entry
func (e *ArchiveEntry) Summary() string {
	cmd := e.Command
	if len(cmd) > 50 {
		cmd = cmd[:47] + "..."
	}

	return fmt.Sprintf("Archive[%s]: %s (saved: %s, %d%%)",
		e.Hash[:8],
		cmd,
		formatBytes(e.SpaceSaved()),
		int((1-e.CompressionRatio())*100),
	)
}

// ToMap converts the entry to a map for flexible serialization
func (e *ArchiveEntry) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"hash":              e.Hash,
		"command":           e.Command,
		"working_directory": e.WorkingDirectory,
		"project_path":      e.ProjectPath,
		"agent":             e.Agent,
		"category":          e.Category,
		"original_size":     e.OriginalSize,
		"compressed_size":   e.CompressedSize,
		"compression_ratio": e.CompressionRatio(),
		"space_saved":       e.SpaceSaved(),
		"created_at":        e.CreatedAt,
		"accessed_at":       e.AccessedAt,
		"expires_at":        e.ExpiresAt,
		"access_count":      e.AccessCount,
		"tags":              e.Tags,
		"metadata":          e.Metadata,
	}
}

// NewArchiveEntry creates a new archive entry with default values
func NewArchiveEntry(content []byte, command string) *ArchiveEntry {
	now := time.Now()
	return &ArchiveEntry{
		OriginalContent: content,
		OriginalSize:    int64(len(content)),
		CompressedSize:  int64(len(content)),
		Compression:     CompressionNone,
		Command:         command,
		Category:        CategoryCommand,
		CreatedAt:       now,
		Tags:            []string{},
		Metadata: ArchiveMetadata{
			Custom: make(map[string]interface{}),
		},
	}
}

// WithFiltered sets the filtered content
func (e *ArchiveEntry) WithFiltered(content []byte) *ArchiveEntry {
	e.FilteredContent = content
	return e
}

// WithCategory sets the category
func (e *ArchiveEntry) WithCategory(category ArchiveCategory) *ArchiveEntry {
	e.Category = category
	return e
}

// WithAgent sets the agent
func (e *ArchiveEntry) WithAgent(agent string) *ArchiveEntry {
	e.Agent = agent
	return e
}

// WithWorkingDirectory sets the working directory
func (e *ArchiveEntry) WithWorkingDirectory(dir string) *ArchiveEntry {
	e.WorkingDirectory = dir
	return e
}

// WithTags sets the tags
func (e *ArchiveEntry) WithTags(tags ...string) *ArchiveEntry {
	e.Tags = tags
	return e
}

// WithExpiration sets the expiration time
func (e *ArchiveEntry) WithExpiration(duration time.Duration) *ArchiveEntry {
	expires := e.CreatedAt.Add(duration)
	e.ExpiresAt = &expires
	return e
}

// WithMetadata sets custom metadata
func (e *ArchiveEntry) WithMetadata(key string, value interface{}) *ArchiveEntry {
	if e.Metadata.Custom == nil {
		e.Metadata.Custom = make(map[string]interface{})
	}
	e.Metadata.Custom[key] = value
	return e
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ArchiveListResult contains the results of a list operation
type ArchiveListResult struct {
	Entries []ArchiveEntry `json:"entries"`
	Total   int64          `json:"total"`
	HasMore bool           `json:"has_more"`
}

// ArchiveSearchResult contains search results
type ArchiveSearchResult struct {
	Entries []ArchiveEntry `json:"entries"`
	Total   int64          `json:"total"`
	Query   string         `json:"query"`
}

// ArchiveStats contains aggregated statistics
type ArchiveEntryStats struct {
	TotalArchives       int64                     `json:"total_archives"`
	TotalOriginalSize   int64                     `json:"total_original_size"`
	TotalCompressedSize int64                     `json:"total_compressed_size"`
	AvgCompressionRatio float64                   `json:"avg_compression_ratio"`
	TotalSpaceSaved     int64                     `json:"total_space_saved"`
	TotalAccesses       int64                     `json:"total_accesses"`
	OldestArchive       *time.Time                `json:"oldest_archive,omitempty"`
	NewestArchive       *time.Time                `json:"newest_archive,omitempty"`
	CategoryBreakdown   map[ArchiveCategory]int64 `json:"category_breakdown"`
	AgentBreakdown      map[string]int64          `json:"agent_breakdown"`
}

// CalculateAvgCompressionRatio calculates the average compression ratio
func (s *ArchiveEntryStats) CalculateAvgCompressionRatio() {
	if s.TotalOriginalSize == 0 {
		s.AvgCompressionRatio = 1.0
		return
	}
	s.AvgCompressionRatio = float64(s.TotalCompressedSize) / float64(s.TotalOriginalSize)
}
