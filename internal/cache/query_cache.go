// Package cache provides persistent query caching for TokMan.
// Caches filtered command outputs for instant retrieval on repeated commands.
package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// QueryCache provides persistent caching of filtered outputs
type QueryCache struct {
	db     *sql.DB
	mu     sync.RWMutex
	hits   int64
	misses int64
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Key              string    `json:"key"`
	Command          string    `json:"command"`
	Args             string    `json:"args"`
	WorkingDir       string    `json:"working_dir"`
	FileHashes       string    `json:"file_hashes"`
	FilteredOutput   string    `json:"filtered_output"`
	OriginalTokens   int       `json:"original_tokens"`
	FilteredTokens   int       `json:"filtered_tokens"`
	CompressionRatio float64   `json:"compression_ratio"`
	CreatedAt        time.Time `json:"created_at"`
	AccessedAt       time.Time `json:"accessed_at"`
	HitCount         int       `json:"hit_count"`
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalEntries int64   `json:"total_entries"`
	TotalHits    int64   `json:"total_hits"`
	TotalMisses  int64   `json:"total_misses"`
	HitRate      float64 `json:"hit_rate"`
	TotalSaved   int64   `json:"total_tokens_saved"`
}

// NewQueryCache creates a new query cache
func NewQueryCache(dbPath string) (*QueryCache, error) {
	if dbPath == "" {
		dbPath = defaultCachePath()
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create cache directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open cache database: %w", err)
	}

	// Set pragmas for performance
	if _, err := db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = 10000;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure database: %w", err)
	}

	qc := &QueryCache{db: db}
	if err := qc.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate cache: %w", err)
	}

	return qc, nil
}

// defaultCachePath returns the default cache database path
func defaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "tokman", "cache.db")
}

// migrate creates the cache schema
func (c *QueryCache) migrate() error {
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS query_cache (
			key TEXT PRIMARY KEY,
			command TEXT NOT NULL,
			args TEXT,
			working_dir TEXT NOT NULL,
			file_hashes TEXT,
			filtered_output TEXT NOT NULL,
			original_tokens INTEGER NOT NULL,
			filtered_tokens INTEGER NOT NULL,
			compression_ratio REAL NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			hit_count INTEGER DEFAULT 1
		);

		CREATE INDEX IF NOT EXISTS idx_cache_command ON query_cache(command);
		CREATE INDEX IF NOT EXISTS idx_cache_accessed ON query_cache(accessed_at);
		CREATE INDEX IF NOT EXISTS idx_cache_created ON query_cache(created_at);
	`)
	return err
}

// GenerateKey creates a cache key from command context
func GenerateKey(command string, args []string, workingDir string, fileHashes map[string]string) string {
	h := sha256.New()

	// Hash command
	h.Write([]byte(command))
	h.Write([]byte("\x00"))

	// Hash args
	h.Write([]byte(strings.Join(args, "\x00")))
	h.Write([]byte("\x00"))

	// Hash working directory
	h.Write([]byte(workingDir))
	h.Write([]byte("\x00"))

	// Hash file hashes (sorted for consistency)
	if len(fileHashes) > 0 {
		keys := make([]string, 0, len(fileHashes))
		for k := range fileHashes {
			keys = append(keys, k)
		}
		// Simple sort (not ideal but deterministic)
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte("="))
			h.Write([]byte(fileHashes[k]))
			h.Write([]byte("\x00"))
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// Get retrieves a cached entry
func (c *QueryCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var entry CacheEntry
	err := c.db.QueryRow(`
		SELECT key, command, args, working_dir, file_hashes,
		       filtered_output, original_tokens, filtered_tokens,
		       compression_ratio, created_at, accessed_at, hit_count
		FROM query_cache
		WHERE key = ?
	`, key).Scan(
		&entry.Key, &entry.Command, &entry.Args, &entry.WorkingDir,
		&entry.FileHashes, &entry.FilteredOutput, &entry.OriginalTokens,
		&entry.FilteredTokens, &entry.CompressionRatio, &entry.CreatedAt,
		&entry.AccessedAt, &entry.HitCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.misses++
			return nil, false
		}
		// Log error but treat as miss
		c.misses++
		return nil, false
	}

	// Update access stats asynchronously
	c.hits++
	go c.updateAccessStats(key)

	return &entry, true
}

// updateAccessStats updates hit count and access time
func (c *QueryCache) updateAccessStats(key string) {
	_, _ = c.db.Exec(`
		UPDATE query_cache
		SET hit_count = hit_count + 1,
		    accessed_at = CURRENT_TIMESTAMP
		WHERE key = ?
	`, key)
}

// Set stores a new cache entry
func (c *QueryCache) Set(key string, command string, args []string, workingDir string,
	fileHashes map[string]string, filteredOutput string,
	originalTokens, filteredTokens int) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	argsJSON, _ := json.Marshal(args)
	hashesJSON, _ := json.Marshal(fileHashes)

	compressionRatio := 0.0
	if originalTokens > 0 {
		compressionRatio = float64(originalTokens-filteredTokens) / float64(originalTokens)
	}

	_, err := c.db.Exec(`
		INSERT INTO query_cache (
			key, command, args, working_dir, file_hashes,
			filtered_output, original_tokens, filtered_tokens,
			compression_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			filtered_output = excluded.filtered_output,
			original_tokens = excluded.original_tokens,
			filtered_tokens = excluded.filtered_tokens,
			compression_ratio = excluded.compression_ratio,
			created_at = CURRENT_TIMESTAMP,
			accessed_at = CURRENT_TIMESTAMP,
			hit_count = 1
	`, key, command, string(argsJSON), workingDir, string(hashesJSON),
		filteredOutput, originalTokens, filteredTokens, compressionRatio)

	return err
}

// Invalidate removes entries matching a predicate
func (c *QueryCache) Invalidate(predicate func(*CacheEntry) bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := c.db.Query(`
		SELECT key, command, args, working_dir, file_hashes,
		       filtered_output, original_tokens, filtered_tokens,
		       compression_ratio, created_at, accessed_at, hit_count
		FROM query_cache
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var toDelete []string

	for rows.Next() {
		var entry CacheEntry
		err := rows.Scan(
			&entry.Key, &entry.Command, &entry.Args, &entry.WorkingDir,
			&entry.FileHashes, &entry.FilteredOutput, &entry.OriginalTokens,
			&entry.FilteredTokens, &entry.CompressionRatio, &entry.CreatedAt,
			&entry.AccessedAt, &entry.HitCount,
		)
		if err != nil {
			continue
		}

		if predicate(&entry) {
			toDelete = append(toDelete, entry.Key)
		}
	}

	// Batch delete
	if len(toDelete) > 0 {
		placeholders := make([]string, len(toDelete))
		args := make([]interface{}, len(toDelete))
		for i, key := range toDelete {
			placeholders[i] = "?"
			args[i] = key
		}

		query := fmt.Sprintf("DELETE FROM query_cache WHERE key IN (%s)",
			strings.Join(placeholders, ","))
		_, err = c.db.Exec(query, args...)
		return err
	}

	return nil
}

// InvalidateByCommand removes all entries for a command
func (c *QueryCache) InvalidateByCommand(command string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("DELETE FROM query_cache WHERE command = ?", command)
	return err
}

// InvalidateByPrefix removes entries with working dir prefix
func (c *QueryCache) InvalidateByPrefix(prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("DELETE FROM query_cache WHERE working_dir LIKE ?", prefix+"%")
	return err
}

// Stats returns cache statistics
func (c *QueryCache) Stats() (*CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var stats CacheStats

	// Get entry count and total saved
	err := c.db.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(original_tokens - filtered_tokens), 0)
		FROM query_cache
	`).Scan(&stats.TotalEntries, &stats.TotalSaved)
	if err != nil {
		return nil, err
	}

	// Get hit counts from entries
	var totalHits int64
	err = c.db.QueryRow(`
		SELECT COALESCE(SUM(hit_count - 1), 0)
		FROM query_cache
	`).Scan(&totalHits)
	if err != nil {
		return nil, err
	}

	stats.TotalHits = totalHits
	stats.TotalMisses = stats.TotalEntries // Approximate

	if stats.TotalHits+stats.TotalMisses > 0 {
		stats.HitRate = float64(stats.TotalHits) / float64(stats.TotalHits+stats.TotalMisses)
	}

	return &stats, nil
}

// GetStats returns runtime hit/miss stats
func (c *QueryCache) GetRuntimeStats() (hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// Close closes the cache database
func (c *QueryCache) Close() error {
	return c.db.Close()
}

// Cleanup removes old entries
func (c *QueryCache) Cleanup(maxAge time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	_, err := c.db.Exec(
		"DELETE FROM query_cache WHERE accessed_at < ?",
		cutoff.Format(time.RFC3339),
	)
	return err
}

// GetTopQueries returns most frequently accessed queries
func (c *QueryCache) GetTopQueries(limit int) ([]*CacheEntry, error) {
	rows, err := c.db.Query(`
		SELECT key, command, args, working_dir, file_hashes,
		       filtered_output, original_tokens, filtered_tokens,
		       compression_ratio, created_at, accessed_at, hit_count
		FROM query_cache
		ORDER BY hit_count DESC, accessed_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*CacheEntry
	for rows.Next() {
		var entry CacheEntry
		err := rows.Scan(
			&entry.Key, &entry.Command, &entry.Args, &entry.WorkingDir,
			&entry.FileHashes, &entry.FilteredOutput, &entry.OriginalTokens,
			&entry.FilteredTokens, &entry.CompressionRatio, &entry.CreatedAt,
			&entry.AccessedAt, &entry.HitCount,
		)
		if err != nil {
			continue
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}
