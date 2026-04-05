// Package reversible provides SQLite-backed storage for compressed entries.
package reversible

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db         *sql.DB
	config     Config
	compressor Compressor
	encryptor  Encryptor
}

// NewSQLiteStore creates a new SQLite-backed store.
func NewSQLiteStore(config Config) (*SQLiteStore, error) {
	// Expand ~ in path
	storePath := expandPath(config.StorePath)
	if storePath == "" {
		return nil, fmt.Errorf("store path is required")
	}

	if err := os.MkdirAll(filepath.Dir(storePath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	db, err := sql.Open("sqlite", storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{
		db:     db,
		config: config,
	}

	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize compressor
	store.compressor = &ZstdCompressor{}

	// Initialize encryptor if key provided
	if len(config.EncryptionKey) > 0 {
		encryptor, err := NewAESEncryptor(config.EncryptionKey)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
		}
		store.encryptor = encryptor
	}

	return store, nil
}

// Close closes the store.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// migrate creates tables and indexes.
func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS entries (
			hash TEXT PRIMARY KEY,
			original_preview TEXT,
			command TEXT,
			content_type INTEGER DEFAULT 0,
			compression_alg TEXT DEFAULT 'zstd',
			encrypted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			access_count INTEGER DEFAULT 0,
			size_original INTEGER DEFAULT 0,
			size_compressed INTEGER DEFAULT 0,
			compressed_data BLOB
		);

		CREATE INDEX IF NOT EXISTS idx_entries_command ON entries(command);
		CREATE INDEX IF NOT EXISTS idx_entries_created ON entries(created_at);
		CREATE INDEX IF NOT EXISTS idx_entries_accessed ON entries(accessed_at);

		CREATE TABLE IF NOT EXISTS command_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			hash TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			duration_ms INTEGER,
			compressed INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_history_command ON command_history(command);
		CREATE INDEX IF NOT EXISTS idx_history_timestamp ON command_history(timestamp);
	`)
	return err
}

// Save stores an entry and returns its hash.
func (s *SQLiteStore) Save(entry *Entry) (string, error) {
	if entry == nil {
		return "", fmt.Errorf("entry is required")
	}
	if entry.SizeOriginal > s.config.MaxEntrySize {
		return "", fmt.Errorf("entry size %d exceeds maximum %d", entry.SizeOriginal, s.config.MaxEntrySize)
	}

	hash := entry.Hash
	if hash == "" {
		hash = ComputeHash(entry.Original)
		entry.Hash = hash
	}
	if entry.SizeOriginal == 0 {
		entry.SizeOriginal = int64(len(entry.Original))
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if entry.CompressionAlg == "" {
		entry.CompressionAlg = s.config.DefaultAlgorithm
	}

	// Compress the original content
	compressedBytes, err := s.compressor.Compress([]byte(entry.Original))
	if err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}

	// Encrypt if configured
	encrypted := false
	if s.encryptor != nil {
		compressedBytes, err = s.encryptor.Encrypt(compressedBytes)
		if err != nil {
			return "", fmt.Errorf("encryption failed: %w", err)
		}
		encrypted = true
	}

	// Store only a preview of the original (first 200 chars)
	preview := entry.Original
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}

	_, err = s.db.Exec(`
		INSERT INTO entries (
			hash, original_preview, command, content_type, compression_alg,
			encrypted, created_at, accessed_at, access_count,
			size_original, size_compressed, compressed_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			accessed_at = CURRENT_TIMESTAMP,
			access_count = access_count + 1
	`,
		hash, preview, entry.Command, entry.ContentType,
		entry.CompressionAlg, encrypted, entry.CreatedAt, time.Now(), 1,
		entry.SizeOriginal, int64(len(compressedBytes)), compressedBytes,
	)

	if err != nil {
		return "", fmt.Errorf("failed to save entry: %w", err)
	}

	return hash, nil
}

// Retrieve gets an entry by hash (short or full).
func (s *SQLiteStore) Retrieve(hash string) (*Entry, error) {
	// Try full hash first
	row := s.db.QueryRow(`
		SELECT hash, original_preview, command, content_type, compression_alg,
		       encrypted, created_at, accessed_at, access_count,
		       size_original, size_compressed, compressed_data
		FROM entries WHERE hash = ?
	`, hash)

	entry, err := s.scanEntry(row)
	if err == sql.ErrNoRows && len(hash) == 16 {
		// Try short hash match
		row = s.db.QueryRow(`
			SELECT hash, original_preview, command, content_type, compression_alg,
			       encrypted, created_at, accessed_at, access_count,
			       size_original, size_compressed, compressed_data
			FROM entries WHERE substr(hash, 1, 16) = ?
		`, hash)
		entry, err = s.scanEntry(row)
	}

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry not found: %s", hash)
	}
	if err != nil {
		return nil, err
	}

	// Update access stats
	if _, err := s.db.Exec(`
		UPDATE entries SET accessed_at = CURRENT_TIMESTAMP, access_count = access_count + 1
		WHERE hash = ?
	`, entry.Hash); err != nil {
		log.Printf("failed to update access stats: %v", err)
	}

	// Decompress
	if len(entry.CompressedData) > 0 {
		data := entry.CompressedData

		// Decrypt if encrypted
		if entry.Encrypted && s.encryptor != nil {
			data, err = s.encryptor.Decrypt(data)
			if err != nil {
				return nil, fmt.Errorf("decryption failed: %w", err)
			}
		}

		// Decompress
		original, err := s.compressor.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("decompression failed: %w", err)
		}

		entry.Original = string(original)
	}

	return entry, nil
}

// Delete removes an entry.
func (s *SQLiteStore) Delete(hash string) error {
	result, err := s.db.Exec(`DELETE FROM entries WHERE hash = ? OR substr(hash, 1, 16) = ?`, hash, hash)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("entry not found: %s", hash)
	}

	if s.config.AutoVacuum {
		if err := s.Vacuum(); err != nil {
			log.Printf("auto-vacuum failed: %v", err)
		}
	}

	return nil
}

// List returns entries matching the filter.
func (s *SQLiteStore) List(filter ListFilter) ([]*Entry, error) {
	query := `
		SELECT hash, original_preview, command, content_type, compression_alg,
		       encrypted, created_at, accessed_at, access_count,
		       size_original, size_compressed, compressed_data
		FROM entries WHERE 1=1
	`
	var args []interface{}
	var conditions []string

	if filter.Command != "" {
		conditions = append(conditions, "command = ?")
		args = append(args, filter.Command)
	}
	if !filter.Since.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.Since)
	}
	if !filter.Before.IsZero() {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.Before)
	}
	if filter.MinSize > 0 {
		conditions = append(conditions, "size_original >= ?")
		args = append(args, filter.MinSize)
	}
	if filter.MaxSize > 0 {
		conditions = append(conditions, "size_original <= ?")
		args = append(args, filter.MaxSize)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		entry, err := s.scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Stats returns store statistics.
func (s *SQLiteStore) Stats() (StoreStats, error) {
	var stats StoreStats

	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(size_original), 0), COALESCE(SUM(size_compressed), 0),
		       MIN(created_at), MAX(created_at)
		FROM entries
	`).Scan(&stats.TotalEntries, &stats.TotalSizeOrig, &stats.TotalSizeComp,
		&stats.OldestEntry, &stats.NewestEntry)
	if err != nil {
		return stats, err
	}

	// Get breakdown by content type
	stats.ByContentType = make(map[ContentType]int64)
	rows, err := s.db.Query(`SELECT content_type, COUNT(*) FROM entries GROUP BY content_type`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var ct int
		var count int64
		if err := rows.Scan(&ct, &count); err == nil {
			stats.ByContentType[ContentType(ct)] = count
		}
	}

	// Get breakdown by command
	stats.ByCommand = make(map[string]int64)
	rows, err = s.db.Query(`SELECT command, COUNT(*) FROM entries WHERE command != '' GROUP BY command`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var cmd string
		var count int64
		if err := rows.Scan(&cmd, &count); err == nil {
			stats.ByCommand[cmd] = count
		}
	}

	return stats, nil
}

// Vacuum reclaims space.
func (s *SQLiteStore) Vacuum() error {
	_, err := s.db.Exec("VACUUM")
	return err
}

// DeleteOlderThan deletes entries older than duration.
func (s *SQLiteStore) DeleteOlderThan(d time.Duration) (int64, error) {
	cutoff := time.Now().Add(-d)
	result, err := s.db.Exec(`DELETE FROM entries WHERE created_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// RecordCommand records command execution.
func (s *SQLiteStore) RecordCommand(record *CommandRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO command_history (command, hash, timestamp, duration_ms, compressed)
		VALUES (?, ?, ?, ?, ?)
	`, record.Command, record.Hash, record.Timestamp, record.Duration.Milliseconds(), record.Compressed)
	return err
}

// GetCommandHistory returns command history.
func (s *SQLiteStore) GetCommandHistory(limit int) ([]*CommandRecord, error) {
	query := `SELECT command, hash, timestamp, duration_ms, compressed FROM command_history ORDER BY timestamp DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*CommandRecord
	for rows.Next() {
		var r CommandRecord
		var durationMs int64
		if err := rows.Scan(&r.Command, &r.Hash, &r.Timestamp, &durationMs, &r.Compressed); err != nil {
			continue
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
		records = append(records, &r)
	}

	return records, rows.Err()
}

// scanEntry scans an entry from a row.
func (s *SQLiteStore) scanEntry(scanner interface {
	Scan(dest ...interface{}) error
}) (*Entry, error) {
	var entry Entry
	var encrypted int
	var compressedData []byte

	err := scanner.Scan(
		&entry.Hash, &entry.Original, &entry.Command, &entry.ContentType,
		&entry.CompressionAlg, &encrypted, &entry.CreatedAt, &entry.AccessedAt,
		&entry.AccessCount, &entry.SizeOriginal, &entry.SizeCompressed, &compressedData,
	)
	if err != nil {
		return nil, err
	}

	entry.Encrypted = encrypted != 0
	entry.CompressedData = compressedData

	return &entry, nil
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
