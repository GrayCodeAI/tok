// Package persistence provides SQLite-based persistence for MCP cache.
package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed persistence for cache metadata.
type Store struct {
	db *sql.DB
}

// CacheMetadata represents persisted cache entry metadata (without content).
type CacheMetadata struct {
	Hash      string    `json:"hash"`
	FilePath  string    `json:"file_path"`
	Timestamp time.Time `json:"timestamp"`
	Accessed  time.Time `json:"accessed"`
	HitCount  int       `json:"hit_count"`
	Size      int64     `json:"size"`
}

// NewStore creates a new SQLite-backed store.
func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the necessary tables.
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_metadata (
			hash TEXT PRIMARY KEY,
			file_path TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			accessed DATETIME NOT NULL,
			hit_count INTEGER DEFAULT 0,
			size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_cache_path ON cache_metadata(file_path);
		CREATE INDEX IF NOT EXISTS idx_cache_accessed ON cache_metadata(accessed);

		CREATE TABLE IF NOT EXISTS store_metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// SaveMetadata saves cache entry metadata.
func (s *Store) SaveMetadata(meta *CacheMetadata) error {
	_, err := s.db.Exec(`
		INSERT INTO cache_metadata (hash, file_path, timestamp, accessed, hit_count, size)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			file_path = excluded.file_path,
			timestamp = excluded.timestamp,
			accessed = excluded.accessed,
			hit_count = excluded.hit_count,
			size = excluded.size,
			updated_at = CURRENT_TIMESTAMP
	`, meta.Hash, meta.FilePath, meta.Timestamp, meta.Accessed, meta.HitCount, meta.Size)
	return err
}

// GetMetadata retrieves metadata by hash.
func (s *Store) GetMetadata(hash string) (*CacheMetadata, error) {
	var meta CacheMetadata
	err := s.db.QueryRow(`
		SELECT hash, file_path, timestamp, accessed, hit_count, size
		FROM cache_metadata
		WHERE hash = ?
	`, hash).Scan(&meta.Hash, &meta.FilePath, &meta.Timestamp, &meta.Accessed, &meta.HitCount, &meta.Size)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// LoadAllMetadata loads all cached metadata entries.
func (s *Store) LoadAllMetadata() ([]*CacheMetadata, error) {
	rows, err := s.db.Query(`
		SELECT hash, file_path, timestamp, accessed, hit_count, size
		FROM cache_metadata
		ORDER BY accessed DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*CacheMetadata
	for rows.Next() {
		var meta CacheMetadata
		if err := rows.Scan(&meta.Hash, &meta.FilePath, &meta.Timestamp, &meta.Accessed, &meta.HitCount, &meta.Size); err != nil {
			return nil, err
		}
		results = append(results, &meta)
	}
	return results, rows.Err()
}

// DeleteMetadata removes metadata by hash.
func (s *Store) DeleteMetadata(hash string) error {
	_, err := s.db.Exec(`DELETE FROM cache_metadata WHERE hash = ?`, hash)
	return err
}

// DeleteOlderThan removes entries older than the given duration.
func (s *Store) DeleteOlderThan(d time.Duration) (int64, error) {
	cutoff := time.Now().Add(-d)
	result, err := s.db.Exec(`DELETE FROM cache_metadata WHERE accessed < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetStats returns store statistics.
func (s *Store) GetStats() (totalEntries, totalSize int64, oldest, newest time.Time, err error) {
	var oldestStr, newestStr sql.NullString
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(size), 0), MIN(timestamp), MAX(timestamp)
		FROM cache_metadata
	`).Scan(&totalEntries, &totalSize, &oldestStr, &newestStr)
	if err != nil {
		return 0, 0, time.Time{}, time.Time{}, err
	}
	if oldestStr.Valid {
		oldest, _ = time.Parse(time.RFC3339Nano, oldestStr.String)
	}
	if newestStr.Valid {
		newest, _ = time.Parse(time.RFC3339Nano, newestStr.String)
	}
	return totalEntries, totalSize, oldest, newest, nil
}

// SetStoreMetadata sets a metadata value.
func (s *Store) SetStoreMetadata(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO store_metadata (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

// GetStoreMetadata retrieves a metadata value.
func (s *Store) GetStoreMetadata(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM store_metadata WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// Backup creates a backup of the store to a JSON file.
func (s *Store) Backup(backupPath string) error {
	metaList, err := s.LoadAllMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	data, err := json.MarshalIndent(metaList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// Restore restores the store from a JSON backup.
func (s *Store) Restore(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	var metaList []*CacheMetadata
	if err := json.Unmarshal(data, &metaList); err != nil {
		return fmt.Errorf("failed to unmarshal backup: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO cache_metadata (hash, file_path, timestamp, accessed, hit_count, size)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			file_path = excluded.file_path,
			timestamp = excluded.timestamp,
			accessed = excluded.accessed,
			hit_count = excluded.hit_count,
			size = excluded.size
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, meta := range metaList {
		if _, err := stmt.Exec(meta.Hash, meta.FilePath, meta.Timestamp, meta.Accessed, meta.HitCount, meta.Size); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Vacuum runs SQLite VACUUM to reclaim space.
func (s *Store) Vacuum() error {
	_, err := s.db.Exec("VACUUM")
	return err
}
