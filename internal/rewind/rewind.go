// Package rewind provides zero-loss compression storage.
//
// RewindStore archives original command outputs before compression,
// allowing users to retrieve the full uncompressed output at any time.
// Inspired by OMNI's RewindStore architecture.
//
// Each entry is identified by a SHA-256 hash and contains both the
// original and filtered output along with metadata about the compression.
package rewind

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single stored original/filtered output pair.
type Entry struct {
	Hash           string    `json:"hash"`
	Command        string    `json:"command"`
	Args           string    `json:"args"`
	OriginalOutput string    `json:"original_output"`
	FilteredOutput string    `json:"filtered_output"`
	OriginalTokens int       `json:"original_tokens"`
	FilteredTokens int       `json:"filtered_tokens"`
	TokensSaved    int       `json:"tokens_saved"`
	CompressionPct float64   `json:"compression_pct"`
	Timestamp      time.Time `json:"timestamp"`
	SessionID      string    `json:"session_id"`
}

// Stats contains aggregate statistics for the RewindStore.
type Stats struct {
	TotalEntries   int     `json:"total_entries"`
	TotalOriginal  int     `json:"total_original_tokens"`
	TotalFiltered  int     `json:"total_filtered_tokens"`
	TotalSaved     int     `json:"total_saved_tokens"`
	AvgCompression float64 `json:"avg_compression_pct"`
	DatabaseSize   int64   `json:"database_size_bytes"`
	OldestEntry    time.Time `json:"oldest_entry"`
	NewestEntry    time.Time `json:"newest_entry"`
}

// Store provides persistent storage for original command outputs.
// It uses SQLite for storage and supports concurrent access.
type Store struct {
	db       *sql.DB
	dbPath   string
	mu       sync.RWMutex
	maxSize  int64 // Maximum database size in bytes
	ttl      time.Duration // Time-to-live for entries
}

// Config holds configuration for the RewindStore.
type Config struct {
	DatabasePath string        `json:"database_path"`
	MaxSize      int64         `json:"max_size"`      // Max DB size in bytes (default: 100MB)
	TTL          time.Duration `json:"ttl"`           // Entry TTL (default: 7 days)
	Enabled      bool          `json:"enabled"`
}

// DefaultConfig returns a default RewindStore configuration.
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		DatabasePath: filepath.Join(homeDir, ".local", "share", "tokman", "rewind.db"),
		MaxSize:      100 * 1024 * 1024, // 100MB
		TTL:          7 * 24 * time.Hour, // 7 days
		Enabled:      true,
	}
}

// New creates a new RewindStore with the given configuration.
func New(cfg Config) (*Store, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create rewind directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open rewind database: %w", err)
	}

	store := &Store{
		db:      db,
		dbPath:  cfg.DatabasePath,
		maxSize: cfg.MaxSize,
		ttl:     cfg.TTL,
	}

	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate rewind database: %w", err)
	}

	return store, nil
}

// migrate creates the database schema.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS rewind_entries (
		hash TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		args TEXT DEFAULT '',
		original_output TEXT NOT NULL,
		filtered_output TEXT NOT NULL,
		original_tokens INTEGER DEFAULT 0,
		filtered_tokens INTEGER DEFAULT 0,
		tokens_saved INTEGER DEFAULT 0,
		compression_pct REAL DEFAULT 0.0,
		session_id TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_rewind_created_at ON rewind_entries(created_at);
	CREATE INDEX IF NOT EXISTS idx_rewind_command ON rewind_entries(command);
	CREATE INDEX IF NOT EXISTS idx_rewind_session ON rewind_entries(session_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// GenerateHash creates a SHA-256 hash for the given content.
func GenerateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16] // Short hash (first 16 chars)
}

// Save stores an original/filtered output pair.
func (s *Store) Save(entry Entry) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate hash from original output
	if entry.Hash == "" {
		entry.Hash = GenerateHash(entry.OriginalOutput)
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Calculate compression percentage
	if entry.OriginalTokens > 0 {
		entry.TokensSaved = entry.OriginalTokens - entry.FilteredTokens
		entry.CompressionPct = float64(entry.TokensSaved) / float64(entry.OriginalTokens) * 100
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO rewind_entries 
		(hash, command, args, original_output, filtered_output, 
		 original_tokens, filtered_tokens, tokens_saved, compression_pct, 
		 session_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.Hash, entry.Command, entry.Args,
		entry.OriginalOutput, entry.FilteredOutput,
		entry.OriginalTokens, entry.FilteredTokens,
		entry.TokensSaved, entry.CompressionPct,
		entry.SessionID, entry.Timestamp,
	)
	if err != nil {
		return "", fmt.Errorf("save rewind entry: %w", err)
	}

	return entry.Hash, nil
}

// Retrieve gets the original output for a given hash.
func (s *Store) Retrieve(hash string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entry Entry
	var createdAt string

	err := s.db.QueryRow(`
		SELECT hash, command, args, original_output, filtered_output,
		       original_tokens, filtered_tokens, tokens_saved, compression_pct,
		       session_id, created_at
		FROM rewind_entries WHERE hash = ?`, hash,
	).Scan(
		&entry.Hash, &entry.Command, &entry.Args,
		&entry.OriginalOutput, &entry.FilteredOutput,
		&entry.OriginalTokens, &entry.FilteredTokens,
		&entry.TokensSaved, &entry.CompressionPct,
		&entry.SessionID, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry not found: %s", hash)
	}
	if err != nil {
		return nil, fmt.Errorf("retrieve rewind entry: %w", err)
	}

	entry.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &entry, nil
}

// List returns recent entries, limited by count.
func (s *Store) List(limit int) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT hash, command, args, original_tokens, filtered_tokens,
		       tokens_saved, compression_pct, session_id, created_at
		FROM rewind_entries
		ORDER BY created_at DESC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list rewind entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var createdAt string

		if err := rows.Scan(
			&entry.Hash, &entry.Command, &entry.Args,
			&entry.OriginalTokens, &entry.FilteredTokens,
			&entry.TokensSaved, &entry.CompressionPct,
			&entry.SessionID, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan rewind entry: %w", err)
		}

		entry.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Delete removes an entry by hash.
func (s *Store) Delete(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.Exec("DELETE FROM rewind_entries WHERE hash = ?", hash)
	if err != nil {
		return fmt.Errorf("delete rewind entry: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("entry not found: %s", hash)
	}

	return nil
}

// Prune removes entries older than the configured TTL.
func (s *Store) Prune() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().UTC().Add(-s.ttl)

	result, err := s.db.Exec(
		"DELETE FROM rewind_entries WHERE created_at < ?",
		cutoff.Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return 0, fmt.Errorf("prune rewind entries: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// GetStats returns aggregate statistics for the RewindStore.
func (s *Store) GetStats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var stats Stats
	var oldestStr, newestStr sql.NullString

	err := s.db.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(original_tokens), 0),
			COALESCE(SUM(filtered_tokens), 0),
			COALESCE(SUM(tokens_saved), 0),
			COALESCE(AVG(compression_pct), 0),
			MIN(created_at),
			MAX(created_at)
		FROM rewind_entries`,
	).Scan(
		&stats.TotalEntries,
		&stats.TotalOriginal,
		&stats.TotalFiltered,
		&stats.TotalSaved,
		&stats.AvgCompression,
		&oldestStr,
		&newestStr,
	)
	if err != nil {
		return nil, fmt.Errorf("get rewind stats: %w", err)
	}

	if oldestStr.Valid {
		stats.OldestEntry, _ = time.Parse("2006-01-02 15:04:05", oldestStr.String)
	}
	if newestStr.Valid {
		stats.NewestEntry, _ = time.Parse("2006-01-02 15:04:05", newestStr.String)
	}

	// Get database file size
	if info, err := os.Stat(s.dbPath); err == nil {
		stats.DatabaseSize = info.Size()
	}

	return &stats, nil
}

// Search finds entries matching a command pattern.
func (s *Store) Search(pattern string) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT hash, command, args, original_tokens, filtered_tokens,
		       tokens_saved, compression_pct, session_id, created_at
		FROM rewind_entries
		WHERE command LIKE ? OR args LIKE ?
		ORDER BY created_at DESC
		LIMIT 50`, "%"+pattern+"%", "%"+pattern+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search rewind entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var createdAt string

		if err := rows.Scan(
			&entry.Hash, &entry.Command, &entry.Args,
			&entry.OriginalTokens, &entry.FilteredTokens,
			&entry.TokensSaved, &entry.CompressionPct,
			&entry.SessionID, &createdAt,
		); err != nil {
			continue
		}

		entry.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Close closes the RewindStore database.
func (s *Store) Close() error {
	return s.db.Close()
}
