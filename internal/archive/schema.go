package archive

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SchemaVersion represents the current database schema version
const SchemaVersion = 2

// SchemaDefinition contains all SQL statements to create the archive database schema
var SchemaDefinition = `
-- Main archives table storing all archived content
CREATE TABLE IF NOT EXISTS archives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT UNIQUE NOT NULL,
    original_content BLOB NOT NULL,
    filtered_content BLOB,
    command TEXT,
    working_directory TEXT,
    project_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    accessed_at DATETIME,
    access_count INTEGER DEFAULT 0,
    category TEXT DEFAULT 'command',
    agent TEXT,
    compression_type TEXT DEFAULT 'gzip',
    original_size INTEGER NOT NULL,
    compressed_size INTEGER NOT NULL,
    metadata TEXT,
    
    -- Indexes
    UNIQUE(hash)
);

-- Index for hash lookups
CREATE INDEX IF NOT EXISTS idx_archives_hash ON archives(hash);

-- Index for expiration queries
CREATE INDEX IF NOT EXISTS idx_archives_expires_at ON archives(expires_at) WHERE expires_at IS NOT NULL;

-- Index for category filtering
CREATE INDEX IF NOT EXISTS idx_archives_category ON archives(category);

-- Index for agent filtering
CREATE INDEX IF NOT EXISTS idx_archives_agent ON archives(agent);

-- Index for created_at sorting
CREATE INDEX IF NOT EXISTS idx_archives_created_at ON archives(created_at DESC);

-- Index for access patterns
CREATE INDEX IF NOT EXISTS idx_archives_accessed_at ON archives(accessed_at DESC);

-- Archive tags for categorization
CREATE TABLE IF NOT EXISTS archive_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    archive_id INTEGER NOT NULL,
    tag TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (archive_id) REFERENCES archives(id) ON DELETE CASCADE,
    UNIQUE(archive_id, tag)
);

-- Index for tag lookups
CREATE INDEX IF NOT EXISTS idx_archive_tags_tag ON archive_tags(tag);
CREATE INDEX IF NOT EXISTS idx_archive_tags_archive_id ON archive_tags(archive_id);

-- Access log for tracking retrievals
CREATE TABLE IF NOT EXISTS archive_access_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    archive_id INTEGER NOT NULL,
    accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accessed_by TEXT,
    context TEXT,
    
    FOREIGN KEY (archive_id) REFERENCES archives(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_access_log_archive_id ON archive_access_log(archive_id);
CREATE INDEX IF NOT EXISTS idx_access_log_accessed_at ON archive_access_log(accessed_at DESC);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert current schema version (use INSERT OR REPLACE to ensure it exists)
INSERT OR REPLACE INTO schema_version (version, applied_at) VALUES (2, CURRENT_TIMESTAMP);

-- Archive statistics (aggregated, updated periodically)
CREATE TABLE IF NOT EXISTS archive_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_archives INTEGER DEFAULT 0,
    total_original_size INTEGER DEFAULT 0,
    total_compressed_size INTEGER DEFAULT 0,
    total_accesses INTEGER DEFAULT 0,
    category_breakdown TEXT,
    agent_breakdown TEXT,
    daily_stats TEXT
);

-- Quota and limits configuration
CREATE TABLE IF NOT EXISTS archive_quotas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quota_type TEXT NOT NULL, -- 'global', 'user', 'category', 'agent'
    quota_key TEXT NOT NULL,  -- actual value (user_id, category_name, etc.)
    max_size_bytes INTEGER,
    max_count INTEGER,
    max_age_days INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(quota_type, quota_key)
);

-- Default global quota
INSERT OR IGNORE INTO archive_quotas (quota_type, quota_key, max_size_bytes, max_count, max_age_days)
VALUES ('global', 'default', 10737418240, 100000, 90); -- 10GB, 100k archives, 90 days

-- Cleanup job tracking
CREATE TABLE IF NOT EXISTS cleanup_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    archives_deleted INTEGER DEFAULT 0,
    bytes_freed INTEGER DEFAULT 0,
    status TEXT DEFAULT 'running', -- 'running', 'completed', 'failed'
    error_message TEXT
);

-- Full-text search virtual table (if supported)
CREATE VIRTUAL TABLE IF NOT EXISTS archives_fts USING fts5(
    hash,
    command,
    working_directory,
    content='archives',
    content_rowid='id'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS archives_ai AFTER INSERT ON archives BEGIN
    INSERT INTO archives_fts(rowid, hash, command, working_directory)
    VALUES (new.id, new.hash, new.command, new.working_directory);
END;

CREATE TRIGGER IF NOT EXISTS archives_ad AFTER DELETE ON archives BEGIN
    INSERT INTO archives_fts(archives_fts, rowid, hash, command, working_directory)
    VALUES ('delete', old.id, old.hash, old.command, old.working_directory);
END;

CREATE TRIGGER IF NOT EXISTS archives_au AFTER UPDATE ON archives BEGIN
    INSERT INTO archives_fts(archives_fts, rowid, hash, command, working_directory)
    VALUES ('delete', old.id, old.hash, old.command, old.working_directory);
    INSERT INTO archives_fts(rowid, hash, command, working_directory)
    VALUES (new.id, new.hash, new.command, new.working_directory);
END;
`

// SchemaManager handles database schema creation and migrations
type SchemaManager struct {
	db *sql.DB
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(db *sql.DB) *SchemaManager {
	return &SchemaManager{db: db}
}

// Initialize creates all necessary tables and indexes
func (sm *SchemaManager) Initialize(ctx context.Context) error {
	// Execute schema definition
	if _, err := sm.db.ExecContext(ctx, SchemaDefinition); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// GetVersion returns the current schema version
func (sm *SchemaManager) GetVersion(ctx context.Context) (int, error) {
	var version int
	err := sm.db.QueryRowContext(ctx,
		"SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// Migrate performs schema migrations from current version to target
func (sm *SchemaManager) Migrate(ctx context.Context, targetVersion int) error {
	currentVersion, err := sm.GetVersion(ctx)
	if err != nil {
		return err
	}

	if currentVersion >= targetVersion {
		return nil // Already up to date
	}

	// Apply migrations sequentially
	for version := currentVersion + 1; version <= targetVersion; version++ {
		if err := sm.applyMigration(ctx, version); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", version, err)
		}
	}

	return nil
}

// applyMigration applies a specific migration version
func (sm *SchemaManager) applyMigration(ctx context.Context, version int) error {
	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch version {
	case 1:
		// Initial schema - already applied by Initialize
		// Just record the version
	case 2:
		// Migration to version 2: Add project_path column
		if _, err := tx.ExecContext(ctx,
			`ALTER TABLE archives ADD COLUMN IF NOT EXISTS project_path TEXT`); err != nil {
			return fmt.Errorf("failed to add project_path column: %w", err)
		}
		// Add index for project_path
		if _, err := tx.ExecContext(ctx,
			`CREATE INDEX IF NOT EXISTS idx_archives_project_path ON archives(project_path)`); err != nil {
			return fmt.Errorf("failed to create project_path index: %w", err)
		}

	default:
		return fmt.Errorf("unknown schema version: %d", version)
	}

	// Record migration
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_version (version) VALUES (?)", version); err != nil {
		return err
	}

	return tx.Commit()
}

// Reset drops all tables (USE WITH CAUTION)
func (sm *SchemaManager) Reset(ctx context.Context) error {
	tables := []string{
		"archive_access_log",
		"archive_tags",
		"archives_fts",
		"archives",
		"archive_stats",
		"archive_quotas",
		"cleanup_jobs",
		"schema_version",
	}

	for _, table := range tables {
		if _, err := sm.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// Verify checks schema integrity
func (sm *SchemaManager) Verify(ctx context.Context) error {
	// Check if all required tables exist
	tables := []string{
		"archives",
		"archive_tags",
		"archive_access_log",
		"schema_version",
		"archive_stats",
		"archive_quotas",
		"cleanup_jobs",
	}

	for _, table := range tables {
		var name string
		err := sm.db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == sql.ErrNoRows {
			return fmt.Errorf("required table missing: %s", table)
		}
		if err != nil {
			return fmt.Errorf("failed to verify table %s: %w", table, err)
		}
	}

	// Check indexes
	indexes := []string{
		"idx_archives_hash",
		"idx_archives_category",
		"idx_archive_tags_tag",
	}

	for _, index := range indexes {
		var name string
		err := sm.db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
		if err == sql.ErrNoRows {
			return fmt.Errorf("required index missing: %s", index)
		}
		if err != nil {
			return fmt.Errorf("failed to verify index %s: %w", index, err)
		}
	}

	return nil
}

// Stats returns current database statistics
func (sm *SchemaManager) Stats(ctx context.Context) (*DBStats, error) {
	stats := &DBStats{}

	// Count archives
	if err := sm.db.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(original_size), 0), COALESCE(SUM(compressed_size), 0) FROM archives").
		Scan(&stats.TotalArchives, &stats.TotalOriginalSize, &stats.TotalCompressedSize); err != nil {
		return nil, fmt.Errorf("failed to get archive stats: %w", err)
	}

	// Count tags
	if err := sm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM archive_tags").
		Scan(&stats.TotalTags); err != nil {
		return nil, fmt.Errorf("failed to get tag stats: %w", err)
	}

	// Count access log entries
	if err := sm.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM archive_access_log").
		Scan(&stats.TotalAccesses); err != nil {
		return nil, fmt.Errorf("failed to get access stats: %w", err)
	}

	// Get schema version
	version, err := sm.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	stats.SchemaVersion = version
	stats.LastUpdated = time.Now()

	return stats, nil
}

// DBStats contains database statistics
type DBStats struct {
	TotalArchives       int64
	TotalOriginalSize   int64
	TotalCompressedSize int64
	TotalTags           int64
	TotalAccesses       int64
	SchemaVersion       int
	LastUpdated         time.Time
}

// CompressionRatio returns the overall compression ratio
func (s *DBStats) CompressionRatio() float64 {
	if s.TotalOriginalSize == 0 {
		return 1.0
	}
	return float64(s.TotalCompressedSize) / float64(s.TotalOriginalSize)
}

// SpaceSaved returns bytes saved by compression
func (s *DBStats) SpaceSaved() int64 {
	return s.TotalOriginalSize - s.TotalCompressedSize
}
