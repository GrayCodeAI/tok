package rightsizing

import (
	"database/sql"
	"time"
)

type RightSizingRecord struct {
	ID               int64     `json:"id"`
	Command          string    `json:"command"`
	CurrentModel     string    `json:"current_model"`
	RecommendedModel string    `json:"recommended_model"`
	ComplexityScore  int       `json:"complexity_score"`
	EstimatedSavings float64   `json:"estimated_savings"`
	Accepted         bool      `json:"accepted"`
	CreatedAt        time.Time `json:"created_at"`
}

type RightSizingStore struct {
	db *sql.DB
}

func NewRightSizingStore(db *sql.DB) *RightSizingStore {
	return &RightSizingStore{db: db}
}

func (s *RightSizingStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS rightsizing_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		current_model TEXT NOT NULL,
		recommended_model TEXT NOT NULL,
		complexity_score INTEGER NOT NULL,
		estimated_savings REAL NOT NULL,
		accepted BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_rightsizing_command ON rightsizing_records(command);
	CREATE INDEX IF NOT EXISTS idx_rightsizing_model ON rightsizing_records(current_model);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *RightSizingStore) Record(rec *RightSizingRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO rightsizing_records (command, current_model, recommended_model, complexity_score, estimated_savings)
		VALUES (?, ?, ?, ?, ?)
	`, rec.Command, rec.CurrentModel, rec.RecommendedModel, rec.ComplexityScore, rec.EstimatedSavings)
	return err
}

func (s *RightSizingStore) AcceptRecommendation(id int64) error {
	_, err := s.db.Exec("UPDATE rightsizing_records SET accepted = 1 WHERE id = ?", id)
	return err
}

func (s *RightSizingStore) GetAccuracy() (float64, error) {
	var total, accepted int
	err := s.db.QueryRow("SELECT COUNT(*), SUM(CASE WHEN accepted = 1 THEN 1 ELSE 0 END) FROM rightsizing_records").Scan(&total, &accepted)
	if err != nil || total == 0 {
		return 0, err
	}
	return float64(accepted) / float64(total) * 100, nil
}

func (s *RightSizingStore) GetRecent(limit int) ([]RightSizingRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, command, current_model, recommended_model, complexity_score, estimated_savings, accepted, created_at
		FROM rightsizing_records ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RightSizingRecord
	for rows.Next() {
		var r RightSizingRecord
		err := rows.Scan(&r.ID, &r.Command, &r.CurrentModel, &r.RecommendedModel, &r.ComplexityScore, &r.EstimatedSavings, &r.Accepted, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
