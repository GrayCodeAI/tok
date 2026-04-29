package tracking

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/core"
	"github.com/GrayCodeAI/tok/internal/retry"
	"github.com/GrayCodeAI/tok/internal/utils"
)

// HistoryRetentionDays is the number of days to retain tracking data.
// Records older than this are automatically cleaned up on each write.
const HistoryRetentionDays = 90

// TrackerInterface defines the contract for command tracking.
// Implementations can use SQLite, in-memory stores, or mocks for testing.
type TrackerInterface interface {
	Record(record *CommandRecord) error
	GetSavings(projectPath string) (*SavingsSummary, error)
	GetRecentCommands(projectPath string, limit int) ([]CommandRecord, error)
	Query(query string, args ...any) (*sql.Rows, error)
	Close() error
}

// Tracker manages token tracking persistence.
type Tracker struct {
	db            *sql.DB
	lastCleanupMs int64          // atomic: unix timestamp of last cleanup
	cleanupCh     chan struct{}  // non-blocking cleanup trigger
	doneCh        chan struct{}  // signals shutdown to cleanup worker
	cleanupWg     sync.WaitGroup // waits for cleanup goroutine to finish
	closed        atomic.Bool
	closeOnce     sync.Once
}

// TimedExecution tracks execution time and token savings.
type TimedExecution struct {
	startTime time.Time
	once      sync.Once
}

var (
	globalTracker *Tracker
	trackerMu     sync.Mutex
)

// Start begins a timed execution for tracking.
func Start() *TimedExecution {
	return &TimedExecution{
		startTime: time.Now(),
	}
}

// Track records the execution with token savings.
// Automatically captures AI agent attribution from environment variables:
//   - TOK_AGENT: AI agent name (e.g., "Claude Code", "OpenCode", "Cursor")
//   - TOK_MODEL: Model name (e.g., "claude-3-opus", "gpt-4")
//   - TOK_PROVIDER: Provider name (e.g., "Anthropic", "OpenAI")
func (t *TimedExecution) Track(command, tokCmd string, originalTokens, filteredTokens int) {
	t.once.Do(func() {
		execTime := time.Since(t.startTime)
		saved := originalTokens - filteredTokens
		if saved < 0 {
			saved = 0
		}

		// Get or create global tracker
		tracker := getGlobalTracker()
		if tracker == nil {
			return
		}

		projectPath := config.ProjectPath()
		tracker.Record(&CommandRecord{
			Command:        command,
			OriginalTokens: originalTokens,
			FilteredTokens: filteredTokens,
			SavedTokens:    saved,
			ProjectPath:    projectPath,
			ExecTimeMs:     execTime.Milliseconds(),
			Timestamp:      time.Now(),
			ParseSuccess:   true,
			// AI Agent attribution from environment
			AgentName:   os.Getenv("TOK_AGENT"),
			ModelName:   os.Getenv("TOK_MODEL"),
			Provider:    os.Getenv("TOK_PROVIDER"),
			ModelFamily: utils.GetModelFamily(os.Getenv("TOK_MODEL")),
		})
	})
}

// getGlobalTracker returns the global tracker instance.
func getGlobalTracker() *Tracker {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	if globalTracker != nil {
		return globalTracker
	}

	// Initialize tracker
	dbPath := DatabasePath()
	if dbPath == "" {
		slog.Info("tracker: no database path configured, tracking disabled")
		return nil
	}

	tracker, err := NewTracker(dbPath)
	if err != nil {
		slog.Warn("tracker: failed to initialize database", "path", dbPath, "error", err)
		return nil
	}

	globalTracker = tracker
	return globalTracker
}

// GetGlobalTracker returns the global tracker instance (exported for external use).
func GetGlobalTracker() *Tracker {
	return getGlobalTracker()
}

// CloseGlobalTracker closes the global tracker without creating it when it has
// not been used in the current process.
func CloseGlobalTracker() error {
	trackerMu.Lock()
	tracker := globalTracker
	globalTracker = nil
	trackerMu.Unlock()

	if tracker == nil {
		return nil
	}
	return tracker.Close()
}

// DatabasePath returns the effective tracking database path using config file
// resolution when available, with environment/default fallback.
var (
	cachedConfig     *config.Config
	cachedConfigOnce sync.Once
)

func loadConfigCached() *config.Config {
	cachedConfigOnce.Do(func() {
		var err error
		cachedConfig, err = config.Load("")
		if err != nil {
			cachedConfig = config.Defaults()
		}
	})
	return cachedConfig
}

func DatabasePath() string {
	cfg := loadConfigCached()
	if cfg != nil {
		return cfg.GetDatabasePath()
	}
	return config.DatabasePath()
}

// NewTracker creates a new Tracker with the given database path.
func NewTracker(dbPath string) (*Tracker, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create tracker directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for concurrent access
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(25)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(2 * time.Minute) // Maximum idle time before closing

	// Enable WAL mode for better performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to retry on locked database
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run versioned migrations.
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Safely add optional metadata columns (idempotent)
	if err := addOptionalCommandColumns(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to add command metadata columns: %w", err)
	}

	t := &Tracker{
		db:        db,
		cleanupCh: make(chan struct{}, 1),
		doneCh:    make(chan struct{}),
	}
	t.cleanupWg.Add(1)
	go t.cleanupWorker()
	return t, nil
}

// Close closes the database connection.
func (t *Tracker) Close() error {
	var err error
	t.closeOnce.Do(func() {
		t.closed.Store(true)
		close(t.doneCh)    // signal worker to exit
		t.cleanupWg.Wait() // wait for cleanup goroutine to finish before closing DB
		err = t.db.Close()
	})
	return err
}

// cleanupWorker processes cleanup triggers from the channel.
func (t *Tracker) cleanupWorker() {
	defer t.cleanupWg.Done()
	for {
		select {
		case <-t.cleanupCh:
			t.cleanupOld()
		case <-t.doneCh:
			return
		}
	}
}

// addOptionalCommandColumns safely adds optional metadata columns if they don't exist.
// This is idempotent and won't fail if columns already exist.
func addOptionalCommandColumns(db *sql.DB) error {
	// Get existing columns
	rows, err := db.Query("PRAGMA table_info(commands)")
	if err != nil {
		return err
	}
	defer rows.Close()

	existingCols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("scanning table_info: %w", err)
		}
		existingCols[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating table_info: %w", err)
	}

	// Add missing columns
	for _, col := range CommandColumnDefs {
		if !existingCols[col.Name] {
			if err := addColumnSafe(db, col.Name, col.Type); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.Name, err)
			}
		}
	}

	// Create indexes (IF NOT EXISTS handles duplicates)
	if _, err := db.Exec(AgentAttributionIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// validColumnNames is the allowlist of column names that may be added dynamically.
// Only names in this set are accepted by addColumnSafe.
var validColumnNames = map[string]bool{
	"agent_name":            true,
	"model_name":            true,
	"provider":              true,
	"model_family":          true,
	"context_kind":          true,
	"context_mode":          true,
	"context_resolved_mode": true,
	"context_target":        true,
	"context_related_files": true,
	"context_bundle":        true,
}

// addColumnSafe adds a column using a strict allowlist and quoted identifiers.
func addColumnSafe(db *sql.DB, name, colType string) error {
	if !validColumnNames[name] {
		return fmt.Errorf("column name %q is not in the allowlist", name)
	}
	quoted := sqliteQuoteIdentifier(name)
	_, err := db.Exec(fmt.Sprintf("ALTER TABLE commands ADD COLUMN %s %s", quoted, colType))
	return err
}

// sqliteQuoteIdentifier wraps an identifier in double quotes and escapes embedded quotes.
func sqliteQuoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// Query executes a raw SQL query and returns the rows.
// This is exposed for custom aggregations in the economics package.
func (t *Tracker) Query(query string, args ...any) (*sql.Rows, error) {
	return t.db.Query(query, args...)
}

// QueryRow executes a raw SQL query expected to return at most one row.
func (t *Tracker) QueryRow(query string, args ...any) *sql.Row {
	return t.db.QueryRow(query, args...)
}

// EstimateTokens provides a BPE-accurate token count for persisted savings
// records. User-visible dashboards (tok gain, tok session) read from this
// table, so we always use the precise path rather than the short-string
// heuristic fast path.
func EstimateTokens(text string) int {
	return core.EstimateTokensPrecise(text)
}

// Record saves a command execution to the database.
func (t *Tracker) Record(record *CommandRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.RecordContext(ctx, record)
}

// RecordContext saves a command execution to the database with context support.
// P3: Enables cancellation for long-running database operations.
func (t *Tracker) RecordContext(ctx context.Context, record *CommandRecord) error {
	projectPath := normalizeProjectPath(record.ProjectPath)
	record.ProjectPath = projectPath

	query := `
		INSERT INTO commands (
			command, original_output, filtered_output,
			original_tokens, filtered_tokens, saved_tokens,
			project_path, session_id, exec_time_ms, parse_success,
			agent_name, model_name, provider, model_family,
			context_kind, context_mode, context_resolved_mode,
			context_target, context_related_files, context_bundle
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Use retry logic for database operations
	err := retry.Do(ctx, retry.DefaultConfig(), func() error {
		result, err := t.db.ExecContext(ctx, query,
			record.Command,
			record.OriginalOutput,
			record.FilteredOutput,
			record.OriginalTokens,
			record.FilteredTokens,
			record.SavedTokens,
			record.ProjectPath,
			record.SessionID,
			record.ExecTimeMs,
			record.ParseSuccess,
			record.AgentName,
			record.ModelName,
			record.Provider,
			record.ModelFamily,
			record.ContextKind,
			record.ContextMode,
			record.ContextResolvedMode,
			record.ContextTarget,
			record.ContextRelatedFiles,
			record.ContextBundle,
		)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err == nil {
			record.ID = id
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to record command", "error", err, "command", record.Command)
		return fmt.Errorf("failed to record command: %w", err)
	}

	// Record checkpoint trigger telemetry for this command.
	if record.ID > 0 {
		if err := t.autoRecordCheckpointEvents(record); err != nil {
			slog.Warn("tracking checkpoint telemetry failed", "error", err)
		}
	}

	// Fan out to live subscribers (TUI live mode). Non-blocking — the
	// SQL row is canonical; a dropped event just means the UI takes
	// the fallback-tick path instead of updating instantly.
	notifySubscribers(record)

	// Run cleanup after recording (throttled - at most once per minute)
	if !t.closed.Load() {
		select {
		case t.cleanupCh <- struct{}{}:
		default:
		}
	}

	return nil
}

func (t *Tracker) autoRecordCheckpointEvents(record *CommandRecord) error {
	if record == nil {
		return nil
	}
	triggers := evaluateCheckpointTriggers(record)
	if len(triggers) == 0 {
		return nil
	}
	sessionID := strings.TrimSpace(record.SessionID)
	if sessionID == "" {
		sessionID = "global"
	}
	fillPct := estimateFillPercent(record.OriginalTokens, record.ModelName)
	quality := estimateQualityScore(record)

	for _, trigger := range triggers {
		allowed, cooldownSec, err := t.checkTriggerCooldown(sessionID, trigger, 10*time.Minute)
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		ev := CheckpointEventRecord{
			CommandID:   record.ID,
			SessionID:   sessionID,
			Trigger:     trigger,
			Reason:      checkpointReason(trigger),
			FillPct:     fillPct,
			Quality:     quality,
			CooldownSec: cooldownSec,
		}
		if err := t.RecordCheckpointEvent(&ev); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracker) checkTriggerCooldown(sessionID, trigger string, cooldown time.Duration) (bool, int, error) {
	var lastEpoch int64
	err := t.db.QueryRow(
		`SELECT COALESCE(CAST(strftime('%s', created_at) AS INTEGER), 0) FROM checkpoint_events
		 WHERE session_id = ? AND trigger = ?
		 ORDER BY created_at DESC LIMIT 1`,
		sessionID, trigger,
	).Scan(&lastEpoch)
	if err != nil && err != sql.ErrNoRows {
		return false, 0, err
	}
	if lastEpoch <= 0 {
		return true, int(cooldown.Seconds()), nil
	}
	last := time.Unix(lastEpoch, 0)
	if time.Since(last) < cooldown {
		return false, int(cooldown.Seconds()), nil
	}
	return true, int(cooldown.Seconds()), nil
}

func evaluateCheckpointTriggers(record *CommandRecord) []string {
	out := make([]string, 0, 6)
	if record.OriginalTokens >= 20_000 {
		out = append(out, "progressive-20")
	}
	if record.OriginalTokens >= 50_000 {
		out = append(out, "progressive-50")
	}
	if record.OriginalTokens >= 100_000 {
		out = append(out, "progressive-100")
	}
	q := estimateQualityScore(record)
	if q < 80 {
		out = append(out, "quality-80")
	}
	if q < 70 {
		out = append(out, "quality-70")
	}
	if isMilestoneCommand(record.Command, record.ContextRelatedFiles) {
		out = append(out, "milestone-edit-batch")
	}
	return dedupeStrings(out)
}

func isMilestoneCommand(command string, relatedFiles int) bool {
	c := strings.ToLower(strings.TrimSpace(command))
	if relatedFiles >= 3 {
		return true
	}
	return strings.HasPrefix(c, "git commit") ||
		strings.HasPrefix(c, "git push") ||
		strings.HasPrefix(c, "git merge")
}

func estimateFillPercent(originalTokens int, model string) float64 {
	window := 200000.0
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "gpt-5.4"):
		window = 1100000
	case strings.Contains(m, "gpt-4.1"), strings.Contains(m, "sonnet"), strings.Contains(m, "opus"), strings.Contains(m, "gemini-2.5-pro"):
		window = 1000000
	case strings.Contains(m, "haiku"), strings.Contains(m, "o3"), strings.Contains(m, "o4"):
		window = 200000
	}
	fill := (float64(originalTokens) / window) * 100
	return math.Max(0, math.Min(100, fill))
}

func estimateQualityScore(record *CommandRecord) float64 {
	reduction := 0.0
	if record.OriginalTokens > 0 {
		reduction = float64(record.SavedTokens) / float64(record.OriginalTokens) * 100
	}
	parseBoost := 0.0
	if record.ParseSuccess {
		parseBoost = 30
	}
	score := reduction*0.7 + parseBoost
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

func checkpointReason(trigger string) string {
	switch trigger {
	case "progressive-20":
		return "session crossed 20k-token band"
	case "progressive-50":
		return "session crossed 50k-token band"
	case "progressive-100":
		return "session crossed 100k-token band"
	case "quality-80":
		return "quality proxy dropped below 80"
	case "quality-70":
		return "quality proxy dropped below 70"
	case "milestone-edit-batch":
		return "edit/milestone command batch detected"
	default:
		return "trigger condition matched"
	}
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, item := range in {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

// RecordCheckpointEvent persists a checkpoint trigger event.
