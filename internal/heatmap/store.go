package heatmap

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type HeatmapRecord struct {
	ID          int64     `json:"id"`
	RequestID   string    `json:"request_id"`
	TotalTokens int       `json:"total_tokens"`
	WasteScore  float64   `json:"waste_score"`
	SystemPct   float64   `json:"system_pct"`
	ToolsPct    float64   `json:"tools_pct"`
	ContextPct  float64   `json:"context_pct"`
	HistoryPct  float64   `json:"history_pct"`
	QueryPct    float64   `json:"query_pct"`
	CreatedAt   time.Time `json:"created_at"`
}

type HeatmapStore struct {
	db *sql.DB
}

func NewHeatmapStore(db *sql.DB) *HeatmapStore {
	return &HeatmapStore{db: db}
}

func (s *HeatmapStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS heatmap_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id TEXT NOT NULL,
		total_tokens INTEGER NOT NULL,
		waste_score REAL NOT NULL DEFAULT 0,
		system_pct REAL NOT NULL DEFAULT 0,
		tools_pct REAL NOT NULL DEFAULT 0,
		context_pct REAL NOT NULL DEFAULT 0,
		history_pct REAL NOT NULL DEFAULT 0,
		query_pct REAL NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_heatmap_request ON heatmap_records(request_id);
	CREATE INDEX IF NOT EXISTS idx_heatmap_created ON heatmap_records(created_at);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *HeatmapStore) Record(data *HeatmapData, requestID string) error {
	var systemPct, toolsPct, contextPct, historyPct, queryPct float64
	for _, sec := range data.Sections {
		switch sec.Type {
		case SectionSystem:
			systemPct += sec.Percentage
		case SectionTools:
			toolsPct += sec.Percentage
		case SectionContext:
			contextPct += sec.Percentage
		case SectionHistory:
			historyPct += sec.Percentage
		case SectionQuery:
			queryPct += sec.Percentage
		}
	}

	_, err := s.db.Exec(`
		INSERT INTO heatmap_records (request_id, total_tokens, waste_score, system_pct, tools_pct, context_pct, history_pct, query_pct)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, requestID, data.TotalTokens, data.WasteScore, systemPct, toolsPct, contextPct, historyPct, queryPct)
	return err
}

func (s *HeatmapStore) GetAverages(days int) (map[string]float64, error) {
	query := `
		SELECT AVG(system_pct), AVG(tools_pct), AVG(context_pct), AVG(history_pct), AVG(query_pct), AVG(total_tokens)
		FROM heatmap_records
		WHERE created_at >= datetime('now', ?)
	`
	interval := fmt.Sprintf("-%d days", days)
	row := s.db.QueryRow(query, interval)
	var sys, tools, ctx, hist, qry float64
	var avgTokens float64
	err := row.Scan(&sys, &tools, &ctx, &hist, &qry, &avgTokens)
	if err != nil {
		return nil, err
	}
	return map[string]float64{
		"system":  sys,
		"tools":   tools,
		"context": ctx,
		"history": hist,
		"query":   qry,
		"tokens":  avgTokens,
	}, nil
}

func (s *HeatmapStore) GetRecent(limit int) ([]HeatmapRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, request_id, total_tokens, waste_score, system_pct, tools_pct, context_pct, history_pct, query_pct, created_at
		FROM heatmap_records ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HeatmapRecord
	for rows.Next() {
		var r HeatmapRecord
		err := rows.Scan(&r.ID, &r.RequestID, &r.TotalTokens, &r.WasteScore,
			&r.SystemPct, &r.ToolsPct, &r.ContextPct, &r.HistoryPct, &r.QueryPct, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *HeatmapStore) ExportJSON() ([]byte, error) {
	records, err := s.GetRecent(10000)
	if err != nil {
		return nil, err
	}
	return json.Marshal(records)
}

func (s *HeatmapStore) ExportCSV() string {
	records, _ := s.GetRecent(10000)
	var sb strings.Builder
	sb.WriteString("id,request_id,total_tokens,waste_score,system_pct,tools_pct,context_pct,history_pct,query_pct,created_at\n")
	for _, r := range records {
		sb.WriteString(fmt.Sprintf("%d,%s,%d,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s\n",
			r.ID, r.RequestID, r.TotalTokens, r.WasteScore,
			r.SystemPct, r.ToolsPct, r.ContextPct, r.HistoryPct, r.QueryPct,
			r.CreatedAt.Format(time.RFC3339)))
	}
	return sb.String()
}
