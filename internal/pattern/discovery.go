package pattern

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/lakshmanpatel/tok/internal/config"
)

// PatternDiscoveryEngine automatically discovers patterns in content
type PatternDiscoveryEngine struct {
	db            *sql.DB
	patterns      map[string]*DiscoveredPattern
	sampleQueue   chan *ContentSample
	stopChan      chan struct{}
	minFrequency  int
	minConfidence float64
	mu            sync.RWMutex
	running       bool
}

// DiscoveredPattern represents an auto-discovered pattern
type DiscoveredPattern struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	Type        string    `json:"type"`
	Regex       string    `json:"regex"`
	Frequency   int       `json:"frequency"`
	Confidence  float64   `json:"confidence"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	SourceFiles []string  `json:"source_files"`
	Status      string    `json:"status"`
	FilterRule  string    `json:"filter_rule,omitempty"`
}

// ContentSample represents a sample of content for analysis
type ContentSample struct {
	Content   string
	Source    string
	Timestamp time.Time
}

// NewPatternDiscoveryEngine creates a new pattern discovery engine
func NewPatternDiscoveryEngine() (*PatternDiscoveryEngine, error) {
	dataDir := config.DataPath()
	dbPath := filepath.Join(dataDir, "patterns.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open patterns database: %w", err)
	}

	engine := &PatternDiscoveryEngine{
		db:            db,
		patterns:      make(map[string]*DiscoveredPattern),
		sampleQueue:   make(chan *ContentSample, 1000),
		stopChan:      make(chan struct{}),
		minFrequency:  5,
		minConfidence: 0.7,
	}

	if err := engine.initializeSchema(); err != nil {
		return nil, err
	}

	// Load existing patterns
	if err := engine.loadPatterns(); err != nil {
		return nil, err
	}

	return engine, nil
}

// initializeSchema creates database tables
func (pde *PatternDiscoveryEngine) initializeSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS discovered_patterns (
    id TEXT PRIMARY KEY,
    pattern TEXT NOT NULL,
    type TEXT NOT NULL,
    regex TEXT,
    frequency INTEGER DEFAULT 0,
    confidence REAL DEFAULT 0.0,
    first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    source_files TEXT, -- JSON array
    status TEXT DEFAULT 'active',
    filter_rule TEXT
);

CREATE INDEX IF NOT EXISTS idx_patterns_type ON discovered_patterns(type);
CREATE INDEX IF NOT EXISTS idx_patterns_status ON discovered_patterns(status);
CREATE INDEX IF NOT EXISTS idx_patterns_frequency ON discovered_patterns(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_patterns_last_seen ON discovered_patterns(last_seen);

CREATE TABLE IF NOT EXISTS pattern_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern_id TEXT,
    sample_hash TEXT UNIQUE,
    content_preview TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pattern_id) REFERENCES discovered_patterns(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_samples_pattern ON pattern_samples(pattern_id);
`
	_, err := pde.db.Exec(schema)
	return err
}

// Start starts the background sampling process
func (pde *PatternDiscoveryEngine) Start() {
	pde.mu.Lock()
	defer pde.mu.Unlock()

	if pde.running {
		return
	}

	pde.stopChan = make(chan struct{})
	pde.running = true
	go pde.samplingWorker()

	slog.Info("Pattern discovery engine started")
}

// Stop stops the background sampling
func (pde *PatternDiscoveryEngine) Stop() {
	pde.mu.Lock()
	defer pde.mu.Unlock()

	if !pde.running {
		return
	}

	close(pde.stopChan)
	pde.running = false

	slog.Info("Pattern discovery engine stopped")
}

// samplingWorker processes samples in the background
func (pde *PatternDiscoveryEngine) samplingWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case sample := <-pde.sampleQueue:
			pde.analyzeSample(sample)
		case <-ticker.C:
			pde.consolidatePatterns()
		case <-pde.stopChan:
			return
		}
	}
}

// SubmitSample submits content for pattern analysis
func (pde *PatternDiscoveryEngine) SubmitSample(content, source string) {
	select {
	case pde.sampleQueue <- &ContentSample{
		Content:   content,
		Source:    source,
		Timestamp: time.Now(),
	}:
		// Sample queued successfully
	default:
		// Queue full, drop sample
		slog.Debug("Pattern sample queue full, dropping sample")
	}
}

// analyzeSample analyzes a single sample for patterns
func (pde *PatternDiscoveryEngine) analyzeSample(sample *ContentSample) {
	lines := strings.Split(sample.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 10 {
			continue
		}

		// Detect different types of patterns
		pde.detectLogPattern(line, sample.Source)
		pde.detectErrorPattern(line, sample.Source)
		pde.detectPathPattern(line, sample.Source)
		pde.detectHashPattern(line, sample.Source)
		pde.detectTimestampPattern(line, sample.Source)
		pde.detectStackTracePattern(line, sample.Source)
	}
}

// detectLogPattern detects common log patterns
func (pde *PatternDiscoveryEngine) detectLogPattern(line, source string) {
	// Common log patterns
	patterns := []struct {
		name  string
		regex string
	}{
		{"timestamp_level", `\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}\s+(INFO|DEBUG|WARN|ERROR|FATAL)`},
		{"log_level", `\[(INFO|DEBUG|WARN|ERROR|FATAL)\]`},
		{"http_method", `(GET|POST|PUT|DELETE|PATCH)\s+/`},
		{"response_code", `HTTP/\d\.\d"\s+\d{3}`},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		if matches := re.FindAllString(line, -1); len(matches) > 0 {
			for _, match := range matches {
				pde.recordPattern(p.name, match, p.regex, source)
			}
		}
	}
}

// detectErrorPattern detects error patterns
func (pde *PatternDiscoveryEngine) detectErrorPattern(line, source string) {
	errorPatterns := []struct {
		name  string
		regex string
	}{
		{"error_keyword", `(?i)error[:\s]`},
		{"exception", `(?i)exception[:\s]`},
		{"failed", `(?i)failed[:\s]`},
		{"panic", `(?i)panic[:\s]`},
	}

	for _, p := range errorPatterns {
		re := regexp.MustCompile(p.regex)
		if re.MatchString(line) {
			pde.recordPattern(p.name, "error_indicator", p.regex, source)
		}
	}
}

// detectPathPattern detects file paths
func (pde *PatternDiscoveryEngine) detectPathPattern(line, source string) {
	pathRegex := `[\/\\][\w\-\.]+[\/\\][\w\-\.\/\\]*`
	re := regexp.MustCompile(pathRegex)

	if matches := re.FindAllString(line, -1); len(matches) > 0 {
		for _, match := range matches {
			// Normalize path
			normalized := regexp.MustCompile(`[\w\-]+`).ReplaceAllString(match, "WORD")
			pde.recordPattern("file_path", normalized, pathRegex, source)
		}
	}
}

// detectHashPattern detects hash patterns
func (pde *PatternDiscoveryEngine) detectHashPattern(line, source string) {
	hashRegex := `[a-f0-9]{32,64}`
	re := regexp.MustCompile(hashRegex)

	if matches := re.FindAllString(line, -1); len(matches) > 0 {
		for _, match := range matches {
			var hashType string
			switch len(match) {
			case 32:
				hashType = "md5"
			case 40:
				hashType = "sha1"
			case 64:
				hashType = "sha256"
			}
			if hashType != "" {
				pde.recordPattern("hash_"+hashType, match[:8]+"...", hashRegex, source)
			}
		}
	}
}

// detectTimestampPattern detects timestamp patterns
func (pde *PatternDiscoveryEngine) detectTimestampPattern(line, source string) {
	timestampRegexes := []string{
		`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d{3})?(?:Z|[+-]\d{2}:\d{2})?`,
		`\d{2}/\d{2}/\d{4}\s+\d{2}:\d{2}:\d{2}`,
		`\d{2}-\d{2}-\d{4}\s+\d{2}:\d{2}:\d{2}`,
	}

	for _, regex := range timestampRegexes {
		re := regexp.MustCompile(regex)
		if matches := re.FindAllString(line, -1); len(matches) > 0 {
			for _, match := range matches {
				normalized := regexp.MustCompile(`\d`).ReplaceAllString(match, "N")
				pde.recordPattern("timestamp", normalized, regex, source)
			}
		}
	}
}

// detectStackTracePattern detects stack trace patterns
func (pde *PatternDiscoveryEngine) detectStackTracePattern(line, source string) {
	stackTracePatterns := []struct {
		name  string
		regex string
	}{
		{"at_file_line", `\s+at\s+\S+\s*\([^)]+:\d+\)`},
		{"file_line", `[\w\-\.]+:\d+`},
		{"func_call", `[\w\-]+\([^)]*\)`},
	}

	for _, p := range stackTracePatterns {
		re := regexp.MustCompile(p.regex)
		if re.MatchString(line) {
			pde.recordPattern(p.name, p.name+"_pattern", p.regex, source)
		}
	}
}

// recordPattern records a discovered pattern
func (pde *PatternDiscoveryEngine) recordPattern(patternType, pattern, regex, source string) {
	// Create hash for deduplication
	hash := sha256.Sum256([]byte(patternType + ":" + pattern))
	patternID := hex.EncodeToString(hash[:8])

	pde.mu.Lock()
	defer pde.mu.Unlock()

	existing, found := pde.patterns[patternID]
	if found {
		// Update existing pattern
		existing.Frequency++
		existing.LastSeen = time.Now()
		existing.Confidence = pde.calculateConfidence(existing.Frequency)

		// Add source if new
		if !contains(existing.SourceFiles, source) {
			existing.SourceFiles = append(existing.SourceFiles, source)
		}
	} else {
		// Create new pattern
		pde.patterns[patternID] = &DiscoveredPattern{
			ID:          patternID,
			Pattern:     pattern,
			Type:        patternType,
			Regex:       regex,
			Frequency:   1,
			Confidence:  pde.calculateConfidence(1),
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
			SourceFiles: []string{source},
			Status:      "active",
		}
	}
}

// calculateConfidence calculates confidence based on frequency
func (pde *PatternDiscoveryEngine) calculateConfidence(frequency int) float64 {
	// Logistic function for confidence
	// confidence = 1 / (1 + e^(-k*(f - f0)))
	k := 0.3
	f0 := float64(pde.minFrequency) / 2

	return 1.0 / (1.0 + math.Exp(-k*(float64(frequency)-f0)))
}

// consolidatePatterns consolidates and saves patterns
func (pde *PatternDiscoveryEngine) consolidatePatterns() {
	pde.mu.Lock()
	patterns := make([]*DiscoveredPattern, 0, len(pde.patterns))
	for _, p := range pde.patterns {
		if p.Frequency >= pde.minFrequency && p.Confidence >= pde.minConfidence {
			patterns = append(patterns, p)
		}
	}
	pde.mu.Unlock()

	// Sort by confidence
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Confidence > patterns[j].Confidence
	})

	// Save top patterns
	for i, p := range patterns {
		if i >= 100 { // Keep top 100
			break
		}
		if err := pde.savePattern(p); err != nil {
			slog.Error("Failed to save pattern", "id", p.ID, "error", err)
		}
	}

	slog.Info("Pattern consolidation complete", "patterns", len(patterns))
}

// savePattern saves a pattern to database
func (pde *PatternDiscoveryEngine) savePattern(p *DiscoveredPattern) error {
	sourceFilesJSON, err := json.Marshal(p.SourceFiles)
	if err != nil {
		return fmt.Errorf("marshal pattern source files: %w", err)
	}

	_, err = pde.db.Exec(`
		INSERT OR REPLACE INTO discovered_patterns
		(id, pattern, type, regex, frequency, confidence, first_seen, last_seen, source_files, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, p.Pattern, p.Type, p.Regex, p.Frequency, p.Confidence,
		p.FirstSeen, p.LastSeen, sourceFilesJSON, p.Status)

	return err
}

// loadPatterns loads patterns from database
func (pde *PatternDiscoveryEngine) loadPatterns() error {
	rows, err := pde.db.Query(`
		SELECT id, pattern, type, regex, frequency, confidence, first_seen, last_seen, source_files, status
		FROM discovered_patterns WHERE status = 'active'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p DiscoveredPattern
		var sourceFilesJSON []byte

		err := rows.Scan(&p.ID, &p.Pattern, &p.Type, &p.Regex, &p.Frequency, &p.Confidence,
			&p.FirstSeen, &p.LastSeen, &sourceFilesJSON, &p.Status)
		if err != nil {
			continue
		}

		if len(sourceFilesJSON) > 0 {
			if err := json.Unmarshal(sourceFilesJSON, &p.SourceFiles); err != nil {
				return fmt.Errorf("unmarshal pattern source files for %s: %w", p.ID, err)
			}
		}
		pde.patterns[p.ID] = &p
	}

	return rows.Err()
}

// GetPatterns returns all discovered patterns
func (pde *PatternDiscoveryEngine) GetPatterns(minConfidence float64) []*DiscoveredPattern {
	pde.mu.RLock()
	defer pde.mu.RUnlock()

	var patterns []*DiscoveredPattern
	for _, p := range pde.patterns {
		if p.Confidence >= minConfidence {
			patterns = append(patterns, p)
		}
	}

	// Sort by confidence
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Confidence > patterns[j].Confidence
	})

	return patterns
}

// GetPatternByID returns a specific pattern
func (pde *PatternDiscoveryEngine) GetPatternByID(id string) (*DiscoveredPattern, bool) {
	pde.mu.RLock()
	defer pde.mu.RUnlock()

	p, ok := pde.patterns[id]
	return p, ok
}

// DeletePattern deletes a pattern
func (pde *PatternDiscoveryEngine) DeletePattern(id string) error {
	pde.mu.Lock()
	defer pde.mu.Unlock()

	delete(pde.patterns, id)

	_, err := pde.db.Exec("DELETE FROM discovered_patterns WHERE id = ?", id)
	return err
}

// GenerateFilter generates a filter rule from a pattern
func (p *DiscoveredPattern) GenerateFilter() string {
	switch p.Type {
	case "timestamp":
		return fmt.Sprintf("remove_lines_matching: '%s'", p.Regex)
	case "hash_md5", "hash_sha1", "hash_sha256":
		return fmt.Sprintf("replace_pattern: '%s' -> '[HASH]'", p.Regex)
	case "file_path":
		return fmt.Sprintf("replace_pattern: '%s' -> '[PATH]'", p.Regex)
	default:
		return fmt.Sprintf("# Pattern: %s (confidence: %.2f)", p.Pattern, p.Confidence)
	}
}

// Close closes the discovery engine
func (pde *PatternDiscoveryEngine) Close() error {
	pde.Stop()
	return pde.db.Close()
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
