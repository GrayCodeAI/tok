package waste

import (
	"database/sql"
	"time"
)

type WasteRecord struct {
	ID           int64     `json:"id"`
	Command      string    `json:"command"`
	TotalTokens  int       `json:"total_tokens"`
	WasteTokens  int       `json:"waste_tokens"`
	WasteScore   float64   `json:"waste_score"`
	FindingCount int       `json:"finding_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type WasteStore struct {
	db *sql.DB
}

func NewWasteStore(db *sql.DB) *WasteStore {
	return &WasteStore{db: db}
}

func (s *WasteStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS waste_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		total_tokens INTEGER NOT NULL,
		waste_tokens INTEGER NOT NULL,
		waste_score REAL NOT NULL,
		finding_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_waste_command ON waste_records(command);
	CREATE INDEX IF NOT EXISTS idx_waste_created ON waste_records(created_at);
	CREATE INDEX IF NOT EXISTS idx_waste_score ON waste_records(waste_score);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *WasteStore) Record(command string, report *WasteReport) error {
	_, err := s.db.Exec(`
		INSERT INTO waste_records (command, total_tokens, waste_tokens, waste_score, finding_count)
		VALUES (?, ?, ?, ?, ?)
	`, command, report.TotalTokens, report.WasteTokens, report.WasteScore, len(report.Findings))
	return err
}

func (s *WasteStore) GetTrend(days int) ([]WasteRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, command, total_tokens, waste_tokens, waste_score, finding_count, created_at
		FROM waste_records
		WHERE created_at >= datetime('now', '-' || ? || ' days')
		ORDER BY created_at DESC
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []WasteRecord
	for rows.Next() {
		var r WasteRecord
		err := rows.Scan(&r.ID, &r.Command, &r.TotalTokens, &r.WasteTokens, &r.WasteScore, &r.FindingCount, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *WasteStore) GetAverageScore(days int) (float64, error) {
	var avg sql.NullFloat64
	err := s.db.QueryRow(`
		SELECT AVG(waste_score) FROM waste_records
		WHERE created_at >= datetime('now', '-' || ? || ' days')
	`, days).Scan(&avg)
	if err != nil || !avg.Valid {
		return 0, err
	}
	return avg.Float64, nil
}
