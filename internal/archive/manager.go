package archive

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/compression"
	"github.com/GrayCodeAI/tokman/internal/config"
	_ "modernc.org/sqlite"
)

// ArchiveManager handles all archive operations
type ArchiveManager struct {
	db          *sql.DB
	hasher      *HashCalculator
	schema      *SchemaManager
	compressor  *CompressionEngine
	config      ArchiveConfig
	dbPath      string
	mu          sync.RWMutex
	initialized bool
}

// NewArchiveManager creates a new archive manager
func NewArchiveManager(cfg ArchiveConfig) (*ArchiveManager, error) {
	// Determine database path
	dataDir := config.DataPath()

	dbPath := filepath.Join(dataDir, "archive.db")

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	manager := &ArchiveManager{
		db:         db,
		hasher:     NewHashCalculator(),
		schema:     NewSchemaManager(db),
		compressor: NewCompressionEngine(),
		config:     cfg,
		dbPath:     dbPath,
	}

	return manager, nil
}

// Initialize sets up the database schema
func (am *ArchiveManager) Initialize(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.initialized {
		return nil
	}

	// Initialize schema
	if err := am.schema.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	am.initialized = true
	slog.Info("archive manager initialized", "db_path", am.dbPath)

	return nil
}

// Close closes the database connection
func (am *ArchiveManager) Close() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.db != nil {
		return am.db.Close()
	}
	return nil
}

// Archive stores content in the archive
func (am *ArchiveManager) Archive(ctx context.Context, entry *ArchiveEntry) (string, error) {
	if !am.initialized {
		return "", fmt.Errorf("archive manager not initialized")
	}

	// Calculate hash from original content
	hash := am.hasher.Calculate(entry.OriginalContent)
	entry.Hash = hash

	// Check if already exists
	var existingID int64
	err := am.db.QueryRowContext(ctx, "SELECT id FROM archives WHERE hash = ?", hash).Scan(&existingID)
	if err == nil {
		// Already exists, return existing hash
		slog.Debug("archive already exists", "hash", hash[:8])
		return hash, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check existing archive: %w", err)
	}

	// Marshal metadata
	metadataJSON, err := entry.MarshalMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Compress content if enabled
	compressedOrig := entry.OriginalContent
	compressedFilt := entry.FilteredContent
	var origResult *compression.CompressionResult

	if am.config.EnableCompression {
		compressedOrig, origResult, err = am.compressor.Compress(entry.OriginalContent)
		if err != nil {
			slog.Error("failed to compress original content", "error", err)
			compressedOrig = entry.OriginalContent
		}

		if entry.FilteredContent != nil {
			compressedFilt, _, err = am.compressor.Compress(entry.FilteredContent)
			if err != nil {
				slog.Error("failed to compress filtered content", "error", err)
				compressedFilt = entry.FilteredContent
			}
		}
	}

	// Update compression info
	if origResult != nil && origResult.WasCompressed {
		entry.Compression = CompressionBrotli
		entry.CompressedSize = int64(len(compressedOrig))
	}

	// Insert archive (use compressed content)
	result, err := am.db.ExecContext(ctx, `
		INSERT INTO archives (
			hash, original_content, filtered_content, command, working_directory,
			project_path, agent, category, compression_type, original_size,
			compressed_size, expires_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.Hash,
		compressedOrig,
		compressedFilt,
		entry.Command,
		entry.WorkingDirectory,
		entry.ProjectPath,
		entry.Agent,
		entry.Category,
		entry.Compression,
		entry.OriginalSize,
		entry.CompressedSize,
		entry.ExpiresAt,
		metadataJSON,
	)

	if err != nil {
		return "", fmt.Errorf("failed to insert archive: %w", err)
	}

	archiveID, _ := result.LastInsertId()
	entry.ID = archiveID

	// Insert tags
	if len(entry.Tags) > 0 {
		for _, tag := range entry.Tags {
			_, err := am.db.ExecContext(ctx,
				"INSERT INTO archive_tags (archive_id, tag) VALUES (?, ?)",
				archiveID, tag)
			if err != nil {
				slog.Error("failed to insert tag", "tag", tag, "error", err)
			}
		}
	}

	slog.Debug("archived content",
		"hash", hash[:8],
		"size", entry.OriginalSize,
		"compressed", entry.CompressedSize,
	)

	return hash, nil
}

// Retrieve retrieves an archive by hash
func (am *ArchiveManager) Retrieve(ctx context.Context, hash string) (*ArchiveEntry, error) {
	if !am.initialized {
		return nil, fmt.Errorf("archive manager not initialized")
	}

	// Validate hash
	if !IsValidHash(hash) {
		return nil, fmt.Errorf("invalid hash format")
	}

	// Query archive
	entry := &ArchiveEntry{}
	var metadataJSON string
	var filteredContent, originalContent []byte
	var expiresAt sql.NullTime
	var accessedAt sql.NullTime

	err := am.db.QueryRowContext(ctx, `
		SELECT id, hash, original_content, filtered_content, command,
			working_directory, project_path, agent, category, compression_type,
			original_size, compressed_size, created_at, accessed_at, expires_at,
			access_count, metadata
		FROM archives WHERE hash = ?
	`, hash).Scan(
		&entry.ID, &entry.Hash, &originalContent, &filteredContent,
		&entry.Command, &entry.WorkingDirectory, &entry.ProjectPath,
		&entry.Agent, &entry.Category, &entry.Compression,
		&entry.OriginalSize, &entry.CompressedSize, &entry.CreatedAt,
		&accessedAt, &expiresAt, &entry.AccessCount, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("archive not found: %s", hash)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve archive: %w", err)
	}

	// Set nullable fields
	if accessedAt.Valid {
		entry.AccessedAt = &accessedAt.Time
	}
	if expiresAt.Valid {
		entry.ExpiresAt = &expiresAt.Time
	}

	// Decompress content if needed
	if entry.Compression == CompressionBrotli && am.config.EnableCompression {
		decompressedOrig, err := am.compressor.Decompress(originalContent)
		if err != nil {
			slog.Error("failed to decompress original content", "error", err)
			entry.OriginalContent = originalContent
		} else {
			entry.OriginalContent = decompressedOrig
		}

		if filteredContent != nil {
			decompressedFilt, err := am.compressor.Decompress(filteredContent)
			if err != nil {
				slog.Error("failed to decompress filtered content", "error", err)
				entry.FilteredContent = filteredContent
			} else {
				entry.FilteredContent = decompressedFilt
			}
		}
	} else {
		entry.OriginalContent = originalContent
		entry.FilteredContent = filteredContent
	}

	// Unmarshal metadata
	if err := entry.UnmarshalMetadata(metadataJSON); err != nil {
		slog.Error("failed to unmarshal metadata", "error", err)
	}

	// Load tags
	tags, err := am.GetTags(ctx, entry.ID)
	if err != nil {
		slog.Error("failed to load tags", "error", err)
	} else {
		entry.Tags = tags
	}

	// Update access log
	am.markAccessed(ctx, entry.ID)

	return entry, nil
}

// markAccessed updates the access timestamp and count
func (am *ArchiveManager) markAccessed(ctx context.Context, archiveID int64) {
	now := time.Now()

	// Update archive access info
	_, err := am.db.ExecContext(ctx,
		"UPDATE archives SET accessed_at = ?, access_count = access_count + 1 WHERE id = ?",
		now, archiveID)
	if err != nil {
		slog.Error("failed to update archive access", "error", err)
		return
	}

	// Log access
	_, err = am.db.ExecContext(ctx,
		"INSERT INTO archive_access_log (archive_id, accessed_at) VALUES (?, ?)",
		archiveID, now)
	if err != nil {
		slog.Error("failed to log archive access", "error", err)
	}
}

// GetTags retrieves tags for an archive
func (am *ArchiveManager) GetTags(ctx context.Context, archiveID int64) ([]string, error) {
	rows, err := am.db.QueryContext(ctx,
		"SELECT tag FROM archive_tags WHERE archive_id = ? ORDER BY tag", archiveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// List retrieves a list of archives
func (am *ArchiveManager) List(ctx context.Context, opts ArchiveListOptions) (*ArchiveListResult, error) {
	if !am.initialized {
		return nil, fmt.Errorf("archive manager not initialized")
	}

	// Build query
	query := `
		SELECT id, hash, command, working_directory, agent, category,
			original_size, compressed_size, created_at, accessed_at, expires_at,
			access_count
		FROM archives WHERE 1=1
	`
	var args []interface{}

	// Apply filters
	if opts.Category != "" {
		query += " AND category = ?"
		args = append(args, opts.Category)
	}
	if opts.Agent != "" {
		query += " AND agent = ?"
		args = append(args, opts.Agent)
	}
	if opts.ProjectPath != "" {
		query += " AND project_path = ?"
		args = append(args, opts.ProjectPath)
	}
	if opts.CreatedAfter != nil {
		query += " AND created_at >= ?"
		args = append(args, *opts.CreatedAfter)
	}
	if opts.CreatedBefore != nil {
		query += " AND created_at <= ?"
		args = append(args, *opts.CreatedBefore)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM archives WHERE 1=1" + query[len("SELECT id, hash, command, working_directory, agent, category, original_size, compressed_size, created_at, accessed_at, expires_at, access_count FROM archives WHERE 1=1"):]
	var total int64
	if err := am.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count archives: %w", err)
	}

	// Apply sorting
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := opts.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply pagination
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	// Execute query
	rows, err := am.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query archives: %w", err)
	}
	defer rows.Close()

	var entries []ArchiveEntry
	for rows.Next() {
		entry := ArchiveEntry{}
		var accessedAt, expiresAt sql.NullTime

		err := rows.Scan(
			&entry.ID, &entry.Hash, &entry.Command, &entry.WorkingDirectory,
			&entry.Agent, &entry.Category, &entry.OriginalSize, &entry.CompressedSize,
			&entry.CreatedAt, &accessedAt, &expiresAt, &entry.AccessCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archive: %w", err)
		}

		if accessedAt.Valid {
			entry.AccessedAt = &accessedAt.Time
		}
		if expiresAt.Valid {
			entry.ExpiresAt = &expiresAt.Time
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ArchiveListResult{
		Entries: entries,
		Total:   total,
		HasMore: total > int64(opts.Offset+len(entries)),
	}, nil
}

// Delete removes an archive by hash
func (am *ArchiveManager) Delete(ctx context.Context, hash string) error {
	if !am.initialized {
		return fmt.Errorf("archive manager not initialized")
	}

	result, err := am.db.ExecContext(ctx, "DELETE FROM archives WHERE hash = ?", hash)
	if err != nil {
		return fmt.Errorf("failed to delete archive: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("archive not found: %s", hash)
	}

	slog.Debug("deleted archive", "hash", hash[:8])
	return nil
}

// Verify checks the integrity of an archive
func (am *ArchiveManager) Verify(ctx context.Context, hash string) (bool, error) {
	if !am.initialized {
		return false, fmt.Errorf("archive manager not initialized")
	}

	entry, err := am.Retrieve(ctx, hash)
	if err != nil {
		return false, err
	}

	// Verify hash matches content
	calculatedHash := am.hasher.Calculate(entry.OriginalContent)
	if calculatedHash != entry.Hash {
		slog.Error("archive integrity check failed",
			"hash", hash[:8],
			"calculated", calculatedHash[:8],
		)
		return false, fmt.Errorf("hash mismatch: archive may be corrupted")
	}

	return true, nil
}

// Stats returns database statistics
func (am *ArchiveManager) Stats(ctx context.Context) (*DBStats, error) {
	if !am.initialized {
		return nil, fmt.Errorf("archive manager not initialized")
	}

	return am.schema.Stats(ctx)
}

// CleanupExpired removes expired archives
func (am *ArchiveManager) CleanupExpired(ctx context.Context) (int64, error) {
	if !am.initialized {
		return 0, fmt.Errorf("archive manager not initialized")
	}

	result, err := am.db.ExecContext(ctx,
		"DELETE FROM archives WHERE expires_at IS NOT NULL AND expires_at < ?",
		time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired archives: %w", err)
	}

	deleted, _ := result.RowsAffected()
	slog.Info("cleaned up expired archives", "deleted", deleted)

	return deleted, nil
}

// GetDBPath returns the database file path
func (am *ArchiveManager) GetDBPath() string {
	return am.dbPath
}

// IsInitialized returns whether the manager has been initialized
func (am *ArchiveManager) IsInitialized() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.initialized
}
