package rewind

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
)

type RewindStore struct {
	db      *sql.DB
	maxSize int
	ttl     time.Duration
}

type StoredContent struct {
	MarkerID   string    `json:"marker_id"`
	Hash       string    `json:"hash"`
	Original   string    `json:"original"`
	Compressed string    `json:"compressed"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewRewindStore(db *sql.DB) *RewindStore {
	return &RewindStore{
		db:      db,
		maxSize: 10000,
		ttl:     24 * time.Hour,
	}
}

func (s *RewindStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS rewind_store (
		marker_id TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		original TEXT NOT NULL,
		compressed TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_rewind_hash ON rewind_store(hash);
	CREATE INDEX IF NOT EXISTS idx_rewind_created ON rewind_store(created_at);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *RewindStore) Store(original string, compressed string) (string, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(original)))
	markerID := fmt.Sprintf("[RW:%s]", hash[:8])

	existing := s.db.QueryRow("SELECT marker_id FROM rewind_store WHERE hash = ?", hash)
	var existingID string
	if err := existing.Scan(&existingID); err == nil {
		return existingID, nil
	}

	_, err := s.db.Exec(`
		INSERT INTO rewind_store (marker_id, hash, original, compressed)
		VALUES (?, ?, ?, ?)
	`, markerID, hash, original, compressed)

	s.cleanup()
	return markerID, err
}

func (s *RewindStore) Retrieve(markerID string) (string, error) {
	var original string
	err := s.db.QueryRow("SELECT original FROM rewind_store WHERE marker_id = ?", markerID).Scan(&original)
	return original, err
}

func (s *RewindStore) RetrieveByHash(hash string) (string, error) {
	var original string
	err := s.db.QueryRow("SELECT original FROM rewind_store WHERE hash = ?", hash).Scan(&original)
	return original, err
}

func (s *RewindStore) ReplaceMarkers(input string) (string, error) {
	output := input
	rows, err := s.db.Query("SELECT marker_id, original FROM rewind_store")
	if err != nil {
		return output, err
	}
	defer rows.Close()

	for rows.Next() {
		var markerID, original string
		if err := rows.Scan(&markerID, &original); err == nil {
			output = replaceMarker(output, markerID, original)
		}
	}
	return output, rows.Err()
}

func replaceMarker(input, markerID, original string) string {
	return ""
}

func (s *RewindStore) Size() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM rewind_store").Scan(&count)
	return count, err
}

func (s *RewindStore) cleanup() {
	size, _ := s.Size()
	if size > s.maxSize {
		s.db.Exec("DELETE FROM rewind_store WHERE created_at < ?", time.Now().Add(-s.ttl))
	}
}

func (s *RewindStore) Stats() (map[string]interface{}, error) {
	var totalOriginal, totalCompressed int
	var count int
	rows, _ := s.db.Query("SELECT original, compressed FROM rewind_store")
	defer rows.Close()
	for rows.Next() {
		var o, c string
		if err := rows.Scan(&o, &c); err == nil {
			totalOriginal += len(o)
			totalCompressed += len(c)
			count++
		}
	}
	ratio := 0.0
	if totalOriginal > 0 {
		ratio = float64(totalCompressed) / float64(totalOriginal) * 100
	}
	return map[string]interface{}{
		"count":             count,
		"total_original":    totalOriginal,
		"total_compressed":  totalCompressed,
		"compression_ratio": ratio,
	}, nil
}
