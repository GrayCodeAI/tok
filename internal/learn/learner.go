// Package learn provides automatic noise pattern discovery and filter generation.
//
// Learning mode monitors command outputs over time, identifies repeated
// patterns that could be filtered, and generates filter suggestions.
// Inspired by OMNI's pattern discovery architecture.
package learn

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Pattern represents a discovered noise pattern.
type Pattern struct {
	ID          int       `json:"id"`
	Pattern     string    `json:"pattern"`
	Example     string    `json:"example"`
	Command     string    `json:"command"`
	Frequency   int       `json:"frequency"`
	Confidence  float64   `json:"confidence"`
	Category    string    `json:"category"` // "noise", "boilerplate", "progress", "debug"
	Status      string    `json:"status"`   // "pending", "approved", "rejected"
	CreatedAt   time.Time `json:"created_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	Description string    `json:"description"`
}

// Sample represents a collected output sample for analysis.
type Sample struct {
	Command   string
	Args      string
	Output    string
	Timestamp time.Time
}

// FilterSuggestion represents a generated filter from discovered patterns.
type FilterSuggestion struct {
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Patterns    []string `json:"patterns"`
	StripLines  []string `json:"strip_lines"`
	MaxLines    int      `json:"max_lines"`
	Confidence  float64  `json:"confidence"`
	TOMLOutput  string   `json:"toml_output"`
}

// Stats contains learning mode statistics.
type Stats struct {
	SamplesCollected  int     `json:"samples_collected"`
	PatternsFound     int     `json:"patterns_found"`
	PendingPatterns   int     `json:"pending_patterns"`
	ApprovedPatterns  int     `json:"approved_patterns"`
	RejectedPatterns  int     `json:"rejected_patterns"`
	AvgConfidence     float64 `json:"avg_confidence"`
	CommandsCovered   int     `json:"commands_covered"`
	LearningActive    bool    `json:"learning_active"`
}

// Config holds configuration for the learning mode.
type Config struct {
	DatabasePath       string        `json:"database_path"`
	SamplingRate       float64       `json:"sampling_rate"`       // 0.0-1.0 (default: 1.0 = collect all)
	MinFrequency       int           `json:"min_frequency"`       // Minimum occurrences to suggest (default: 3)
	MinConfidence      float64       `json:"min_confidence"`      // Minimum confidence to suggest (default: 0.7)
	MaxSamples         int           `json:"max_samples"`         // Max samples to keep (default: 1000)
	Enabled            bool          `json:"enabled"`
	AutoApply          bool          `json:"auto_apply"`          // Auto-apply high-confidence patterns
	AutoApplyThreshold float64       `json:"auto_apply_threshold"` // Confidence for auto-apply (default: 0.95)
	RetentionPeriod    time.Duration `json:"retention_period"`
}

// DefaultConfig returns default learning mode configuration.
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		DatabasePath:       filepath.Join(homeDir, ".local", "share", "tokman", "learn.db"),
		SamplingRate:       1.0,
		MinFrequency:       3,
		MinConfidence:      0.7,
		MaxSamples:         1000,
		Enabled:            false, // Off by default
		AutoApply:          false,
		AutoApplyThreshold: 0.95,
		RetentionPeriod:    30 * 24 * time.Hour, // 30 days
	}
}

// Learner manages pattern discovery and filter generation.
type Learner struct {
	db     *sql.DB
	dbPath string
	cfg    Config
	mu     sync.RWMutex
	active bool

	// Common noise patterns for bootstrapping
	noisePatterns []*regexp.Regexp
}

// New creates a new Learner with the given configuration.
func New(cfg Config) (*Learner, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create learn directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open learn database: %w", err)
	}

	l := &Learner{
		db:     db,
		dbPath: cfg.DatabasePath,
		cfg:    cfg,
		active: true,
		noisePatterns: compileNoisePatterns(),
	}

	if err := l.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate learn database: %w", err)
	}

	return l, nil
}

func compileNoisePatterns() []*regexp.Regexp {
	patterns := []string{
		`^\s*$`,                           // Empty lines
		`^\[INFO\]`,                       // Info log lines
		`^\[DEBUG\]`,                      // Debug log lines
		`^\[TRACE\]`,                      // Trace log lines
		`^(Downloading|Fetching|Loading)`, // Progress indicators
		`^=== RUN\s`,                      // Go test run lines
		`^--- PASS:`,                      // Go test pass lines
		`^\s*\d+\.\d+s$`,                 // Bare timing lines
		`^Compiling\s`,                    // Compilation progress
		`^\s*warning:.*unused`,           // Unused warnings
		`^\s*#\s`,                         // Comment-like lines
		`^npm WARN`,                       // NPM warnings
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

func (l *Learner) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS learn_samples (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		args TEXT DEFAULT '',
		output_hash TEXT NOT NULL,
		output_preview TEXT NOT NULL,
		line_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS learn_patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL UNIQUE,
		example TEXT NOT NULL,
		command TEXT NOT NULL,
		frequency INTEGER DEFAULT 1,
		confidence REAL DEFAULT 0.0,
		category TEXT DEFAULT 'noise',
		status TEXT DEFAULT 'pending',
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_learn_command ON learn_samples(command);
	CREATE INDEX IF NOT EXISTS idx_learn_patterns_status ON learn_patterns(status);
	CREATE INDEX IF NOT EXISTS idx_learn_patterns_freq ON learn_patterns(frequency DESC);
	`
	_, err := l.db.Exec(schema)
	return err
}

// CollectSample records a command output sample for analysis.
func (l *Learner) CollectSample(sample Sample) error {
	if !l.active {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Hash the output for dedup
	h := sha256.New()
	h.Write([]byte(sample.Output))
	hash := hex.EncodeToString(h.Sum(nil))[:16]

	// Preview (first 200 chars)
	preview := sample.Output
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}

	lineCount := strings.Count(sample.Output, "\n") + 1

	_, err := l.db.Exec(`
		INSERT INTO learn_samples (command, args, output_hash, output_preview, line_count)
		VALUES (?, ?, ?, ?, ?)`,
		sample.Command, sample.Args, hash, preview, lineCount,
	)
	if err != nil {
		return fmt.Errorf("collect sample: %w", err)
	}

	// Analyze the sample for patterns
	l.analyzeSample(sample)

	return nil
}

// analyzeSample processes a sample to identify noise patterns.
func (l *Learner) analyzeSample(sample Sample) {
	lines := strings.Split(sample.Output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check against known noise patterns
		for _, re := range l.noisePatterns {
			if re.MatchString(trimmed) {
				l.recordPattern(re.String(), trimmed, sample.Command, "noise")
				break
			}
		}

		// Detect repetitive lines (same prefix, different suffix)
		l.detectRepetition(trimmed, sample.Command)
	}
}

func (l *Learner) detectRepetition(line, command string) {
	// Look for common repetitive patterns
	repetitivePatterns := []struct {
		prefix   string
		category string
	}{
		{"ok  \t", "boilerplate"},        // Go test OK lines
		{"Compiling ", "progress"},        // Rust compiling
		{"  adding: ", "progress"},        // Zip/archive adding
		{"Downloading ", "progress"},      // Download progress
		{"Installing ", "progress"},       // Package install
		{"Checking ", "progress"},         // Lint checking
	}

	for _, rp := range repetitivePatterns {
		if strings.HasPrefix(line, rp.prefix) {
			pattern := "^" + regexp.QuoteMeta(rp.prefix)
			l.recordPattern(pattern, line, command, rp.category)
			return
		}
	}
}

func (l *Learner) recordPattern(pattern, example, command, category string) {
	_, err := l.db.Exec(`
		INSERT INTO learn_patterns (pattern, example, command, category, frequency, confidence)
		VALUES (?, ?, ?, ?, 1, 0.5)
		ON CONFLICT(pattern) DO UPDATE SET
			frequency = frequency + 1,
			last_seen_at = CURRENT_TIMESTAMP,
			confidence = MIN(1.0, 0.5 + (frequency * 0.05))`,
		pattern, example, command, category,
	)
	if err != nil {
		// Silently fail - learning shouldn't break the pipeline
		return
	}
}

// GetPatterns returns discovered patterns, optionally filtered by status.
func (l *Learner) GetPatterns(status string) ([]Pattern, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, pattern, example, command, frequency, confidence, 
				category, status, description, created_at, last_seen_at
				FROM learn_patterns WHERE status = ? ORDER BY frequency DESC`
		args = append(args, status)
	} else {
		query = `SELECT id, pattern, example, command, frequency, confidence, 
				category, status, description, created_at, last_seen_at
				FROM learn_patterns ORDER BY frequency DESC`
	}

	rows, err := l.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get patterns: %w", err)
	}
	defer rows.Close()

	var patterns []Pattern
	for rows.Next() {
		var p Pattern
		var createdAt, lastSeen string

		if err := rows.Scan(
			&p.ID, &p.Pattern, &p.Example, &p.Command,
			&p.Frequency, &p.Confidence, &p.Category,
			&p.Status, &p.Description, &createdAt, &lastSeen,
		); err != nil {
			continue
		}

		p.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
		p.LastSeenAt, _ = time.Parse("2006-01-02T15:04:05Z", lastSeen)
		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

// ApprovePattern marks a pattern as approved for filtering.
func (l *Learner) ApprovePattern(id int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.db.Exec("UPDATE learn_patterns SET status = 'approved' WHERE id = ?", id)
	return err
}

// RejectPattern marks a pattern as rejected.
func (l *Learner) RejectPattern(id int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.db.Exec("UPDATE learn_patterns SET status = 'rejected' WHERE id = ?", id)
	return err
}

// GenerateFilters creates TOML filter suggestions from approved patterns.
func (l *Learner) GenerateFilters() ([]FilterSuggestion, error) {
	patterns, err := l.GetPatterns("pending")
	if err != nil {
		return nil, err
	}

	// Group by command
	cmdPatterns := make(map[string][]Pattern)
	for _, p := range patterns {
		if p.Confidence >= l.cfg.MinConfidence && p.Frequency >= l.cfg.MinFrequency {
			cmdPatterns[p.Command] = append(cmdPatterns[p.Command], p)
		}
	}

	var suggestions []FilterSuggestion
	for cmd, pats := range cmdPatterns {
		var stripLines []string
		var patternStrs []string
		var totalConfidence float64

		for _, p := range pats {
			stripLines = append(stripLines, p.Pattern)
			patternStrs = append(patternStrs, p.Pattern)
			totalConfidence += p.Confidence
		}

		avgConfidence := totalConfidence / float64(len(pats))

		// Generate TOML
		toml := generateTOML(cmd, stripLines)

		suggestions = append(suggestions, FilterSuggestion{
			Command:     cmd,
			Description: fmt.Sprintf("Auto-discovered filter for %s (%d patterns)", cmd, len(pats)),
			Patterns:    patternStrs,
			StripLines:  stripLines,
			MaxLines:    50,
			Confidence:  avgConfidence,
			TOMLOutput:  toml,
		})
	}

	// Sort by confidence (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	return suggestions, nil
}

func generateTOML(command string, stripLines []string) string {
	safeName := strings.ReplaceAll(command, " ", "_")
	safeName = strings.ReplaceAll(safeName, "-", "_")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Auto-generated filter for %s\n", command))
	sb.WriteString(fmt.Sprintf("[%s_learned]\n", safeName))
	sb.WriteString(fmt.Sprintf("match = \"^%s\"\n", regexp.QuoteMeta(command)))
	sb.WriteString(fmt.Sprintf("description = \"Auto-discovered patterns for %s\"\n", command))
	sb.WriteString("strip_lines_matching = [\n")
	for _, line := range stripLines {
		sb.WriteString(fmt.Sprintf("    %q,\n", line))
	}
	sb.WriteString("]\n")
	sb.WriteString("max_lines = 50\n")

	return sb.String()
}

// GetStats returns learning mode statistics.
func (l *Learner) GetStats() (*Stats, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var stats Stats
	stats.LearningActive = l.active

	// Count samples
	l.db.QueryRow("SELECT COUNT(*) FROM learn_samples").Scan(&stats.SamplesCollected)

	// Count patterns by status
	l.db.QueryRow("SELECT COUNT(*) FROM learn_patterns").Scan(&stats.PatternsFound)
	l.db.QueryRow("SELECT COUNT(*) FROM learn_patterns WHERE status = 'pending'").Scan(&stats.PendingPatterns)
	l.db.QueryRow("SELECT COUNT(*) FROM learn_patterns WHERE status = 'approved'").Scan(&stats.ApprovedPatterns)
	l.db.QueryRow("SELECT COUNT(*) FROM learn_patterns WHERE status = 'rejected'").Scan(&stats.RejectedPatterns)

	// Average confidence
	l.db.QueryRow("SELECT COALESCE(AVG(confidence), 0) FROM learn_patterns").Scan(&stats.AvgConfidence)

	// Unique commands
	l.db.QueryRow("SELECT COUNT(DISTINCT command) FROM learn_patterns").Scan(&stats.CommandsCovered)

	return &stats, nil
}

// Start enables learning mode.
func (l *Learner) Start() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.active = true
}

// Stop disables learning mode.
func (l *Learner) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.active = false
}

// Clear removes all learned data.
func (l *Learner) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.db.Exec("DELETE FROM learn_samples")
	if err != nil {
		return err
	}
	_, err = l.db.Exec("DELETE FROM learn_patterns")
	return err
}

// Close closes the learner database.
func (l *Learner) Close() error {
	return l.db.Close()
}
